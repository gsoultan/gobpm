package gorms

import (
	"context"
	"sync"

	"github.com/gsoultan/gobpm/server/repositories/contracts"

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
func ResolveDB(db *gorm.DB) *gorm.DB {
	dbOverrideMu.RLock()
	defer dbOverrideMu.RUnlock()
	if dbOverride != nil {
		return dbOverride
	}
	return db
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
	return ResolveDB(u.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

// GetTx retrieves the transaction from the context, if present.
func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return ResolveDB(db).WithContext(ctx)
}
