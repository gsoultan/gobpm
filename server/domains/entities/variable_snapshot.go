package entities

import (
	"time"

	"github.com/google/uuid"
)

// VariableSnapshot captures the full variable state of a process instance at a
// specific point in time. One snapshot is written each time UpdateInstance is called,
// enabling audit compliance and time-travel debugging of variable values.
type VariableSnapshot struct {
	// ID is the unique snapshot record ID.
	ID uuid.UUID
	// Instance is the process instance whose variables were captured.
	Instance *ProcessInstance
	// NodeID identifies the BPMN node that triggered the snapshot (e.g., the task that just completed).
	Node *Node
	// Variables is the full variable map at the moment of capture.
	Variables map[string]any
	// CapturedAt is when the snapshot was taken.
	CapturedAt time.Time
}
