package models

import (
	"github.com/google/uuid"
)

// DecisionDefinitionModel represents the GORM model for decision definitions.
type DecisionDefinitionModel struct {
	Base
	ProjectID         uuid.UUID        `gorm:"type:uuid;index" json:"project_id,omitzero"`
	Key               string           `gorm:"index" json:"key"`
	Name              string           `json:"name"`
	Version           int              `json:"version"`
	HitPolicy         string           `json:"hit_policy"`
	Aggregation       string           `json:"aggregation,omitzero"`
	RequiredDecisions []string         `gorm:"type:text;serializer:json" json:"required_decisions,omitzero"`
	Inputs            []DecisionInput  `gorm:"type:text;serializer:json" json:"inputs,omitzero"`
	Outputs           []DecisionOutput `gorm:"type:text;serializer:json" json:"outputs,omitzero"`
	Rules             []DecisionRule   `gorm:"type:text;serializer:json" json:"rules,omitzero"`
}

// DecisionInput represents an input column in a decision table in the database.
type DecisionInput struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Expression string `json:"expression"`
	Type       string `json:"type"` // string, number, boolean
}

// DecisionOutput represents an output column in a decision table in the database.
type DecisionOutput struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// DecisionRule represents a rule in a decision table in the database.
type DecisionRule struct {
	ID          string   `json:"id"`
	Inputs      []string `json:"inputs,omitzero"`
	Outputs     []any    `json:"outputs,omitzero"`
	Description string   `json:"description,omitzero"`
}

// TableName overrides the table name for DecisionDefinitionModel.
func (DecisionDefinitionModel) TableName() string {
	return "decision_definitions"
}
