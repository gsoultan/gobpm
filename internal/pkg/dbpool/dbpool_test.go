package dbpool

import (
	"testing"
	"time"
)

func TestSQLiteIsAlwaysOneConnection(t *testing.T) {
	// Even if somebody sets the knobs: SQLite is a single-writer file database
	// and a second pooled connection inside a transaction deadlocks on a lock
	// upgrade that busy_timeout does not cover.
	t.Setenv(envMaxOpenConns, "50")
	got := Resolve("sqlite")
	if got.MaxOpenConns != 1 || got.MaxIdleConns != 1 {
		t.Fatalf("sqlite pool is %+v, want one connection", got)
	}
}

func TestServerDialectsGetABoundedPoolByDefault(t *testing.T) {
	t.Setenv(envMaxOpenConns, "")
	t.Setenv(envMaxIdleConns, "")
	got := Resolve("postgres")
	if got.MaxOpenConns != defaultMaxOpenConns {
		t.Fatalf("max open %d, want %d", got.MaxOpenConns, defaultMaxOpenConns)
	}
	// Idle tracks open, so a steady workload does not reconnect between bursts.
	if got.MaxIdleConns != defaultMaxOpenConns {
		t.Fatalf("max idle %d, want it to track max open", got.MaxIdleConns)
	}
	if got.ConnMaxLifetime != defaultConnLifetime {
		t.Fatalf("lifetime %s, want %s", got.ConnMaxLifetime, defaultConnLifetime)
	}
}

func TestKnobsAreRead(t *testing.T) {
	t.Setenv(envMaxOpenConns, "40")
	t.Setenv(envMaxIdleConns, "10")
	t.Setenv(envConnLifetime, "1h")
	t.Setenv(envConnIdleTime, "90s")
	got := Resolve("postgres")
	if got.MaxOpenConns != 40 || got.MaxIdleConns != 10 {
		t.Fatalf("counts are %+v", got)
	}
	if got.ConnMaxLifetime != time.Hour || got.ConnMaxIdleTime != 90*time.Second {
		t.Fatalf("durations are %+v", got)
	}
}

func TestIdleNeverExceedsOpen(t *testing.T) {
	// database/sql would silently reduce it; making it explicit means the
	// logged numbers are the numbers in force.
	t.Setenv(envMaxOpenConns, "5")
	t.Setenv(envMaxIdleConns, "50")
	if got := Resolve("postgres"); got.MaxIdleConns != 5 {
		t.Fatalf("max idle %d, want it capped at max open", got.MaxIdleConns)
	}
}

func TestNonsenseFallsBackToTheDefault(t *testing.T) {
	// A typo in a deployment must not produce an unbounded pool, which is the
	// exact failure this package exists to prevent.
	for _, value := range []string{"lots", "-1", "0", "12abc"} {
		t.Setenv(envMaxOpenConns, value)
		if got := Resolve("postgres"); got.MaxOpenConns != defaultMaxOpenConns {
			t.Fatalf("%q gave max open %d, want the default", value, got.MaxOpenConns)
		}
	}
	t.Setenv(envMaxOpenConns, "")
	t.Setenv(envConnLifetime, "half an hour")
	if got := Resolve("postgres"); got.ConnMaxLifetime != defaultConnLifetime {
		t.Fatalf("unparseable duration gave %s", got.ConnMaxLifetime)
	}
}
