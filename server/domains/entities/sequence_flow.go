package entities

// SequenceFlow represents a connection between two nodes.
type SequenceFlow struct {
	ID            string `json:"id"`
	SourceRef     string `json:"source_ref"`
	TargetRef     string `json:"target_ref"`
	Condition     string `json:"condition,omitzero"`
	Documentation string `json:"documentation,omitzero"`
}
