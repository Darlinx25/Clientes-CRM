package controllers

import (
	"net/http"

	apperrors "meerkat/errors"
	"meerkat/middleware"
	"meerkat/models"
	"meerkat/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateBackup generates a snapshot archive of the database and profile photos
// in the server's backup folder. Safe to call while the app is serving: the
// snapshot is taken via VACUUM INTO on a separate connection.
//
//	POST /api/v1/admin/backup (admin only)
func CreateBackup(c *gin.Context) {
	cfg := currentConfig(c)

	result, err := services.CreateBackup(cfg.DBPath, cfg.ProfilePhotoDir, cfg.BackupDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create backup: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAutoBackupSchedule returns the configured automatic backup schedule.
//
//	GET /api/v1/admin/auto-backup (admin only)
func GetAutoBackupSchedule(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	schedule, err := services.GetAutoBackupConfig(db)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to load automatic backup settings").WithError(err))
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// UpdateAutoBackupSchedule validates and persists the automatic backup
// schedule. The running poller applies it on its next tick, so no restart is
// needed for the change to take effect.
//
//	PUT /api/v1/admin/auto-backup (admin only)
func UpdateAutoBackupSchedule(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	input, gvErr := middleware.GetValidated[models.AutoBackupInput](c)
	if gvErr != nil {
		apperrors.AbortWithError(c, gvErr)
		return
	}

	schedule := services.AutoBackupConfig{
		Enabled: input.Enabled,
		Weekday: input.Weekday,
		Time:    input.Time,
	}
	if err := services.SaveAutoBackupConfig(db, schedule); err != nil {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("time", err.Error()))
		return
	}

	saved, err := services.GetAutoBackupConfig(db)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to reload automatic backup settings").WithError(err))
		return
	}
	c.JSON(http.StatusOK, saved)
}
