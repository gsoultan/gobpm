package contracts

import (
	"context"
	"time"
)

// DistributedLocker defines the Strategy interface for acquiring and releasing
// distributed locks across multiple engine replicas. Implementations must be
// swappable (PostgreSQL advisory locks, Redis SET NX, etc.) without changing callers.
type DistributedLocker interface {
	// TryAcquire attempts to acquire a lock for the given key with the given TTL.
	// Returns true if the lock was acquired, false if it is already held.
	TryAcquire(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Release releases the lock for the given key.
	Release(ctx context.Context, key string) error
}
