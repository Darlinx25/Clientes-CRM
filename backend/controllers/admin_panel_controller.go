package controllers

import (
	"errors"
	"meerkat/adminauth"
	"meerkat/adminpanel"
	apperrors "meerkat/errors"
	"meerkat/middleware"
	"meerkat/models"
	"meerkat/services"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminPanelSession(c *gin.Context) {
	cfg := currentConfig(c)
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := sessionUserID(c)
	if !ok {
		return
	}

	input, gvErr := middleware.GetValidated[models.AdminPanelReauthInput](c)
	if gvErr != nil {
		apperrors.AbortWithError(c, gvErr)
		return
	}

	var user models.User
	if err := db.Select("id, password").First(&user, userID).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrUnauthorized("Authentication required"))
		return
	}
	if !services.VerifyPassword(user.Password, input.Password) {
		apperrors.AbortWithError(c, apperrors.ErrInvalidCredentials())
		return
	}

	token, err := adminauth.GenerateToken(user.ID, &cfg, adminauth.SessionTTL)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrInternal("Failed to create admin session").WithError(err))
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		adminauth.CookieName,
		token,
		int(adminauth.SessionTTL.Seconds()),
		adminauth.CookiePath,
		cfg.CookieDomain,
		cfg.CookieSecure,
		true,
	)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"expires_in": int(adminauth.SessionTTL.Seconds())})
}

func AdminPanelSessionDelete(c *gin.Context) {
	cfg := currentConfig(c)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(adminauth.CookieName, "", -1, adminauth.CookiePath, cfg.CookieDomain, cfg.CookieSecure, true)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminPanelOverview(c *gin.Context) {
	cfg := currentConfig(c)
	db := c.MustGet("db").(*gorm.DB)

	overview, err := adminpanel.BuildOverview(db, cfg)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to build admin overview").WithError(err))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, overview)
}

func AdminPanelTables(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	tables, err := adminpanel.ListTables(db)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to list admin tables").WithError(err))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

func AdminPanelTableRows(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	pagination := GetPaginationParams(c)
	includeDeleted := c.Query("include_deleted") == "true" || c.Query("include_deleted") == "1"

	rows, err := adminpanel.QueryTable(
		db,
		c.Param("table"),
		pagination.Page,
		pagination.Limit,
		c.Query("search"),
		includeDeleted,
	)
	if err != nil {
		if errors.Is(err, adminpanel.ErrUnknownTable) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Table"))
			return
		}
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to read admin table").WithError(err))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, rows)
}

func AdminPanelBackups(c *gin.Context) {
	cfg := currentConfig(c)

	backups, err := adminpanel.ListBackups(cfg.BackupDir)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrInternal("Failed to list backups").WithError(err))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"backups":    backups,
		"backup_dir": cfg.BackupDir,
	})
}

func AdminPanelDownloadBackup(c *gin.Context) {
	cfg := currentConfig(c)

	path, err := adminpanel.ResolveBackup(cfg.BackupDir, c.Param("filename"))
	if err != nil {
		if errors.Is(err, adminpanel.ErrInvalidBackup) || os.IsNotExist(err) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Backup"))
			return
		}
		apperrors.AbortWithError(c, apperrors.ErrInternal("Failed to open backup").WithError(err))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.FileAttachment(path, filepath.Base(path))
}
