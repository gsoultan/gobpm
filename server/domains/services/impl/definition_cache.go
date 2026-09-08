package impl

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/envvar"
	"github.com/gsoultan/metis/internal/pkg/lru"
	"github.com/gsoultan/metis/server/domains/adapters"
	"github.com/gsoultan/metis/server/domains/entities"
)

// defaultDefinitionCacheSize bounds the number of process definitions held in
// memory. Each is a decoded graph of nodes and flows; a few hundred is a small
// number of megabytes and covers every definition a busy installation runs.
const defaultDefinitionCacheSize = 256

const envDefinitionCacheSize = "METIS_DEFINITION_CACHE_SIZE"

// definitionCache remembers decoded process definitions.
//
// A definition is immutable once deployed — a change files a new version — but
// it was read and decoded from scratch on every job, every message, every timer
// and twice per task completion. Each read is a SELECT of a row whose node and
// flow columns are large JSON documents, followed by decoding them and
// rebuilding four lookup maps. On the engine's hot path that is the single
// largest avoidable cost.
//
// Entries are keyed by tenant as well as by id. Reads go through the repository,
// which scopes by tenant, so serving a cached definition to a caller from
// another organization would hand them a row the scope had refused. Including
// the tenant in the key makes a hit impossible across that boundary, rather than
// relying on every future call site to remember.
type definitionCache struct {
	entries *lru.Cache[string, *entities.ProcessDefinition]
}

func newDefinitionCache() *definitionCache {
	return &definitionCache{entries: lru.New[string, *entities.ProcessDefinition](definitionCacheSize())}
}

func definitionCacheSize() int {
	if raw := envvar.Get(envDefinitionCacheSize); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return defaultDefinitionCacheSize
}

// scopeKey identifies the visibility the cached copy was read under.
//
// System work spans every tenant and is marked as such, so it gets its own
// bucket. A context carrying neither a tenant nor that marker is the fail-open
// case the repository already treats as unscoped; it keys separately again, so
// it can never be served a tenant-scoped entry or vice versa.
func scopeKey(ctx context.Context) string {
	if entities.IsSystemContext(ctx) {
		return "system"
	}
	if tc, ok := entities.TenantContextFrom(ctx); ok {
		return "t:" + tc.TenantID
	}
	return "unscoped"
}

// get returns a decoded definition, reading it only on a miss.
//
// The returned pointer is shared. Definitions are read-only after loading —
// the only code that assembles one is the BPMN importer, which builds a fresh
// entity — and ProcessDefinition guards its own lazy index with a sync.Once, so
// concurrent readers are safe. Never mutate what this returns, and never copy
// the struct by value: it embeds that Once.
func (c *definitionCache) get(
	ctx context.Context,
	id uuid.UUID,
	read func(context.Context, uuid.UUID) (*entities.ProcessDefinition, error),
) (*entities.ProcessDefinition, error) {
	key := scopeKey(ctx) + "|" + id.String()
	if cached, ok := c.entries.Get(key); ok {
		return cached, nil
	}
	def, err := read(ctx, id)
	if err != nil {
		return nil, err
	}
	c.entries.Put(key, def)
	return def, nil
}

// forget drops every cached copy of a definition.
//
// Called when one is deleted. The scope is part of the key, so a single
// definition may be held under more than one; clearing is the honest way to
// drop all of them without keeping a second index just for this.
func (c *definitionCache) forget() {
	c.entries.Clear()
}

// loadDefinition reads a definition through the engine's cache.
func (e *Engine) loadDefinition(ctx context.Context, id uuid.UUID) (*entities.ProcessDefinition, error) {
	return e.definitions.get(ctx, id, func(ctx context.Context, id uuid.UUID) (*entities.ProcessDefinition, error) {
		m, err := e.repo.Definition().Get(ctx, id)
		if err != nil {
			return nil, err
		}
		return adapters.DefinitionEntityAdapter{Model: m}.ToEntity(), nil
	})
}

// LoadDefinition is loadDefinition for collaborators that hold an engine rather
// than a repository — the job service reads the same definitions on the same
// hot path.
func (e *Engine) LoadDefinition(ctx context.Context, id uuid.UUID) (*entities.ProcessDefinition, error) {
	return e.loadDefinition(ctx, id)
}

// ForgetDefinitions drops the cache. Called after a definition is deleted, so a
// running engine does not keep executing something an administrator removed.
func (e *Engine) ForgetDefinitions() {
	e.definitions.forget()
}
