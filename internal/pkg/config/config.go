package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gsoultan/metis/internal/pkg/crypto"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigPath is the default location for the configuration file.
	DefaultConfigPath = "config.yaml"

	// DriverPostgres is the database this runs on. It is the only one.
	//
	// SQLite, MySQL and SQL Server were supported and are not any more. The
	// engine's storage layer is compiled rather than assembled at run time,
	// which is what lets a query's shape be checked before it runs and a
	// soft-delete predicate be a property of the schema rather than a rule every
	// call site remembers — and that compiler emits PostgreSQL. Four dialects
	// also meant four spellings of every constraint, three of which were
	// exercised by a test suite that skipped unless somebody had a server
	// running, so "the tests pass" routinely meant "SQLite passes".
	//
	// The constant remains rather than being inlined because a stored config
	// names its driver, and an installation upgrading into this needs its file
	// to still parse so it can be told what changed.
	DriverPostgres = "postgres"
)

// DatabaseConfig holds the database connection settings.
type DatabaseConfig struct {
	Driver              string `yaml:"driver" json:"driver"`
	EncryptedConnection string `yaml:"encrypted_connection" json:"encrypted_connection"`
}

// Config represents the top-level application configuration stored in config.yaml.
//
// Security note: While it is recommended to supply the encryption key via the
// ENCRYPTION_KEY environment variable, it can also be provided in config.yaml
// for development convenience. Storing the key alongside the encrypted
// connection string in production is NOT recommended.
type Config struct {
	Database      DatabaseConfig `yaml:"database" json:"database"`
	EncryptionKey string         `yaml:"encryption_key" json:"encryption_key,omitzero"`
	JWTSecret     string         `yaml:"jwt_secret" json:"jwt_secret,omitzero"`
}

// Save writes the configuration to the specified file path as YAML.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DecryptConnectionString decrypts the stored connection string using the
// provided passphrase (typically the value of the ENCRYPTION_KEY env var)
// or the key stored in the configuration itself.
func (c *Config) DecryptConnectionString(passphrase string) (string, error) {
	if c.Database.EncryptedConnection == "" {
		return "", nil
	}

	keyToUse := passphrase
	if keyToUse == "" {
		keyToUse = c.EncryptionKey
	}

	if keyToUse == "" {
		return "", fmt.Errorf("ENCRYPTION_KEY environment variable or config encryption_key is required to decrypt the database connection string")
	}

	key := crypto.DeriveKey(keyToUse)
	plaintext, err := crypto.DecryptWithKey(c.Database.EncryptedConnection, key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt connection string: %w", err)
	}

	return plaintext, nil
}

// Load reads and parses a config.yaml file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	// KnownFields: an unrecognised key is a typo, and a typo in a database
	// setting is a server quietly pointed at something nobody intended.
	// yaml.Unmarshal ignores them by default, so `databse:` read as no database
	// at all and the caller fell through to a fresh local one.
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Exists checks whether a config file exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DatabaseFields holds the individual database connection parameters.
type DatabaseFields struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	DBName     string `json:"db_name"`
	SSLEnabled bool   `json:"ssl_enabled"`
}

// DefaultPort returns the port PostgreSQL listens on unless told otherwise.
//
// It still takes a driver so that a stored config naming one this no longer
// supports gets zero rather than 5432 — a wrong port is a connection error that
// names the host, which is a better thing to read than a refusal to start.
func DefaultPort(driver string) int {
	if driver == DriverPostgres {
		return 5432
	}
	return 0
}

// BuildConnectionString assembles a PostgreSQL connection string from fields.
//
// A driver this does not recognise returns the empty string rather than a
// best-effort guess. Opening on an empty DSN fails immediately and says so,
// where a guess would connect to something — the local socket, a default
// database — and the first sign of trouble would be data in the wrong place.
func BuildConnectionString(driver string, fields DatabaseFields) string {
	if driver != DriverPostgres {
		return ""
	}
	sslMode := "disable"
	if fields.SSLEnabled {
		sslMode = "require"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		fields.Host, fields.Port, fields.Username, fields.Password, fields.DBName, sslMode,
	)
}

// SupportedDriver reports whether a stored or submitted driver is one this can
// open, so the refusal happens where somebody can read it rather than at the
// first query.
func SupportedDriver(driver string) bool { return driver == DriverPostgres }

// NewConfig creates a new Config by encrypting the provided connection string
// with the supplied passphrase.
func NewConfig(driver, connectionString, encryptionKey, jwtSecret string) (*Config, error) {
	if encryptionKey == "" {
		return nil, fmt.Errorf("encryption key must not be empty")
	}
	key := crypto.DeriveKey(encryptionKey)

	encrypted, err := crypto.EncryptWithKey(connectionString, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt connection string: %w", err)
	}

	return &Config{
		Database: DatabaseConfig{
			Driver:              driver,
			EncryptedConnection: encrypted,
		},
		EncryptionKey: encryptionKey,
		JWTSecret:     jwtSecret,
	}, nil
}

// PostgresURL converts a key/value connection string into the URL form pgx
// takes.
//
// Two formats for one database is not a choice anybody made; it is what the two
// drivers accept. Converting in one place means an environment is configured
// once and both layers reach the same database — resolving it twice is how one
// ends up on the configured database and the other somewhere else.
func PostgresURL(dsn string) string {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return dsn
	}
	fields := map[string]string{}
	for _, pair := range strings.Fields(dsn) {
		key, value, ok := strings.Cut(pair, "=")
		if ok {
			fields[key] = value
		}
	}
	sslMode := fields["sslmode"]
	if sslMode == "" {
		sslMode = "disable"
	}
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		fields["user"], fields["password"], fields["host"], fields["port"],
		fields["dbname"], sslMode)
	if searchPath := fields["search_path"]; searchPath != "" {
		url += "&search_path=" + searchPath
	}
	return url
}
