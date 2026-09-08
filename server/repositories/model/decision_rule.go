package model

// DecisionRule is one row of a decision table.
//
// Not a table: rules live inside DecisionDefinition.Rules as jsonb, in order,
// because for several hit policies the order is the meaning.
type DecisionRule struct {
	ID string `json:"id"`
	// Inputs are the cell expressions, positionally matched to the table's
	// input columns. An empty cell matches anything.
	Inputs []string `json:"inputs,omitzero"`
	// Outputs are the values this rule produces, positionally matched to the
	// output columns.
	Outputs     []any  `json:"outputs,omitzero"`
	Description string `json:"description,omitzero"`
}
