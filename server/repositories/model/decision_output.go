package model

// DecisionOutput is one column on the right of a decision table.
//
// Not a table: it lives inside DecisionDefinition.Outputs as jsonb.
type DecisionOutput struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Name is what the result is called in the variables the decision returns.
	Name string `json:"name"`
	Type string `json:"type"`
	// Values constrains the output to a list, when the column is an enumeration.
	Values []string `json:"values,omitzero"`
}
