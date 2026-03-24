package models

import (
	"github.com/google/uuid"
)

type AuditModel struct {
	Base
	ProjectID  uuid.UUID      `gorm:"type:uuid;index" json:"project_id,omitzero"`
	InstanceID uuid.UUID      `gorm:"type:uuid;index" json:"instance_id,omitzero"`
	Type       string         `json:"type"`
	NodeID     string         `json:"node_id,omitzero"`
	NodeName   string         `json:"node_name,omitzero"`
	Message    string         `json:"message"`
	Narrative  string         `json:"narrative,omitzero"`
	Data       map[string]any `gorm:"type:text;serializer:json" json:"data,omitzero"`
}

func (AuditModel) TableName() string {
	return "audit_logs"
}
