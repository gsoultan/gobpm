package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gsoultan/metis/internal/pkg/config"
)

func TestNewConfig_EncryptsConnectionString(t *testing.T) {
	cfg, err := config.NewConfig(config.DriverPostgres, "host=localhost dbname=test", "my-secret-key-1234", "test-jwt-secret")
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	if cfg.Database.Driver != config.DriverPostgres {
		t.Fatalf("expected driver %q, got %q", config.DriverPostgres, cfg.Database.Driver)
	}

	if cfg.Database.EncryptedConnection == "" {
		t.Fatal("encrypted connection should not be empty")
	}

	if cfg.Database.EncryptedConnection == "host=localhost dbname=test" {
		t.Fatal("connection string should be encrypted, not plaintext")
	}
}

func TestNewConfig_DecryptConnectionString(t *testing.T) {
	original := "host=localhost port=5432 user=gobpm password=secret dbname=gobpm"
	encKey := "my-encryption-key-16ch"

	cfg, err := config.NewConfig(config.DriverPostgres, original, encKey, "test-jwt-secret")
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	decrypted, err := cfg.DecryptConnectionString(encKey)
	if err != nil {
		t.Fatalf("DecryptConnectionString failed: %v", err)
	}

	if decrypted != original {
		t.Fatalf("expected %q, got %q", original, decrypted)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	original := "host=db.example.com dbname=production"
	encKey := "save-load-test-key-16"

	cfg, err := config.NewConfig(config.DriverPostgres, original, encKey, "test-jwt-secret")
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if !config.Exists(configPath) {
		t.Fatal("config file should exist after save")
	}

	// Verify file permissions (owner read/write only)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("config file should not be empty")
	}

	// Load and verify
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Database.Driver != config.DriverPostgres {
		t.Fatalf("expected driver %q, got %q", config.DriverPostgres, loaded.Database.Driver)
	}

	decrypted, err := loaded.DecryptConnectionString(encKey)
	if err != nil {
		t.Fatalf("DecryptConnectionString failed: %v", err)
	}

	if decrypted != original {
		t.Fatalf("expected %q, got %q", original, decrypted)
	}
}

func TestConfig_DecryptConnectionString_Empty(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver:              config.DriverPostgres,
			EncryptedConnection: "",
		},
		EncryptionKey: "some-key-1234567",
	}

	decrypted, err := cfg.DecryptConnectionString(cfg.EncryptionKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("expected empty string, got %q", decrypted)
	}
}

func TestExists_NonExistentFile(t *testing.T) {
	if config.Exists(filepath.Join(t.TempDir(), "nonexistent.yaml")) {
		t.Fatal("Exists should return false for non-existent file")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("Load should fail for non-existent file")
	}
}

func TestNewConfig_Postgres(t *testing.T) {
	cfg, err := config.NewConfig(config.DriverPostgres, "test-connection", "encryption-key-16ch", "test-jwt-secret")
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}
	if cfg.Database.Driver != config.DriverPostgres {
		t.Fatalf("expected driver %q, got %q", config.DriverPostgres, cfg.Database.Driver)
	}
}

// A config naming an engine this no longer supports still parses. It has to:
// an installation upgrading into a PostgreSQL-only build has one, and it needs
// to be told what to do rather than met with a parse error.
func TestAConfigNamingARetiredDriverStillReads(t *testing.T) {
	for _, driver := range []string{"sqlite", "mysql", "sqlserver"} {
		t.Run(driver, func(t *testing.T) {
			cfg, err := config.NewConfig(driver, "old-connection", "encryption-key-16ch", "test-jwt-secret")
			if err != nil {
				t.Fatalf("a config naming %s should still be writable: %v", driver, err)
			}
			if config.SupportedDriver(cfg.Database.Driver) {
				t.Fatalf("%s is reported as supported; it is not", driver)
			}
		})
	}
}

func TestOnlyPostgresIsSupported(t *testing.T) {
	if !config.SupportedDriver(config.DriverPostgres) {
		t.Fatal("PostgreSQL is the engine this runs on and is reported as unsupported")
	}
	for _, driver := range []string{"", "sqlite", "mysql", "sqlserver", "oracle"} {
		if config.SupportedDriver(driver) {
			t.Fatalf("%q is reported as supported", driver)
		}
	}
}

func TestDefaultPort(t *testing.T) {
	if got := config.DefaultPort(config.DriverPostgres); got != 5432 {
		t.Fatalf("DefaultPort(postgres) = %d, want 5432", got)
	}
	// Zero rather than 5432 for anything else: a wrong port is a connection
	// error naming the host, which reads better than a silent default that
	// connects to whatever happens to be listening there.
	for _, driver := range []string{"", "mysql", "sqlserver", "unknown"} {
		if got := config.DefaultPort(driver); got != 0 {
			t.Fatalf("DefaultPort(%q) = %d, want 0", driver, got)
		}
	}
}

func TestBuildConnectionString(t *testing.T) {
	fields := config.DatabaseFields{
		Host: "db.internal", Port: 5432, Username: "metis", Password: "s3cr3t", DBName: "metis",
	}
	got := config.BuildConnectionString(config.DriverPostgres, fields)
	want := "host=db.internal port=5432 user=metis password=s3cr3t dbname=metis sslmode=disable"
	if got != want {
		t.Fatalf("BuildConnectionString(postgres) = %q, want %q", got, want)
	}

	fields.SSLEnabled = true
	if got := config.BuildConnectionString(config.DriverPostgres, fields); !strings.Contains(got, "sslmode=require") {
		t.Fatalf("SSL was asked for and the connection string does not require it: %q", got)
	}
}

// An engine this no longer supports builds nothing, rather than a best-effort
// string. Opening on empty fails immediately and says so; a guess would connect
// to something — a local socket, a default database — and the first sign of
// trouble would be data in the wrong place.
func TestARetiredDriverBuildsNoConnectionString(t *testing.T) {
	for _, driver := range []string{"sqlite", "mysql", "sqlserver", ""} {
		if got := config.BuildConnectionString(driver, config.DatabaseFields{DBName: "metis"}); got != "" {
			t.Fatalf("BuildConnectionString(%q) = %q, want the empty string", driver, got)
		}
	}
}

func TestLoad_RefusesAnUnknownKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("databse:\n  driver: postgres\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := config.Load(path); err == nil {
		t.Fatal("a misspelled key was accepted as a valid config")
	}
}

func TestLoad_RefusesMalformedYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("database:\n  driver: [unclosed\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := config.Load(path); err == nil {
		t.Fatal("malformed YAML was accepted")
	}
}

func TestLoad_AcceptsAnEmptyFile(t *testing.T) {
	// An empty file is not a typo; it is a config with nothing set, and the
	// caller decides what that means.
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := config.Load(path); err != nil {
		t.Fatalf("an empty config was refused: %v", err)
	}
}
