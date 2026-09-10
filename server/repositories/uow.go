package repositories

import (
	"context"

	"github.com/gsoultan/metis/server/repositories/contracts"
	stormdb "github.com/gsoultan/metis/server/repositories/db"
)

// spanningUnitOfWork runs work inside a transaction on *both* persistence
// layers.
//
// While the port is under way a single logical operation can touch repositories
// on either side — advancing a token writes the instance through storm and the
// task through GORM — and the two hold separate connections, so a GORM rollback
// left storm's write committed. That is how a compensation that failed was
// recorded as done: the failure rolled back the GORM half and the storm half
// stayed.
//
// So both are opened, and both are ended together. It is not two-phase commit:
// a crash between the two commits leaves them inconsistent, and nothing here
// pretends otherwise. What it buys is the case that actually happens — an error
// in the middle of a unit of work undoes all of it — and the window closes
// entirely when the last GORM repository moves, at which point this collapses
// back to one transaction.
type spanningUnitOfWork struct {
	gorm  contracts.UnitOfWork
	storm *stormdb.Conn
}

func newSpanningUnitOfWork(gorm contracts.UnitOfWork, storm *stormdb.Conn) contracts.UnitOfWork {
	return &spanningUnitOfWork{gorm: gorm, storm: storm}
}

// Do runs fn with both transactions open.
//
// storm's is the outer one, so its rollback covers the GORM work as well as its
// own: if the inner GORM transaction commits and fn then fails, storm still
// rolls back, and the operation is left half-applied in exactly one direction
// rather than two. Neither ordering is atomic; this one at least fails the same
// way every time.
func (u *spanningUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.storm.Transact(ctx, func(stormCtx context.Context) error {
		return u.gorm.Do(stormCtx, fn)
	})
}

// Attempt is Do with a savepoint, for work the caller intends to retry.
//
// On PostgreSQL a failed statement poisons its transaction, so a retry inside
// one runs against a connection that refuses everything until rollback. A
// savepoint rolls back to just before the attempt and leaves the enclosing
// transaction usable.
func (u *spanningUnitOfWork) Attempt(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.storm.Attempt(ctx, func(stormCtx context.Context) error {
		return u.gorm.Attempt(stormCtx, fn)
	})
}
