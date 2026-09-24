package controllers

import (
	"errors"
	"net/http"
	"strings"

	apperrors "meerkat/errors"
	"meerkat/middleware"
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

// CreateCompanyType adds a new type to the shared list if it does not already
// exist, and returns it (existing or new). Used by the company dialogs when the
// user picks "Otro" and types a brand-new label.
func CreateCompanyType(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	input, gvErr := middleware.GetValidated[models.CompanyTypeInput](c)
	if gvErr != nil {
		apperrors.AbortWithError(c, gvErr)
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("name", "Type name is required"))
		return
	}

	var ct models.CompanyType
	err := db.Where("LOWER(name) = LOWER(?)", name).First(&ct).Error
	if err == nil {
		c.JSON(http.StatusOK, ct)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to check company types").WithError(err))
		return
	}

	ct = models.CompanyType{Name: name}
	if cErr := db.Create(&ct).Error; cErr != nil {
		// A concurrent create may have won the unique index race.
		if fErr := db.Where("LOWER(name) = LOWER(?)", name).First(&ct).Error; fErr != nil {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to create company type").WithError(cErr))
			return
		}
	}

	c.JSON(http.StatusCreated, ct)
}
