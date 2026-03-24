package entities

import "time"

// HeatmapNode represents aggregated execution statistics for a single BPMN node,
// used to render the live process heatmap overlay on the designer canvas.
type HeatmapNode struct {
	// NodeID is the BPMN flow node identifier.
	Node *Node
	// ActiveCount is the number of process instances currently waiting at this node.
	ActiveCount int
	// CompletedCount is the total number of times this node has been completed.
	CompletedCount int
	// AvgDurationMs is the average time in milliseconds instances spend at this node.
	AvgDurationMs int64
	// MaxDurationMs is the longest time any instance has spent at this node.
	MaxDurationMs int64
	// SLABreached is true when AvgDurationMs exceeds the configured SLA threshold.
	SLABreached bool
	// LastActivity is when this node was last reached by any instance.
	LastActivity time.Time
}

// HeatmapColor returns a traffic-light colour code based on average duration:
// green < 1 hour, yellow < 24 hours, red >= 24 hours.
func (n *HeatmapNode) HeatmapColor() string {
	const oneHourMs = int64(3_600_000)
	const oneDayMs = int64(86_400_000)
	switch {
	case n.AvgDurationMs < oneHourMs:
		return "green"
	case n.AvgDurationMs < oneDayMs:
		return "yellow"
	default:
		return "red"
	}
}
