package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/storm/runtime"
)

type (
	txKey          struct{}
	environmentKey struct{}
)

// withTx puts an open transaction on the context, so everything downstream
// joins it rather than opening its own.
func withTx(ctx context.Context, tx runtime.Executor) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFrom(ctx context.Context) (runtime.Executor, bool) {
	tx, ok := ctx.Value(txKey{}).(runtime.Executor)
	return tx, ok
}

// WithEnvironment binds work to one of a project's runtimes.
//
// Applied by the listener a request arrived on: each environment is served on
// its own port, so which runtime a caller is working in is decided by where
// they connected, not by anything they send. An environment named in a header
// would be a value the caller chooses, and choosing production is exactly what
// a staging user must not be able to do.
func WithEnvironment(ctx context.Context, environmentID uuid.UUID) context.Context {
	return context.WithValue(ctx, environmentKey{}, environmentID)
}

// EnvironmentFrom reports which runtime this work belongs to. Absent means the
// main database, which is where accounts, projects and the environment registry
// itself live.
func EnvironmentFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(environmentKey{}).(uuid.UUID)
	return id, ok && id != uuid.Nil
}
