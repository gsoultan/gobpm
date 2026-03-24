package impl

import (
	"context"
	"time"
)

// NoOpLocker is a Null Object implementation of DistributedLocker.
// It always grants the lock and is safe to use in single-instance deployments
// or tests where distributed coordination is not required.
type NoOpLocker struct{}

// NewNoOpLocker creates a new NoOpLocker.
func NewNoOpLocker() *NoOpLocker {
	return &NoOpLocker{}
}

func (l *NoOpLocker) TryAcquire(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

func (l *NoOpLocker) Release(_ context.Context, _ string) error {
	return nil
}
