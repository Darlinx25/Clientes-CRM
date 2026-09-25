package services

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/glebarez/sqlite"
)

// BackupResult describes a created backup archive.
type BackupResult struct {
	Path      string `json:"path"`
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
}

// CreateBackup writes a consistent snapshot of the SQLite database plus the
// profile photos into a timestamped archive (historial-<timestamp>.zip) under
// backupDir. The database snapshot uses VACUUM INTO, which reads a consistent
// state of the source even while the server is serving other requests, so it
// is safe to trigger from the web UI without stopping anything.
func CreateBackup(dbPath, photosDir, backupDir string) (*BackupResult, error) {
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "historial-backup-")
	if err != nil {
		return nil, fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	snapshotPath := filepath.Join(tmpDir, "meerkat.db")
	if err := vacuumInto(dbPath, snapshotPath); err != nil {
		return nil, fmt.Errorf("snapshot database: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_150405")
	zipPath := filepath.Join(backupDir, fmt.Sprintf("historial-%s.zip", timestamp))
	if err := writeBackupZip(zipPath, snapshotPath, photosDir); err != nil {
		return nil, fmt.Errorf("create archive: %w", err)
	}

	info, err := os.Stat(zipPath)
	if err != nil {
		return nil, fmt.Errorf("stat archive: %w", err)
	}

	return &BackupResult{
		Path:      zipPath,
		Filename:  filepath.Base(zipPath),
		SizeBytes: info.Size(),
	}, nil
}

// vacuumInto runs VACUUM INTO on the source database producing destPath. A
// plain pooled connection with a busy timeout is used so the snapshot waits for
// any in-flight write instead of failing with "database is locked", and no
// journaling pragma is applied to avoid touching the source storage mode.
func vacuumInto(sourcePath, destPath string) error {
	dsn := sourcePath + "?_pragma=busy_timeout(10000)"
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	escaped := strings.ReplaceAll(destPath, "'", "''")
	_, err = sqlDB.Exec(fmt.Sprintf("VACUUM INTO '%s'", escaped))
	return err
}

// writeBackupZip archives the database snapshot as "meerkat.db" and, when the
// photos folder exists, its contents under "photos/".
func writeBackupZip(zipPath, snapshotPath, photosDir string) (err error) {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()

	zw := zip.NewWriter(file)
	defer zw.Close()

	if err := addToZip(zw, "meerkat.db", snapshotPath); err != nil {
		return err
	}

	if _, statErr := os.Stat(photosDir); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil
		}
		return statErr
	}

	return filepath.WalkDir(photosDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(photosDir, path)
		if err != nil {
			return err
		}
		return addToZip(zw, filepath.Join("photos", filepath.ToSlash(rel)), path)
	})
}

func addToZip(zw *zip.Writer, name, sourcePath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = name
	hdr.Method = zip.Deflate

	writer, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = io.Copy(writer, source)
	return err
}
