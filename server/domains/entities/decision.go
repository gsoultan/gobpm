package entities

import (
	"github.com/google/uuid"
	"time"
)

const (
	HitPolicyUnique   = "UNIQUE"
	HitPolicyFirst    = "FIRST"
	HitPolicyPriority = "PRIORITY"
	HitPolicyAny      = "ANY"
	HitPolicyCollect  = "COLLECT"

	AggregationSum   = "SUM"
	AggregationCount = "COUNT"
	AggregationMin   = "MIN"
	AggregationMax   = "MAX"
)

// DecisionDefinition represents a DMN decision table or expression.
type DecisionDefinition struct {
	ID                uuid.UUID        `json:"id"`
	Project           *Project         `json:"project,omitzero"`
	Key               string           `json:"key"`
	Name              string           `json:"name"`
	Version           int              `json:"version"`
	HitPolicy         string           `json:"hit_policy"`
	Aggregation       string           `json:"aggregation,omitzero"`
	RequiredDecisions []string         `json:"required_decisions,omitzero"`
	Inputs            []DecisionInput  `json:"inputs,omitzero"`
	Outputs           []DecisionOutput `json:"outputs,omitzero"`
	Rules             []DecisionRule   `json:"rules,omitzero"`
	CreatedAt         time.Time        `json:"created_at,omitzero"`
}

// DecisionInput represents an input column in a decision table.
type DecisionInput struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Expression string `json:"expression"`
	Type       string `json:"type"` // string, number, boolean
}

// DecisionOutput represents an output column in a decision table.
type DecisionOutput struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// DecisionRule represents a rule in a decision table.
type DecisionRule struct {
	ID          string   `json:"id"`
	Inputs      []string `json:"inputs,omitzero"`
	Outputs     []any    `json:"outputs,omitzero"`
	Description string   `json:"description,omitzero"`
}

// DecisionResult is the result of evaluating a decision.
type DecisionResult struct {
	Values map[string]any `json:"values"`
}
