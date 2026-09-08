package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// ParticipantSource is a directory a project keeps its participants in step
// with: an endpoint to call or a query to run, and how often.
//
// A file upload is not one of these. An upload is a thing somebody did once;
// a source is a standing statement that another system is where the answer
// lives, which is what makes syncing it on a schedule meaningful.
type ParticipantSource struct {
	storm.Model

	Project Project

	Name string
	// Kind is http or postgres. A source has to be something that can be read
	// again without a person present.
	Kind string

	// Config holds the endpoint and its headers, or the connection string and
	// the query. Encrypted by the repository: both carry credentials, and a
	// database connection string is every credential in that system at once.
	Config storm.JSON

	// Schedule is an ISO 8601 repeating interval — "R/PT1H" for hourly — which
	// is the same vocabulary BPMN timers already use here. Empty means the
	// source is only ever synced when somebody asks.
	Schedule *string

	// OnMissing decides what happens to somebody the source has stopped
	// naming: "leave" or "deactivate".
	//
	// Defaulting to leave, and stated per source rather than assumed, because
	// the two are very different and neither is obviously right. A directory
	// that is genuinely the source of truth should deactivate; a partial feed
	// covering one department must not, or syncing it would take work away from
	// everybody else.
	OnMissing string

	Enabled bool

	// The outcome of the last run, so somebody can see a source has been
	// failing without reading logs. A sync that has been quietly failing for a
	// month is the failure this exists to make visible.
	LastRunAt      *time.Time
	LastRunOK      *bool
	LastRunDetail  *string
	LastRunCreated int
	LastRunUpdated int

	DeletedAt *time.Time
}

func (s *ParticipantSource) Schema(t *storm.Table) {
	t.Col(&s.Name).Size(255)
	// Live rows only: a removed source's name is available again.
	t.Unique(&s.Project, &s.Name)
	t.Col(&s.Kind).Size(32)
	t.Col(&s.OnMissing).Size(16)
	t.Col(&s.Schedule).Size(64)
	t.Col(&s.Enabled).Index()
	t.Col(&s.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&s.DeletedAt)
}
