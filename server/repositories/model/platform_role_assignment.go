package model

import "github.com/gsoultan/storm"

// PlatformRoleAssignment grants a role to a platform account.
//
// The pair is the primary key, so granting the same role twice is refused by
// the database rather than by whoever remembered to check first — and revoking
// once is enough.
type PlatformRoleAssignment struct {
	PlatformUser PlatformUser
	PlatformRole PlatformRole
}

func (a *PlatformRoleAssignment) Schema(t *storm.Table) {
	t.PrimaryKey(&a.PlatformUser, &a.PlatformRole)
	// Removing an account takes its grants with it. Leaving them would let a
	// later account reusing the id inherit them.
	t.Col(&a.PlatformUser).OnDelete(storm.Cascade)
	t.Col(&a.PlatformRole).OnDelete(storm.Cascade)
}
