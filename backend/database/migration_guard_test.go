package database

import "testing"

func TestMigrationChecksumsMatchManifest(t *testing.T) {
	if err := VerifyMigrationChecksums(); err != nil {
		t.Fatalf("migration files changed without updating migrations.lock: %v", err)
	}
}

func TestMigrationsLintClean(t *testing.T) {
	if problems := CheckMigrations(); len(problems) > 0 {
		for _, problem := range problems {
			t.Errorf("migration lint: %v", problem)
		}
	}
}

func TestLatestMigrationVersion(t *testing.T) {
	version, err := LatestMigrationVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != 39 {
		t.Fatalf("expected latest migration 39, got %d (update this test only when adding migrations)", version)
	}
}

func TestDatabaseLockIsExclusive(t *testing.T) {
	path := t.TempDir() + "/test.db.lock"
	release, err := acquireDBLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	if releaseSecond, err := acquireDBLock(path); err == nil {
		releaseSecond()
		t.Fatal("second lock acquisition should fail while the first is held")
	}
}
