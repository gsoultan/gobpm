package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// AuditEntry is one line of the business timeline: what happened, to which
// node, and in words a person can read.
//
// Narrative is stored rather than generated at read time so the account of what
// happened cannot change when the code that describes it does.
type AuditEntry struct {
	storm.Model

	Project  Project
	Instance ProcessInstance

	Type     string
	NodeID   *string
	NodeName *string
	Message  string
	// Narrative is the plain-English account shown on the timeline.
	Narrative *string
	// Data is the structured detail behind the line, as jsonb.
	Data storm.JSON

	DeletedAt *time.Time
}

func (a *AuditEntry) Schema(t *storm.Table) {
	t.Name("audit_logs")
	t.Col(&a.Type).Size(64)
	t.Col(&a.NodeID).Size(255)
	t.Col(&a.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&a.DeletedAt)
}
