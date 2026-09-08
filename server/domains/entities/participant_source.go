package entities

import (
	"time"

	"github.com/google/uuid"
)

// What a source does about somebody it has stopped naming.
const (
	// OnMissingLeave keeps them. The safe default, and the right one for a feed
	// that covers part of an organization: syncing a finance directory must not
	// take work away from everybody outside finance.
	OnMissingLeave = "leave"
	// OnMissingDeactivate stops them receiving new work, keeping their history.
	// For a source that genuinely is where the answer lives.
	OnMissingDeactivate = "deactivate"
)

// ParticipantSource is a directory a project keeps its participants in step
// with.
type ParticipantSource struct {
	ID      uuid.UUID `json:"id"`
	Project *Project  `json:"project,omitzero"`

	Name string `json:"name"`
	// Kind is http or postgres.
	Kind string `json:"kind"`

	// Config is the endpoint and headers, or the connection string and query.
	// Leaving the service, credentials in it are masked.
	Config map[string]any `json:"config,omitzero"`

	// Schedule is an ISO 8601 repeating interval — "R/PT1H". Empty means manual
	// only.
	Schedule string `json:"schedule,omitzero"`
	// OnMissing is leave or deactivate.
	OnMissing string `json:"on_missing"`
	Enabled   bool   `json:"enabled"`

	LastRun *SourceRun `json:"last_run,omitzero"`
}

// SourceRun is how a source's last sync went.
//
// Kept on the source rather than only in logs, because a sync that has been
// failing quietly for a month is exactly what nobody notices.
type SourceRun struct {
	At      time.Time `json:"at"`
	OK      bool      `json:"ok"`
	Detail  string    `json:"detail,omitzero"`
	Created int       `json:"created"`
	Updated int       `json:"updated"`
}
