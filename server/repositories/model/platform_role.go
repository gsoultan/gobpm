package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// PlatformRole is a named set of things an administrator may do — ADMIN,
// DESIGNER.
//
// A table rather than the JSON array of strings the old user model carried.
// The array had two problems worth fixing here: nothing constrained what went
// into it, so a typo was a role that silently granted nothing; and a role's
// meaning lived only in the Go constants that checked for it, so nobody could
// see what a role actually permitted without reading the interceptors.
type PlatformRole struct {
	storm.Model

	// Name is the identifier the authorization checks compare against.
	Name        string
	Description *string

	// Permissions is a jsonb array of the endpoint names this role admits.
	// Held as data so what a role grants is answerable by reading a row rather
	// than by reading the code that enforces it.
	Permissions storm.JSON

	// BuiltIn marks the roles the installation ships with. They can be granted
	// and revoked but not deleted: an installation with no ADMIN role is one
	// nobody can administer.
	BuiltIn bool

	DeletedAt *time.Time
}

func (r *PlatformRole) Schema(t *storm.Table) {
	t.Col(&r.Name).Size(64)
	// Across the deleted rows: EnsureBuiltInRoles upserts on this name at every
	// boot, and a marked row that stopped conflicting would give the
	// installation two roles called ADMIN, one of which grants nothing.
	t.UniqueAcrossDeleted(&r.Name)
	t.Col(&r.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&r.DeletedAt)
}
