package database

import (
	"database/sql"
	"embed"
	"fmt"
	"meerkat/logger"

	"github.com/glebarez/sqlite"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// pragmaDSN appends SQLite pragmas suited to a small machine shared by several
// concurrent users: WAL journal mode (concurrent readers + a writer), a busy
// timeout so concurrent writes wait instead of failing with "database is
// locked", and foreign-key integrity checks (safe because the app only does
// soft deletes). synchronous=NORMAL keeps writes fast while remaining durable
// enough under WAL.
func pragmaDSN(dbPath string) string {
	return dbPath + "?_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(10000)" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=synchronous(NORMAL)"
}

// InitDB initializes the database connection and runs migrations.
// A single pooled connection is used on purpose: SQLite serialises writers
// through one connection, which removes "database is locked" failures when
// several users read and write at the same time.
//
// If backupDir is provided and the database needs migrations (or is dirty), a
// consistent snapshot is written before anything is changed. This is the safety
// net behind the repository's upgrade contract: an instance in use can always
// be updated without losing data, and a failed migration is recoverable from
// that snapshot.
func InitDB(dbPath string, backupDir ...string) (*gorm.DB, error) {
	// Open database connection for migrations
	sqlDB, err := sql.Open("sqlite", pragmaDSN(dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	release, err := acquireDBLock(dbPath + ".lock")
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	defer release()

	latest, err := LatestMigrationVersion()
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("read migration version: %w", err)
	}
	current, dirty, err := readSchemaVersion(sqlDB)
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("read current schema version: %w", err)
	}

	needsUpgrade := current < int(latest)
	backup := ""
	if dirty || (needsUpgrade && current >= 0) {
		dir := ""
		if len(backupDir) > 0 {
			dir = backupDir[0]
		}
		backup, err = CreatePreUpgradeSnapshot(dbPath, dir, current, latest)
		if err != nil {
			_ = sqlDB.Close()
			return nil, fmt.Errorf("create pre-upgrade backup: %w", err)
		}
		if backup != "" {
			logger.Warn().Str("backup", backup).Int("from", current).Uint("to", latest).Msg("Created pre-upgrade database snapshot")
		}
	}

	if dirty {
		_ = sqlDB.Close()
		if backup != "" {
			return nil, fmt.Errorf("database schema is dirty at version %d; refusing to force it. A snapshot was saved to %s; restore it or inspect the failed migration before retrying", current, backup)
		}
		return nil, fmt.Errorf("database schema is dirty at version %d; refusing to force it. Restore the last backup before retrying", current)
	}

	if current > int(latest) {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("database schema version %d is newer than this build's version %d; update the application instead of opening it with an older binary", current, latest)
	}

	// Run migrations
	if err := RunMigrations(sqlDB); err != nil {
		_ = sqlDB.Close()
		if backup != "" {
			return nil, fmt.Errorf("failed to run migrations (pre-upgrade snapshot: %s): %w", backup, err)
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := quickCheck(sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("database integrity check after migrations failed: %w", err)
	}

	// Open GORM connection on the same pool so the pragmas apply to every query
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to connect with GORM: %w", err)
	}

	return db, nil
}

// RunMigrations runs all pending database migrations
func RunMigrations(db *sql.DB) error {
	driver, err := withInstance(db, &sqliteConfig{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Never force a dirty version automatically: doing so can skip a migration
	// that failed halfway and silently corrupt data. InitDB reports dirty
	// databases to the user with the pre-upgrade snapshot path instead.
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}
	if dirty {
		return fmt.Errorf("database schema is dirty at version %d; refusing to run migrations automatically", version)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Get final version
	version, _, err = m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get final version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		logger.Info().Msg("No migrations applied (database is empty)")
	} else {
		logger.Info().Uint("version", version).Msg("Migrations applied successfully")
	}

	return nil
}

// MigrateDown rolls back the last migration (use with caution)
func MigrateDown(dbPath string) error {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer sqlDB.Close()

	driver, err := withInstance(sqlDB, &sqliteConfig{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	logger.Info().Msg("Migration rolled back successfully")
	return nil
}
