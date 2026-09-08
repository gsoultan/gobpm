package userimport_test

import (
	"strings"
	"testing"

	"github.com/gsoultan/metis/server/domains/logic/userimport"
)

func fromJSON(t *testing.T, doc string) userimport.Result {
	t.Helper()
	result, err := userimport.FromJSON([]byte(doc))
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}
	return result
}

// A plain array of participants, with JSON's own types.
func TestAJSONDirectoryImports(t *testing.T) {
	result := fromJSON(t, `[
	  {"username":"ada","display_name":"Ada Lovelace","email":"ada@example.com","groups":["approvers","finance"],"active":true},
	  {"username":"bob","email":"bob@example.com","active":false}
	]`)
	if len(result.Problems) != 0 {
		t.Fatalf("expected no problems, got %v", result.Problems)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected two participants, got %d", len(result.Rows))
	}
	// A JSON list of groups means the same thing as the CSV's semicolons.
	if len(result.Rows[0].Groups) != 2 || result.Rows[0].Groups[1] != "finance" {
		t.Errorf("groups read wrongly: %v", result.Rows[0].Groups)
	}
	// A JSON boolean means the same thing as the CSV's yes/no.
	if !result.Rows[0].Active || result.Rows[1].Active {
		t.Errorf("active read wrongly: %v / %v", result.Rows[0].Active, result.Rows[1].Active)
	}
}

// Most APIs wrap their results, and refusing that would mean asking somebody to
// put a transforming proxy in front of their own directory.
func TestAWrappedListIsAccepted(t *testing.T) {
	for _, key := range []string{"data", "items", "results", "users", "participants"} {
		t.Run(key, func(t *testing.T) {
			result := fromJSON(t, `{"`+key+`":[{"username":"ada"}],"total":1}`)
			if len(result.Rows) != 1 || result.Rows[0].Username != "ada" {
				t.Fatalf("a list under %q should be read, got %+v", key, result.Rows)
			}
		})
	}
}

// The wrapper key is named rather than guessed: guessing picks a different key
// when an endpoint adds one, and the import silently starts reading something
// else.
func TestAnUnrecognisedWrapperIsRefusedRatherThanGuessed(t *testing.T) {
	_, err := userimport.FromJSON([]byte(`{"payload":[{"username":"ada"}]}`))
	if err == nil {
		t.Fatal("an unrecognised wrapper should be refused")
	}
	if !strings.Contains(err.Error(), "data") {
		t.Errorf("the refusal should say which keys are understood, got %q", err)
	}
}

// The same validation as the CSV, because it is the same code.
func TestJSONGetsTheSameValidationAsCSV(t *testing.T) {
	result := fromJSON(t, `[
	  {"username":"ada","email":"ada@example.com"},
	  {"username":"","email":"orphan@example.com"},
	  {"username":"bob","email":"not-an-address"},
	  {"username":"ADA"}
	]`)
	if len(result.Rows) != 1 {
		t.Fatalf("only ada is usable, got %d rows", len(result.Rows))
	}
	if len(result.Problems) != 3 {
		t.Fatalf("expected three problems, got %v", result.Problems)
	}
	// Reported by position, the same way a file is reported by line.
	if result.Problems[0].Line != 2 || !strings.Contains(result.Problems[0].Reason, "username") {
		t.Errorf("the nameless entry should be reported by position, got %v", result.Problems[0])
	}
	if !strings.Contains(result.Problems[2].Reason, "already named") {
		t.Errorf("a duplicate should be caught case-insensitively, got %v", result.Problems[2])
	}
}

// A number where a string was expected is read as the number, not as a float.
func TestANumericFieldIsNotReadBackAsAFloat(t *testing.T) {
	result := fromJSON(t, `[{"username":1024}]`)
	if len(result.Rows) != 1 || result.Rows[0].Username != "1024" {
		t.Fatalf("expected the username 1024, got %+v (%v)", result.Rows, result.Problems)
	}
}

// Keys nothing understands are ignored, the same as unknown CSV columns: a
// directory endpoint returns what it returns.
func TestUnknownKeysAreIgnored(t *testing.T) {
	result := fromJSON(t, `[{"username":"ada","employee_id":"E-1","cost_centre":"CC-9"}]`)
	if len(result.Rows) != 1 || result.Rows[0].Username != "ada" {
		t.Fatalf("unknown keys should be ignored, got %+v (%v)", result.Rows, result.Problems)
	}
}

func TestAResponseThatIsNotADirectoryIsRefused(t *testing.T) {
	for _, tc := range []struct{ name, doc string }{
		{"empty", ""},
		{"not json", "<html>nope</html>"},
		{"a bare string", `"ada"`},
		{"malformed array", `[{"username":]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := userimport.FromJSON([]byte(tc.doc)); err == nil {
				t.Fatal("expected a refusal")
			}
		})
	}
}
