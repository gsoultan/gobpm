package entities

// ExecutionPath represents the historical path of a process instance.
type ExecutionPath struct {
	Nodes       []*Node        `json:"nodes"`
	Frequencies map[string]int `json:"frequencies"`
}
