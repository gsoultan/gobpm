package gorms

import (
	"fmt"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/gsoultan/metis/internal/pkg/config"
	"github.com/gsoultan/metis/internal/pkg/dbpool"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// Dialector builds the driver for a connection string.
//
// One answer rather than one per call site, for the same reason Config is: the
// application's own boot, the setup wizard and every environment connection all
// have to open a database the same way, and a setting one of them does not share
// is a setting nothing checks. An unrecognised driver falls back to SQLite,
// which is what the configuration loader treats as the default engine.
func Dialector(driver, dsn string) gorm.Dialector {
	switch driver {
	case config.DriverPostgres:
		return postgres.Open(dsn)
	case config.DriverMySQL:
		return mysql.Open(dsn)
	case config.DriverSQLServer:
		return sqlserver.Open(dsn)
	default:
		return sqlite.Open(sqliteDSN(dsn))
	}
}

// Open connects to a database with this codebase's settings applied.
//
// Opening in GORM does not contact the server — the first query does — so a
// successful return here means the connection string parsed, not that the
// database answered. Callers that need to know it is reachable have to ask it
// something; see Ping.
func Open(driver, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(Dialector(driver, dsn), Config())
	if err != nil {
		return nil, fmt.Errorf("could not open the %s database: %w", driver, err)
	}
	// Pool sizing, including SQLite's single connection: it is a single-writer
	// file database, and two pooled connections inside transactions deadlock on
	// a lock upgrade that busy_timeout does not cover.
	dbpool.Apply(db)
	return db, nil
}

// Ping asks the database a question, which is the only way to learn it is
// reachable — opening a connection does not.
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("could not reach the connection pool: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("the database did not answer: %w", err)
	}
	return nil
}

// sqliteDSN gives a SQLite file the busy timeout the rest of the application
// assumes. Without it a freshly created database fails its second concurrent
// request with "database is locked".
func sqliteDSN(dsn string) string {
	if strings.Contains(dsn, "_pragma=busy_timeout") {
		return dsn
	}
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return dsn + separator + "_pragma=busy_timeout(5000)"
}
