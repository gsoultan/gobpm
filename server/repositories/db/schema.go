package db

import (
	"context"
	"fmt"

	"github.com/gsoultan/storm/compile/pgddl"
	"github.com/gsoultan/storm/migrate"
	"github.com/gsoultan/storm/schema"
	pgschema "github.com/gsoultan/storm/schema/pg"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureTables creates the tables in want that the database does not have, and
// nothing else.
//
// Create, never alter. The port runs storm and GORM against one database, and
// most of these tables were made by GORM's numbered migrations — reconciling
// them from the model layer would mean two things deciding the shape of one
// table, and the one that ran last would win. So an existing table is left
// exactly as it is, including one whose columns have drifted from the model:
// that is a drift the numbered migrations exist to fix, and silently rewriting
// it here would make the migration history a fiction.
//
// What this does cover is the tables only storm knows about — the participant
// directory, the platform accounts — which no GORM migration creates because no
// GORM model describes them. Without this they simply would not exist, and
// every request that touched one would fail with "relation does not exist"
// while the health check stayed green.
//
// Returns the tables it created, so boot can say so.
func EnsureTables(ctx context.Context, pool *pgxpool.Pool, want *schema.Schema) ([]string, error) {
	if pool == nil || want == nil {
		return nil, nil
	}

	missing := make([]*schema.Table, 0, len(want.Tables))
	for _, table := range want.Tables {
		present, err := tableExists(ctx, pool, table.Name)
		if err != nil {
			return nil, err
		}
		if !present {
			missing = append(missing, table)
		}
	}
	if len(missing) == 0 {
		return nil, nil
	}

	created := make([]string, 0, len(missing))
	for _, table := range missing {
		if _, err := pool.Exec(ctx, pgddl.CreateTable(table)); err != nil {
			return created, fmt.Errorf("could not create the %s table: %w", table.Name, err)
		}
		created = append(created, table.Name)
	}

	// Foreign keys after every table, so a pair that reference each other does
	// not depend on which was created first.
	for _, table := range missing {
		for _, fk := range table.ForeignKeys {
			present, err := tableExists(ctx, pool, fk.RefTable)
			if err != nil {
				return created, err
			}
			if !present {
				// The reference points at a table nothing has created — a model
				// half-ported, most likely. The table it is on is usable; what
				// is missing is the database's guarantee that the reference
				// points at something real.
				return created, fmt.Errorf(
					"%s references %s, which does not exist", table.Name, fk.RefTable)
			}
			if _, err := pool.Exec(ctx, pgddl.AddForeignKey(table, fk)); err != nil {
				return created, fmt.Errorf("could not add %s on %s: %w", fk.Name, table.Name, err)
			}
		}
	}
	return created, nil
}

// tableExists asks PostgreSQL whether a name resolves, on the current
// search_path — which matters, because the test harnesses give each run its own
// schema and two of them must not see each other's tables.
func tableExists(ctx context.Context, pool *pgxpool.Pool, name string) (bool, error) {
	var found bool
	err := pool.QueryRow(ctx,
		`SELECT to_regclass(quote_ident(current_schema()) || '.' || quote_ident($1)) IS NOT NULL`,
		name).Scan(&found)
	if err != nil {
		return false, fmt.Errorf("could not check whether %s exists: %w", name, err)
	}
	return found, nil
}

// ReportDrift compares the live schema against the model and returns the
// statements that would reconcile them, without running any of them.
//
// EnsureTables creates what is missing and never alters what is there, which is
// right — the numbered migrations decide the shape of an existing table, and two
// things deciding it means the one that ran last wins. But "never alter" with no
// report means a table whose shape has moved on stays wrong silently, and the
// symptom arrives much later and somewhere else: a unique key that is scoped to
// the live rows in the model and to every row in the database looks identical
// until the first time somebody deletes a row and cannot recreate it.
//
// So the drift is named at boot rather than fixed. Destructive steps are marked,
// because a report that reads like a to-do list should say which items would
// lose data.
//
// Reads only the tables in want. A database with tables storm does not model —
// every GORM table, during the port — is not drift.
func ReportDrift(ctx context.Context, pool *pgxpool.Pool, want *schema.Schema) ([]string, error) {
	if pool == nil || want == nil {
		return nil, nil
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not borrow a connection to check the schema: %w", err)
	}
	defer conn.Release()

	var namespace string
	if err := conn.QueryRow(ctx, "SELECT current_schema()").Scan(&namespace); err != nil {
		return nil, fmt.Errorf("could not read the current schema: %w", err)
	}

	live, err := pgschema.Introspect(ctx, conn.Conn(), namespace)
	if err != nil {
		return nil, fmt.Errorf("could not read the live schema: %w", err)
	}

	// Narrowed to the tables the model knows about. Diffing the whole namespace
	// would propose dropping every table GORM owns, which is not drift — it is
	// the port being unfinished.
	modelled := map[string]struct{}{}
	for _, table := range want.Tables {
		modelled[table.Name] = struct{}{}
	}
	narrowed := &schema.Schema{Enums: live.Enums}
	for _, table := range live.Tables {
		if _, ok := modelled[table.Name]; ok {
			narrowed.Tables = append(narrowed.Tables, table)
		}
	}

	var drift []string
	for _, change := range migrate.Diff(narrowed, want).Changes {
		line := change.SQL
		if change.Destructive {
			line = "would lose data: " + line
		}
		drift = append(drift, line)
	}
	return drift, nil
}

// EnsureColumnDefaults gives the columns storm writes the database defaults its
// own DDL would have given them.
//
// storm omits a column from an INSERT when the model says the database supplies
// it — an identifier, a created_at — and reads the value back from RETURNING.
// GORM fills those in Go instead and creates the column with no default, so a
// table GORM made and storm writes to hands back NULL and the decoder panics on
// a zero-length timestamp.
//
// This is the reconciliation, and it is additive: it adds a default where there
// is none and never changes one that is already there. Both layers then agree
// about who supplies the value, which is the property that lets a repository
// move from one to the other without its table changing underneath it.
//
// Derived from the model rather than a hand-written list, so a table added to
// the model layer is covered without anybody remembering to add it here.
func EnsureColumnDefaults(ctx context.Context, pool *pgxpool.Pool, want *schema.Schema) ([]string, error) {
	if pool == nil || want == nil {
		return nil, nil
	}

	var applied []string
	for _, table := range want.Tables {
		present, err := tableExists(ctx, pool, table.Name)
		if err != nil {
			return applied, err
		}
		if !present {
			continue
		}
		for _, column := range table.Columns {
			if column.Default == "" {
				continue
			}
			has, err := columnHasDefault(ctx, pool, table.Name, column.Name)
			if err != nil {
				return applied, err
			}
			if has {
				continue
			}
			statement := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
				pgddl.Ident(table.Name), pgddl.Ident(column.Name), column.Default)
			if _, err := pool.Exec(ctx, statement); err != nil {
				return applied, fmt.Errorf("could not default %s.%s: %w", table.Name, column.Name, err)
			}
			applied = append(applied, table.Name+"."+column.Name)
		}
	}
	return applied, nil
}

func columnHasDefault(ctx context.Context, pool *pgxpool.Pool, table, column string) (bool, error) {
	var has bool
	err := pool.QueryRow(ctx,
		`SELECT column_default IS NOT NULL
		   FROM information_schema.columns
		  WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2`,
		table, column).Scan(&has)
	if err != nil {
		return false, fmt.Errorf("could not read the default on %s.%s: %w", table, column, err)
	}
	return has, nil
}
