package controllers

import (
	"net/http"

	apperrors "meerkat/errors"
	"meerkat/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCompanyTypes returns the fixed list of company types, shared by all users.
func GetCompanyTypes(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var types []models.CompanyType
	if err := db.Order("name ASC").Find(&types).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve company types").WithError(err))
		return
	}

	c.JSON(http.StatusOK, types)
}
