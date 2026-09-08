package entities

import (
	"context"

	"github.com/google/uuid"
)

// SSEScope is who an event is for.
//
// The live event stream carries process variables — an amount, an applicant's
// name, an approval decision. A client registry that does not know who is
// listening delivers all of it to every browser with the stream open, which is
// every organization's business data handed to every signed-in user.
//
// Two dimensions, because there are two ways an event can reach somebody it is
// not for:
//
//   - Organization: the tenant the work belongs to.
//   - Environment: which of a project's runtimes it happened in. A browser on
//     the staging port must not see production, and the whole point of giving
//     each environment its own database is that its data never appears
//     elsewhere.
//
// The zero value is not "everybody". It is a scope that matches only other
// zero values, which is what makes an unresolved scope a delivery that does not
// happen rather than one that goes everywhere.
type SSEScope struct {
	Organization uuid.UUID
	Environment  uuid.UUID
}

// Delivers reports whether an event in this scope should reach a client in the
// other.
//
// Equality, deliberately. There is no wildcard and no "unscoped means all":
// every widening rule anybody could add here is a rule that, misapplied once,
// puts one tenant's variables on another tenant's screen.
func (s SSEScope) Delivers(to SSEScope) bool {
	return s == to
}

// Resolved reports whether this scope names an organization.
//
// An event with no organization cannot be delivered to anybody, because no
// signed-in client has an empty one. Callers use this to say so out loud rather
// than to silently drop it.
func (s SSEScope) Resolved() bool {
	return s.Organization != uuid.Nil
}

// SSEScopeFrom reads the scope a request or a unit of work belongs to.
//
// The environment comes from the listener the request arrived on, and the
// organization from the token. Neither is anything the caller chooses.
//
// System work — the job worker, the timer sweep — carries no tenant, so this
// returns an unresolved scope for it. That is not a gap to paper over: an event
// produced by background work has to have its organization resolved from the
// work itself, which is what SSEObserver's project resolver does.
func SSEScopeFrom(ctx context.Context) SSEScope {
	var scope SSEScope
	if tenant, ok := TenantContextFrom(ctx); ok {
		if id, err := uuid.Parse(tenant.TenantID); err == nil {
			scope.Organization = id
		}
	}
	if environmentID, bound := EnvironmentFrom(ctx); bound {
		scope.Environment = environmentID
	}
	return scope
}
