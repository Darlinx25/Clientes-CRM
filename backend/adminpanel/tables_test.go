package adminpanel

import (
	"meerkat/models"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTablesDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Contact{},
		&models.Note{},
		&models.Activity{},
		&models.Reminder{},
		&models.ReminderCompletion{},
		&models.Relationship{},
		&models.JobExecution{},
	))
	return db
}

func TestListTablesIncludesKnownTables(t *testing.T) {
	db := setupTablesDB(t)
	require.NoError(t, db.Create(&models.User{Username: "admin", Password: "hash", Email: "admin@example.com"}).Error)

	tables, err := ListTables(db)
	require.NoError(t, err)
	require.NotEmpty(t, tables)

	byName := make(map[string]TableSummary, len(tables))
	for _, table := range tables {
		byName[table.Name] = table
	}
	assert.Equal(t, int64(1), byName["users"].RowCount)
	assert.Contains(t, byName, "schema_migrations")
}

func TestQueryTableFiltersSearchAndDeletedRows(t *testing.T) {
	db := setupTablesDB(t)
	user := models.User{Username: "admin", Password: "hash", Email: "admin@example.com"}
	require.NoError(t, db.Create(&user).Error)

	active := models.Contact{UserID: user.ID, Firstname: "Ana", Lastname: "Perez"}
	deleted := models.Contact{UserID: user.ID, Firstname: "Bruno", Lastname: "Gomez"}
	require.NoError(t, db.Create(&active).Error)
	require.NoError(t, db.Create(&deleted).Error)
	require.NoError(t, db.Delete(&deleted).Error)

	result, err := QueryTable(db, "contacts", 1, 25, "Perez", false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	require.Len(t, result.Rows, 1)
	assert.Equal(t, "Ana", result.Rows[0]["firstname"])

	result, err = QueryTable(db, "contacts", 1, 25, "", true)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
}

func TestQueryTableDoesNotExposeUserSecrets(t *testing.T) {
	db := setupTablesDB(t)
	require.NoError(t, db.Create(&models.User{Username: "admin", Password: "hash", Email: "admin@example.com"}).Error)

	result, err := QueryTable(db, "users", 1, 25, "", false)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"id", "username", "email", "is_admin", "created_at", "updated_at", "deleted_at"}, result.Columns)
	require.Len(t, result.Rows, 1)
	assert.NotContains(t, result.Rows[0], "password")
}

func TestQueryTablePaginates(t *testing.T) {
	db := setupTablesDB(t)
	user := models.User{Username: "admin", Password: "hash", Email: "admin@example.com"}
	require.NoError(t, db.Create(&user).Error)
	for i := 0; i < 3; i++ {
		require.NoError(t, db.Create(&models.Contact{UserID: user.ID, Firstname: "Contact", Lastname: "Test"}).Error)
	}

	result, err := QueryTable(db, "contacts", 2, 2, "", false)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 2, result.Limit)
	assert.Equal(t, 2, result.TotalPages)
	assert.Len(t, result.Rows, 1)
}

func TestQueryTableUnknownTable(t *testing.T) {
	db := setupTablesDB(t)
	_, err := QueryTable(db, "api_tokens", 1, 25, "", false)
	assert.ErrorIs(t, err, ErrUnknownTable)
}

func TestQueryTableJobExecution(t *testing.T) {
	db := setupTablesDB(t)
	job := models.JobExecution{JobName: "daily_reminders", LastRunAt: time.Now(), LockedBy: "worker"}
	require.NoError(t, db.Create(&job).Error)

	result, err := QueryTable(db, "job_executions", 1, 25, "daily", false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "daily_reminders", result.Rows[0]["job_name"])
}
