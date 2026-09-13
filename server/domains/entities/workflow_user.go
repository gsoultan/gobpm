package entities

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowUser is somebody a process can assign work to.
//
// Scoped to a project and unique within it: two projects naming an approver
// "ada" mean two different people. It carries no roles — what a participant may
// do is decided by the process that names them, and access control lives on the
// platform account, which is a different population.
type WorkflowUser struct {
	ID      uuid.UUID `json:"id"`
	Project *Project  `json:"project,omitzero"`

	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitzero"`
	Email       string `json:"email,omitzero"`

	// Active is how somebody stops receiving work without deleting the record
	// of what they did.
	Active bool `json:"active"`

	// HasCredentials reports whether this participant can sign in yet. A
	// directory imported from a CSV is five hundred names, not five hundred
	// passwords, so a participant routinely exists before they can log in.
	//
	// The hash itself never leaves the repository.
	HasCredentials bool `json:"has_credentials"`

	CreatedAt time.Time `json:"created_at,omitzero"`
}

// WorkflowGroup is a named set of participants a task can be offered to.
//
// Not a role: belonging to one grants nothing except the chance to claim work
// addressed to it.
type WorkflowGroup struct {
	ID          uuid.UUID `json:"id"`
	Project     *Project  `json:"project,omitzero"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitzero"`
}

// ImportSummary is what an import of a participant directory did.
//
// Created and Updated are counted separately because they answer different
// questions: one says how many people are new, the other says how many rows
// overwrote somebody who was already there — which is the number worth checking
// when a file was meant to be additive.
type ImportSummary struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	// Groups is how many groups had to be created to satisfy the file.
	Groups int `json:"groups"`
	// Deactivated is how many people the source stopped naming and was told to
	// stand down. Zero unless the source asks for it — an import leaves people
	// alone, because a partial file is not a statement about everybody it
	// omits.
	Deactivated int `json:"deactivated,omitzero"`
	// Problems are the rows that were not imported, each with its line.
	Problems []ImportProblem `json:"problems,omitzero"`
}

// ImportProblem is one row that could not be imported, and why.
type ImportProblem struct {
	Line     int    `json:"line"`
	Username string `json:"username,omitzero"`
	Reason   string `json:"reason"`
}
