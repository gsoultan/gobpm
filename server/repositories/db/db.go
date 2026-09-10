// Package db is the connection layer under the storm repositories.
//
// It answers one question — which database does this work belong to — and it is
// the only place that answers it. Repositories ask for an executor and get the
// right one: the transaction they are already inside, or their environment's
// database, or the main one.
//
// It replaces gorms.GetTx, which did the same job for GORM. The shape is kept
// deliberately: every read and write goes through a single resolver, so the
// choice is a property of the request rather than something each of the thirty
// repositories has to remember.
package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gsoultan/storm/runtime"
	"github.com/gsoultan/storm/runtime/pgxdrv"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ErrEnvironmentUnavailable is returned for work bound to an environment whose
// database is not open.
//
// It fails closed rather than falling back to the main database, because the
// fallback would be silent, would succeed, and would put one runtime's rows in
// another's store — which is precisely what running separate environments is
// meant to make impossible.
var ErrEnvironmentUnavailable = errors.New("this environment's database is not available")

// Conn holds the open pools: the main database, and one per environment.
type Conn struct {
	main *pgxpool.Pool

	mu   sync.RWMutex
	envs map[uuid.UUID]*pgxpool.Pool
}

// Open connects to the main database.
//
// The pool comes from storm's constructor rather than pgxpool directly, because
// that is what installs the fast parameter encoders the generated code assumes.
func Open(ctx context.Context, dsn string) (*Conn, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("could not read the connection string: %w", err)
	}

	// A pooled connection whose TCP has died looks alive until something is
	// sent on it, so the first request after a network blip fails on a
	// connection that was already gone. database/sql hid this by retrying once
	// on driver.ErrBadConn; pgx has no equivalent, and adding one here would
	// mean retrying statements that may already have run.
	//
	// So the pool checks instead. Idle connections are reaped and every one is
	// health-checked, which turns "the first caller after an outage gets an
	// error" into "the pool noticed before anybody asked".
	config.HealthCheckPeriod = 5 * time.Second
	config.MaxConnIdleTime = 30 * time.Second
	// Bounded lifetime as well, because a connection can also be quietly broken
	// by something between here and the server that keeps the socket open — a
	// load balancer, a firewall idle timeout — which no health check on this end
	// will notice until it is used.
	config.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("could not open the main database: %w", err)
	}
	return &Conn{main: pool, envs: map[uuid.UUID]*pgxpool.Pool{}}, nil
}

// Main is the pool for the main database, for the migration runner and for
// anything that legitimately works across environments.
func (c *Conn) Main() *pgxpool.Pool { return c.main }

// RegisterEnvironment records the open pool for one environment, returning any
// it displaced so the caller can close it once nothing is using it. Closing it
// here would break requests still in flight on it.
func (c *Conn) RegisterEnvironment(id uuid.UUID, pool *pgxpool.Pool) (replaced *pgxpool.Pool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	replaced = c.envs[id]
	c.envs[id] = pool
	return replaced
}

// ForgetEnvironment drops an environment's pool and hands it back to be closed.
func (c *Conn) ForgetEnvironment(id uuid.UUID) (*pgxpool.Pool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pool, ok := c.envs[id]
	delete(c.envs, id)
	return pool, ok
}

// Close releases every pool.
func (c *Conn) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, pool := range c.envs {
		pool.Close()
	}
	c.envs = map[uuid.UUID]*pgxpool.Pool{}
	c.main.Close()
}

// Executor resolves which database this work belongs to.
//
// A transaction already on the context wins: work inside a unit of work must
// join it, not open a second connection alongside it and deadlock against
// itself.
func (c *Conn) Executor(ctx context.Context) (runtime.Executor, error) {
	if tx, ok := txFrom(ctx); ok {
		return tx, nil
	}
	pool, err := c.poolFor(ctx)
	if err != nil {
		return nil, err
	}
	return pgxdrv.Pool{P: pool}, nil
}

// Outside resolves an executor that deliberately ignores an enclosing
// transaction.
//
// For work that is a fact about something which has already happened and must
// not be undone with the caller's write: the SSE bus is the case, where joining
// the transaction would mean a rollback silently un-notifying browsers about
// work that did commit earlier in the same handler, and would hold the row
// invisible until commit — exactly when it is least useful.
//
// The environment binding still applies. Which database the work belongs to is
// not a transaction question.
func (c *Conn) Outside(ctx context.Context) (runtime.Executor, error) {
	pool, err := c.poolFor(ctx)
	if err != nil {
		return nil, err
	}
	return pgxdrv.Pool{P: pool}, nil
}

// poolFor picks the pool for this work, without regard to transactions.
func (c *Conn) poolFor(ctx context.Context) (*pgxpool.Pool, error) {
	environmentID, bound := EnvironmentFrom(ctx)
	if !bound {
		return c.main, nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	pool, open := c.envs[environmentID]
	if !open {
		return nil, fmt.Errorf("%w: %s", ErrEnvironmentUnavailable, environmentID)
	}
	return pool, nil
}

// Transact runs fn inside a database transaction.
//
// Already being in one reuses it rather than nesting: a repository that opened
// its own would have its writes commit or roll back independently of the work
// around them, which is the bug that made five GORM repositories silently
// escape every transaction they ran inside.
//
// Rollback on panic as well as on error, because a panic that leaves a
// transaction open holds its locks until the connection is reaped.
func (c *Conn) Transact(ctx context.Context, fn func(context.Context) error) (err error) {
	if _, inTx := txFrom(ctx); inTx {
		return fn(ctx)
	}

	pool, err := c.poolFor(ctx)
	if err != nil {
		return err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin: %w", err)
	}

	// The rollback reads the named result, which every return below assigns
	// before deferred functions run — so the inner errors do not need to be
	// hoisted into it by hand.
	defer func() {
		if recovered := recover(); recovered != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				// Logged rather than discarded: the panic is on its way up and
				// cannot carry this, but a transaction that would not roll back
				// is holding locks until the connection is reaped.
				log.Warn().Err(rollbackErr).Msg("Could not roll back a transaction while panicking")
			}
			panic(recovered)
		}
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	if err := fn(withTx(ctx, pgxdrv.Tx{T: tx})); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not commit: %w", err)
	}
	return nil
}
