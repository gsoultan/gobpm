package impl

import (
	"time"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/envvar"
	"github.com/gsoultan/metis/internal/pkg/lru"
	"github.com/gsoultan/metis/server/domains/entities"
)

// defaultPrincipalCacheTTL is how long a resolved caller is reused.
//
// Deliberately seconds, not minutes. The cached value includes the moment a
// user's credentials last changed, which is what invalidates their existing
// tokens — so a stale entry extends a compromised session by exactly this long.
// Somebody changing their password because they believe they are compromised is
// doing it to end an attacker's access, and "in a few minutes" is not an answer.
//
// Five seconds still removes almost all of the work: a user making ten requests
// a second goes from ten identity lookups per second to one every five.
const defaultPrincipalCacheTTL = 5 * time.Second

// defaultPrincipalCacheSize bounds the map. Keyed by user id, which is not
// attacker-supplied — a caller must already hold a valid signature to reach it —
// but a bound and an eviction are cheap and the alternative grows with every
// account that has ever signed in.
const defaultPrincipalCacheSize = 10_000

const (
	envPrincipalCacheTTL  = "METIS_AUTH_CACHE_TTL"
	envPrincipalCacheSize = "METIS_AUTH_CACHE_SIZE"
)

// cachedPrincipal is a resolved caller and the credential cutoff their tokens
// are checked against.
type cachedPrincipal struct {
	user            entities.User
	tokensValidFrom *time.Time
	expires         time.Time
}

// principalCache remembers who a token belongs to, briefly.
//
// Validating a token read the account twice — once to find the credential
// cutoff and once to build the caller — and each read preloads the user's
// organizations and projects, so a single authenticated request cost about six
// queries before it reached the handler it was for. That is on every request,
// under every latency target in the roadmap.
//
// Entries are dropped explicitly whenever the account changes, so the time-based
// expiry is a backstop for the case this process cannot see: another replica
// making the change. One replica is the supported topology, so in practice the
// explicit path is the one that runs.
type principalCache struct {
	entries *lru.Cache[uuid.UUID, cachedPrincipal]
	ttl     time.Duration
	// now is injectable so the expiry can be tested without sleeping.
	now func() time.Time
}

func newPrincipalCache() *principalCache {
	return &principalCache{
		entries: lru.New[uuid.UUID, cachedPrincipal](principalCacheSize()),
		ttl:     principalCacheTTL(),
		now:     time.Now,
	}
}

func principalCacheTTL() time.Duration {
	if raw := envvar.Get(envPrincipalCacheTTL); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d >= 0 {
			return d
		}
	}
	return defaultPrincipalCacheTTL
}

func principalCacheSize() int {
	if raw := envvar.Get(envPrincipalCacheSize); raw != "" {
		n := 0
		valid := true
		for _, c := range raw {
			if c < '0' || c > '9' {
				valid = false
				break
			}
			n = n*10 + int(c-'0')
		}
		if valid && n > 0 {
			return n
		}
	}
	return defaultPrincipalCacheSize
}

// get returns a cached principal if one is held and still fresh.
//
// A zero TTL disables the cache outright rather than caching forever, so an
// installation that wants every request to re-read the account can say so.
func (c *principalCache) get(id uuid.UUID) (cachedPrincipal, bool) {
	if c.ttl <= 0 {
		return cachedPrincipal{}, false
	}
	entry, ok := c.entries.Get(id)
	if !ok {
		return cachedPrincipal{}, false
	}
	if !c.now().Before(entry.expires) {
		// Expired entries are dropped on read rather than swept: the cache is
		// bounded and evicting, so the only cost of leaving one is a map slot.
		c.entries.Remove(id)
		return cachedPrincipal{}, false
	}
	return entry, true
}

func (c *principalCache) put(id uuid.UUID, user entities.User, tokensValidFrom *time.Time) {
	if c.ttl <= 0 {
		return
	}
	c.entries.Put(id, cachedPrincipal{
		user:            user,
		tokensValidFrom: tokensValidFrom,
		expires:         c.now().Add(c.ttl),
	})
}

// forget drops an account.
//
// Called on every change to a user: a password change, an edit, a deletion, an
// organization being granted or withdrawn. Each of those alters what the cached
// value says, and two of them are security decisions — ending sessions, and
// removing somebody's access to a tenant — where waiting out a TTL is the wrong
// behaviour.
func (c *principalCache) forget(id uuid.UUID) {
	c.entries.Remove(id)
}
