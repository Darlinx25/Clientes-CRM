//go:build !windows

package database

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// acquireDBLock takes an advisory exclusive lock on path for the duration of
// the migration. It prevents two processes from migrating the same database at
// the same time (for example, a double-click plus a scheduled task).
func acquireDBLock(path string) (func(), error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open database lock: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, fmt.Errorf("another instance is already updating the database (lock %s)", path)
		}
		return nil, fmt.Errorf("lock database: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
	}, nil
}
