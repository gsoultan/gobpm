package models

import (
	"github.com/google/uuid"
)

type FormModel struct {
	Base
	ProjectID uuid.UUID      `gorm:"type:uuid;index" json:"project_id,omitzero"`
	Key       string         `gorm:"index" json:"key"`
	Name      string         `json:"name"`
	Schema    map[string]any `gorm:"type:text;serializer:json" json:"schema,omitzero"`
}

func (FormModel) TableName() string {
	return "forms"
}
