package adminpanel

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ErrInvalidBackup = errors.New("invalid backup filename")

func ListBackups(backupDir string) ([]Backup, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) || backupDir == "" {
			return []Backup{}, nil
		}
		return nil, err
	}

	backups := make([]Backup, 0)
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		kind, ok := backupKind(entry.Name())
		if !ok {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		backups = append(backups, Backup{
			Filename:  entry.Name(),
			Kind:      kind,
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		if backups[i].CreatedAt.Equal(backups[j].CreatedAt) {
			return backups[i].Filename > backups[j].Filename
		}
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})
	return backups, nil
}

func ResolveBackup(backupDir, filename string) (string, error) {
	if backupDir == "" || filename == "" || filename != filepath.Base(filename) || strings.ContainsAny(filename, `/\`) {
		return "", ErrInvalidBackup
	}
	if _, ok := backupKind(filename); !ok {
		return "", ErrInvalidBackup
	}

	path := filepath.Clean(filepath.Join(backupDir, filename))
	rel, err := filepath.Rel(backupDir, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", ErrInvalidBackup
	}

	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", ErrInvalidBackup
	}
	return path, nil
}

func backupKind(filename string) (string, bool) {
	switch {
	case strings.HasPrefix(filename, "historial-") && strings.HasSuffix(filename, ".zip"):
		return "backup", true
	case strings.HasPrefix(filename, "pre-upgrade-") && strings.HasSuffix(filename, ".db"):
		return "pre_upgrade", true
	default:
		return "", false
	}
}
