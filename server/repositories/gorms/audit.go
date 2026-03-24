package gorms

import (
	"context"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/repositories/contracts"
	"github.com/gsoultan/gobpm/server/repositories/models"
	"gorm.io/gorm"
)

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) contracts.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, entry models.AuditModel) error {
	return GetTx(ctx, r.db).Create(&entry).Error
}

func (r *auditRepository) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]models.AuditModel, error) {
	var modelsList []models.AuditModel
	err := GetTx(ctx, r.db).Where("instance_id = ?", instanceID).Order("created_at desc").Find(&modelsList).Error
	if err != nil {
		return nil, err
	}
	return modelsList, nil
}

func (r *auditRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.AuditModel, error) {
	var modelsList []models.AuditModel
	err := GetTx(ctx, r.db).Where("project_id = ?", projectID).Order("created_at desc").Find(&modelsList).Error
	if err != nil {
		return nil, err
	}
	return modelsList, nil
}
