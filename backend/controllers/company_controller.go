package controllers

import (
	"errors"
	"net/http"
	"strconv"

	apperrors "meerkat/errors"
	"meerkat/middleware"
	"meerkat/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAllCompanies returns every company across the current user's contacts,
// with its type labels and owning client. Used by the timeline note picker.
func GetAllCompanies(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var companies []models.Company
	if err := db.Model(&models.Company{}).
		Joins("JOIN contacts ON contacts.id = companies.contact_id").
		Where("contacts.user_id = ?", userID).
		Preload("Types").
		Preload("Contact", func(db *gorm.DB) *gorm.DB {
			return db.Where("user_id = ?", userID)
		}).
		Find(&companies).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve companies").WithError(err))
		return
	}

	c.JSON(http.StatusOK, companies)
}

// GetCompanies returns all companies for a contact
func GetCompanies(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	contactID, err := parseContactID(c)
	if err != nil {
		return
	}

	var companies []models.Company
	if err := db.Preload("Types").Where("contact_id = ?", contactID).Find(&companies).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve companies").WithError(err))
		return
	}

	c.JSON(http.StatusOK, companies)
}

// CreateCompany adds a new company to a contact
func CreateCompany(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	contactID, err := parseContactID(c)
	if err != nil {
		return
	}

	companyInput, gvErr := middleware.GetValidated[models.CompanyInput](c)
	if gvErr != nil {
		apperrors.AbortWithError(c, gvErr)
		return
	}

	company := models.Company{
		ContactID:     contactID,
		CompanyNumber: companyInput.CompanyNumber,
		Aportacion:    companyInput.Aportacion,
		Types:         resolveCompanyTypes(db, companyInput.TypeIDs),
	}

	if err := db.Create(&company).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to create company").WithError(err))
		return
	}

	if err := db.Preload("Types").First(&company, company.ID).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve company").WithError(err))
		return
	}

	c.JSON(http.StatusCreated, company)
}

// UpdateCompany updates an existing company
func UpdateCompany(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	contactID, err := parseContactID(c)
	if err != nil {
		return
	}

	companyID, err := strconv.ParseUint(c.Param("cid"), 10, 64)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("id", "Invalid company ID"))
		return
	}

	var company models.Company
	if err := db.Where("id = ? AND contact_id = ?", companyID, contactID).First(&company).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Company").WithDetails("id", c.Param("cid")))
		} else {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve company").WithError(err))
		}
		return
	}

	companyInput, gvErr := middleware.GetValidated[models.CompanyInput](c)
	if gvErr != nil {
		apperrors.AbortWithError(c, gvErr)
		return
	}

	company.CompanyNumber = companyInput.CompanyNumber
	company.Aportacion = companyInput.Aportacion

	if err := db.Save(&company).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to update company").WithError(err))
		return
	}

	if err := db.Model(&company).Association("Types").Replace(resolveCompanyTypes(db, companyInput.TypeIDs)); err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to update company types").WithError(err))
		return
	}

	if err := db.Preload("Types").First(&company, company.ID).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve company").WithError(err))
		return
	}

	c.JSON(http.StatusOK, company)
}

// resolveCompanyTypes loads the CompanyType rows matching the given IDs.
func resolveCompanyTypes(db *gorm.DB, typeIDs []uint) []models.CompanyType {
	if len(typeIDs) == 0 {
		return nil
	}
	types := make([]models.CompanyType, 0, len(typeIDs))
	db.Where("id IN ?", typeIDs).Find(&types)
	return types
}

// DeleteCompany removes a company from a contact
func DeleteCompany(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	contactID, err := parseContactID(c)
	if err != nil {
		return
	}

	companyID, err := strconv.ParseUint(c.Param("cid"), 10, 64)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("id", "Invalid company ID"))
		return
	}

	var company models.Company
	if err := db.Where("id = ? AND contact_id = ?", companyID, contactID).First(&company).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Company").WithDetails("id", c.Param("cid")))
		} else {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve company").WithError(err))
		}
		return
	}

	if err := db.Delete(&company).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to delete company").WithError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company deleted"})
}

func parseContactID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("id", "Invalid contact ID"))
		return 0, err
	}
	return uint(id), nil
}
