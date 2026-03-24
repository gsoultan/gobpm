package models

import (
	"time"

	"github.com/google/uuid"
)

type IncidentStatus string

const (
	IncidentOpen     IncidentStatus = "open"
	IncidentResolved IncidentStatus = "resolved"
)

type IncidentModel struct {
	Base
	JobID        uuid.UUID      `gorm:"type:uuid;index" json:"job_id,omitzero"`
	InstanceID   uuid.UUID      `gorm:"type:uuid;index" json:"instance_id,omitzero"`
	DefinitionID uuid.UUID      `gorm:"type:uuid" json:"definition_id,omitzero"`
	NodeID       string         `json:"node_id"`
	Error        string         `json:"error"`
	Status       IncidentStatus `gorm:"index" json:"status"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitzero"`
}

func (IncidentModel) TableName() string {
	return "incidents"
}
