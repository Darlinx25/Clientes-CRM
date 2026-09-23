package models

import (
	"time"

	"gorm.io/gorm"
)

// Note struct to represent notes attached to a contact
type Note struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index" json:"-"`
	Title     string    `json:"title" validate:"max=200"`
	Content   string    `json:"content" validate:"required,min=1,max=5000"`
	Date      time.Time `json:"date" validate:"required"`
	ContactID *uint     `json:"contact_id" validate:"omitempty,gt=0"`
	Contact   Contact   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"contact,omitempty"`

	CompanyID *uint    `gorm:"index" json:"company_id"`
	Company   *Company `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"company,omitempty"`

	// AuthorName is the username of the user who created the note.
	// Not persisted (UserID is); filled at read time for display.
	AuthorName string `gorm:"-" json:"author_name"`
}
