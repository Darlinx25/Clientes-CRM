package adminauth

import (
	"meerkat/config"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndParseToken(t *testing.T) {
	cfg := &config.Config{JWTSecretKey: "test-secret"}

	token, err := GenerateToken(42, cfg, time.Minute)
	require.NoError(t, err)

	userID, err := ParseToken(token, cfg)
	require.NoError(t, err)
	assert.Equal(t, uint(42), userID)
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	token, err := GenerateToken(42, &config.Config{JWTSecretKey: "test-secret"}, time.Minute)
	require.NoError(t, err)

	_, err = ParseToken(token, &config.Config{JWTSecretKey: "other-secret"})
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	cfg := &config.Config{JWTSecretKey: "test-secret"}
	claims := jwt.MapClaims{
		"scope":      tokenScope,
		"token_type": TokenType,
		"user_id":    42,
		"exp":        time.Now().Add(-time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecretKey))
	require.NoError(t, err)

	_, err = ParseToken(signed, cfg)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParseTokenRejectsUnscopedToken(t *testing.T) {
	cfg := &config.Config{JWTSecretKey: "test-secret"}
	claims := jwt.MapClaims{
		"user_id": 42,
		"exp":     time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecretKey))
	require.NoError(t, err)

	_, err = ParseToken(signed, cfg)
	assert.ErrorIs(t, err, ErrInvalidToken)
}
