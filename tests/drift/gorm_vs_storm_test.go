package drift_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gsoultan/metis/server/repositories/db"
	"github.com/gsoultan/metis/server/repositories/model"
	"github.com/gsoultan/metis/tests/testutils"
	"github.com/gsoultan/storm"
)

// TestTheTwoLayersDescribeTheSameTables is the port's safety net.
//
// GORM's migrations own the shape of every table that has not moved yet, and
// the storm model describes what the ported repositories write against. Where
// the two disagree — a column one calls `schema` and the other calls `fields` —
// nothing fails until a query runs, and then it fails as a SQL error from
// whichever repository moved most recently.
//
// So the disagreement is named here instead, for every table at once. Anything
// this reports is either a storm model that needs to match the column GORM
// created, or a migration that needs to change the column.
func TestTheTwoLayersDescribeTheSameTables(t *testing.T) {
	gormDB, conn := testutils.SetupTestStore(t)
	_ = gormDB

	want, err := storm.Build(model.All()...)
	if err != nil {
		t.Fatalf("build the model: %v", err)
	}
	drift, err := db.ReportDrift(context.Background(), conn.Main(), want)
	if err != nil {
		t.Fatalf("report drift: %v", err)
	}

	// Only the differences that break a query: a column the model names and the
	// table does not have. Those are an error on every read of that table, and
	// they read as an add-and-drop pair because a rename looks like one.
	//
	// Type differences are deliberately not failed on. GORM created text where
	// the model says varchar(64) or jsonb, and text accepts everything both of
	// them do — a storm write of raw JSON bytes into a text column round-trips.
	// They are real debt, and the shape of paying it is one migration that
	// alters every column at once, at the end of the port, rather than a lock
	// on a live table for each repository that moves.
	var breaking []string
	for _, statement := range drift {
		if strings.Contains(statement, "ADD COLUMN") || strings.Contains(statement, "DROP COLUMN") {
			breaking = append(breaking, statement)
		}
	}
	if len(breaking) > 0 {
		t.Fatalf("the storm model and the GORM schema disagree about %d columns:\n%s",
			len(breaking), strings.Join(breaking, "\n"))
	}
}
