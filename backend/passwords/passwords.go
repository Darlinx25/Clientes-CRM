package passwords

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordTooLong = errors.New("password must not exceed 72 characters")

// bcrypt only uses the first 72 bytes of a password. Passwords with no length
// limit are supported by pre-hashing anything longer than that with SHA-256
// before bcrypt; VerifyPassword checks plain bcrypt first so legacy hashes
// keep verifying.
const bcryptMaxBytes = 72

func prehash(password string) []byte {
	raw := []byte(password)
	if len(raw) > bcryptMaxBytes {
		sum := sha256.Sum256(raw)
		raw = []byte(hex.EncodeToString(sum[:]))
	}
	return raw
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(prehash(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword checks a plaintext password against a stored hash. It accepts
// both plain bcrypt hashes (legacy) and the pre-hashed form used for passwords
// longer than 72 bytes.
func VerifyPassword(hashed, password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil {
		return true
	}
	if len([]byte(password)) > bcryptMaxBytes {
		if bcrypt.CompareHashAndPassword([]byte(hashed), prehash(password)) == nil {
			return true
		}
	}
	return false
}