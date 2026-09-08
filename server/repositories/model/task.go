package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// Task is a unit of human work an instance is waiting on.
//
// Assignee and the candidate lists are usernames, not foreign keys to User: a
// process is authored against people who may have no account, and the runtime
// moves between environments without the account list.
type Task struct {
	storm.Model

	Project  Project
	Instance ProcessInstance

	NodeID      string
	Name        string
	Description *string
	Type        NodeType
	Status      TaskStatus

	Assignee        *string
	CandidateUsers  storm.JSON
	CandidateGroups storm.JSON

	// Priority and DueDate are what the inbox computes urgency from. Both are
	// authored on the node and copied here when the task is created.
	Priority int
	DueDate  *time.Time

	FormKey        *string
	FormDefinition *string

	// Variables is jsonb, encrypted by the repository for the same reason an
	// instance's are.
	Variables storm.JSON

	DeletedAt *time.Time
}

func (t2 *Task) Schema(t *storm.Table) {
	t.Col(&t2.NodeID).Size(255)
	t.Col(&t2.Type).Size(64)
	t.Col(&t2.Type).Index()
	t.Col(&t2.Status).Size(32)
	t.Col(&t2.Status).Index()
	t.Col(&t2.Assignee).Size(255)
	t.Col(&t2.Assignee).Index()
	t.Col(&t2.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&t2.DeletedAt)
}
