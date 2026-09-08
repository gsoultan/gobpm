package gorms

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/repositories/contracts"

	"gorm.io/gorm"
)

var (
	dbOverrideMu sync.RWMutex
	dbOverride   *gorm.DB
)

// SetDBOverride replaces the active database connection used by all repositories.
// This is called after first-time setup to switch from the temporary SQLite database
// to the user-configured target database without requiring an application restart.
func SetDBOverride(db *gorm.DB) {
	dbOverrideMu.Lock()
	defer dbOverrideMu.Unlock()
	dbOverride = db
}

// ResolveDB returns the override database if set, otherwise the original database.
// ResolveDB applies the hot-swap override. Repository methods must not call it
// directly — use GetTx, which both applies the override and joins any active
// unit-of-work transaction. Five repositories called this directly and their
// writes silently escaped every transaction they ran inside; with SQLite's
// single connection the same call graph deadlocked against itself, which is
// how it was finally noticed.
func ResolveDB(db *gorm.DB) *gorm.DB {
	dbOverrideMu.RLock()
	defer dbOverrideMu.RUnlock()
	if dbOverride != nil {
		return dbOverride
	}
	return db
}

// Config is the GORM configuration every connection to this schema must use.
//
// It exists so there is one answer rather than one per gorm.Open call site.
// Test harnesses must use it too: a setting the tests do not share is a setting
// the tests do not check.
func Config() *gorm.Config {
	return &gorm.Config{
		// TranslateError turns each driver's own way of saying "that row already
		// exists" into gorm.ErrDuplicatedKey. Without it, recognising a unique
		// constraint means matching error text per dialect — four spellings to
		// keep in step, and the wrong one is discovered in production. The
		// definition version allocator depends on it.
		TranslateError: true,
	}
}

type contextKey string

const (
	txKey contextKey = "gorm_tx"
)

type gormUnitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork creates a new GORM-based UnitOfWork.
func NewUnitOfWork(db *gorm.DB) contracts.UnitOfWork {
	return &gormUnitOfWork{db: db}
}

func (u *gormUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey).(*gorm.DB); ok {
		// Already in a transaction; reuse it.
		return fn(ctx)
	}
	db, err := resolveForWrite(ctx, u.db)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

// resolveForWrite picks the database a unit of work belongs to.
//
// Separate from ResolveDBFor because a transaction cannot be opened on the
// error-carrying session that one returns for an unavailable environment: the
// session would hand its error to the first statement inside the transaction,
// which is after the transaction has already begun on the wrong connection.
// Here the refusal happens before anything opens.
//
// This used to be a plain ResolveDB, which is how the environments feature
// ended up reading one database and writing another: every read outside a
// transaction went to the environment, and every write — all of them are inside
// a unit of work — went to the main database. A model deployed on the staging
// port landed in production's tables and was invisible to staging.
func resolveForWrite(ctx context.Context, db *gorm.DB) (*gorm.DB, error) {
	environmentID, bound := entities.EnvironmentFrom(ctx)
	if !bound {
		return ResolveDB(db), nil
	}
	environmentDB, open := EnvironmentDB(environmentID)
	if !open {
		return nil, fmt.Errorf("%w: %s", ErrEnvironmentUnavailable, environmentID)
	}
	return environmentDB, nil
}

// Attempt runs fn so that its failure can be recovered from.
//
// Do reuses an enclosing transaction, which is right for work that must succeed
// or roll back with everything around it. It is wrong for work the caller
// intends to retry: on PostgreSQL a failed statement poisons its transaction, so
// the retry runs inside a connection that will refuse everything until rollback,
// and the loop returns the same error until it gives up.
//
// Attempt takes a savepoint instead, so a failed try rolls back to the point
// just before it and leaves the enclosing transaction usable. Outside a
// transaction it behaves exactly like Do.
func (u *gormUnitOfWork) Attempt(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx.Transaction(func(saved *gorm.DB) error {
			return fn(context.WithValue(ctx, txKey, saved))
		})
	}
	return u.Do(ctx, fn)
}

// GetTx retrieves the transaction from the context, if present.
//
// Otherwise it resolves which database this work belongs to — the main one, or
// one of a project's environments. Every repository read and write goes through
// here, which is what makes the choice a property of the request rather than
// something each of them has to remember.
func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return ResolveDBFor(ctx, db).WithContext(ctx)
}

// ErrEnvironmentUnavailable is returned by every query made under an
// environment whose database is not open.
//
// It exists because the alternative is worse than an error. An environment with
// no connection could fall back to the main database, and the query would
// succeed — writing one runtime's data into another's store, which is precisely
// what running separate environments is meant to make impossible. So the
// fallback is a refusal, and it says which environment.
var ErrEnvironmentUnavailable = errors.New("this environment's database is not available")

// ResolveDBFor picks the database this work belongs to.
//
// A context bound to an environment resolves to that environment's connection.
// Anything else — a request against the main database, background work, the
// migration runner — resolves as it always did.
//
// The unavailable case fails closed rather than falling back. See
// ErrEnvironmentUnavailable: a fallback here would be silent, would succeed, and
// would put one runtime's rows in another's database.
func ResolveDBFor(ctx context.Context, db *gorm.DB) *gorm.DB {
	environmentID, bound := entities.EnvironmentFrom(ctx)
	if !bound {
		return ResolveDB(db)
	}
	environmentDB, open := EnvironmentDB(environmentID)
	if open {
		return environmentDB
	}
	// A session carrying an error: GORM returns it from whatever is chained
	// next, so the refusal reaches the caller as an error from their own query
	// rather than as a panic or a wrong answer.
	refused := ResolveDB(db).Session(&gorm.Session{NewDB: true})
	// AddError returns the error it was just handed. Recording it is the whole
	// point — every operation chained off this session now fails with it — so
	// the return value is the input and there is nothing left to decide.
	//nolint:errcheck // The recorded error is the return value; see above.
	refused.AddError(fmt.Errorf("%w: %s", ErrEnvironmentUnavailable, environmentID))
	return refused
}

// MainTx is GetTx for the tables that live in the main database whatever port
// a request arrived on.
//
// Accounts, groups, organizations, projects and the environment registry are
// installation-wide facts. An environment holds a runtime — deployed models,
// instances, tasks — and its schema has those identity tables too, empty,
// because one migration list is easier to keep honest than two.
//
// Without this every request on an environment's port would authenticate
// against that environment's empty users table and be refused: the token is
// valid, the account exists, and the database being asked is the wrong one.
// Worse, an installation that had seeded accounts into an environment database
// would authenticate against a second, divergent copy of who may do what.
//
// An ambient transaction wins only when it is a transaction on the main
// database. Work bound to an environment runs its unit of work on that
// environment's connection, and joining it would ask the empty copy of these
// tables — which is how deploying a model on a staging port failed with "record
// not found" while the project it named was sitting in the main database.
//
// Stepping outside that transaction costs nothing here: these are reads of
// installation-wide rows that the transaction is not writing, so there is no
// atomicity to break. A unit of work on the main database still joins, exactly
// as GetTx does, because there the rows and the transaction are in the same
// place.
func MainTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if _, bound := entities.EnvironmentFrom(ctx); !bound {
		if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
			return tx
		}
	}
	return ResolveDB(db).WithContext(ctx)
}
