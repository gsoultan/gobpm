package entities

// ProcessEvent represents an event in the process lifecycle.
type ProcessEvent struct {
	Type      string           `json:"type"`
	Instance  *ProcessInstance `json:"instance,omitzero"`
	Project   *Project         `json:"project,omitzero"`
	Node      *Node            `json:"node,omitzero"`
	Timestamp int64            `json:"timestamp"`
	Variables map[string]any   `json:"variables,omitzero"`
}

const (
	EventProcessStarted   = "ProcessStarted"
	EventNodeReached      = "NodeReached"
	EventTaskCreated      = "TaskCreated"
	EventTaskCompleted    = "TaskCompleted"
	EventTaskUpdated      = "TaskUpdated"
	EventTaskClaimed      = "TaskClaimed"
	EventProcessCompleted = "ProcessCompleted"
)
