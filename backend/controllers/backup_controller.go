package controllers

import (
	"net/http"

	"meerkat/services"

	"github.com/gin-gonic/gin"
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
