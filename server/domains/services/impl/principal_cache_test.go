package impl

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/entities"
)

func fixedClock(at *time.Time) func() time.Time {
	return func() time.Time { return *at }
}

func TestACachedPrincipalIsReused(t *testing.T) {
	now := time.Now()
	c := newPrincipalCache()
	c.now = fixedClock(&now)

	id := uuid.New()
	c.put(id, entities.User{Username: "alice"}, nil)

	got, ok := c.get(id)
	if !ok || got.user.Username != "alice" {
		t.Fatalf("got %+v %v", got, ok)
	}
}

/*
 * The lifetime is the security property.
 *
 * The cached value carries the moment a user's credentials last changed, which
 * is what invalidates their existing tokens. A stale entry extends a
 * compromised session by exactly this long, which is why it is seconds.
 */
func TestACachedPrincipalExpires(t *testing.T) {
	now := time.Now()
	c := newPrincipalCache()
	c.now = fixedClock(&now)

	id := uuid.New()
	c.put(id, entities.User{Username: "alice"}, nil)

	now = now.Add(defaultPrincipalCacheTTL - time.Millisecond)
	if _, ok := c.get(id); !ok {
		t.Fatal("the entry expired early")
	}

	now = now.Add(2 * time.Millisecond)
	if _, ok := c.get(id); ok {
		t.Fatal("the entry outlived its lifetime")
	}
}

// Ending sessions is the point of a password change; waiting out a lifetime is
// not an acceptable answer to somebody who believes they are compromised.
func TestForgettingIsImmediate(t *testing.T) {
	now := time.Now()
	c := newPrincipalCache()
	c.now = fixedClock(&now)

	id := uuid.New()
	c.put(id, entities.User{Username: "alice"}, nil)
	c.forget(id)

	if _, ok := c.get(id); ok {
		t.Fatal("a forgotten principal was still served")
	}
}

func TestTheCacheCanBeTurnedOff(t *testing.T) {
	// A zero lifetime disables it rather than caching forever, so an
	// installation that wants every request to re-read the account can say so.
	t.Setenv(envPrincipalCacheTTL, "0s")
	c := newPrincipalCache()

	id := uuid.New()
	c.put(id, entities.User{Username: "alice"}, nil)
	if _, ok := c.get(id); ok {
		t.Fatal("a disabled cache served an entry")
	}
}

func TestTheCacheIsBounded(t *testing.T) {
	t.Setenv(envPrincipalCacheSize, "8")
	c := newPrincipalCache()
	for range 100 {
		c.put(uuid.New(), entities.User{Username: "u"}, nil)
	}
	if c.entries.Len() > 8 {
		t.Fatalf("holding %d entries, want at most 8", c.entries.Len())
	}
}

func TestSettingsFallBackOnNonsense(t *testing.T) {
	t.Setenv(envPrincipalCacheTTL, "a while")
	if principalCacheTTL() != defaultPrincipalCacheTTL {
		t.Fatalf("ttl %s", principalCacheTTL())
	}
	t.Setenv(envPrincipalCacheSize, "many")
	if principalCacheSize() != defaultPrincipalCacheSize {
		t.Fatalf("size %d", principalCacheSize())
	}
}

// The cutoff check is now a pure function over what was read once, rather than
// a second query of its own.
func TestCredentialCutoffCheck(t *testing.T) {
	changed := time.Now()
	before := changed.Add(-time.Hour)
	after := changed.Add(time.Hour)

	cases := []struct {
		name    string
		cutoff  *time.Time
		issued  time.Time
		refused bool
	}{
		{"no change recorded", nil, before, false},
		{"issued before the change", &changed, before, true},
		{"issued after the change", &changed, after, false},
		// A token minted in the same second as the change is the one the user
		// is being handed; refusing it signs them out of the session they just
		// re-authenticated.
		{"issued in the same second", &changed, changed, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := rejectIfIssuedBeforeCredentialsChanged(c.cutoff, claimsIssuedAt(c.issued))
			if c.refused && err == nil {
				t.Fatal("expected the token to be refused")
			}
			if !c.refused && err != nil {
				t.Fatalf("token refused: %v", err)
			}
		})
	}
}

// claimsIssuedAt builds the only claim the cutoff check reads.
func claimsIssuedAt(at time.Time) jwt.MapClaims {
	return jwt.MapClaims{"iat": float64(at.Unix())}
}
