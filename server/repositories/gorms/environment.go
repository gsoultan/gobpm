package gorms

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/repositories/contracts"
	"github.com/gsoultan/metis/server/repositories/models"

	"gorm.io/gorm"
)

type gormEnvironmentRepository struct {
	db *gorm.DB
}

// NewEnvironmentRepository creates a new GORM-based EnvironmentRepository.
func NewEnvironmentRepository(db *gorm.DB) contracts.EnvironmentRepository {
	return &gormEnvironmentRepository{db: db}
}

// tableEnvironments is the SQL table behind EnvironmentModel, needed by name so
// the tenant scope can build its clauses.
const tableEnvironments = "environments"

// Get returns one environment, scoped to the caller's tenant.
//
// The row carries an encrypted database connection, so an unscoped read hands
// over the credentials to another organization's runtime — every process
// variable, task and audit entry it holds, at once.
func (r *gormEnvironmentRepository) Get(ctx context.Context, id uuid.UUID) (models.EnvironmentModel, error) {
	var m models.EnvironmentModel
	db := tenantScopeDB(ctx, MainTx(ctx, r.db), tableEnvironments)
	if err := db.First(&m, QualifiedByID(tableEnvironments), id).Error; err != nil {
		return models.EnvironmentModel{}, lookupError(err, "environment")
	}
	return m, nil
}

// ListByProject returns a project's environments, oldest first.
//
// Creation order rather than name: it is the order they were set up in, which
// for most projects is development, then staging, then production — the order
// somebody reading the list expects to promote through.
func (r *gormEnvironmentRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.EnvironmentModel, error) {
	var modelsList []models.EnvironmentModel
	db := tenantScopeDB(ctx, MainTx(ctx, r.db), tableEnvironments)
	if err := db.
		Where(QualifiedByProjectID(tableEnvironments), projectID).
		Order(tableEnvironments + ".created_at ASC").
		Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list environments: %w", err)
	}
	return modelsList, nil
}

// ListAll returns every environment in the installation.
//
// Deliberately unscoped, and called only at boot: the server has to bind a
// listener for each environment before any request exists to carry a tenant.
// Everything a request can reach goes through the scoped reads above.
func (r *gormEnvironmentRepository) ListAll(ctx context.Context) ([]models.EnvironmentModel, error) {
	var modelsList []models.EnvironmentModel
	if err := MainTx(ctx, r.db).Order("created_at ASC").Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list environments: %w", err)
	}
	return modelsList, nil
}

// Create records a new environment, refusing one planted in another
// organization's project.
func (r *gormEnvironmentRepository) Create(ctx context.Context, m models.EnvironmentModel) error {
	if err := requireProjectInTenant(ctx, MainTx(ctx, r.db), uuid.UUID(m.ProjectID)); err != nil {
		return err
	}
	if err := MainTx(ctx, r.db).Create(&m).Error; err != nil {
		return fmt.Errorf("could not create environment: %w", err)
	}
	return nil
}

// Update saves an environment, refusing an ID outside the caller's tenant.
func (r *gormEnvironmentRepository) Update(ctx context.Context, m models.EnvironmentModel) error {
	db := MainTx(ctx, r.db)
	if err := requireVisibleToTenant(ctx, db, tableEnvironments, &models.EnvironmentModel{}, uuid.UUID(m.ID)); err != nil {
		return err
	}
	if err := db.Save(&m).Error; err != nil {
		return fmt.Errorf("could not update environment: %w", err)
	}
	return nil
}

// Delete removes an environment, refusing an ID outside the caller's tenant.
//
// The database it names is left alone. Dropping somebody's production data as a
// side effect of tidying a list is not a thing this should be able to do; the
// row going away means the runtime is no longer served, not that it never
// existed.
func (r *gormEnvironmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := MainTx(ctx, r.db)
	if err := requireVisibleToTenant(ctx, db, tableEnvironments, &models.EnvironmentModel{}, id); err != nil {
		return err
	}
	if err := db.Delete(&models.EnvironmentModel{}, QualifiedByID(tableEnvironments), id).Error; err != nil {
		return fmt.Errorf("could not delete environment: %w", err)
	}
	return nil
}

// PortTaken reports whether another environment already claims a port.
//
// Checked in the service before saving so the caller is told which environment
// holds it, rather than being handed a unique-constraint violation — and long
// before the server tries to bind two listeners to one port at boot, where the
// failure would take the whole installation down rather than one form.
//
// Unscoped on purpose: ports are an installation-wide resource, and a
// collision with another organization's environment is still a collision. The
// answer is a boolean and an id, never the other tenant's row.
func (r *gormEnvironmentRepository) PortTaken(ctx context.Context, port int, excluding uuid.UUID) (bool, error) {
	query := MainTx(ctx, r.db).
		Model(&models.EnvironmentModel{}).
		Where("port = ?", port)
	if excluding != uuid.Nil {
		query = query.Where("id <> ?", models.FromUUID(excluding))
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("could not check whether port %d is in use: %w", port, err)
	}
	return count > 0, nil
}
