package database

import (
	"errors"

	"meerkat/models"
	"meerkat/services"

	"gorm.io/gorm"
)

// contentTables are the tables whose rows carry a user_id and represent CRM
// content that the app displays as the dataset (contacts, notes, activities,
// reminders, relationships and reminder completions). Integration rows (API
// tokens, webhooks, calendar subscriptions, CardDAV links) stay per-user.
var contentTables = []string{
	"contacts",
	"notes",
	"activities",
	"reminders",
	"relationships",
	"reminder_completions",
}

// SeedAdminUser ensures the built-in "admin" account exists with the
// admin/admin credentials on first install. On later starts it only promotes an
// existing "admin" username to the admin role; it never resets the password, so
// updating a running instance does not undo the password the user already set.
func SeedAdminUser(db *gorm.DB) error {
	var admin models.User
	err := db.Where("username = ?", services.AdminUsername).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hashed, hashErr := services.HashPassword("admin")
		if hashErr != nil {
			return hashErr
		}
		admin = models.User{
			Username: services.AdminUsername,
			Email:    services.PlaceholderEmail(services.AdminUsername),
			Password: hashed,
			Language: "es",
			IsAdmin:  true,
		}
		return db.Create(&admin).Error
	}
	if err != nil {
		return err
	}

	if !admin.IsAdmin {
		return db.Model(&admin).Update("is_admin", true).Error
	}
	return nil
}

// ReassignContentToAdmin moves every content row to the shared admin account so
// all users see the same clients, notes, activities, reminders and relationships.
func ReassignContentToAdmin(db *gorm.DB) error {
	var admin models.User
	if err := db.Where("username = ?", services.AdminUsername).Select("id").First(&admin).Error; err != nil {
		return err
	}

	for _, table := range contentTables {
		if err := db.Exec("UPDATE "+table+" SET user_id = ? WHERE user_id IS NOT NULL AND user_id != ?", admin.ID, admin.ID).Error; err != nil {
			return err
		}
	}
	return nil
}
