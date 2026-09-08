package entities

import "github.com/google/uuid"

// MigrationPlan is what moving running instances onto another version would do,
// worked out without writing anything.
//
// It exists because the alternative is asking somebody to authorise a change to
// durable business commitments — instances that are somebody's purchase order,
// somebody's leave request — from a form with no preview. The refusals in
// particular are worth seeing before committing: a mapping that strands a token
// is refused either way, and finding that out from a dry run costs nothing while
// finding it out from a failed apply costs a half-finished cutover.
type MigrationPlan struct {
	SourceKey     string    `json:"source_key"`
	SourceVersion int       `json:"source_version"`
	TargetVersion int       `json:"target_version"`
	TargetID      uuid.UUID `json:"target_id"`

	// Instances is how many running instances would move.
	Instances int `json:"instances"`

	// Moves is where the work currently sits and where it would land, one entry
	// per distinct node, so a plan over a thousand instances is still readable.
	Moves []NodeMove `json:"moves,omitzero"`

	// Refusals are the reasons this would not be applied. A plan with any is
	// one the apply would reject; they are returned rather than raised so the
	// caller sees all of them at once instead of one per attempt.
	Refusals []string `json:"refusals,omitzero"`
}

// Applicable reports whether applying this plan would be accepted.
func (p MigrationPlan) Applicable() bool { return len(p.Refusals) == 0 }

// NodeMove is one node's worth of a plan.
type NodeMove struct {
	From string `json:"from"`
	To   string `json:"to"`
	// Tokens, Tasks and Jobs are how much work sits on From. Separated because
	// they read differently to whoever is deciding: a task is somebody's inbox
	// item and a job is a timer that will fire.
	Tokens int `json:"tokens"`
	Tasks  int `json:"tasks"`
	Jobs   int `json:"jobs"`
	// Mapped is false when From is carried across unchanged because the target
	// has a node of the same id. That is the common case and needs no mapping
	// entry; showing it is how somebody confirms they did not need one.
	Mapped bool `json:"mapped"`
}
