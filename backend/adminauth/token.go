package adminauth

import (
	"errors"
	"meerkat/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	CookieName = "admin_token"
	CookiePath = "/api/v1/admin/panel"
	SessionTTL = 15 * time.Minute
	TokenType  = "admin_panel"
	tokenScope = "admin_panel"
)

var ErrInvalidToken = errors.New("invalid admin panel token")

func GenerateToken(userID uint, cfg *config.Config, ttl time.Duration) (string, error) {
	if cfg == nil || cfg.JWTSecretKey == "" {
		return "", errors.New("JWT secret key is empty")
	}
	if userID == 0 {
		return "", errors.New("invalid user ID")
	}
	if ttl <= 0 {
		ttl = SessionTTL
	}

	claims := jwt.MapClaims{
		"scope":      tokenScope,
		"token_type": TokenType,
		"user_id":    userID,
		"exp":        time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}

func ParseToken(tokenString string, cfg *config.Config) (uint, error) {
	if tokenString == "" || cfg == nil || cfg.JWTSecretKey == "" {
		return 0, ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(cfg.JWTSecretKey), nil
	})
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidToken
	}
	if scope, _ := claims["scope"].(string); scope != tokenScope {
		return 0, ErrInvalidToken
	}
	if claimType, _ := claims["token_type"].(string); claimType != TokenType {
		return 0, ErrInvalidToken
	}

	idValue, exists := claims["user_id"]
	if !exists {
		return 0, ErrInvalidToken
	}
	switch v := idValue.(type) {
	case float64:
		if v <= 0 {
			return 0, ErrInvalidToken
		}
		return uint(v), nil
	case int:
		if v <= 0 {
			return 0, ErrInvalidToken
		}
		return uint(v), nil
	case uint:
		if v == 0 {
			return 0, ErrInvalidToken
		}
		return v, nil
	default:
		return 0, ErrInvalidToken
	}
}
