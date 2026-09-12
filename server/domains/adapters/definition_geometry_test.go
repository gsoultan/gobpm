package adapters

import (
	"testing"

	"github.com/gsoultan/metis/server/domains/entities"
)

// TestGeometrySurvivesTheAdapters is the save path for an imported diagram.
//
// The adapters copy field by field, so a field with no counterpart on the other
// side is dropped without a word — the same way ErrorCode was, which turned
// every error boundary event into a catch-all. Here the symptom would be a
// diagram that imports and renders correctly and loses its layout the first
// time anybody presses save.
func TestGeometrySurvivesTheAdapters(t *testing.T) {
	def := &entities.ProcessDefinition{
		Key:  "geo",
		Name: "Geometry",
		Nodes: []*entities.Node{
			{
				ID: "sub", Type: entities.SubProcess,
				X: 100, Y: 200, Width: 350, Height: 180, IsExpanded: true,
				Nodes: []*entities.Node{
					{ID: "inner", Type: entities.UserTask, X: 130, Y: 240, Width: 100, Height: 80},
				},
				Flows: []*entities.SequenceFlow{
					{ID: "inner-f", SourceRef: "inner", TargetRef: "inner",
						Waypoints: []entities.Waypoint{{X: 1, Y: 2}, {X: 3, Y: 4}}},
				},
			},
		},
		Flows: []*entities.SequenceFlow{
			{ID: "f1", SourceRef: "a", TargetRef: "sub",
				Waypoints: []entities.Waypoint{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}},
		},
	}

	back := DefinitionEntityAdapter{Model: DefinitionModelAdapter{Definition: def}.ToModel()}.ToEntity()

	sub := back.FindNode("sub")
	if sub == nil {
		t.Fatal("the sub-process did not survive the adapters")
	}
	if sub.X != 100 || sub.Y != 200 || sub.Width != 350 || sub.Height != 180 {
		t.Errorf("sub-process bounds: got x=%d y=%d w=%d h=%d, want 100,200,350,180",
			sub.X, sub.Y, sub.Width, sub.Height)
	}
	if !sub.IsExpanded {
		t.Error("the sub-process came back collapsed, so its children would render on top of it")
	}

	inner := back.FindNode("inner")
	if inner == nil {
		t.Fatal("the node inside the sub-process did not survive the adapters")
	}
	if inner.Width != 100 || inner.Height != 80 {
		t.Errorf("nested node size: got w=%d h=%d, want 100,80", inner.Width, inner.Height)
	}

	if len(back.Flows) != 1 || len(back.Flows[0].Waypoints) != 3 {
		t.Fatalf("top-level flow waypoints: got %d, want 3", len(back.Flows[0].Waypoints))
	}
	if back.Flows[0].Waypoints[2] != (entities.Waypoint{X: 50, Y: 60}) {
		t.Errorf("last waypoint: got %+v, want {50 60}", back.Flows[0].Waypoints[2])
	}
	if len(sub.Flows) != 1 || len(sub.Flows[0].Waypoints) != 2 {
		t.Errorf("flow inside the sub-process lost its waypoints: got %d, want 2", len(sub.Flows[0].Waypoints))
	}
}
