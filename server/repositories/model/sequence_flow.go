package model

// SequenceFlow is an arrow between two nodes.
//
// Not a table: it lives inside ProcessDefinition.Flows as jsonb, beside the
// nodes it connects.
type SequenceFlow struct {
	ID        string `json:"id"`
	SourceRef string `json:"source_ref"`
	TargetRef string `json:"target_ref"`
	// Condition decides whether this arrow is taken. Empty means unconditional.
	Condition     string `json:"condition,omitzero"`
	Documentation string `json:"documentation,omitzero"`
}
