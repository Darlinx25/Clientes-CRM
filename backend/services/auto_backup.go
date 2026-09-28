package services

import (
	"errors"
	"fmt"
	"meerkat/config"
	"meerkat/logger"
	"meerkat/models"
	"regexp"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Automatic backup schedule stored in the app_settings table. A single poller
// job is registered at server start and reads the schedule from the database on
// every tick, so admin changes apply on the next tick without ever mutating a
// running gocron scheduler (which is not thread-safe).
const (
	backupSettingEnabled   = "auto_backup_enabled"
	backupSettingWeekday   = "auto_backup_weekday"
	backupSettingTime      = "auto_backup_time"
	backupSettingTimezone  = "auto_backup_timezone"
	defaultBackupWeekday   = "friday"
	defaultBackupTime      = "18:00"
	backupPollInterval     = 30 * time.Second
)

// validWeekdays maps the accepted weekday names to true. Weekday is stored in
// English to stay locale-independent between the API and the scheduler.
var validWeekdays = map[string]bool{
	"sunday": true, "monday": true, "tuesday": true, "wednesday": true,
	"thursday": true, "friday": true, "saturday": true,
}

var backupTimeRe = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// AutoBackupConfig is the server-wide automatic backup schedule.
// Timezone is an optional IANA name (e.g. "America/Santiago"); when empty the
// scheduler falls back to the server's configured reminder timezone.
type AutoBackupConfig struct {
	Enabled  bool   `json:"enabled"`
	Weekday  string `json:"weekday"`  // "sunday".."saturday"
	Time     string `json:"time"`     // 24h "HH:MM"
	Timezone string `json:"timezone"` // IANA name, or "" for the server default
}

// SaveAutoBackupConfig validates and persists the automatic backup schedule.
func SaveAutoBackupConfig(db *gorm.DB, in AutoBackupConfig) error {
	weekday := strings.ToLower(strings.TrimSpace(in.Weekday))
	if !validWeekdays[weekday] {
		return fmt.Errorf("el día no es válido: %q (use lunes..domingo en inglés: monday..sunday)", in.Weekday)
	}
	timeVal := strings.TrimSpace(in.Time)
	if !backupTimeRe.MatchString(timeVal) {
		return fmt.Errorf("la hora no es válida: %q (use el formato HH:MM, ej. 18:00)", in.Time)
	}
	timezone := strings.TrimSpace(in.Timezone)
	if timezone != "" {
		if _, err := time.LoadLocation(timezone); err != nil {
			return fmt.Errorf("la zona horaria no es válida: %q (use un nombre IANA, ej. America/Santiago o UTC)", in.Timezone)
		}
	}

	values := map[string]string{
		backupSettingEnabled:  "0",
		backupSettingWeekday:  weekday,
		backupSettingTime:     timeVal,
		backupSettingTimezone: timezone,
	}
	if in.Enabled {
		values[backupSettingEnabled] = "1"
	}

	for key, val := range values {
		if err := setSetting(db, key, val); err != nil {
			return err
		}
	}
	return nil
}

// GetAutoBackupConfig reads the schedule from the database, falling back to a
// sane default (disabled, Friday 18:00) for keys that have never been written.
func GetAutoBackupConfig(db *gorm.DB) (AutoBackupConfig, error) {
	cfg := AutoBackupConfig{Enabled: false, Weekday: defaultBackupWeekday, Time: defaultBackupTime}

	var rows []models.AppSetting
	if err := db.Find(&rows).Error; err != nil {
		return cfg, err
	}
	for _, row := range rows {
		switch row.Key {
		case backupSettingEnabled:
			cfg.Enabled = row.Value == "1" || strings.EqualFold(row.Value, "true")
		case backupSettingWeekday:
			if validWeekdays[strings.ToLower(row.Value)] {
				cfg.Weekday = strings.ToLower(row.Value)
			}
		case backupSettingTime:
			if backupTimeRe.MatchString(row.Value) {
				cfg.Time = row.Value
			}
		case backupSettingTimezone:
			if _, err := time.LoadLocation(strings.TrimSpace(row.Value)); err == nil {
				cfg.Timezone = strings.TrimSpace(row.Value)
			}
		}
	}
	return cfg, nil
}

func setSetting(db *gorm.DB, key, value string) error {
	var s models.AppSetting
	if err := db.Where("key = ?", key).First(&s).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return db.Create(&models.AppSetting{Key: key, Value: value}).Error
	}
	if s.Value == value {
		return nil
	}
	s.Value = value
	return db.Save(&s).Error
}

// AutoBackupScheduler polls the configured schedule and creates the weekly
// backup. Register CSV-style Poll() once per server run with
// s.Every(30).Seconds(); it compares the wall clock (in the reminder timezone)
// to the stored "weekday HH:MM" and fires at most once per minute.
type AutoBackupScheduler struct {
	db      *gorm.DB
	cfg     *config.Config
	loc     *time.Location
	mu      sync.Mutex
	lastRun string // "YYYY-MM-DD HH:MM" of the last automatic backup
}

// NewAutoBackupScheduler builds a poller for the given DB, config and location.
func NewAutoBackupScheduler(db *gorm.DB, cfg *config.Config, loc *time.Location) *AutoBackupScheduler {
	return &AutoBackupScheduler{db: db, cfg: cfg, loc: loc}
}

// PollInterval is how often the auto-backup schedule is re-checked. Kept short
// (30s) so the backup fires within the configured minute without putting load
// on the DB: each tick reads a handful of rows from app_settings.
func PollInterval() time.Duration { return backupPollInterval }

// Poll checks the configured schedule against the wall clock and triggers the
// backup when the weekday and HH:MM match. Safe to call from the scheduler
// goroutine; the actual backup runs in its own goroutine.
func (a *AutoBackupScheduler) Poll() {
	schedule, err := GetAutoBackupConfig(a.db)
	if err != nil || !schedule.Enabled {
		return
	}
	if !validWeekdays[schedule.Weekday] || !backupTimeRe.MatchString(schedule.Time) {
		return
	}

	now := time.Now().In(a.scheduleLocation(schedule))
	if !matchesSchedule(schedule, now) {
		return
	}

	if !a.tryClaim(now.Format("2006-01-02 15:04")) {
		return
	}

	key := now.Format("2006-01-02 15:04")
	go a.runBackup(key)
}

// scheduleLocation resolves the timezone the schedule is evaluated in: the
// stored IANA zone when set, otherwise the server's reminder timezone.
func (a *AutoBackupScheduler) scheduleLocation(schedule AutoBackupConfig) *time.Location {
	if schedule.Timezone != "" {
		if loc, err := time.LoadLocation(schedule.Timezone); err == nil {
			return loc
		}
	}
	return a.loc
}

// tryClaim reserves the given wall-clock minute for a backup run. It reports
// false if that minute has already been claimed (previous run still happening,
// or already done within the same minute during a continuing poll window).
func (a *AutoBackupScheduler) tryClaim(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lastRun == key {
		return false
	}
	a.lastRun = key
	return true
}

// matchesSchedule reports whether a wall-clock time falls on the configured
// weekday and HH:MM minute, ignoring seconds.
func matchesSchedule(schedule AutoBackupConfig, now time.Time) bool {
	if !validWeekdays[schedule.Weekday] || !backupTimeRe.MatchString(schedule.Time) {
		return false
	}
	if strings.ToLower(now.Weekday().String()) != schedule.Weekday {
		return false
	}
	return now.Format("15:04") == schedule.Time
}

func (a *AutoBackupScheduler) runBackup(key string) {
	result, err := CreateBackup(a.cfg.DBPath, a.cfg.ProfilePhotoDir, a.cfg.BackupDir)
	if err != nil {
		a.runBackupFailed(key)
		logger.Error().Err(err).Msg("Automatic backup failed")
		return
	}
	logger.Info().Str("path", result.Path).Int64("size_bytes", result.SizeBytes).Msg("Automatic backup completed")
}

// runBackupFailed releases the minute guard after a failed backup so the next
// tick within the same minute can retry.
func (a *AutoBackupScheduler) runBackupFailed(key string) {
	a.mu.Lock()
	if a.lastRun == key {
		a.lastRun = ""
	}
	a.mu.Unlock()
}
