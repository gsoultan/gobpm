package model_test

import (
	"context"
	"os"
	"testing"

	"github.com/gsoultan/metis/server/repositories/model"
	"github.com/gsoultan/storm"
	"github.com/gsoultan/storm/compile/pgddl"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Platform accounts and workflow participants are two populations, and nothing
// in the schema joins them.
//
// They used to be one table, which meant one role list answered two unrelated
// questions — who may deploy a process model, and who may approve an invoice —
// so giving somebody an inbox gave them the platform. The separation is only
// real if no foreign key reaches across it, and that is what this asserts:
// structure, not intention.
func TestThePlatformAndWorkflowPopulationsAreSeparate(t *testing.T) {
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
	if _, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS metis_split CASCADE; CREATE SCHEMA metis_split"); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := pool.Exec(ctx, "SET search_path TO metis_split; "+pgddl.Create(s)); err != nil {
		t.Fatalf("apply: %v", err)
	}

	platform := map[string]bool{"platform_users": true, "platform_roles": true, "platform_role_assignments": true}
	workflow := map[string]bool{"workflow_users": true, "workflow_groups": true, "workflow_group_memberships": true}

	rows, err := pool.Query(ctx, `
		SELECT tc.table_name, ccu.table_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.constraint_column_usage ccu
		  ON tc.constraint_name = ccu.constraint_name AND tc.constraint_schema = ccu.constraint_schema
		WHERE tc.constraint_schema = 'metis_split' AND tc.constraint_type = 'FOREIGN KEY'`)
	if err != nil {
		t.Fatalf("read foreign keys: %v", err)
	}
	defer rows.Close()

	var crossings int
	for rows.Next() {
		var from, to string
		if err := rows.Scan(&from, &to); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if (platform[from] && workflow[to]) || (workflow[from] && platform[to]) {
			t.Errorf("a foreign key crosses the split: %s -> %s", from, to)
			crossings++
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	if crossings == 0 {
		t.Log("no foreign key connects a platform account to a workflow participant")
	}

	// The two populations are addressed by the same kind of name, and the same
	// name in each is two different people. Neither table constrains the other.
	var orgID, projectA, projectB string
	if err := pool.QueryRow(ctx,
		`INSERT INTO metis_split.organizations (name) VALUES ('Acme') RETURNING id`).Scan(&orgID); err != nil {
		t.Fatalf("seed organization: %v", err)
	}
	for _, seed := range []struct {
		name string
		into *string
	}{{"Payments", &projectA}, {"Invoicing", &projectB}} {
		if err := pool.QueryRow(ctx,
			`INSERT INTO metis_split.projects (organization_id, name) VALUES ($1, $2) RETURNING id`,
			orgID, seed.name).Scan(seed.into); err != nil {
			t.Fatalf("seed project %s: %v", seed.name, err)
		}
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_split.platform_users (username, password_hash) VALUES ('ada', 'x')`); err != nil {
		t.Fatalf("insert platform account: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_split.workflow_users (project_id, username, active) VALUES ($1, 'ada', true)`,
		projectA); err != nil {
		t.Fatalf("a participant may share a username with an unrelated platform account: %v", err)
	}
}

// A participant's name identifies them within a project, not across the
// installation.
//
// Two projects naming an approver "ada" mean two different people, and a
// directory shared across projects would make onboarding one team an act that
// touches every other team's inbox.
func TestAParticipantIsUniquePerProject(t *testing.T) {
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
	if _, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS metis_scope CASCADE; CREATE SCHEMA metis_scope"); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	if _, err := pool.Exec(ctx, "SET search_path TO metis_scope; "+pgddl.Create(s)); err != nil {
		t.Fatalf("apply: %v", err)
	}

	var orgID, projectA, projectB string
	if err := pool.QueryRow(ctx,
		`INSERT INTO metis_scope.organizations (name) VALUES ('Acme') RETURNING id`).Scan(&orgID); err != nil {
		t.Fatalf("seed organization: %v", err)
	}
	for _, seed := range []struct {
		name string
		into *string
	}{{"Payments", &projectA}, {"Invoicing", &projectB}} {
		if err := pool.QueryRow(ctx,
			`INSERT INTO metis_scope.projects (organization_id, name) VALUES ($1, $2) RETURNING id`,
			orgID, seed.name).Scan(seed.into); err != nil {
			t.Fatalf("seed project %s: %v", seed.name, err)
		}
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_scope.workflow_users (project_id, username, active) VALUES ($1, 'ada', true)`,
		projectA); err != nil {
		t.Fatalf("first participant: %v", err)
	}

	// The same name in a different project is a different person.
	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_scope.workflow_users (project_id, username, active) VALUES ($1, 'ada', true)`,
		projectB); err != nil {
		t.Fatalf("two projects may each have an 'ada': %v", err)
	}

	// The same name twice in one project is not.
	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_scope.workflow_users (project_id, username, active) VALUES ($1, 'ada', true)`,
		projectA); err == nil {
		t.Error("a project cannot have two participants called 'ada'")
	}

	// Groups are scoped the same way, and being in one grants nothing.
	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_scope.workflow_groups (project_id, name) VALUES ($1, 'approvers')`, projectA); err != nil {
		t.Fatalf("first group: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_scope.workflow_groups (project_id, name) VALUES ($1, 'approvers')`, projectB); err != nil {
		t.Fatalf("two projects may each have an 'approvers' group: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO metis_scope.workflow_groups (project_id, name) VALUES ($1, 'approvers')`, projectA); err == nil {
		t.Error("a project cannot have two groups called 'approvers'")
	}
}

// A task names its assignee, and nothing makes that name a foreign key.
//
// Deliberate: a process is authored against people who may have no account yet,
// and refusing to deploy a model until every name in it exists would make
// authoring depend on onboarding.
func TestATaskNamesItsAssigneeWithoutReferencingThem(t *testing.T) {
	s, err := storm.Build(model.All()...)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	ddl := pgddl.Create(s)
	if !contains(ddl, `"assignee" varchar(255)`) {
		t.Fatal("a task's assignee should be a name, not a reference")
	}
	if contains(ddl, `FOREIGN KEY ("assignee")`) {
		t.Fatal("a task's assignee must not be a foreign key to the participant table")
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
