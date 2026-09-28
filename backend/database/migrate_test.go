package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationsApplyAndRollback runs the full migration chain on a fresh
// database, then rolls back one step at a time verifying the newest migration
// (company type rename), the auto-backup settings table and the newest data
// migration (Exonerado company type) are all reversible.
func TestMigrationsApplyAndRollback(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mig.db")

	db, err := InitDB(dbPath)
	require.NoError(t, err)
	conn, gormErr := db.DB()
	require.NoError(t, gormErr)
	defer conn.Close()

	assertRowExists(t, dbPath, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='app_settings'", 1)
	assertRowExists(t, dbPath, "SELECT COUNT(*) FROM company_types WHERE name = 'Exonerado'", 1)
	assertRowExists(t, dbPath, "SELECT COUNT(*) FROM company_types WHERE name = 'RtasFinExter'", 1)

	// Roll back 000039 (Rentas Exterior -> RtasFinExter rename).
	require.NoError(t, MigrateDown(dbPath))
	assertRowExists(t, dbPath, "SELECT COUNT(*) FROM company_types WHERE name = 'Rentas Exterior'", 1)

	// Roll back 000038 (app_settings) — the backup feature starts disabled.
	require.NoError(t, MigrateDown(dbPath))
	assertRowExists(t, dbPath, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='app_settings'", 0)

	// Roll back 000037 (Exonerado company type).
	require.NoError(t, MigrateDown(dbPath))
	assertRowExists(t, dbPath, "SELECT COUNT(*) FROM company_types WHERE name = 'Exonerado'", 0)
}

func assertRowExists(t *testing.T, dbPath, sqlQuery string, want int) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	var got int
	require.NoError(t, db.QueryRow(sqlQuery).Scan(&got), "query: "+sqlQuery)
	assert.Equal(t, want, got, fmt.Sprintf("query: %s", sqlQuery))
}
