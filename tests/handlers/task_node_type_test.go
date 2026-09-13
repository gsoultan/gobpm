package handlers_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/adapters"
	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/repositories/models"
)

// Every task response named its node by id alone. A client holding one could
// not tell a user task from a manual task without reading Task.Type and knowing
// that Node.Type — the obvious place — is never filled in.
//
// The type is on the task row already, so this costs nothing to answer.
func TestTaskEntityAdapter_CarriesTheNodeType(t *testing.T) {
	got := adapters.TaskEntityAdapter{Model: models.TaskModel{
		Base:       models.Base{ID: models.UUID(uuid.Must(uuid.NewV7()))},
		InstanceID: models.UUID(uuid.Must(uuid.NewV7())),
		NodeID:     "approve",
		Name:       "Approve the refund",
		Type:       models.NodeType(entities.UserTask),
		Status:     models.TaskStatus(entities.TaskUnclaimed),
	}}.ToEntity()

	if got.Node == nil {
		t.Fatal("the task carries no node")
	}
	if got.Node.ID != "approve" {
		t.Errorf("node id = %q", got.Node.ID)
	}
	if got.Node.Type != entities.UserTask {
		t.Errorf("node type = %q, want %q", got.Node.Type, entities.UserTask)
	}
	// Task.Type stays the authoritative copy; the two must agree rather than
	// one of them being a stale guess.
	if got.Type != got.Node.Type {
		t.Errorf("Task.Type = %q but Node.Type = %q", got.Type, got.Node.Type)
	}
}

// Node.Name is deliberately left empty rather than copied from the task.
//
// UpdateTask can rename a task, after which the task's name is no longer the
// diagram's label for its node. Filling Node.Name from it would report the
// newer of the two with no way for a caller to tell which they were given, so
// the label to display is Task.Name and Node stays an identifier plus a kind.
func TestTaskEntityAdapter_DoesNotGuessTheNodeName(t *testing.T) {
	got := adapters.TaskEntityAdapter{Model: models.TaskModel{
		Base:   models.Base{ID: models.UUID(uuid.Must(uuid.NewV7()))},
		NodeID: "approve",
		Name:   "Renamed by an operator",
		Type:   models.NodeType(entities.UserTask),
	}}.ToEntity()

	if got.Node.Name != "" {
		t.Errorf("node name = %q; it is the task's name, which a rename detaches from the diagram", got.Node.Name)
	}
	if got.Name != "Renamed by an operator" {
		t.Errorf("task name = %q", got.Name)
	}
}

// A manual task is a different element from a user task, and an inbox renders
// them differently — the type has to survive for tasks that are not userTask.
func TestTaskEntityAdapter_CarriesEveryNodeType(t *testing.T) {
	for _, nodeType := range []entities.NodeType{
		entities.UserTask, entities.ManualTask, entities.ServiceTask,
	} {
		got := adapters.TaskEntityAdapter{Model: models.TaskModel{
			Base:   models.Base{ID: models.UUID(uuid.Must(uuid.NewV7()))},
			NodeID: "n1",
			Type:   models.NodeType(nodeType),
		}}.ToEntity()

		if got.Node.Type != nodeType {
			t.Errorf("node type = %q, want %q", got.Node.Type, nodeType)
		}
	}
}
