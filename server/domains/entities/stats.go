package entities

// ProcessStatistics represents high-level metrics for a project or the system.
type ProcessStatistics struct {
	ActiveInstances    int            `json:"active_instances"`
	CompletedInstances int            `json:"completed_instances"`
	FailedInstances    int            `json:"failed_instances"`
	TotalTasks         int            `json:"total_tasks"`
	PendingTasks       int            `json:"pending_tasks"`
	NodeFrequencies    map[string]int `json:"node_frequencies,omitzero"`
}
