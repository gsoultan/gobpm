package gorms

import (
	"context"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/repositories/contracts"
	"github.com/gsoultan/gobpm/server/repositories/models"
	"gorm.io/gorm"
)

type gormFormRepository struct {
	db *gorm.DB
}

func NewFormRepository(db *gorm.DB) contracts.FormRepository {
	return &gormFormRepository{db: db}
}

func (r *gormFormRepository) Create(ctx context.Context, f models.FormModel) error {
	return GetTx(ctx, r.db).Create(&f).Error
}

func (r *gormFormRepository) Get(ctx context.Context, id uuid.UUID) (models.FormModel, error) {
	var m models.FormModel
	if err := GetTx(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		return models.FormModel{}, err
	}
	return m, nil
}

func (r *gormFormRepository) GetByKey(ctx context.Context, projectID uuid.UUID, key string) (models.FormModel, error) {
	var m models.FormModel
	if err := GetTx(ctx, r.db).Where("project_id = ? AND key = ?", projectID, key).First(&m).Error; err != nil {
		return models.FormModel{}, err
	}
	return m, nil
}

func (r *gormFormRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.FormModel, error) {
	var modelsList []models.FormModel
	query := GetTx(ctx, r.db)
	if projectID != uuid.Nil {
		query = query.Where("project_id = ?", projectID)
	}
	if err := query.Find(&modelsList).Error; err != nil {
		return nil, err
	}
	return modelsList, nil
}

func (r *gormFormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return GetTx(ctx, r.db).Delete(&models.FormModel{}, "id = ?", id).Error
}
