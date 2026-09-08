package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/entities"
)

// PlatformUserRepository stores the accounts that administer Metis, and what
// they may do.
type PlatformUserRepository interface {
	List(ctx context.Context) ([]entities.PlatformUser, error)
	Get(ctx context.Context, id uuid.UUID) (entities.PlatformUser, error)

	// Create adds an account. The hash is passed separately from the entity so
	// a password cannot travel on a struct that is otherwise returned to
	// callers.
	Create(ctx context.Context, account entities.PlatformUser, passwordHash string) (uuid.UUID, error)
	Update(ctx context.Context, account entities.PlatformUser) error

	// Delete removes an account. Refused for the last remaining administrator:
	// an installation nobody can administer cannot be repaired from inside it.
	Delete(ctx context.Context, id uuid.UUID) error

	// SetRoles replaces an account's grants.
	SetRoles(ctx context.Context, id uuid.UUID, roles []string) error

	ListRoles(ctx context.Context) ([]entities.PlatformRole, error)

	// EnsureBuiltInRoles creates the roles the installation ships with, if they
	// are not there. Idempotent, and called at boot.
	EnsureBuiltInRoles(ctx context.Context) error

	// CountAdministrators reports how many accounts still hold the admin role.
	CountAdministrators(ctx context.Context) (int, error)
}
