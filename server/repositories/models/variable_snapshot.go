package models

import (
	"time"

	"github.com/google/uuid"
)

// VariableSnapshotModel is the GORM persistence model for a process variable snapshot.
type VariableSnapshotModel struct {
	Base
	InstanceID uuid.UUID      `gorm:"type:uuid;index"           json:"instance_id"`
	NodeID     string         `json:"node_id,omitzero"`
	Variables  map[string]any `gorm:"type:text;serializer:json" json:"variables,omitzero"`
	CapturedAt time.Time      `gorm:"index"                     json:"captured_at"`
}

func (VariableSnapshotModel) TableName() string {
	return "variable_snapshots"
}
