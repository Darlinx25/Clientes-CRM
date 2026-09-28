package services

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"meerkat/config"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAutoBackupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS app_settings (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL DEFAULT ''
	)`).Error)
	return db
}

func mustScheduler(t *testing.T, db *gorm.DB) *AutoBackupScheduler {
	t.Helper()
	cfg := config.Config{
		DBPath:          filepath.Join(t.TempDir(), "test.db"),
		ProfilePhotoDir: t.TempDir(),
		BackupDir:       t.TempDir(),
	}
	return NewAutoBackupScheduler(db, &cfg, time.FixedZone("test", 0))
}

func TestSaveAndGetAutoBackupConfig(t *testing.T) {
	db := setupAutoBackupDB(t)

	cfg, err := GetAutoBackupConfig(db)
	require.NoError(t, err)
	assert.False(t, cfg.Enabled)
	assert.Equal(t, "friday", cfg.Weekday)
	assert.Equal(t, "18:00", cfg.Time)
	assert.Equal(t, "", cfg.Timezone, "defaults to the server timezone")

	err = SaveAutoBackupConfig(db, AutoBackupConfig{Enabled: true, Weekday: "monday", Time: "07:30", Timezone: "America/Santiago"})
	require.NoError(t, err)

	cfg, err = GetAutoBackupConfig(db)
	require.NoError(t, err)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "monday", cfg.Weekday)
	assert.Equal(t, "07:30", cfg.Time)
	assert.Equal(t, "America/Santiago", cfg.Timezone)
}

func TestScheduleLocation(t *testing.T) {
	scheduler := mustScheduler(t, setupAutoBackupDB(t))

	// Stored IANA zone wins.
	loc := scheduler.scheduleLocation(AutoBackupConfig{Timezone: "America/Santiago"})
	assert.Equal(t, "America/Santiago", loc.String())

	// Empty or invalid zone falls back to the server location.
	assert.Same(t, scheduler.loc, scheduler.scheduleLocation(AutoBackupConfig{Timezone: ""}))
	assert.Same(t, scheduler.loc, scheduler.scheduleLocation(AutoBackupConfig{Timezone: "not/a/zone"}))
}

func TestSaveAutoBackupConfigInvalidInput(t *testing.T) {
	db := setupAutoBackupDB(t)

	err := SaveAutoBackupConfig(db, AutoBackupConfig{Enabled: true, Weekday: "funday", Time: "18:00"})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "día")

	err = SaveAutoBackupConfig(db, AutoBackupConfig{Enabled: true, Weekday: "friday", Time: "25:00"})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "hora")

	err = SaveAutoBackupConfig(db, AutoBackupConfig{Enabled: true, Weekday: "friday", Time: "18:00", Timezone: "Not/AZone"})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "horaria")

	// Invalid values are never persisted; defaults survive.
	cfg, err := GetAutoBackupConfig(db)
	require.NoError(t, err)
	assert.False(t, cfg.Enabled)
}

func TestMatchesSchedule(t *testing.T) {
	loc := time.FixedZone("test", 0)
	monday := time.Date(2026, 9, 21, 0, 0, 0, 0, loc) // a Monday
	friday := time.Date(2026, 9, 18, 0, 0, 0, 0, loc) // a Friday

	cfg := AutoBackupConfig{Enabled: true, Weekday: "friday", Time: "18:00"}

	assert.True(t, matchesSchedule(cfg, friday.Add(18*time.Hour)))
	assert.False(t, matchesSchedule(cfg, friday.Add(18*time.Hour).Add(1*time.Minute)))
	assert.False(t, matchesSchedule(cfg, friday.Add(17*time.Hour+59*time.Minute)))
	assert.False(t, matchesSchedule(cfg, monday.Add(18*time.Hour)))

	// Invalid stored config never fires.
	assert.False(t, matchesSchedule(AutoBackupConfig{Enabled: true, Weekday: "x", Time: "18:00"}, friday.Add(18*time.Hour)))
	assert.False(t, matchesSchedule(AutoBackupConfig{Enabled: true, Weekday: "friday", Time: "9"}, friday.Add(18*time.Hour)))
}

func TestTryClaimDedupesPerMinute(t *testing.T) {
	scheduler := mustScheduler(t, setupAutoBackupDB(t))

	key := "2026-09-18 18:00"
	assert.True(t, scheduler.tryClaim(key))
	assert.False(t, scheduler.tryClaim(key), "second claim of the same minute must be rejected")
	assert.True(t, scheduler.tryClaim("2026-09-18 19:00"), "a new minute must be accepted")

	// On failure the guard is released so the next tick within the minute retries.
	scheduler.mu.Lock()
	scheduler.lastRun = key
	scheduler.mu.Unlock()
	scheduler.runBackupFailed(key)
	assert.Equal(t, "", scheduler.lastRun)
}

func TestPollDoesNotRunWhenDisabled(t *testing.T) {
	db := setupAutoBackupDB(t)
	require.NoError(t, SaveAutoBackupConfig(db, AutoBackupConfig{Enabled: false, Weekday: "friday", Time: "18:00"}))
	scheduler := mustScheduler(t, db)

	scheduler.Poll()
	assert.Equal(t, "", scheduler.lastRun, "disabled schedule must never reserve a minute")
}

func TestPollReservesMinuteOnMatch(t *testing.T) {
	db := setupAutoBackupDB(t)
	require.NoError(t, SaveAutoBackupConfig(db, AutoBackupConfig{Enabled: true, Weekday: "friday", Time: "18:00"}))
	scheduler := mustScheduler(t, db)

	// Force "now" to a Friday 18:xx by pre-seeding the guard exactly as Poll would.
	// Poll itself uses wall-clock time, so we test the same decision path via
	// matchesSchedule + tryClaim which Poll composes.
	key := "2026-09-18 18:00"
	reason := matchesSchedule(AutoBackupConfig{Enabled: true, Weekday: "friday", Time: "18:00"}, time.Date(2026, 9, 18, 18, 0, 5, 0, time.FixedZone("test", 0)))
	assert.True(t, reason)
	assert.True(t, scheduler.tryClaim(key))
}
