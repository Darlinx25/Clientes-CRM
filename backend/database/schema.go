package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

// SchemaInfo is a read-only snapshot of the database migration state.
type SchemaInfo struct {
	Current  int  `json:"current"`
	Expected uint `json:"expected"`
	Dirty    bool `json:"dirty"`
	UpToDate bool `json:"up_to_date"`
}

// readSchemaVersion reads the current migration version without creating the
// migrations table. A missing table or an empty table means "new database" and
// returns -1.
func readSchemaVersion(db *sql.DB) (int, bool, error) {
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&exists); err != nil {
		return -1, false, err
	}
	if exists == 0 {
		return -1, false, nil
	}

	var version int
	var dirty bool
	err := db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, false, nil
	}
	if err != nil {
		return -1, false, err
	}
	return version, dirty, nil
}

// GetSchemaInfo returns the current and expected migration versions for health
// reporting and upgrade diagnostics.
func GetSchemaInfo(db *gorm.DB) (SchemaInfo, error) {
	expected, err := LatestMigrationVersion()
	if err != nil {
		return SchemaInfo{}, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return SchemaInfo{}, err
	}
	current, dirty, err := readSchemaVersion(sqlDB)
	if err != nil {
		return SchemaInfo{}, err
	}
	return SchemaInfo{
		Current:  current,
		Expected: expected,
		Dirty:    dirty,
		UpToDate: !dirty && current >= 0 && current == int(expected),
	}, nil
}

// MigrateUpTo applies embedded migrations up to target. It is used by the
// fixture generator and tests; production code goes through InitDB so it gets
// the pre-upgrade snapshot and lock.
func MigrateUpTo(dbPath string, target uint) error {
	sqlDB, err := sql.Open("sqlite", pragmaDSN(dbPath))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer sqlDB.Close()

	driver, err := withInstance(sqlDB, &sqliteConfig{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}
	if err := m.Migrate(target); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate to version %d: %w", target, err)
	}
	return nil
}
