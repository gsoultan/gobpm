package userimport_test

import (
	"strings"
	"testing"

	"github.com/gsoultan/metis/server/domains/logic/userimport"
)

func parse(t *testing.T, csv string) userimport.Result {
	t.Helper()
	result, err := userimport.Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return result
}

func TestAWholeDirectoryImports(t *testing.T) {
	result := parse(t, `username,display_name,email,groups,active
ada,Ada Lovelace,ada@example.com,approvers;finance,yes
bob,Bob Vance,bob@example.com,approvers,no
`)
	if len(result.Problems) != 0 {
		t.Fatalf("expected no problems, got %v", result.Problems)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected two participants, got %d", len(result.Rows))
	}
	ada := result.Rows[0]
	if ada.Username != "ada" || ada.DisplayName != "Ada Lovelace" || ada.Email != "ada@example.com" {
		t.Errorf("ada read wrongly: %+v", ada)
	}
	if len(ada.Groups) != 2 || ada.Groups[0] != "approvers" || ada.Groups[1] != "finance" {
		t.Errorf("groups are semicolon-separated, got %v", ada.Groups)
	}
	if !ada.Active || result.Rows[1].Active {
		t.Errorf("active read wrongly: %v / %v", ada.Active, result.Rows[1].Active)
	}
}

// A bad row is reported and the rest of the file still imports.
//
// All-or-nothing on a five-hundred-row file fails at row four hundred and
// leaves whoever ran it to work out which rows landed — the answer being none,
// which they then have to take on trust.
func TestOneBadRowDoesNotLoseTheFile(t *testing.T) {
	result := parse(t, `username,email
ada,ada@example.com
,orphan@example.com
bob,not-an-address
carol,carol@example.com
`)
	if len(result.Rows) != 2 {
		t.Fatalf("the two good rows should import, got %d", len(result.Rows))
	}
	if len(result.Problems) != 2 {
		t.Fatalf("expected two problems, got %v", result.Problems)
	}
	// Reported where somebody can find them.
	if result.Problems[0].Line != 3 || !strings.Contains(result.Problems[0].Reason, "username") {
		t.Errorf("the nameless row should be reported by line, got %v", result.Problems[0])
	}
	if result.Problems[1].Line != 4 || result.Problems[1].Username != "bob" {
		t.Errorf("a bad address should name the person, got %v", result.Problems[1])
	}
}

// An address that will not parse is refused rather than dropped: a participant
// with a silently blank address is one who never gets told they have work.
func TestABadAddressIsRefusedNotBlanked(t *testing.T) {
	result := parse(t, "username,email\nada,\"not an address\"\n")
	if len(result.Rows) != 0 {
		t.Fatalf("the row should be refused, got %+v", result.Rows)
	}
	if len(result.Problems) != 1 || !strings.Contains(result.Problems[0].Reason, "email address") {
		t.Fatalf("expected an address problem, got %v", result.Problems)
	}
}

// Somebody named twice is reported once, pointing at where they first appeared.
func TestADuplicateNamesWhereItFirstAppeared(t *testing.T) {
	result := parse(t, "username\nada\nbob\nADA\n")
	if len(result.Rows) != 2 {
		t.Fatalf("the duplicate should not import, got %d rows", len(result.Rows))
	}
	if len(result.Problems) != 1 {
		t.Fatalf("expected one problem, got %v", result.Problems)
	}
	// Case-insensitively: two rows differing only in case are one person as far
	// as anybody reading the file is concerned.
	if !strings.Contains(result.Problems[0].Reason, "line 2") {
		t.Errorf("the duplicate should point at the first appearance, got %q", result.Problems[0].Reason)
	}
}

// Columns arrive in whatever order the exporting system chose, and carry
// columns that mean nothing here.
func TestColumnsAreFoundByNameNotPosition(t *testing.T) {
	result := parse(t, `employee_id,email,cost_centre,username
E-1,ada@example.com,CC-9,ada
`)
	if len(result.Rows) != 1 {
		t.Fatalf("expected one participant, got %d (%v)", len(result.Rows), result.Problems)
	}
	if result.Rows[0].Username != "ada" || result.Rows[0].Email != "ada@example.com" {
		t.Errorf("columns read by position rather than name: %+v", result.Rows[0])
	}
}

// Excel writes a byte order mark before the first header, which glues itself to
// the column name — so "username" is not "username" and the file is refused for
// having no such column. That is the first thing an Excel export would hit.
func TestAnExcelExportIsNotRefusedForItsByteOrderMark(t *testing.T) {
	result := parse(t, "\ufeffusername,email\nada,ada@example.com\n")
	if len(result.Rows) != 1 {
		t.Fatalf("a file with a byte order mark should import, got %d rows", len(result.Rows))
	}
}

// Absent means active: importing a directory is adding people who should
// receive work, and defaulting to inactive imports participants nothing can
// assign to.
func TestAnAbsentActiveColumnMeansActive(t *testing.T) {
	result := parse(t, "username\nada\n")
	if len(result.Rows) != 1 || !result.Rows[0].Active {
		t.Fatalf("a participant with no active column should be active, got %+v", result.Rows)
	}
}

func TestBlankLinesAreSkippedNotReported(t *testing.T) {
	result := parse(t, "username\nada\n\n   \nbob\n")
	if len(result.Rows) != 2 {
		t.Fatalf("expected two participants, got %d", len(result.Rows))
	}
	if len(result.Problems) != 0 {
		t.Fatalf("a blank line is not a problem to report, got %v", result.Problems)
	}
}

// A file that cannot be used at all fails as a file, not as a list of problems.
func TestAFileThatCannotBeUsedIsRefusedWhole(t *testing.T) {
	for _, tc := range []struct {
		name, csv, says string
	}{
		{"empty", "", "empty"},
		{"no username column", "email,display_name\nada@example.com,Ada\n", "username"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := userimport.Parse(strings.NewReader(tc.csv))
			if err == nil {
				t.Fatal("expected the file to be refused")
			}
			if !strings.Contains(err.Error(), tc.says) {
				t.Errorf("the refusal should say %q, got %q", tc.says, err)
			}
		})
	}
}

// The file is caller-supplied and read into memory, so it needs a bound that is
// not "whatever fits".
func TestAnOversizedFileIsRefused(t *testing.T) {
	var b strings.Builder
	b.WriteString("username\n")
	for i := range userimport.MaxRows + 10 {
		b.WriteString("user")
		b.WriteString(strings.Repeat("0", 1))
		b.WriteString(itoa(i))
		b.WriteString("\n")
	}
	_, err := userimport.Parse(strings.NewReader(b.String()))
	if err == nil {
		t.Fatal("a file past the row limit should be refused")
	}
	if !strings.Contains(err.Error(), "split it") {
		t.Errorf("the refusal should say what to do, got %q", err)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
