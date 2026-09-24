package controllers

import (
	"meerkat/config"
	apperrors "meerkat/errors"
	"meerkat/models"
	"meerkat/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	defaultPage  = 1
	defaultLimit = 25
	maxLimit     = 100
)

// PaginationParams represents sanitized pagination query values.
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// sessionUserID returns the ID of the currently authenticated user (from the
// JWT or API token). Identity endpoints (preferences, password, admin checks)
// must use this so each account keeps its own settings.
func sessionUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("userID")
	if !exists {
		apperrors.AbortWithError(c, apperrors.ErrUnauthorized("Authentication required"))
		return 0, false
	}

	userID, ok := value.(uint)
	if !ok {
		apperrors.AbortWithError(c, apperrors.ErrUnauthorized("Authentication required"))
		return 0, false
	}

	return userID, true
}

// currentUserID resolves the owner of the shared content dataset: the seeded
// "admin" account when it exists, otherwise the session user (single-user or
// test databases). All content queries read and write through this owner, which
// is what makes clients, notes, activities, reminders and relationships visible
// to every user.
func currentUserID(c *gin.Context) (uint, bool) {
	sessionID, ok := sessionUserID(c)
	if !ok {
		return 0, false
	}

	dbValue, exists := c.Get("db")
	if !exists {
		return sessionID, true
	}
	gdb, ok := dbValue.(*gorm.DB)
	if !ok {
		return sessionID, true
	}

	var owner models.User
	if err := gdb.Where("username = ?", services.AdminUsername).Select("id").First(&owner).Error; err == nil {
		return owner.ID, true
	}
	return sessionID, true
}

// sessionUserIDOrZero returns the session user's ID, or 0 when the request is
// unauthenticated. Used to stamp authorship on records that also carry an owner
// (shared) user_id, so the real creator is never lost on unified data.
func sessionUserIDOrZero(c *gin.Context) uint {
	if id, ok := sessionUserID(c); ok {
		return id
	}
	return 0
}

func currentConfig(c *gin.Context) config.Config {
	if val, exists := c.Get("cfg"); exists {
		if cfg, ok := val.(config.Config); ok {
			return cfg
		}
	}
	return config.Config{}
}

// GetPaginationParams extracts pagination query params using shared defaults and bounds.
func GetPaginationParams(c *gin.Context) PaginationParams {
	page := parsePositiveOrDefault(c.DefaultQuery("page", "1"), defaultPage)
	limit := parsePositiveOrDefault(c.DefaultQuery("limit", "25"), defaultLimit)
	if limit > maxLimit {
		limit = maxLimit
	}

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func parsePositiveOrDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
