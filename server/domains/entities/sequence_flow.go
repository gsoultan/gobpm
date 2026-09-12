package entities

// SequenceFlow represents a connection between two nodes.
type SequenceFlow struct {
	ID            string `json:"id"`
	SourceRef     string `json:"source_ref"`
	TargetRef     string `json:"target_ref"`
	Condition     string `json:"condition,omitzero"`
	Documentation string `json:"documentation,omitzero"`
	// Waypoints are the bend points of the edge as the diagram's author drew it.
	// BPMN requires at least two on an edge, so an empty slice means the flow
	// came from somewhere with no diagram and export has to synthesise them.
	Waypoints []Waypoint `json:"waypoints,omitzero"`
}

// Waypoint is one point on a sequence flow's drawn route.
type Waypoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}
