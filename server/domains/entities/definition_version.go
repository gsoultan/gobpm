package entities

import (
	"time"

	"github.com/google/uuid"
)

// DefinitionVersionStatus is one version of a process key, with enough context
// to decide whether it is safe to retire.
//
// The two counts are the point. A version stops receiving work the moment
// another one is promoted, but it keeps executing everything already started on
// it — instances pin their definition by ID and never move. "Has v2 finished?"
// is therefore a question about Running, not about which version is live, and
// without it the answer was only discoverable by reading the instance list and
// counting by eye.
type DefinitionVersionStatus struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at,omitzero"`

	// Live is true for the single version new instances start on.
	Live bool `json:"live"`

	// RunningInstances is what still has to finish before this version has
	// drained. Suspended instances count: they have not finished, and they
	// resume on this version.
	RunningInstances int64 `json:"running_instances"`

	// TotalInstances is every instance ever started on this version.
	TotalInstances int64 `json:"total_instances"`

	// ScheduledFor is when this version is arranged to take over, when a cutover
	// naming it is still in the future. Zero when none is.
	//
	// The soonest one, if somebody has arranged more than one: cancelling it
	// reveals the next. Showing every future entry per row would make the common
	// case — one cutover, one date — read like a list.
	ScheduledFor time.Time `json:"scheduled_for,omitzero"`
	// ScheduledReleaseID names the timeline entry behind ScheduledFor, because
	// cancelling has to identify the entry rather than the version.
	ScheduledReleaseID uuid.UUID `json:"scheduled_release_id,omitzero"`
}

// Draining reports whether this version is still executing work it will never
// receive more of — the state between being replaced and being safe to delete.
func (v DefinitionVersionStatus) Draining() bool {
	return !v.Live && v.RunningInstances > 0
}

// ScheduledVersion is a cutover that has been arranged but has not happened yet.
//
// It carries its own ID because cancelling one has to name the entry rather than
// the version: the same version can be scheduled more than once, and "cancel
// v3" would be ambiguous about which of them.
type ScheduledVersion struct {
	ID      uuid.UUID `json:"id"`
	Key     string    `json:"key"`
	Version int       `json:"version"`
	// ActivateAt is when this version takes over, in UTC.
	ActivateAt time.Time `json:"activate_at"`
}
