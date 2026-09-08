package setup_test

import (
	"os"
	"strings"
	"testing"

	"github.com/gsoultan/metis/internal/pkg/config"
	"github.com/gsoultan/metis/server/domains/services/contracts"
	"github.com/gsoultan/metis/server/domains/services/impl"
)

func TestTestConnection_SQLiteSuccess(t *testing.T) {
	svc := impl.NewSetupService(nil)

	result := svc.TestConnection(t.Context(), contracts.TestConnectionRequest{
		DatabaseDriver: "sqlite",
		DBName:         ":memory:",
	})

	if !result.Success {
		t.Fatalf("expected success, got failure: %s", result.Message)
	}
	if result.Message != "Connection successful" {
		t.Errorf("expected 'Connection successful', got %q", result.Message)
	}
}

func TestTestConnection_EmptyDriver(t *testing.T) {
	svc := impl.NewSetupService(nil)

	result := svc.TestConnection(t.Context(), contracts.TestConnectionRequest{
		DatabaseDriver: "",
	})

	if result.Success {
		t.Fatal("expected failure for empty driver")
	}
	if result.Message != "Database driver is required" {
		t.Errorf("expected 'Database driver is required', got %q", result.Message)
	}
}

func TestTestConnection_InvalidHost(t *testing.T) {
	svc := impl.NewSetupService(nil)

	result := svc.TestConnection(t.Context(), contracts.TestConnectionRequest{
		DatabaseDriver: "postgres",
		DBHost:         "invalid-host-that-does-not-exist.local",
		DBPort:         5432,
		DBUsername:     "test",
		DBPassword:     "test",
		DBName:         "test",
	})

	if result.Success {
		t.Fatal("expected failure for invalid host")
	}
	if result.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTestConnection_SQLiteDefaultPath(t *testing.T) {
	svc := impl.NewSetupService(nil)

	// SQLite with empty DBName should use the default file (metis.db, or an
	// existing gobpm.db from before the rename).
	result := svc.TestConnection(t.Context(), contracts.TestConnectionRequest{
		DatabaseDriver: "sqlite",
	})

	if !result.Success {
		t.Fatalf("expected success for SQLite default path, got failure: %s", result.Message)
	}
}

// The public connection test closes as soon as the installation is configured.
//
// It takes a host and a port from an unauthenticated caller and reports exactly
// what happened to the attempt — the raw dial error, distinguishing "connection
// refused" from a timeout from an authentication failure. Before setup that is
// the wizard doing its job. After setup it is a port scanner for whatever
// network the server sits in, and it was open forever.
func TestTheConnectionTestClosesOnceConfigured(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	svc := impl.NewSetupService(nil)
	probe := contracts.TestConnectionRequest{
		DatabaseDriver: "postgres",
		DBHost:         "127.0.0.1",
		DBPort:         1,
		DBUsername:     "probe",
		DBPassword:     "probe",
		DBName:         "probe",
	}

	// Unconfigured: the wizard needs this, and it answers with what it found.
	before := svc.TestConnection(t.Context(), probe)
	if !strings.Contains(before.Message, "connect") && !strings.Contains(before.Message, "refused") {
		t.Fatalf("while unconfigured the wizard should report what happened, got %q", before.Message)
	}

	// Configured is what config.yaml existing means — the same thing Setup checks.
	if err := os.WriteFile(config.DefaultConfigPath, []byte("database:\n  driver: sqlite\n"), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	after := svc.TestConnection(t.Context(), probe)
	if after.Success {
		t.Fatal("a configured installation must not run connection probes for anonymous callers")
	}
	if strings.Contains(after.Message, "127.0.0.1") || strings.Contains(after.Message, "refused") {
		t.Fatalf("the reply must not say what it found at the address, got %q", after.Message)
	}
	if !strings.Contains(after.Message, "already configured") {
		t.Fatalf("the refusal should say why and where to go instead, got %q", after.Message)
	}
}
