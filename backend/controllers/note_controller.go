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

// fillNoteAuthors resolves the username of each note's author (AuthorID) and
// last editor (EditedByID) so the UI can display both. Notes created before
// the author columns existed have AuthorID 0 and fall back to UserID (their
// real author at the time) to avoid showing a blank name.
func fillNoteAuthors(c *gin.Context, notes []models.Note) {
	if len(notes) == 0 {
		return
	}

	seen := make(map[uint]bool)
	ids := make([]uint, 0, len(notes)*2)
	collect := func(id uint) {
		if id != 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, n := range notes {
		if n.AuthorID != 0 {
			collect(n.AuthorID)
		} else {
			collect(n.UserID)
		}
		collect(n.EditedByID)
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
		if notes[i].AuthorID != 0 {
			notes[i].AuthorName = byID[notes[i].AuthorID]
		} else {
			notes[i].AuthorName = byID[notes[i].UserID]
		}
		if notes[i].EditedByID != 0 {
			notes[i].EditedByName = byID[notes[i].EditedByID]
		}
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
		AuthorID:  sessionUserIDOrZero(c),
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
		AuthorID:  sessionUserIDOrZero(c),
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
	note.EditedByName = notes[0].EditedByName

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
		like := "%" + foldTerm(search) + "%"
		baseQuery = baseQuery.Where(accentFoldExpr("content")+" LIKE ?", like)
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
		like := "%" + foldTerm(search) + "%"
		baseQuery = baseQuery.
			Joins("LEFT JOIN contacts ON contacts.id = notes.contact_id").
			Where("("+accentFoldExpr("notes.title")+" LIKE ? OR "+accentFoldExpr("notes.content")+" LIKE ? OR "+accentFoldExpr("contacts.firstname")+" LIKE ? OR "+accentFoldExpr("contacts.lastname")+" LIKE ? OR "+accentFoldExpr("(contacts.firstname || ' ' || contacts.lastname)")+" LIKE ? OR "+accentFoldExpr("(contacts.lastname || ' ' || contacts.firstname)")+" LIKE ?)", like, like, like, like, like, like)
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
	if editorID, ok := sessionUserID(c); ok {
		note.EditedByID = editorID
	}

	if err := resolveNoteTarget(c, db, note.ContactID, note.CompanyID); err != nil {
		return
	}

	db.Updates(&note)

	notes := []models.Note{note}
	fillNoteAuthors(c, notes)
	note.AuthorName = notes[0].AuthorName
	note.EditedByName = notes[0].EditedByName

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

// GetNotesForContact retrieves notes for a given contact, paginated and
// filterable by ?search, ?fromDate and ?toDate. Pass ?deleted=true to list
// soft-deleted notes (borradas) instead.
func GetNotesForContact(c *gin.Context) {
	// Get contact ID from the request URL
	contactID := c.Param("id")

	// Get the database instance from the context
	db := c.MustGet("db").(*gorm.DB)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	// Verify the contact exists and belongs to this user
	var contact models.Contact
	if err := db.Where("user_id = ?", userID).First(&contact, contactID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no contact found, return a 404 error
			apperrors.AbortWithError(c, apperrors.ErrNotFound("Contact").WithDetails("id", contactID))
		} else {
			// For any other errors, return a 500 error
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve contact").WithError(err))
		}
		return
	}

	// Listing soft-deleted notes lets the UI offer a discreet "ver eliminadas"
	// view while the timeline itself only shows live notes.
	if c.Query("deleted") == "true" {
		var notes []models.Note
		if err := db.Unscoped().
			Where("user_id = ? AND contact_id = ? AND deleted_at IS NOT NULL", userID, contact.ID).
			Order("date DESC, id DESC").
			Find(&notes).Error; err != nil {
			apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve deleted notes").WithError(err))
			return
		}
		fillNoteAuthors(c, notes)
		c.JSON(http.StatusOK, gin.H{"notes": notes})
		return
	}

	pagination := GetPaginationParams(c)
	search := strings.ToLower(strings.TrimSpace(c.Query("search")))
	fromDateStr := c.Query("fromDate")
	toDateStr := c.Query("toDate")

	baseQuery := db.Model(&models.Note{}).
		Where("notes.user_id = ? AND notes.contact_id = ?", userID, contact.ID)

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
		like := "%" + foldTerm(search) + "%"
		baseQuery = baseQuery.Where("(" + accentFoldExpr("notes.title") + " LIKE ? OR " + accentFoldExpr("notes.content") + " LIKE ?)", like, like)
	}

	countQuery := baseQuery.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to count notes").WithError(err))
		return
	}

	var notes []models.Note
	if err := baseQuery.Session(&gorm.Session{}).
		Order("notes.date DESC, notes.id DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&notes).Error; err != nil {
		apperrors.AbortWithError(c, apperrors.ErrDatabase("Failed to retrieve notes").WithError(err))
		return
	}

	// Fill each note's author for display
	fillNoteAuthors(c, notes)
	c.JSON(http.StatusOK, gin.H{
		"notes": notes,
		"total": total,
		"page":  pagination.Page,
		"limit": pagination.Limit,
	})
}
