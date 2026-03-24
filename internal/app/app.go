package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/google/uuid"
	pbservices "github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/internal/pkg/auth"
	"github.com/gsoultan/gobpm/internal/pkg/config"
	"github.com/gsoultan/gobpm/internal/pkg/logger"
	"github.com/gsoultan/gobpm/server/domains/observers/impl"
	"github.com/gsoultan/gobpm/server/domains/services"
	"github.com/gsoultan/gobpm/server/endpoints"
	"github.com/gsoultan/gobpm/server/interceptors"
	authinterceptor "github.com/gsoultan/gobpm/server/interceptors/auth"
	"github.com/gsoultan/gobpm/server/repositories"
	gorms "github.com/gsoultan/gobpm/server/repositories/gorms"
	models "github.com/gsoultan/gobpm/server/repositories/models"
	"github.com/gsoultan/gobpm/server/transports/grpcs"
	https "github.com/gsoultan/gobpm/server/transports/https"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type App struct {
	config     *config.Config
	db         *gorm.DB
	repo       repositories.Repository
	svc        services.ServiceFacade
	sse        *impl.SSEObserver
	validator  *auth.TokenValidator
	initDBOnce func()
}

func New() *App {
	return &App{
		initDBOnce: sync.OnceFunc(func() {
			log.Info().Msg("Initializing database connection...")
		}),
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 0. Flag Parsing
	buildUI := flag.Bool("build-ui", false, "Build the UI using bun")
	flag.Parse()

	if *buildUI {
		return a.handleBuildUI()
	}

	// 0. Initialize Logger
	logger.Init()

	a.initDBOnce()

	// 1. Initialize DB with GORM
	if err := a.setupDatabase(); err != nil {
		return err
	}

	// 2. Initialize Domain
	if err := a.setupService(ctx); err != nil {
		return err
	}

	// 3. Setup Transports
	a.setupAuth(ctx)

	// 4. Start Servers using errgroup
	return a.runServers(ctx)
}

func (a *App) handleBuildUI() error {
	fmt.Println("Building UI...")
	if _, err := exec.LookPath("bun"); err != nil {
		return fmt.Errorf("'bun' not found in PATH. Please install Bun to build the UI")
	}
	cmd := exec.Command("bun", "run", "build")
	cmd.Dir = "ui"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building UI: %w", err)
	}
	fmt.Println("UI build successful!")
	return nil
}

func (a *App) setupDatabase() error {
	dialector, err := a.resolveDialector()
	if err != nil {
		return err
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	a.db = db

	return a.migrate()
}

func (a *App) migrate() error {
	if err := a.db.AutoMigrate(models.MigrationModels()...); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}
	return nil
}

func (a *App) setupService(ctx context.Context) error {
	a.repo = repositories.NewRepository(a.db)

	dispatcher := impl.NewEventDispatcher()
	dispatcher.Register(impl.NewAuditLogObserver(a.repo.Audit()))
	a.sse = impl.NewSSEObserver()
	dispatcher.Register(a.sse)

	// Register Webhook Observer if endpoints are provided
	webhookEndpoints := os.Getenv("WEBHOOK_ENDPOINTS")
	if webhookEndpoints != "" {
		endpointsList := strings.Split(webhookEndpoints, ",")
		dispatcher.Register(impl.NewWebhookObserver(endpointsList))
		log.Info().Int("count", len(endpointsList)).Msg("Registered Webhook Observer")
	}

	jwtSecret, err := a.resolveJWTSecret()
	if err != nil {
		return err
	}
	a.svc = services.NewServiceFacade(a.repo, dispatcher, a.sse, jwtSecret, func(targetDB *gorm.DB) {
		log.Info().Msg("Setup complete: hot-swapping database connection to target database")
		gorms.SetDBOverride(targetDB)
	})

	dispatcher.Register(impl.NewNotificationObserver(a.svc))

	a.svc.StartWorkers(ctx)
	return nil
}

func (a *App) setupAuth(ctx context.Context) {
	oidcIssuer := os.Getenv("OIDC_ISSUER")
	oidcClientID := os.Getenv("OIDC_CLIENT_ID")
	if oidcIssuer != "" && oidcClientID != "" {
		v, err := auth.NewTokenValidator(ctx, oidcIssuer, oidcClientID)
		if err != nil {
			log.Error().Err(err).Msg("failed to initialize OIDC validator")
		} else {
			a.validator = v
			log.Info().Str("issuer", oidcIssuer).Msg("OIDC Authentication enabled")
		}
	}
}

func (a *App) runServers(ctx context.Context) error {
	endpts := endpoints.MakeEndpoints(a.svc)
	httpHandler := https.NewHTTPHandler(a.svc, endpts, a.sse)

	f := interceptors.NewInterceptorFactory(a.svc)
	var strategy authinterceptor.SecurityStrategy
	if a.validator != nil {
		strategy = f.NewOIDCStrategy(a.validator)
	} else {
		strategy = f.NewJWTStrategy()
	}

	publicPaths := []string{
		"/api/v1/login",
		"/api/v1/setup/status",
		"/api/v1/setup",
		"/api/v1/setup/test-connection",
	}
	httpHandler = f.NewMandatoryHTTPAuth(strategy, publicPaths).Wrap(httpHandler)

	grpcServer := grpcs.NewGRPCServer(endpts)

	g, ctx := errgroup.WithContext(ctx)

	// HTTP Server
	g.Go(func() error {
		log.Info().Msg("HTTP server listening on :8080")
		server := &http.Server{
			Addr:    ":8080",
			Handler: httpHandler,
		}
		go func() {
			<-ctx.Done()
			server.Shutdown(context.Background())
		}()
		return server.ListenAndServe()
	})

	// gRPC Server
	g.Go(func() error {
		lis, err := net.Listen("tcp", ":8081")
		if err != nil {
			return err
		}
		baseServer := grpc.NewServer()
		a.registerGRPCServices(baseServer, grpcServer)

		log.Info().Msg("gRPC server listening on :8081")

		go func() {
			<-ctx.Done()
			baseServer.GracefulStop()
		}()
		return baseServer.Serve(lis)
	})

	err := g.Wait()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server crashed: %w", err)
	}
	return nil
}

func (a *App) registerGRPCServices(baseServer *grpc.Server, grpcServer any) {
	pbservices.RegisterOrganizationServiceServer(baseServer, grpcServer.(pbservices.OrganizationServiceServer))
	pbservices.RegisterProjectServiceServer(baseServer, grpcServer.(pbservices.ProjectServiceServer))
	pbservices.RegisterProcessServiceServer(baseServer, grpcServer.(pbservices.ProcessServiceServer))
	pbservices.RegisterTaskServiceServer(baseServer, grpcServer.(pbservices.TaskServiceServer))
	pbservices.RegisterDefinitionServiceServer(baseServer, grpcServer.(pbservices.DefinitionServiceServer))
	pbservices.RegisterStatsServiceServer(baseServer, grpcServer.(pbservices.StatsServiceServer))
	pbservices.RegisterExternalTaskServiceServer(baseServer, grpcServer.(pbservices.ExternalTaskServiceServer))
	pbservices.RegisterSignalServiceServer(baseServer, grpcServer.(pbservices.SignalServiceServer))
}

// resolveJWTSecret determines the JWT signing secret.
//
// Priority:
//  1. JWT_SECRET environment variable (always preferred).
//  2. If config.yaml exists (system is configured) and JWT_SECRET is absent,
//     the application refuses to start – a random secret would silently
//     invalidate all sessions on every restart.
//  3. Pre-setup state (no config.yaml): a random ephemeral secret is used
//     because no real users exist yet.
func (a *App) resolveJWTSecret() (string, error) {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret, nil
	}

	if config.Exists(config.DefaultConfigPath) {
		cfg, err := config.Load(config.DefaultConfigPath)
		if err == nil && cfg.JWTSecret != "" {
			return cfg.JWTSecret, nil
		}

		return "", fmt.Errorf(
			"JWT_SECRET environment variable or config jwt_secret is required when config.yaml is present; " +
				"set it to a stable secret to avoid invalidating user sessions on restart",
		)
	}

	// Pre-setup: no persistent users exist yet; ephemeral secret is acceptable.
	log.Warn().Msg("No JWT_SECRET set and no config.yaml found; using a random ephemeral secret (pre-setup mode only)")
	return uuid.NewString(), nil
}

// resolveDialector determines the GORM dialector based on config.yaml or environment variables.
func (a *App) resolveDialector() (gorm.Dialector, error) {
	// Priority 1: Load from config.yaml if it exists
	if config.Exists(config.DefaultConfigPath) {
		cfg, err := config.Load(config.DefaultConfigPath)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load config.yaml, falling back to environment")
		} else {
			return a.dialectorFromConfig(cfg)
		}
	}

	// Priority 2: Fall back to environment variables
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		log.Info().Msg("Using PostgreSQL database from DATABASE_URL...")
		return postgres.Open(dsn), nil
	}

	// Priority 3: Default to SQLite
	log.Info().Msg("Using SQLite database (gobpm.db)...")
	return sqlite.Open("gobpm.db"), nil
}

func (a *App) dialectorFromConfig(cfg *config.Config) (gorm.Dialector, error) {
	encKey := os.Getenv("ENCRYPTION_KEY")
	dsn, err := cfg.DecryptConnectionString(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt database connection string: %w", err)
	}

	switch cfg.Database.Driver {
	case config.DriverPostgres:
		log.Info().Msg("Using PostgreSQL database from config...")
		return postgres.Open(dsn), nil
	case config.DriverMySQL:
		log.Info().Msg("Using MySQL database from config...")
		return mysql.Open(dsn), nil
	case config.DriverSQLServer:
		log.Info().Msg("Using SQL Server database from config...")
		return sqlserver.Open(dsn), nil
	default:
		log.Info().Str("path", dsn).Msg("Using SQLite database from config...")
		return sqlite.Open(dsn), nil
	}
}
