package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// PlatformUser is an account that administers Metis: configures environments,
// authors process models, manages other accounts.
//
// It is deliberately not the same table as WorkflowUser. The two are different
// populations that happened to share a table: a handful of administrators and
// designers, against everyone in the business who is ever named in a process.
// Sharing one table meant one role list served both, so "can deploy a process
// model" and "can approve an invoice" were the same kind of fact — and granting
// somebody an inbox meant creating them an account on the platform.
//
// This row lives in the main database, because administering the installation
// is not something that happens inside one runtime.
type PlatformUser struct {
	storm.Model

	Username     string
	PasswordHash string

	// TokensValidFrom is when this account's credentials last changed. A token
	// issued before it is refused, which is what makes a password change end
	// every session rather than only the one that changed it.
	TokensValidFrom *time.Time

	FullName    *string
	DisplayName *string
	Email       *string

	// Roles is the has-many through PlatformRoleAssignment. What this account
	// may do is a set of rows, not a JSON array — see PlatformRole.
	Roles []PlatformRoleAssignment

	DeletedAt *time.Time
}

func (u *PlatformUser) Schema(t *storm.Table) {
	t.Col(&u.Username).Size(255)
	// Across the deleted rows: the audit trail records who did what, and a
	// username reissued to a different person makes two people indistinguishable
	// in it — for exactly the accounts that can reconfigure the installation.
	t.UniqueAcrossDeleted(&u.Username)
	t.Col(&u.Email).Size(320)
	t.Col(&u.DeletedAt).Index()
	// Declared, so the predicate is compiled into every read of this table
	// rather than written out at each call site. See the note in doc.go.
	t.SoftDelete(&u.DeletedAt)
}
