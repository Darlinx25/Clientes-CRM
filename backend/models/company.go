package models

import "gorm.io/gorm"

// Company represents an empresa number (número de empresa) associated with a
// client (contact) identified by a single RUT. A RUT can have several company
// numbers, and each number can have one or several fixed types (many-to-many
// with CompanyType).
type Company struct {
	gorm.Model
	ContactID     uint          `gorm:"not null;index" json:"contact_id"`
	CompanyNumber string        `gorm:"type:text COLLATE NOCASE" json:"company_number" validate:"max=100"`
	Aportacion    string        `gorm:"type:text COLLATE NOCASE" json:"aportacion" validate:"max=100"`
	Types         []CompanyType `gorm:"many2many:company_company_types;" json:"types"`
	Contact       *Contact      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"contact,omitempty"`
}

// CompanyInput is the DTO for creating/updating a company
type CompanyInput struct {
	CompanyNumber string `json:"company_number" validate:"max=100"`
	Aportacion    string `json:"aportacion" validate:"max=100"`
	TypeIDs       []uint `json:"type_ids" validate:"omitempty,max=19"`
}

// CompanyType is a fixed allowed company type label shared by all users.
type CompanyType struct {
	gorm.Model
	Name string `gorm:"type:text not null COLLATE NOCASE uniqueIndex" json:"name" validate:"required,min=1,max=100"`
}
