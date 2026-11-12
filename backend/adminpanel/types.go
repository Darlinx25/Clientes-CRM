package adminpanel

import "time"

type SchemaInfo struct {
	Current  int  `json:"current"`
	Expected uint `json:"expected"`
	Dirty    bool `json:"dirty"`
	UpToDate bool `json:"up_to_date"`
}

type Counts struct {
	Users               int64 `json:"users"`
	Contacts            int64 `json:"contacts"`
	ArchivedContacts    int64 `json:"archived_contacts"`
	DeletedContacts     int64 `json:"deleted_contacts"`
	Notes               int64 `json:"notes"`
	DeletedNotes        int64 `json:"deleted_notes"`
	Activities          int64 `json:"activities"`
	Reminders           int64 `json:"reminders"`
	ReminderCompletions int64 `json:"reminder_completions"`
	Companies           int64 `json:"companies"`
	Relationships       int64 `json:"relationships"`
}

type Job struct {
	JobName   string     `json:"job_name"`
	LastRunAt time.Time  `json:"last_run_at"`
	LockedAt  *time.Time `json:"locked_at"`
	LockedBy  string     `json:"locked_by"`
}

type BackupSummary struct {
	Count           int        `json:"count"`
	LatestFilename  string     `json:"latest_filename"`
	LatestCreatedAt *time.Time `json:"latest_created_at"`
}

type Overview struct {
	AppVersion   string        `json:"app_version"`
	Schema       SchemaInfo    `json:"schema"`
	DBPath       string        `json:"db_path"`
	DBSizeBytes  int64         `json:"db_size_bytes"`
	WALSizeBytes int64         `json:"wal_size_bytes"`
	BackupDir    string        `json:"backup_dir"`
	Counts       Counts        `json:"counts"`
	Jobs         []Job         `json:"jobs"`
	Backups      BackupSummary `json:"backups"`
}

type TableSummary struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	RowCount int64  `json:"row_count"`
}

type TableRows struct {
	Table      string           `json:"table"`
	Label      string           `json:"label"`
	Columns    []string         `json:"columns"`
	Rows       []map[string]any `json:"rows"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type Backup struct {
	Filename  string    `json:"filename"`
	Kind      string    `json:"kind"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}
