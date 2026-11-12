package adminpanel

import (
	"meerkat/config"
	"meerkat/database"
	"meerkat/models"
	"meerkat/version"
	"os"

	"gorm.io/gorm"
)

func BuildOverview(db *gorm.DB, cfg config.Config) (*Overview, error) {
	schema, err := database.GetSchemaInfo(db)
	if err != nil {
		return nil, err
	}

	counts, err := buildCounts(db)
	if err != nil {
		return nil, err
	}

	jobs, err := buildJobs(db)
	if err != nil {
		return nil, err
	}

	backups, _ := ListBackups(cfg.BackupDir)
	summary := BackupSummary{Count: len(backups)}
	if len(backups) > 0 {
		summary.LatestFilename = backups[0].Filename
		createdAt := backups[0].CreatedAt
		summary.LatestCreatedAt = &createdAt
	}

	overview := &Overview{
		AppVersion: version.Version,
		Schema: SchemaInfo{
			Current:  schema.Current,
			Expected: schema.Expected,
			Dirty:    schema.Dirty,
			UpToDate: schema.UpToDate,
		},
		DBPath:    cfg.DBPath,
		BackupDir: cfg.BackupDir,
		Counts:    counts,
		Jobs:      jobs,
		Backups:   summary,
	}

	if info, statErr := os.Stat(cfg.DBPath); statErr == nil {
		overview.DBSizeBytes = info.Size()
	}
	if info, statErr := os.Stat(cfg.DBPath + "-wal"); statErr == nil {
		overview.WALSizeBytes = info.Size()
	}

	return overview, nil
}

func buildCounts(db *gorm.DB) (Counts, error) {
	var counts Counts

	items := []struct {
		model    any
		dest     *int64
		where    string
		args     []any
		unscoped bool
	}{
		{model: &models.User{}, dest: &counts.Users},
		{model: &models.Contact{}, dest: &counts.Contacts},
		{model: &models.Contact{}, dest: &counts.ArchivedContacts, where: "archived = ?", args: []any{true}},
		{model: &models.Contact{}, dest: &counts.DeletedContacts, where: "deleted_at IS NOT NULL", unscoped: true},
		{model: &models.Note{}, dest: &counts.Notes},
		{model: &models.Note{}, dest: &counts.DeletedNotes, where: "deleted_at IS NOT NULL", unscoped: true},
		{model: &models.Activity{}, dest: &counts.Activities},
		{model: &models.Reminder{}, dest: &counts.Reminders},
		{model: &models.ReminderCompletion{}, dest: &counts.ReminderCompletions},
		{model: &models.Company{}, dest: &counts.Companies},
		{model: &models.Relationship{}, dest: &counts.Relationships},
	}

	for _, item := range items {
		if !db.Migrator().HasTable(item.model) {
			continue
		}
		query := db
		if item.unscoped {
			query = query.Unscoped()
		}
		query = query.Model(item.model)
		if item.where != "" {
			query = query.Where(item.where, item.args...)
		}
		if err := query.Count(item.dest).Error; err != nil {
			return Counts{}, err
		}
	}

	return counts, nil
}

func buildJobs(db *gorm.DB) ([]Job, error) {
	jobs := make([]Job, 0)
	if !db.Migrator().HasTable(&models.JobExecution{}) {
		return jobs, nil
	}

	var rows []models.JobExecution
	if err := db.Order("job_name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	jobs = make([]Job, 0, len(rows))
	for _, row := range rows {
		jobs = append(jobs, Job{
			JobName:   row.JobName,
			LastRunAt: row.LastRunAt,
			LockedAt:  row.LockedAt,
			LockedBy:  row.LockedBy,
		})
	}
	return jobs, nil
}
