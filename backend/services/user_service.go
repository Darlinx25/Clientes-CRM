package services

import (
	"errors"
	"meerkat/config"
	"meerkat/models"
	"meerkat/passwords"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrPasswordTooLong = errors.New("password must not exceed 72 characters")

// AdminUsername is the shared account that owns the CRM's content. All client
// data lives under it after startup unification, which is what makes the data
// visible to every user.
const AdminUsername = "admin"

// noEmailSuffix marks accounts registered without an email address. The email
// column is UNIQUE NOT NULL, so a deterministic placeholder derived from the
// unique username satisfies both constraints without a table rebuild.
const noEmailSuffix = "local.invalid"

// PlaceholderEmail returns a stable, unique, non-routable email for users who
// registered without providing one.
func PlaceholderEmail(username string) string {
	return strings.ToLower(username) + "@" + noEmailSuffix
}

// HasUsableEmail reports whether an account carries a real, deliverable email
// address (as opposed to an empty or placeholder one).
func HasUsableEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	return email != "" && !strings.HasSuffix(email, "@"+noEmailSuffix)
}

func HashPassword(password string) (string, error) {
	return passwords.HashPassword(password)
}

// VerifyPassword checks a plaintext password against a stored hash.
func VerifyPassword(hashed, password string) bool {
	return passwords.VerifyPassword(hashed, password)
}

func GenerateToken(user models.User, cfg *config.Config) (string, error) {
	JWTSecretKey := cfg.JWTSecretKey
	if JWTSecretKey == "" {
		return "", errors.New("JWT secret key is empty")
	}

	JWTExpiryHours := cfg.JWTExpiryHours
	if JWTExpiryHours <= 0 {
		return "", errors.New("JWT expiry hours is invalid")
	}

	// Note: is_admin is intentionally NOT included in the JWT (AdminMiddleware handles this)
	claims := jwt.MapClaims{
		"authorized": true,
		"username":   user.Username,
		"user_id":    user.ID,
		"exp":        time.Now().Add(time.Hour * time.Duration(JWTExpiryHours)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
