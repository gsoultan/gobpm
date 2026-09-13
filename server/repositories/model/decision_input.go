package model

// DecisionInput is one column on the left of a decision table.
//
// Not a table: it lives inside DecisionDefinition.Inputs as jsonb.
type DecisionInput struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Expression is evaluated against the process variables to produce the
	// value this column is matched on.
	Expression string `json:"expression"`
	// Type is string, number or boolean.
	Type string `json:"type"`
}
