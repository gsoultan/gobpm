package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// ProcessInstance is one running execution of a process definition.
//
// Definition is a hard reference to the exact version it started on, and it
// never changes: an instance finishes on the graph it began with, whatever is
// deployed afterwards. That is the guarantee the whole versioning story rests
// on, and it is why deleting a definition an instance references is refused.
type ProcessInstance struct {
	storm.Model

	Project    Project
	Definition ProcessDefinition

	// A sub-process instance knows the instance and node that called it.
	ParentInstance *ProcessInstance
	ParentNodeID   *string

	Status ProcessStatus

	// Variables is the business data, jsonb and encrypted by the repository:
	// it holds whatever the process was started with, which is routinely
	// personal or commercial.
	Variables storm.JSON

	// Tokens are the live points of execution — see the Token payload struct.
	Tokens storm.JSON
	// CompletedNodes and CompensatedNodes are jsonb arrays of node id.
	CompletedNodes   storm.JSON
	CompensatedNodes storm.JSON

	// MultiInstance and Joins are engine bookkeeping, kept out of Variables so
	// they cannot collide with business data or reach the audit trail through
	// it.
	MultiInstance storm.JSON
	Joins         storm.JSON

	DeletedAt *time.Time
}

func (i *ProcessInstance) Schema(t *storm.Table) {
	t.Col(&i.Status).Size(32)
	t.Col(&i.Status).Index()
	// Self-referential: a sub-process instance names the one that called it.
	// Set to null rather than cascading, so removing a parent does not delete
	// the record of work its children actually did.
	t.Col(&i.ParentInstance).OnDelete(storm.SetNull).Index()
	t.Col(&i.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&i.DeletedAt)
}
