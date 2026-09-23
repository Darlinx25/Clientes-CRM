package controllers

import (
	"errors"
	apperrors "meerkat/errors"
	"meerkat/logger"
	"meerkat/middleware"
	"meerkat/models"
	"meerkat/services"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// fillNoteAuthors resolves the username of each note's author (Note.UserID).
func fillNoteAuthors(c *gin.Context, notes []models.Note) {
	if len(notes) == 0 {
		return
	}

	seen := make(map[uint]bool)
	ids := make([]uint, 0, len(notes))
	for _, n := range notes {
		if n.UserID != 0 && !seen[n.UserID] {
			seen[n.UserID] = true
			ids = append(ids, n.UserID)
		}
	}
	if len(ids) == 0 {
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	var users []models.User
	if err := db.Select("id", "username").Where("id IN ?", ids).Find(&users).Error; err != nil {
		return
	}
	byID := make(map[uint]string, len(users))
	for _, u := range users {
		byID[u.ID] = u.Username
	}
	for i := range notes {
		notes[i].AuthorName = byID[notes[i].UserID]
	}
}

func CreateNote(c *gin.Context) {
	// Get the database instance from the context
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	// Get contact ID from the request URL
	contactID := c.Param("id")

	// Find the contact by the ID
	var contact models.Contact
	if err := db.First(&contact, contactID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Contact").WithDetails("id", contactID))
		} else {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve contact").WithError(err))
		}
		return
	}

	// Get validated input from validation middleware
	noteInput, err := middleware.GetValidated[models.NoteInput](c)
	if err != nil {
		apperrors.AbortWithError(c, err)
		return
	}

	// Create note from validated input
	note := models.Note{
		UserID:    userID,
		Title:     noteInput.Title,
		Content:   noteInput.Content,
		Date:      noteInput.Date,
		ContactID: &contact.ID,
	}
	if err := db.Create(&note).Error; err != nil {
		logger.FromContext(c).Error().Err(err).Msg("Error saving note to database")
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to save note").WithError(err))
		return
	}

	if username, ok := c.Get("username"); ok {
		if name, isStr := username.(string); isStr {
			note.AuthorName = name
		}
	}

	go services.TriggerWebhooks(db, currentConfig(c), userID, "note.created", note)
	c.JSON(http.StatusOK, gin.H{"message": "Note created successfully", "note": note})
}

// resolveNoteTarget validates that an optional contact and/or company exist
// for a note, so a note never points at a deleted target.
func resolveNoteTarget(c *gin.Context, db *gorm.DB, contactID, companyID *uint) error {
	if contactID != nil {
		var contact models.Contact
		if err := db.First(&contact, *contactID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				apperrors.AbortWithError(c, apperrors.ErrNotFound("Contact").WithDetails("id", *contactID))
			} else {
				apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve contact").WithError(err))
			}
			return err
		}
	}
	if companyID != nil {
		var company models.Company
		if err := db.First(&company, *companyID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				apperrors.AbortWithError(c, apperrors.ErrNotFound("Company").WithDetails("id", *companyID))
			} else {
				apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve company").WithError(err))
			}
			return err
		}
	}
	return nil
}

func CreateUnassignedNote(c *gin.Context) {
	// Get the database instance from the context
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	// Get validated input from validation middleware
	noteInput, err := middleware.GetValidated[models.NoteInput](c)
	if err != nil {
		apperrors.AbortWithError(c, err)
		return
	}

	if noteInput.ContactID == nil && noteInput.CompanyID == nil {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("company_id", "A note must be linked to a company or a contact"))
		return
	}

	if err := resolveNoteTarget(c, db, noteInput.ContactID, noteInput.CompanyID); err != nil {
		return
	}

	// Create note from validated input
	note := models.Note{
		UserID:    userID,
		Title:     noteInput.Title,
		Content:   noteInput.Content,
		Date:      noteInput.Date,
		ContactID: noteInput.ContactID,
		CompanyID: noteInput.CompanyID,
	}

	if err := db.Create(&note).Error; err != nil {
		logger.FromContext(c).Error().Err(err).Msg("Error saving unassigned note to database")
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to save note").WithError(err))
		return
	}

	if username, ok := c.Get("username"); ok {
		if name, isStr := username.(string); isStr {
			note.AuthorName = name
		}
	}

	go services.TriggerWebhooks(db, currentConfig(c), userID, "note.created", note)
	c.JSON(http.StatusOK, gin.H{"message": "Note created successfully", "note": note})
}

func GetNote(c *gin.Context) {
	id := c.Param("id")
	var note models.Note
	db := c.MustGet("db").(*gorm.DB)

	if err := db.First(&note, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Note").WithDetails("id", id))
		} else {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve note").WithError(err))
		}
		return
	}

	notes := []models.Note{note}
	fillNoteAuthors(c, notes)
	note.AuthorName = notes[0].AuthorName

	c.JSON(http.StatusOK, note)
}

func GetUnassignedNotes(c *gin.Context) {
	var notes []models.Note
	db := c.MustGet("db").(*gorm.DB)

	pagination := GetPaginationParams(c)
	search := strings.ToLower(strings.TrimSpace(c.Query("search")))
	fromDateStr := c.Query("fromDate")
	toDateStr := c.Query("toDate")

	baseQuery := db.Model(&models.Note{}).
		Where("notes.contact_id IS NULL")

	// Apply date filters
	if fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			baseQuery = baseQuery.Where("notes.date >= ?", fromDate)
		}
	}
	if toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			// Add one day to include the entire end date
			toDate = toDate.AddDate(0, 0, 1)
			baseQuery = baseQuery.Where("notes.date < ?", toDate)
		}
	}

	if search != "" {
		like := "%" + search + "%"
		baseQuery = baseQuery.Where("LOWER(content) LIKE ?", like)
	}

	countQuery := baseQuery.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to count notes").WithError(err))
		return
	}

	if err := baseQuery.Session(&gorm.Session{}).
		Order("notes.date DESC, notes.id DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&notes).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve unassigned notes").WithError(err))
		return
	}

	fillNoteAuthors(c, notes)

	c.JSON(http.StatusOK, gin.H{
		"notes": notes,
		"total": total,
		"page":  pagination.Page,
		"limit": pagination.Limit,
	})
}

// GetAllNotes retrieves all notes (assigned or unassigned) with the
// associated contact name, ordered by date. Used by the global timeline view.
func GetAllNotes(c *gin.Context) {
	var notes []models.Note
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	pagination := GetPaginationParams(c)
	search := strings.ToLower(strings.TrimSpace(c.Query("search")))
	fromDateStr := c.Query("fromDate")
	toDateStr := c.Query("toDate")

	baseQuery := db.Model(&models.Note{}).
		Where("notes.user_id = ?", userID)

	// Apply date filters
	if fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			baseQuery = baseQuery.Where("notes.date >= ?", fromDate)
		}
	}
	if toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			// Add one day to include the entire end date
			toDate = toDate.AddDate(0, 0, 1)
			baseQuery = baseQuery.Where("notes.date < ?", toDate)
		}
	}

	if search != "" {
		like := "%" + search + "%"
		baseQuery = baseQuery.
			Joins("LEFT JOIN contacts ON contacts.id = notes.contact_id").
			Where("(LOWER(notes.title) LIKE ? OR LOWER(notes.content) LIKE ? OR LOWER(contacts.firstname) LIKE ? OR LOWER(contacts.lastname) LIKE ?)", like, like, like, like)
	}

	countQuery := baseQuery.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to count notes").WithError(err))
		return
	}

	// Preload the related contact (soft-deletion aware, so a deleted contact
	// still returns its note with an empty contact) and the company with its
	// owning contact for display in the timeline. Both are scoped to the
	// current user to avoid leaking client data across accounts.
	if err := baseQuery.Session(&gorm.Session{}).
		Preload("Contact", func(db *gorm.DB) *gorm.DB {
			return db.Where("user_id = ?", userID)
		}).
		Preload("Company", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN contacts ON contacts.id = companies.contact_id").Where("contacts.user_id = ?", userID)
		}).
		Preload("Company.Types").
		Preload("Company.Contact", func(db *gorm.DB) *gorm.DB {
			return db.Where("user_id = ?", userID)
		}).
		Order("notes.date DESC, notes.id DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&notes).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve notes").WithError(err))
		return
	}

	fillNoteAuthors(c, notes)

	c.JSON(http.StatusOK, gin.H{
		"notes": notes,
		"total": total,
		"page":  pagination.Page,
		"limit": pagination.Limit,
	})
}

func UpdateNote(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var note models.Note

	// Retrieve the existing note from the database
	if err := db.First(&note, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Note").WithDetails("id", id))
		} else {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve note").WithError(err))
		}
		return
	}

	// Get validated input from validation middleware
	updatedNote, err := middleware.GetValidated[models.NoteInput](c)
	if err != nil {
		apperrors.AbortWithError(c, err)
		return
	}

	// Updateable fields
	note.Title = updatedNote.Title
	note.Content = updatedNote.Content
	note.Date = updatedNote.Date
	note.ContactID = updatedNote.ContactID
	note.CompanyID = updatedNote.CompanyID

	if err := resolveNoteTarget(c, db, note.ContactID, note.CompanyID); err != nil {
		return
	}

	db.Updates(&note)

	go services.TriggerWebhooks(db, currentConfig(c), userID, "note.updated", note)
	c.JSON(http.StatusOK, gin.H{"message": "Note updated successfully", "note": note})
}

func DeleteNote(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	// Check if note exists first
	var note models.Note
	if err := db.First(&note, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Note").WithDetails("id", id))
		} else {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve note").WithError(err))
		}
		return
	}

	if err := db.Delete(&note).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to delete note").WithError(err))
		return
	}

	go services.TriggerWebhooks(db, currentConfig(c), userID, "note.deleted", gin.H{"id": note.ID})
	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}

// GetNotesForContact retrieves all notes for a given contact
func GetNotesForContact(c *gin.Context) {
	// Get contact ID from the request URL
	contactID := c.Param("id")

	// Get the database instance from the context
	db := c.MustGet("db").(*gorm.DB)

	// Initialize a variable to store the contact
	var contact models.Contact

	// Fetch the contact and preload associated notes
	if err := db.Preload("Notes").First(&contact, contactID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no contact found, return a 404 error
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Contact").WithDetails("id", contactID))
		} else {
			// For any other errors, return a 500 error
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve contact").WithError(err))
		}
		return
	}

	// If successful, return the contact and its notes as JSON
	fillNoteAuthors(c, contact.Notes)
	c.JSON(http.StatusOK, gin.H{
		"notes": contact.Notes,
	})
}
