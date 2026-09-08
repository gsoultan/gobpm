package model_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gsoultan/metis/server/repositories/model"
	"github.com/gsoultan/storm"
	"github.com/gsoultan/storm/compile/pgddl"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The model layer's DDL applies to a real PostgreSQL.
//
// Building the schema only proves the declarations agree with each other.
// Applying it proves they agree with the database — which is where a bad
// referential action, a too-long index or a reserved word actually shows up.
func TestTheModelLayerApplies(t *testing.T) {
	dsn := os.Getenv("STORM_DSN")
	if dsn == "" {
		t.Skip("STORM_DSN unset")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	s, err := storm.Build(model.All()...)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	ddl := pgddl.Create(s)

	if _, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS metis_storm CASCADE; CREATE SCHEMA metis_storm"); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := pool.Exec(ctx, "SET search_path TO metis_storm; "+ddl); err != nil {
		t.Fatalf("the model layer's DDL was refused by PostgreSQL: %v", err)
	}

	var tables int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM information_schema.tables WHERE table_schema = 'metis_storm'").Scan(&tables); err != nil {
		t.Fatalf("count tables: %v", err)
	}
	var fks int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM information_schema.table_constraints WHERE constraint_schema = 'metis_storm' AND constraint_type = 'FOREIGN KEY'").Scan(&fks); err != nil {
		t.Fatalf("count foreign keys: %v", err)
	}
	t.Logf("applied: %d tables, %d foreign key constraints", tables, len(strings.Fields(""))+fks)
	if tables != len(model.All()) {
		t.Fatalf("declared %d tables, created %d", len(model.All()), tables)
	}
}
