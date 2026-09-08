package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// WorkflowGroup is a named set of participants, so a process can offer a task
// to a team without naming each member in the model.
//
// It is not a role. A group is who a task can be offered to — BPMN's
// candidateGroups — and belonging to one grants nothing except the chance to
// claim work addressed to it. Permissions are PlatformRole's job, on the other
// population entirely.
//
// The old Group table conflated the two: it was the candidate list a task could
// be offered to *and* it carried the platform roles its members inherited.
// Answering both from one row is why granting somebody an inbox also granted
// them the platform.
type WorkflowGroup struct {
	storm.Model

	Project Project

	// Name is what a process model refers to in its candidate groups. Unique
	// within the project, for the same reason a participant's username is.
	Name        string
	Description *string

	Members []WorkflowGroupMembership

	DeletedAt *time.Time
}

func (g *WorkflowGroup) Schema(t *storm.Table) {
	t.Col(&g.Name).Size(255)
	// Across the deleted rows: memberships name the group by id, so a second
	// group of the same name would split one group's membership in two.
	t.UniqueAcrossDeleted(&g.Project, &g.Name)
	t.Col(&g.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&g.DeletedAt)
}
