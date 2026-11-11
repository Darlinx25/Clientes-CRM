package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CreatePreUpgradeSnapshot writes a consistent copy of the SQLite database
// before migrations run. Only the database is copied: migrations never touch
// profile photos, and the snapshot is meant to make a failed upgrade
// recoverable. The destination is a plain .db file (not a zip) so restoring is
// a single file copy.
func CreatePreUpgradeSnapshot(dbPath, backupDir string, from int, to uint) (string, error) {
	if backupDir == "" {
		backupDir = filepath.Dir(dbPath)
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return "", fmt.Errorf("stat database: %w", err)
	}
	if info.Size() == 0 {
		return "", nil
	}

	stamp := time.Now().Format("20060102-150405.000000000")
	name := fmt.Sprintf("pre-upgrade-v%03d-to-v%03d-%s.db", from, to, stamp)
	dest := filepath.Join(backupDir, name)

	dsn := dbPath + "?_pragma=busy_timeout(10000)"
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return "", fmt.Errorf("open database for snapshot: %w", err)
	}
	defer sqlDB.Close()

	escaped := strings.ReplaceAll(dest, "'", "''")
	if _, err := sqlDB.Exec(fmt.Sprintf("VACUUM INTO '%s'", escaped)); err != nil {
		return "", fmt.Errorf("snapshot database: %w", err)
	}
	return dest, nil
}

// quickCheck runs SQLite's integrity quick check. It is deliberately cheaper
// than a full integrity_check but still catches corruption introduced by a bad
// migration before the server starts serving requests.
func quickCheck(db *sql.DB) error {
	var result string
	if err := db.QueryRow("PRAGMA quick_check").Scan(&result); err != nil {
		return fmt.Errorf("run quick_check: %w", err)
	}
	if !strings.EqualFold(result, "ok") {
		return fmt.Errorf("database integrity check failed: %s", result)
	}
	return nil
}
