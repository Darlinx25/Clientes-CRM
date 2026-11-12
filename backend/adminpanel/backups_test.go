package adminpanel

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListBackupsFiltersAndSorts(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "historial-2026-01-01_000000.zip")
	newPath := filepath.Join(dir, "pre-upgrade-20260201-120000.db")
	require.NoError(t, os.WriteFile(oldPath, []byte("zip"), 0o600))
	require.NoError(t, os.WriteFile(newPath, []byte("db"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore"), 0o600))

	oldTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(oldPath, oldTime, oldTime))
	require.NoError(t, os.Chtimes(newPath, newTime, newTime))

	backups, err := ListBackups(dir)
	require.NoError(t, err)
	require.Len(t, backups, 2)
	assert.Equal(t, "pre-upgrade-20260201-120000.db", backups[0].Filename)
	assert.Equal(t, "pre_upgrade", backups[0].Kind)
	assert.Equal(t, "historial-2026-01-01_000000.zip", backups[1].Filename)
	assert.Equal(t, "backup", backups[1].Kind)
}

func TestListBackupsMissingDirectory(t *testing.T) {
	backups, err := ListBackups(filepath.Join(t.TempDir(), "missing"))
	require.NoError(t, err)
	assert.Empty(t, backups)
}

func TestResolveBackupRejectsUnsafeNames(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "historial-2026-01-01_000000.zip"), []byte("zip"), 0o600))

	for _, name := range []string{
		"",
		"../historial-2026-01-01_000000.zip",
		"nested/historial-2026-01-01_000000.zip",
		`nested\historial-2026-01-01_000000.zip`,
		"notes.txt",
		"historial-2026-01-01_000000.db",
	} {
		_, err := ResolveBackup(dir, name)
		assert.ErrorIs(t, err, ErrInvalidBackup, name)
	}
}

func TestResolveBackupRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "historial-2026-01-01_000000.zip")
	require.NoError(t, os.WriteFile(target, []byte("zip"), 0o600))

	link := filepath.Join(dir, "historial-2026-02-01_000000.zip")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	_, err := ResolveBackup(dir, filepath.Base(link))
	assert.ErrorIs(t, err, ErrInvalidBackup)
}

func TestResolveBackupReturnsRegularFile(t *testing.T) {
	dir := t.TempDir()
	name := "historial-2026-01-01_000000.zip"
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte("zip"), 0o600))

	got, err := ResolveBackup(dir, name)
	require.NoError(t, err)
	assert.Equal(t, path, got)
}
