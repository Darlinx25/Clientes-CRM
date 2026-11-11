package database

import (
	"database/sql"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/fixtures/*.db
var fixtureFS embed.FS

func fixtureNames(t *testing.T) []string {
	t.Helper()
	entries, err := fixtureFS.ReadDir("testdata/fixtures")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".db") {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		t.Fatal("no fixtures found; run: go run ./cmd/mkfixture -version 23 -out database/testdata/fixtures/v000023.db")
	}
	return names
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	raw, err := fixtureFS.ReadFile(filepath.Join("testdata/fixtures", name))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readSchemaVersionAt(dbPath string) (int, bool, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return -1, false, err
	}
	defer db.Close()
	return readSchemaVersion(db)
}

func rowCount(t *testing.T, dbPath, table string) int {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists == 0 {
		return -1
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestUpgradeFixturesToLatestPreservesData(t *testing.T) {
	tables := []string{"users", "contacts", "notes", "reminders", "activities", "relationships", "reminder_completions", "companies"}

	for _, name := range fixtureNames(t) {
		name := name
		t.Run(name, func(t *testing.T) {
			dbPath := copyFixture(t, name)
			backupDir := filepath.Join(t.TempDir(), "backups")
			versionBefore, _, err := readSchemaVersionAt(dbPath)
			if err != nil {
				t.Fatal(err)
			}
			before := map[string]int{}
			for _, table := range tables {
				before[table] = rowCount(t, dbPath, table)
			}

			db, err := InitDB(dbPath, backupDir)
			if err != nil {
				t.Fatalf("upgrade failed: %v", err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			defer sqlDB.Close()

			info, err := GetSchemaInfo(db)
			if err != nil {
				t.Fatal(err)
			}
			if info.Dirty {
				t.Fatal("upgraded database is dirty")
			}
			if !info.UpToDate || info.Current != 39 {
				t.Fatalf("expected schema 39 up to date, got %+v", info)
			}

			for _, table := range tables {
				if before[table] < 0 {
					continue
				}
				after := rowCount(t, dbPath, table)
				if after < before[table] {
					t.Errorf("table %s lost rows: %d -> %d", table, before[table], after)
				}
			}

			var firstname string
			if err := sqlDB.QueryRow("SELECT firstname FROM contacts WHERE id = 1").Scan(&firstname); err != nil {
				t.Fatalf("contact 1 not readable after upgrade: %v", err)
			}
			if firstname != "María" {
				t.Errorf("contact 1 firstname changed: %q", firstname)
			}

			backups, err := filepath.Glob(filepath.Join(backupDir, "pre-upgrade-*.db"))
			if err != nil {
				t.Fatal(err)
			}
			if versionBefore < 39 && versionBefore >= 0 && len(backups) == 0 {
				t.Errorf("expected a pre-upgrade snapshot for %s (from version %d)", name, versionBefore)
			}
			if versionBefore == 39 && len(backups) != 0 {
				t.Errorf("latest fixture should not create a snapshot, got %v", backups)
			}
		})
	}
}

func TestDirtyDatabaseIsNeverForced(t *testing.T) {
	dbPath := copyFixture(t, fixtureNames(t)[0])
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE schema_migrations SET dirty = 1"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	backupDir := filepath.Join(t.TempDir(), "backups")
	if _, err := InitDB(dbPath, backupDir); err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("expected dirty database to be refused, got %v", err)
	}

	backups, err := filepath.Glob(filepath.Join(backupDir, "pre-upgrade-*.db"))
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one snapshot of the dirty database, got %d", len(backups))
	}

	backup, err := sql.Open("sqlite", backups[0])
	if err != nil {
		t.Fatal(err)
	}
	defer backup.Close()
	var dirty bool
	if err := backup.QueryRow("SELECT dirty FROM schema_migrations LIMIT 1").Scan(&dirty); err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Fatal("snapshot does not preserve the dirty state")
	}
}

func TestNewDatabaseDoesNotNeedBackup(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "new.db")
	backupDir := filepath.Join(dir, "backups")

	db, err := InitDB(dbPath, backupDir)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	info, err := GetSchemaInfo(db)
	if err != nil {
		t.Fatal(err)
	}
	if !info.UpToDate || info.Current != 39 {
		t.Fatalf("new database not at latest schema: %+v", info)
	}
	backups, err := filepath.Glob(filepath.Join(backupDir, "pre-upgrade-*.db"))
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Fatalf("new database should not create a pre-upgrade snapshot, got %v", backups)
	}
}

func TestNewerSchemaIsRefused(t *testing.T) {
	dbPath := copyFixture(t, fixtureNames(t)[0])
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE schema_migrations SET version = 100, dirty = 0"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := InitDB(dbPath, t.TempDir()); err == nil || !strings.Contains(err.Error(), "newer than this build") {
		t.Fatalf("expected newer schema to be refused, got %v", err)
	}
}
