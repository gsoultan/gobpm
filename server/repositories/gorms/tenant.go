package gorms

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/tenantscope"
	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/repositories/models"
	"gorm.io/gorm"
)

// QueryDenyAll matches nothing. It is how a query with no tenant identity is
// answered once strict scoping is on: an empty result rather than an error,
// because a repository's contract is rows, and every caller already handles
// finding none.
const QueryDenyAll = "1 = 0"

// tableProjects is the table every scoped read joins through. It used to live
// beside the project repository, which has moved to storm.
const tableProjects = "projects"

// unscopedAccessAllowed decides what a context carrying no tenant may see.
//
// The decision, the reporting and the call-site attribution now live in
// internal/pkg/tenantscope, because the storm repositories replacing this
// package need exactly the same answer — and two copies of "may this query run"
// is one answer that will drift.
func unscopedAccessAllowed(ctx context.Context) bool { return tenantscope.Allowed(ctx) }

// DeniedSites is every path that reached a repository with no identity.
// Re-exported because the strict-scope harness and several tests name it here.
func DeniedSites() []string { return tenantscope.DeniedSites() }

// ResetDeniedSites forgets what has been reported.
func ResetDeniedSites() { tenantscope.ResetDeniedSites() }

// denyAll returns a query guaranteed to match nothing.
//
// An empty result rather than an error, because a repository's contract is rows
// and every caller already handles finding none. What makes that safe to do
// quietly is reportUnidentifiedAccess above, which names the path that got here
// without an identity.
func denyAll(db *gorm.DB) *gorm.DB {
	return db.Where(QueryDenyAll)
}

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
		if unscopedAccessAllowed(ctx) {
			return db
		}
		return denyAll(db)
	}

	joinClause := strings.ReplaceAll(QueryTenantScopeViaProject, "{table}", table)
	return db.Joins(joinClause, tc.TenantID)
}

// tenantScopeCondition is tenantScopeDB expressed as a WHERE predicate instead
// of a JOIN, for locking reads and for statements where join syntax is not
// portable. It scopes the same rows; it just does not widen the statement's
// lock footprint to the projects table.
func tenantScopeCondition(ctx context.Context, db *gorm.DB, table string) *gorm.DB {
	tc, ok := entities.TenantContextFrom(ctx)
	if !ok || tc.TenantID == "" {
		if unscopedAccessAllowed(ctx) {
			return db
		}
		return denyAll(db)
	}

	condition := strings.ReplaceAll(QueryTenantScopeViaProjectSubquery, "{table}", table)
	return db.Where(condition, tc.TenantID)
}

// tenantScopeOrganization scopes a table that carries organization_id directly,
// which today means projects alone.
//
// Projects were the one table nothing scoped, because they are what every other
// scope joins through — so List returned every organization's projects, and a
// caller could name any project as the parent of something they created.
func tenantScopeOrganization(ctx context.Context, db *gorm.DB, table string) *gorm.DB {
	tc, ok := entities.TenantContextFrom(ctx)
	if !ok || tc.TenantID == "" {
		if unscopedAccessAllowed(ctx) {
			return db
		}
		return denyAll(db)
	}

	condition := strings.ReplaceAll(QueryTenantScopeDirect, "{table}", table)
	return db.Where(condition, tc.TenantID)
}

// tenantScopeMembership scopes a table whose tenant is a many-to-many membership
// rather than a column — users, who belong to organizations through
// user_organizations and can belong to more than one.
//
// A subquery rather than a join, because joining the membership table would
// return one row per membership: a user in two of the caller's organizations
// would appear twice in a list that is supposed to be of users.
func tenantScopeMembership(ctx context.Context, db *gorm.DB, table, joinTable, foreignKey string) *gorm.DB {
	tc, ok := entities.TenantContextFrom(ctx)
	if !ok || tc.TenantID == "" {
		if unscopedAccessAllowed(ctx) {
			return db
		}
		return denyAll(db)
	}

	return db.Where(
		table+".id IN (SELECT "+foreignKey+" FROM "+joinTable+" WHERE organization_model_id = ?)",
		tc.TenantID)
}

// requireProjectInTenant returns ErrRecordNotFound unless the project belongs to
// the caller's organization.
//
// This is the check that makes Create safe. A create request names its parent
// project, and without this a caller could point one at another organization's
// project — the row would then be scoped to *that* tenant, so the attacker could
// not even read back what they had planted, but it would be there.
func requireProjectInTenant(ctx context.Context, db *gorm.DB, projectID uuid.UUID) error {
	tc, ok := entities.TenantContextFrom(ctx)
	if !ok || tc.TenantID == "" {
		if unscopedAccessAllowed(ctx) {
			return nil
		}
		return gorm.ErrRecordNotFound
	}
	if projectID == uuid.Nil {
		return nil
	}

	var count int64
	if err := tenantScopeOrganization(ctx, db.Model(&models.ProjectModel{}), tableProjects).
		Where(QualifiedByID(tableProjects), projectID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("could not check project ownership: %w", err)
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// requireVisibleToTenant returns ErrRecordNotFound unless the row is inside the
// caller's tenant scope, and nil when there is no tenant to scope by — engine
// and background work is unguarded here for the same reason the read scope lets
// it through.
//
// Writes are guarded with a scoped read rather than a scoped UPDATE or DELETE
// because neither alternative is safe:
//
//   - GORM's Save ignores a preceding Where. Verified on SQLite, PostgreSQL and
//     MySQL: the row updates anyway. A scope written that way would read as
//     applied and enforce nothing, which is worse than none at all.
//   - Rewriting Save as Model().Where().Updates() changes which columns are
//     written, because Updates skips zero values. Clearing a field would quietly
//     stop persisting.
//
// The cost is a check-then-write window. These rows do not change project in
// normal operation, and when a unit of work is active both statements run inside
// its transaction.
func requireVisibleToTenant(ctx context.Context, db *gorm.DB, table string, model any, id uuid.UUID) error {
	return requireVisible(ctx, tenantScopeCondition, db, table, model, id)
}

func requireVisible(
	ctx context.Context,
	scope func(context.Context, *gorm.DB, string) *gorm.DB,
	db *gorm.DB,
	table string,
	model any,
	id uuid.UUID,
) error {
	tc, ok := entities.TenantContextFrom(ctx)
	if !ok || tc.TenantID == "" {
		if unscopedAccessAllowed(ctx) {
			return nil
		}
		return gorm.ErrRecordNotFound
	}

	var count int64
	if err := scope(ctx, db.Model(model), table).
		Where(QualifiedByID(table), id).
		Count(&count).Error; err != nil {
		return fmt.Errorf("could not check tenant ownership: %w", err)
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// tenantScopeDeploymentResources scopes deployment_resources through their
// parent deployment, which is where the project — and therefore the tenant —
// actually lives.
func tenantScopeDeploymentResources(ctx context.Context, db *gorm.DB) *gorm.DB {
	tc, ok := entities.TenantContextFrom(ctx)
	if !ok || tc.TenantID == "" {
		if unscopedAccessAllowed(ctx) {
			return db
		}
		return denyAll(db)
	}

	return db.Joins(QueryTenantScopeViaDeployment, tc.TenantID)
}
