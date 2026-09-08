package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// Resource is one file inside a deployment — a BPMN document, a decision table,
// a form.
type Resource struct {
	storm.Model

	Deployment Deployment

	Name string
	// Content is the file as bytes, whatever it is.
	Content []byte
	Type    string

	DeletedAt *time.Time
}

func (r *Resource) Schema(t *storm.Table) {
	t.Col(&r.Name).Size(255)
	t.Col(&r.Type).Size(64)
	t.Col(&r.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&r.DeletedAt)
}
