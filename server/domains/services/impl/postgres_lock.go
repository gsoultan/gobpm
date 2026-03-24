package impl

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"time"
)

// PostgresLocker is a Strategy implementation of DistributedLocker that uses
// PostgreSQL advisory locks to coordinate job acquisition across engine replicas.
//
// TTL behaviour:
// pg_advisory_lock is session-scoped: the lock is automatically released when
// the database session closes (e.g. on worker crash or connection loss), which
// provides a natural dead-lock-avoidance guarantee.  However, the ttl parameter
// accepted by TryAcquire is intentionally IGNORED because advisory locks have
// no time-based expiry mechanism – the lock exists until Release is called or
// the session ends.
//
// Implications for callers:
//   - Always pair every successful TryAcquire with a Release call.
//   - If you need TTL-based expiry (e.g. to reclaim locks from long-running but
//     alive workers), use a different backend (Redis SET NX PX, etcd, etc.) or
//     implement a heartbeat loop that re-acquires the lock periodically.
type PostgresLocker struct {
	db *sql.DB
}

// NewPostgresLocker creates a new PostgresLocker backed by the given *sql.DB.
func NewPostgresLocker(db *sql.DB) *PostgresLocker {
	return &PostgresLocker{db: db}
}

// TryAcquire attempts a non-blocking PostgreSQL advisory lock using pg_try_advisory_lock.
// The key string is hashed to a 64-bit integer for the advisory lock ID.
// TTL is not enforced by Postgres advisory locks; callers must call Release explicitly.
func (l *PostgresLocker) TryAcquire(ctx context.Context, key string, _ time.Duration) (bool, error) {
	lockID := hashKey(key)
	var acquired bool
	err := l.db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("postgres advisory lock acquire %q: %w", key, err)
	}
	return acquired, nil
}

// Release releases the PostgreSQL advisory lock for the given key.
func (l *PostgresLocker) Release(ctx context.Context, key string) error {
	lockID := hashKey(key)
	_, err := l.db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockID)
	if err != nil {
		return fmt.Errorf("postgres advisory lock release %q: %w", key, err)
	}
	return nil
}

// hashKey converts an arbitrary string key into a stable int64 for advisory lock IDs.
func hashKey(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int64(h.Sum64())
}
