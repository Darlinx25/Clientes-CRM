package middleware

import (
	"meerkat/adminauth"
	"meerkat/config"
	apperrors "meerkat/errors"

	"github.com/gin-gonic/gin"
)

func AdminSessionMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(adminauth.CookieName)
		if err != nil || token == "" {
			abortAdminSession(c)
			return
		}

		tokenUserID, err := adminauth.ParseToken(token, cfg)
		if err != nil {
			abortAdminSession(c)
			return
		}

		sessionID, ok := c.Get("userID")
		if !ok {
			abortAdminSession(c)
			return
		}
		userID, ok := sessionID.(uint)
		if !ok || userID != tokenUserID {
			abortAdminSession(c)
			return
		}

		c.Set("adminPanelUserID", tokenUserID)
		c.Next()
	}
}

func abortAdminSession(c *gin.Context) {
	apperrors.AbortWithError(c, apperrors.ErrAdminSessionRequired())
}
