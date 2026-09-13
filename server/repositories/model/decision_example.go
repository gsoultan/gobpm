package model

// DecisionExample is an input and the result it is expected to produce.
//
// Not a table: examples live inside DecisionDefinition.Tests as jsonb. A
// decision table nobody can test is a spreadsheet with extra steps, so the
// examples travel with the table rather than beside it.
//
// Named DecisionExample rather than DecisionTest: a type whose name ends in
// Test in a _test.go-adjacent package invites confusion with a Go test, and the
// generator would carry the name into a package called decisiontest.
type DecisionExample struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Inputs map[string]any `json:"inputs,omitzero"`
	// Expected is what the table should produce for those inputs.
	Expected map[string]any `json:"expected,omitzero"`
}
