package models

// AppSetting is a small server-wide key/value row (e.g. the automatic backup
// schedule). Values are stored as text; each consumer parses its own keys.
type AppSetting struct {
	Key   string `gorm:"primaryKey;column:key" json:"key"`
	Value string `gorm:"not null;default ''" json:"value"`
}

// TableName returns the table name for AppSetting.
func (AppSetting) TableName() string { return "app_settings" }

// AutoBackupInput is the DTO for updating the automatic backup schedule from
// the admin settings. Weekday is a day name ("monday".."sunday") and Time is a
// 24h "HH:MM" clock; both are validated in services.SaveAutoBackupConfig so the
// user gets a friendly message instead of a schema error.
type AutoBackupInput struct {
	Enabled  bool   `json:"enabled"`
	Weekday  string `json:"weekday" validate:"required,min=3,max=9"`
	Time     string `json:"time" validate:"required,min=5,max=5"`
	Timezone string `json:"timezone"`
}
