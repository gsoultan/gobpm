package contracts

import "context"

// UnitOfWork defines the interface for managing database transactions.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
