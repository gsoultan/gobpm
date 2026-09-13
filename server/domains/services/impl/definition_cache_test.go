package impl

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/entities"
)

// A definition is immutable once deployed but was read and decoded on every
// job, message and timer. These assert the cache does what it is for, and — the
// part that matters more — that it cannot hand one organization's definition to
// another.
func TestCachedDefinitionIsReadOnce(t *testing.T) {
	cache := newDefinitionCache()
	id := uuid.New()
	reads := 0
	read := func(context.Context, uuid.UUID) (*entities.ProcessDefinition, error) {
		reads++
		return &entities.ProcessDefinition{Key: "expense-approval"}, nil
	}

	ctx := entities.WithSystemContext(context.Background())
	for range 5 {
		def, err := cache.get(ctx, id, read)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if def.Key != "expense-approval" {
			t.Fatalf("got %q", def.Key)
		}
	}
	if reads != 1 {
		t.Fatalf("read the definition %d times, want 1", reads)
	}
}

/*
 * The safety property.
 *
 * Reads go through the repository, which scopes by tenant. Serving a cached
 * copy to a caller from another organization would hand them a row the scope
 * had refused — the exact leak the tenant work closed. The tenant is part of
 * the key so a hit across that boundary is impossible by construction, rather
 * than by every future call site remembering.
 */
func TestCacheNeverServesAcrossTenants(t *testing.T) {
	cache := newDefinitionCache()
	id := uuid.New()

	var lastScope string
	read := func(ctx context.Context, _ uuid.UUID) (*entities.ProcessDefinition, error) {
		lastScope = scopeKey(ctx)
		return &entities.ProcessDefinition{Key: lastScope}, nil
	}

	acme := entities.WithTenantContext(context.Background(), entities.TenantContext{TenantID: "acme"})
	other := entities.WithTenantContext(context.Background(), entities.TenantContext{TenantID: "other"})

	first, err := cache.get(acme, id, read)
	if err != nil {
		t.Fatalf("acme read: %v", err)
	}
	second, err := cache.get(other, id, read)
	if err != nil {
		t.Fatalf("other read: %v", err)
	}

	if first.Key == second.Key {
		t.Fatal("the second organization was served the first one's cached definition")
	}
	if second.Key != "t:other" {
		t.Fatalf("second read ran under scope %q", second.Key)
	}
}

func TestSystemAndUnscopedAreDistinctBuckets(t *testing.T) {
	// Background work spans every tenant and is marked as such. A context with
	// neither a tenant nor that marker is the fail-open case the repository
	// treats as unscoped; the two must not share cached rows.
	if scopeKey(entities.WithSystemContext(context.Background())) == scopeKey(context.Background()) {
		t.Fatal("system work and an unscoped context share a cache bucket")
	}
}

func TestAFailedReadIsNotCached(t *testing.T) {
	cache := newDefinitionCache()
	id := uuid.New()
	reads := 0
	read := func(context.Context, uuid.UUID) (*entities.ProcessDefinition, error) {
		reads++
		return nil, context.DeadlineExceeded
	}
	ctx := entities.WithSystemContext(context.Background())
	for range 3 {
		if _, err := cache.get(ctx, id, read); err == nil {
			t.Fatal("expected the read error to surface")
		}
	}
	if reads != 3 {
		t.Fatalf("a failed read was cached: %d reads for 3 calls", reads)
	}
}

func TestForgetDropsEverything(t *testing.T) {
	// A cache that outlives its source keeps executing a definition an
	// administrator deleted.
	cache := newDefinitionCache()
	id := uuid.New()
	reads := 0
	read := func(context.Context, uuid.UUID) (*entities.ProcessDefinition, error) {
		reads++
		return &entities.ProcessDefinition{Key: "k"}, nil
	}
	ctx := entities.WithSystemContext(context.Background())
	if _, err := cache.get(ctx, id, read); err != nil {
		t.Fatalf("first read: %v", err)
	}
	cache.forget()
	if _, err := cache.get(ctx, id, read); err != nil {
		t.Fatalf("second read: %v", err)
	}
	if reads != 2 {
		t.Fatalf("forget did not drop the entry: %d reads", reads)
	}
}
