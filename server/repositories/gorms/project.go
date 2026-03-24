package gorms

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/repositories/contracts"
	"github.com/gsoultan/gobpm/server/repositories/models"

	"gorm.io/gorm"
)

type gormProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new GORM-based ProjectRepository.
func NewProjectRepository(db *gorm.DB) contracts.ProjectRepository {
	return &gormProjectRepository{db: db}
}

func (r *gormProjectRepository) Get(ctx context.Context, id uuid.UUID) (models.ProjectModel, error) {
	var m models.ProjectModel
	if err := GetTx(ctx, r.db).First(&m, QueryByID, id).Error; err != nil {
		return models.ProjectModel{}, fmt.Errorf("could not get project: %w", err)
	}
	return m, nil
}

func (r *gormProjectRepository) List(ctx context.Context) ([]models.ProjectModel, error) {
	var modelsList []models.ProjectModel
	if err := GetTx(ctx, r.db).Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list projects: %w", err)
	}
	return modelsList, nil
}

func (r *gormProjectRepository) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]models.ProjectModel, error) {
	var modelsList []models.ProjectModel
	query := GetTx(ctx, r.db)
	if organizationID != uuid.Nil {
		query = query.Where("organization_id = ?", organizationID)
	}
	if err := query.Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list projects: %w", err)
	}
	return modelsList, nil
}

func (r *gormProjectRepository) Create(ctx context.Context, p models.ProjectModel) error {
	if err := GetTx(ctx, r.db).Create(&p).Error; err != nil {
		return fmt.Errorf("could not create project: %w", err)
	}
	return nil
}

func (r *gormProjectRepository) Update(ctx context.Context, p models.ProjectModel) error {
	result := GetTx(ctx, r.db).Save(&p)
	if result.Error != nil {
		return fmt.Errorf("could not update project: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found: %s", p.ID)
	}
	return nil
}

func (r *gormProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := GetTx(ctx, r.db).Delete(&models.ProjectModel{}, QueryByID, id)
	if result.Error != nil {
		return fmt.Errorf("could not delete project: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("project not found: %s", id)
	}
	return nil
}
