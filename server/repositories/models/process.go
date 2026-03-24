package models

import (
	"time"

	"github.com/google/uuid"
)

// ProcessStatus defines the current state of a process instance in the database.
type ProcessStatus string

const (
	ProcessActive    ProcessStatus = "active"
	ProcessCompleted ProcessStatus = "completed"
	ProcessSuspended ProcessStatus = "suspended"
	ProcessFailed    ProcessStatus = "failed"
)

// TokenStatus represents the current state of a token in the database.
type TokenStatus string

const (
	TokenActive    TokenStatus = "active"
	TokenSuspended TokenStatus = "suspended"
	TokenCompleted TokenStatus = "completed"
)

// Token represents a single point of execution in a process instance in the database.
type Token struct {
	ID          uuid.UUID      `json:"id"`
	InstanceID  uuid.UUID      `json:"instance_id"`
	NodeID      string         `json:"node_id"`
	Status      TokenStatus    `json:"status"`
	IterationID string         `json:"iteration_id,omitzero"`
	Variables   map[string]any `json:"variables,omitzero"`
	CreatedAt   time.Time      `json:"created_at,omitzero"`
}

// ProcessInstanceModel represents the GORM model for process instances.
type ProcessInstanceModel struct {
	Base
	ProjectID        uuid.UUID              `gorm:"type:uuid;index" json:"project_id,omitzero"`
	Project          ProjectModel           `gorm:"foreignKey:ProjectID" json:"project,omitzero"`
	DefinitionID     uuid.UUID              `gorm:"type:uuid;index" json:"definition_id,omitzero"`
	Definition       ProcessDefinitionModel `gorm:"foreignKey:DefinitionID" json:"definition,omitzero"`
	ParentInstanceID *uuid.UUID             `gorm:"type:uuid;index" json:"parent_instance_id,omitzero"`
	ParentNodeID     string                 `json:"parent_node_id,omitzero"`
	Status           ProcessStatus          `gorm:"index" json:"status"`
	Variables        EncryptedMap           `gorm:"type:text" json:"variables,omitzero"`
	Tokens           []Token                `gorm:"type:text;serializer:json" json:"tokens,omitzero"`
	CompletedNodes   []string               `gorm:"type:text;serializer:json" json:"completed_nodes,omitzero"`
	CompensatedNodes []string               `gorm:"type:text;serializer:json" json:"compensated_nodes,omitzero"`
}

// TableName overrides the table name for ProcessInstanceModel.
func (ProcessInstanceModel) TableName() string {
	return "process_instances"
}
