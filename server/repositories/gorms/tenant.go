package gorms

import (
	"context"
	"strings"

	"github.com/gsoultan/gobpm/server/domains/entities"
	"gorm.io/gorm"
)

// tenantScopeDB returns a *gorm.DB scoped to the active tenant (organization)
// extracted from the request context via TenantContext. It joins through the
// projects table so list queries only return records belonging to the caller's
// organization.
//
// If no TenantContext is present (e.g. internal/system calls), the original db
// is returned unchanged so the caller can still function without tenant context.
//
// table must be the SQL table name of the model being queried (e.g. "tasks",
// "process_instances") so the JOIN clause can be built correctly.
func tenantScopeDB(ctx context.Context, db *gorm.DB, table string) *gorm.DB {
	tc, ok := entities.TenantContextFrom(ctx)
	if !ok || tc.TenantID == "" {
		return db
	}

	joinClause := strings.ReplaceAll(QueryTenantScopeViaProject, "{table}", table)
	return db.Joins(joinClause, tc.TenantID)
}
