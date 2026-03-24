package entities

import (
	"time"
)

// SLAStatus represents the compliance state of an SLA entry.
type SLAStatus string

const (
	SLAStatusOnTrack  SLAStatus = "on_track"
	SLAStatusAtRisk   SLAStatus = "at_risk"
	SLAStatusBreached SLAStatus = "breached"
)

// SLAEntry represents the SLA compliance record for a single process instance
// at a specific BPMN node. Used in the SLA compliance dashboard and CSV exports.
type SLAEntry struct {
	// Instance is the process instance being tracked.
	Instance *ProcessInstance
	// NodeID is the BPMN node where the SLA applies.
	Node *Node
	// DueAt is the deadline by which this node must be completed.
	DueAt time.Time
	// CompletedAt is when the node was actually completed (zero if still active).
	CompletedAt *time.Time
	// DurationMs is how long the instance has been (or was) waiting at this node.
	DurationMs int64
	// Status indicates whether the SLA is on track, at risk, or breached.
	Status SLAStatus
}
