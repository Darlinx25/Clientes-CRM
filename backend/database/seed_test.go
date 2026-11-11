package database

import (
	"testing"

	"meerkat/models"
	"meerkat/services"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSeedAdminUserDoesNotResetExistingPassword(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	if err := SeedAdminUser(db); err != nil {
		t.Fatal(err)
	}

	var admin models.User
	if err := db.Where("username = ?", services.AdminUsername).First(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if !services.VerifyPassword(admin.Password, "admin") {
		t.Fatal("first install should use the documented admin/admin credentials")
	}

	customHash, err := services.HashPassword("a-password-the-user-chose")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&admin).Update("password", customHash).Error; err != nil {
		t.Fatal(err)
	}

	if err := SeedAdminUser(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Where("username = ?", services.AdminUsername).First(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if !services.VerifyPassword(admin.Password, "a-password-the-user-chose") {
		t.Fatal("SeedAdminUser reset an existing password on restart")
	}
	if services.VerifyPassword(admin.Password, "admin") {
		t.Fatal("existing password was silently replaced with admin/admin")
	}
	if !admin.IsAdmin {
		t.Fatal("existing admin account lost the admin flag")
	}
}

func TestSeedAdminUserPromotesWithoutResettingPassword(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}

	hash, err := services.HashPassword("already-set")
	if err != nil {
		t.Fatal(err)
	}
	user := models.User{Username: services.AdminUsername, Email: "admin@example.invalid", Password: hash, IsAdmin: false}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	if err := SeedAdminUser(db); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&user, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !user.IsAdmin {
		t.Fatal("expected existing admin username to be promoted")
	}
	if !services.VerifyPassword(user.Password, "already-set") {
		t.Fatal("promotion must not reset the password")
	}
}
