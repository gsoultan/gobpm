package impl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gsoultan/metis/internal/pkg/secrets"

	"github.com/rs/zerolog/log"

	"github.com/gsoultan/metis/server/repositories/gorms"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/config"
	"github.com/gsoultan/metis/internal/pkg/crypto"
	"github.com/gsoultan/metis/internal/pkg/dbpool"
	"github.com/gsoultan/metis/internal/pkg/redaction"
	"github.com/gsoultan/metis/server/domains/services/contracts"
	stormdb "github.com/gsoultan/metis/server/repositories/db"
	"github.com/gsoultan/metis/server/repositories/migrations"
	"github.com/gsoultan/metis/server/repositories/model"
	"github.com/gsoultan/metis/server/repositories/models"
	"github.com/gsoultan/storm"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// OnSetupCompleteFunc is called after setup succeeds, passing the open target database
// so the application can hot-swap its connection without requiring a restart.
type OnSetupCompleteFunc func(targetDB *gorm.DB)

type setupService struct {
	onSetupComplete OnSetupCompleteFunc
}

func NewSetupService(onSetupComplete OnSetupCompleteFunc) contracts.SetupService {
	return &setupService{onSetupComplete: onSetupComplete}
}

func (s *setupService) GetSetupStatus(_ context.Context) (contracts.SetupStatus, error) {
	return contracts.SetupStatus{
		IsInitialized: config.Exists(config.DefaultConfigPath),
	}, nil
}

func (s *setupService) Setup(ctx context.Context, req contracts.SetupRequest) error {
	status, err := s.GetSetupStatus(ctx)
	if err != nil {
		return err
	}
	if status.IsInitialized {
		return errors.New("system already configured: config.yaml exists")
	}

	if err := validateSetupRequest(req); err != nil {
		return err
	}

	// 1. Open a connection to the TARGET database and seed initial data
	targetDB, cleanup, err := openTargetDatabase(req)
	if err != nil {
		return fmt.Errorf("failed to connect to target database: %w", err)
	}

	// 2. Run migrations on the target database
	if err := migrateTargetDatabase(ctx, targetDB, req); err != nil {
		cleanup()
		return fmt.Errorf("failed to migrate target database: %w", err)
	}

	// 3. Create Organization, Project, and Admin User in the target database
	if err := seedTargetDatabase(targetDB, req); err != nil {
		cleanup()
		return err
	}

	// 4. Generate and save config.yaml with encrypted connection string
	// We do this AFTER database operations succeed to ensure consistency.
	if err := saveConfiguration(req); err != nil {
		cleanup()
		return err
	}

	// 5. Hot-swap the database connection so the app uses the target DB immediately
	if s.onSetupComplete != nil {
		s.onSetupComplete(targetDB)
	} else {
		cleanup()
	}

	return nil
}

const testConnectionTimeout = 10 * time.Second

// TestConnection opens a database and reports whether it answered.
//
// Only while the installation is unconfigured. This endpoint is public, because
// the wizard that uses it runs before anyone can sign in — and it takes a host
// and a port from the caller and reports precisely what happened to the attempt.
// On a configured installation that is an unauthenticated port scanner: the
// reply distinguishes "connection refused" from a timeout from an authentication
// failure, which is enough to map whatever network the server sits in.
//
// After setup, the same job is done by the environments endpoint, which requires
// an administrator.
func (s *setupService) TestConnection(ctx context.Context, req contracts.TestConnectionRequest) contracts.TestConnectionResult {
	status, err := s.GetSetupStatus(ctx)
	if err != nil {
		return contracts.TestConnectionResult{Success: false, Message: "Could not determine whether this installation is configured"}
	}
	if status.IsInitialized {
		return contracts.TestConnectionResult{
			Success: false,
			Message: "This installation is already configured. Test a database connection from Settings → Environments.",
		}
	}

	if req.DatabaseDriver == "" {
		return contracts.TestConnectionResult{Success: false, Message: "Database driver is required"}
	}

	fields := config.DatabaseFields{
		Host:       req.DBHost,
		Port:       req.DBPort,
		Username:   req.DBUsername,
		Password:   req.DBPassword,
		DBName:     req.DBName,
		SSLEnabled: req.DBSSLEnabled,
	}
	dsn := config.BuildConnectionString(req.DatabaseDriver, fields)
	dialector, err := gorms.Dialector(req.DatabaseDriver, dsn)
	if err != nil {
		return contracts.TestConnectionResult{Success: false, Message: err.Error()}
	}

	db, err := gorm.Open(dialector, gorms.Config())
	if err != nil {
		return contracts.TestConnectionResult{Success: false, Message: fmt.Sprintf("Failed to open connection: %s", redaction.RedactError(err))}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return contracts.TestConnectionResult{Success: false, Message: fmt.Sprintf("Failed to get database handle: %s", redaction.RedactError(err))}
	}
	defer func() {
		// A connection test that leaks its own pool is a connection test that
		// exhausts the server it was checking.
		if err := sqlDB.Close(); err != nil {
			log.Warn().Err(err).Msg("Could not close the connection opened to test the database")
		}
	}()

	pingCtx, cancel := context.WithTimeout(ctx, testConnectionTimeout)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return contracts.TestConnectionResult{Success: false, Message: fmt.Sprintf("Connection failed: %s", redaction.RedactError(err))}
	}

	return contracts.TestConnectionResult{Success: true, Message: "Connection successful"}
}

func validateSetupRequest(req contracts.SetupRequest) error {
	if req.AdminUsername == "" || req.AdminPassword == "" || req.AdminFullName == "" || req.AdminPublicName == "" || req.OrganizationName == "" {
		return errors.New("admin username, password, full name, public name and organization name are required")
	}
	if req.DatabaseDriver == "" {
		return errors.New("database driver is required")
	}
	if !config.SupportedDriver(req.DatabaseDriver) {
		return fmt.Errorf("this runs on PostgreSQL; %q is not a database engine it supports", req.DatabaseDriver)
	}
	// Validated here, before saveConfiguration encrypts anything with the key.
	//
	// The wizard used to apply a weaker rule of its own — sixteen characters
	// for the encryption key, nothing at all for the JWT secret beyond being
	// present — while the boot path refused anything under thirty-two. So a
	// wizard run could succeed, write the config, and leave a server that
	// refused to start on the next restart: configured successfully, and
	// permanently unable to come back. Sharing the rule is what stops the two
	// disagreeing.
	//
	// This is also the right moment for it. A new installation is the one point
	// where the key can still be chosen freely; once data has been encrypted
	// with it, a weak key has no safe remedy.
	if !secrets.Allowed() {
		if err := secrets.Validate("encryption key", req.EncryptionKey); err != nil {
			return err
		}
		if err := secrets.Validate("JWT secret", req.JWTSecret); err != nil {
			return err
		}
	}
	if req.EncryptionKey == "" {
		return errors.New("encryption key is required")
	}
	if req.JWTSecret == "" {
		return errors.New("jwt secret is required")
	}
	return nil
}

func saveConfiguration(req contracts.SetupRequest) error {
	fields := config.DatabaseFields{
		Host:       req.DBHost,
		Port:       req.DBPort,
		Username:   req.DBUsername,
		Password:   req.DBPassword,
		DBName:     req.DBName,
		SSLEnabled: req.DBSSLEnabled,
	}
	connectionString := config.BuildConnectionString(req.DatabaseDriver, fields)

	cfg, err := config.NewConfig(req.DatabaseDriver, connectionString, req.EncryptionKey, req.JWTSecret)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	if err := cfg.Save(config.DefaultConfigPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Install the key in the running process. Without this the server would
	// have persisted a config it cannot use until the next restart, and every
	// write to an encrypted column would fail with ErrKeyNotConfigured.
	if err := crypto.Configure(req.EncryptionKey); err != nil {
		return fmt.Errorf("failed to install encryption key: %w", err)
	}

	return nil
}

func buildDatabaseFields(req contracts.SetupRequest) config.DatabaseFields {
	return config.DatabaseFields{
		Host:       req.DBHost,
		Port:       req.DBPort,
		Username:   req.DBUsername,
		Password:   req.DBPassword,
		DBName:     req.DBName,
		SSLEnabled: req.DBSSLEnabled,
	}
}

func openTargetDatabase(req contracts.SetupRequest) (*gorm.DB, func(), error) {
	dsn := config.BuildConnectionString(req.DatabaseDriver, buildDatabaseFields(req))
	dialector, err := gorms.Dialector(req.DatabaseDriver, dsn)
	if err != nil {
		return nil, nil, err
	}

	db, err := gorm.Open(dialector, gorms.Config())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open target database: %w", err)
	}

	// Sized the same way the app's own open path sizes it — this database is
	// about to be hot-swapped in as the live one, so it must not run the rest
	// of its life on a pool nobody configured.
	dbpool.Apply(db)

	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Warn().Err(err).Msg("Could not reach the database handle to close it")
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Warn().Err(err).Msg("Could not close the database connection pool")
		}
	}

	return db, cleanup, nil
}

// migrateTargetDatabase brings the freshly configured database up to the
// current schema.
//
// It runs the same versioned migrations the application runs at boot, rather
// than a bare AutoMigrate. Otherwise setup would create the schema without
// recording a single version, and the first boot afterwards would treat a brand
// new database as one that had never been migrated — replaying every data
// migration over it, including the one that walks every process instance.
func migrateTargetDatabase(ctx context.Context, db *gorm.DB, req contracts.SetupRequest) error {
	// Only the schema migrations. The data migrations need the repository layer,
	// which setup does not have here, and they repair rows written by older
	// versions of the engine — of which a database created seconds ago has none.
	// The first boot runs and records them against empty tables, which costs two
	// queries that find nothing.
	if _, err := migrations.Run(ctx, db, migrations.Schema(models.MigrationModels())); err != nil {
		return err
	}
	return ensureStormSchema(ctx, req)
}

// ensureStormSchema creates the tables the GORM migrations do not describe, in
// the database the wizard has just configured.
//
// The same two steps the application performs at boot, run here because setup
// writes into this database immediately afterwards — the admin's organization
// membership lands in a join table the GORM models no longer declare, and
// without this the very first insert of a fresh installation fails with
// "relation user_organizations does not exist".
//
// A second connection rather than the GORM one, because these are storm's own
// DDL and its schema builder speaks pgx.
func ensureStormSchema(ctx context.Context, req contracts.SetupRequest) error {
	want, err := storm.Build(model.All()...)
	if err != nil {
		return fmt.Errorf("the model layer does not build: %w", err)
	}
	dsn := config.BuildConnectionString(req.DatabaseDriver, buildDatabaseFields(req))
	pool, err := pgxpool.New(ctx, config.PostgresURL(dsn))
	if err != nil {
		return fmt.Errorf("could not open the target database for the storm schema: %w", err)
	}
	defer pool.Close()

	if _, err := stormdb.EnsureTables(ctx, pool, want); err != nil {
		return fmt.Errorf("could not create the storm tables: %w", err)
	}
	if _, err := stormdb.EnsureColumnDefaults(ctx, pool, want); err != nil {
		return fmt.Errorf("could not reconcile the storm column defaults: %w", err)
	}
	return nil
}

const defaultProjectName = "Default Project"

func seedTargetDatabase(db *gorm.DB, req contracts.SetupRequest) error {
	return db.Transaction(func(tx *gorm.DB) error {
		orgID := uuid.Must(uuid.NewV7())
		now := time.Now()

		org := models.OrganizationModel{
			Base: models.Base{
				ID:        models.UUID(orgID),
				CreatedAt: now,
			},
			Name: req.OrganizationName,
		}
		if err := tx.Create(&org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		projectName := req.ProjectName
		if projectName == "" {
			projectName = defaultProjectName
		}
		project := models.ProjectModel{
			Base: models.Base{
				ID:        models.UUID(uuid.Must(uuid.NewV7())),
				CreatedAt: now,
			},
			OrganizationID: models.UUID(orgID),
			Name:           projectName,
		}
		if err := tx.Create(&project).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		admin := models.UserModel{
			Base: models.Base{
				ID:        models.UUID(uuid.Must(uuid.NewV7())),
				CreatedAt: now,
			},
			Username:     req.AdminUsername,
			PasswordHash: string(hash),
			FullName:     req.AdminFullName,
			DisplayName:  req.AdminPublicName,
			Email:        req.AdminEmail,
			Roles:        []string{"ADMIN"},
		}
		if err := tx.Create(&admin).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		// The memberships, written explicitly.
		//
		// They used to be GORM associations on the struct above, which wrote
		// the join rows as a side effect of creating the user. The join tables
		// belong to the storm model now and the association tags are gone, so
		// the side effect went with them — silently: setup succeeded, the admin
		// existed, and signing in showed "authenticated principal has no
		// organization membership" on every page.
		//
		// Raw SQL because this runs on the database the wizard is configuring,
		// which is not the one the repositories are connected to yet.
		if err := tx.Exec(
			`INSERT INTO user_organizations (user_id, organization_id) VALUES (?, ?)`,
			admin.ID, org.ID).Error; err != nil {
			return fmt.Errorf("failed to put the admin in the organization: %w", err)
		}
		if err := tx.Exec(
			`INSERT INTO user_projects (user_id, project_id) VALUES (?, ?)`,
			admin.ID, project.ID).Error; err != nil {
			return fmt.Errorf("failed to put the admin in the project: %w", err)
		}

		return nil
	})
}
