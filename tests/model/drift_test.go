package model_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/repositories/db"
	"github.com/gsoultan/metis/server/repositories/model"
	"github.com/gsoultan/storm"
	"github.com/gsoultan/storm/compile/pgddl"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestAFreshSchemaHasNoDrift is the round trip: model → DDL → introspect →
// diff, and the diff must be empty.
//
// A drift report that never converges is one an operator learns to ignore. If
// applying the model's own DDL leaves differences, every boot logs them, and the
// warning stops meaning anything long before a real change arrives.
func TestAFreshSchemaHasNoDrift(t *testing.T) {
	dsn := os.Getenv("STORM_DSN")
	if dsn == "" {
		t.Skip("STORM_DSN unset")
	}
	ctx := context.Background()

	want, err := storm.Build(model.All()...)
	if err != nil {
		t.Fatalf("build the model: %v", err)
	}

	namespace := fmt.Sprintf("drift_%s", strings.ReplaceAll(uuid.NewString()[:8], "-", ""))
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("admin pool: %v", err)
	}
	defer admin.Close()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+namespace); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		cleanup, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			return
		}
		defer cleanup.Close()
		_, _ = cleanup.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+namespace+" CASCADE")
	})
	if _, err := admin.Exec(ctx, "SET search_path TO "+namespace+"; "+pgddl.Create(want)); err != nil {
		t.Fatalf("apply the model's own DDL: %v", err)
	}

	scoped, err := pgxpool.New(ctx, dsn+"&search_path="+namespace)
	if err != nil {
		t.Fatalf("scoped pool: %v", err)
	}
	defer scoped.Close()

	drift, err := db.ReportDrift(ctx, scoped, want)
	if err != nil {
		t.Fatalf("report drift: %v", err)
	}
	if len(drift) > 0 {
		t.Fatalf("a schema built from the model reports %d differences from it:\n%s",
			len(drift), strings.Join(drift, "\n"))
	}
}

// TestAChangedKeyIsReported proves the check reports something, so that an empty
// report means "no drift" rather than "the check is not looking".
func TestAChangedKeyIsReported(t *testing.T) {
	dsn := os.Getenv("STORM_DSN")
	if dsn == "" {
		t.Skip("STORM_DSN unset")
	}
	ctx := context.Background()

	want, err := storm.Build(model.All()...)
	if err != nil {
		t.Fatalf("build the model: %v", err)
	}

	namespace := fmt.Sprintf("drift_%s", strings.ReplaceAll(uuid.NewString()[:8], "-", ""))
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("admin pool: %v", err)
	}
	defer admin.Close()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+namespace); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		cleanup, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			return
		}
		defer cleanup.Close()
		_, _ = cleanup.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+namespace+" CASCADE")
	})
	if _, err := admin.Exec(ctx, "SET search_path TO "+namespace+"; "+pgddl.Create(want)); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// The shape an installation created before the soft-delete upgrade has:
	// a unique key over every row where the model now wants one over the live
	// ones. Nothing about it fails until somebody removes a row.
	if _, err := admin.Exec(ctx,
		"SET search_path TO "+namespace+"; DROP INDEX uq_connectors_key; "+
			"ALTER TABLE connectors ADD CONSTRAINT uq_connectors_key UNIQUE (key)"); err != nil {
		t.Fatalf("age the schema: %v", err)
	}

	scoped, err := pgxpool.New(ctx, dsn+"&search_path="+namespace)
	if err != nil {
		t.Fatalf("scoped pool: %v", err)
	}
	defer scoped.Close()

	drift, err := db.ReportDrift(ctx, scoped, want)
	if err != nil {
		t.Fatalf("report drift: %v", err)
	}
	if len(drift) == 0 {
		t.Fatal("a unique key with the wrong scope was not reported; the check would be silent about the change that introduced it")
	}
}
