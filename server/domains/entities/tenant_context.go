package entities

import "context"

type tenantContextKey struct{}

// TenantContext carries the active tenant identifier through a request context.
// It is injected by TenantMiddleware and validated on every repository query.
type TenantContext struct {
	// TenantID is the unique identifier for the tenant (maps to OrganizationID).
	TenantID string
}

// WithTenantContext returns a new context with the given TenantContext attached.
func WithTenantContext(ctx context.Context, tc TenantContext) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tc)
}

// TenantContextFrom extracts the TenantContext from the context.
// Returns the zero value and false if no TenantContext is present.
func TenantContextFrom(ctx context.Context) (TenantContext, bool) {
	tc, ok := ctx.Value(tenantContextKey{}).(TenantContext)
	return tc, ok
}
