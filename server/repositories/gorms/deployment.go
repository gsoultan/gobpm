package gorms

import (
	"context"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/repositories/contracts"
	"github.com/gsoultan/gobpm/server/repositories/models"
	"gorm.io/gorm"
)

type gormDeploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) contracts.DeploymentRepository {
	return &gormDeploymentRepository{db: db}
}

func (r *gormDeploymentRepository) Create(ctx context.Context, d models.DeploymentModel) error {
	return GetTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&d).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *gormDeploymentRepository) Get(ctx context.Context, id uuid.UUID) (models.DeploymentModel, error) {
	var m models.DeploymentModel
	if err := GetTx(ctx, r.db).Preload("Resources").First(&m, "id = ?", id).Error; err != nil {
		return models.DeploymentModel{}, err
	}
	return m, nil
}

func (r *gormDeploymentRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.DeploymentModel, error) {
	var modelsList []models.DeploymentModel
	query := GetTx(ctx, r.db)
	if projectID != uuid.Nil {
		query = query.Where("project_id = ?", projectID)
	}
	if err := query.Find(&modelsList).Error; err != nil {
		return nil, err
	}
	return modelsList, nil
}

func (r *gormDeploymentRepository) GetResource(ctx context.Context, id uuid.UUID) (models.ResourceModel, error) {
	var m models.ResourceModel
	if err := GetTx(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		return models.ResourceModel{}, err
	}
	return m, nil
}

func (r *gormDeploymentRepository) ListResources(ctx context.Context, deploymentID uuid.UUID) ([]models.ResourceModel, error) {
	var modelsList []models.ResourceModel
	if err := GetTx(ctx, r.db).Find(&modelsList, "deployment_id = ?", deploymentID).Error; err != nil {
		return nil, err
	}
	return modelsList, nil
}
