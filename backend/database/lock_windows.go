//go:build windows

package database

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// acquireDBLock takes an exclusive lock on path for the duration of the
// migration. It prevents two processes from migrating the same database at the
// same time (for example, a double-click plus a scheduled task).
func acquireDBLock(path string) (func(), error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open database lock: %w", err)
	}
	overlapped := new(windows.Overlapped)
	flags := uint32(windows.LOCKFILE_EXCLUSIVE_LOCK | windows.LOCKFILE_FAIL_IMMEDIATELY)
	if err := windows.LockFileEx(windows.Handle(file.Fd()), flags, 0, 1, 0, overlapped); err != nil {
		_ = file.Close()
		if err == windows.ERROR_LOCK_VIOLATION {
			return nil, fmt.Errorf("another instance is already updating the database (lock %s)", path)
		}
		return nil, fmt.Errorf("lock database: %w", err)
	}
	return func() {
		_ = windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, overlapped)
		_ = file.Close()
	}, nil
}
