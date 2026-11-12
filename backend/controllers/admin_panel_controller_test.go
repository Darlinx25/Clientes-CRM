package controllers

import (
	"bytes"
	"encoding/json"
	"meerkat/config"
	"meerkat/middleware"
	"meerkat/models"
	"meerkat/services"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupAdminPanelRouter(t *testing.T) (*gorm.DB, *gin.Engine, config.Config, models.User) {
	t.Helper()
	gin.SetMode(gin.ReleaseMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Contact{},
		&models.Note{},
		&models.Activity{},
		&models.Reminder{},
		&models.ReminderCompletion{},
		&models.Relationship{},
		&models.JobExecution{},
	))

	password, err := services.HashPassword("secret-password")
	require.NoError(t, err)
	user := models.User{Username: "admin", Password: password, Email: "admin@example.com", IsAdmin: true}
	require.NoError(t, db.Create(&user).Error)

	dir := t.TempDir()
	cfg := config.Config{
		DBPath:       filepath.Join(dir, "meerkat.db"),
		BackupDir:    filepath.Join(dir, "backups"),
		JWTSecretKey: "test-secret",
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("cfg", cfg)
		c.Set("userID", user.ID)
		c.Next()
	})

	router.POST("/session", middleware.ValidateJSONMiddleware(&models.AdminPanelReauthInput{}), AdminPanelSession)
	router.DELETE("/session", AdminPanelSessionDelete)
	router.GET("/normal-auth", middleware.AuthMiddleware(&cfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	panel := router.Group("/panel")
	panel.Use(middleware.AdminSessionMiddleware(&cfg))
	{
		panel.GET("/overview", AdminPanelOverview)
		panel.GET("/tables", AdminPanelTables)
		panel.GET("/tables/:table", AdminPanelTableRows)
		panel.GET("/backups", AdminPanelBackups)
	}

	return db, router, cfg, user
}

func performJSONRequest(router *gin.Engine, method, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	var payload bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&payload).Encode(body)
	}
	req := httptest.NewRequest(method, path, &payload)
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func adminSessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "admin_token" {
			return cookie
		}
	}
	t.Fatal("admin session cookie not set")
	return nil
}

func TestAdminPanelSessionRejectsWrongPassword(t *testing.T) {
	_, router, _, _ := setupAdminPanelRouter(t)

	rec := performJSONRequest(router, http.MethodPost, "/session", gin.H{"password": "wrong-password"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, rec.Result().Cookies())
}

func TestAdminPanelSessionAllowsPanelAccess(t *testing.T) {
	_, router, _, _ := setupAdminPanelRouter(t)

	login := performJSONRequest(router, http.MethodPost, "/session", gin.H{"password": "secret-password"})
	require.Equal(t, http.StatusOK, login.Code)
	cookie := adminSessionCookie(t, login)

	overview := performJSONRequest(router, http.MethodGet, "/panel/overview", nil, cookie)
	require.Equal(t, http.StatusOK, overview.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(overview.Body.Bytes(), &body))
	assert.NotEmpty(t, body["app_version"])
	assert.NotNil(t, body["schema"])

	tables := performJSONRequest(router, http.MethodGet, "/panel/tables", nil, cookie)
	require.Equal(t, http.StatusOK, tables.Code)
	assert.Contains(t, tables.Body.String(), "contacts")
}

func TestAdminPanelRequiresSession(t *testing.T) {
	_, router, _, _ := setupAdminPanelRouter(t)

	rec := performJSONRequest(router, http.MethodGet, "/panel/tables", nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "ADMIN_SESSION_REQUIRED")
	assert.Contains(t, rec.Body.String(), "Admin session required")
}

func TestAdminPanelSessionTokenCannotAuthenticateNormalRoutes(t *testing.T) {
	_, router, _, _ := setupAdminPanelRouter(t)

	login := performJSONRequest(router, http.MethodPost, "/session", gin.H{"password": "secret-password"})
	require.Equal(t, http.StatusOK, login.Code)
	cookie := adminSessionCookie(t, login)

	req := httptest.NewRequest(http.MethodGet, "/normal-auth", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: cookie.Value})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAdminPanelUnknownTable(t *testing.T) {
	_, router, _, _ := setupAdminPanelRouter(t)

	login := performJSONRequest(router, http.MethodPost, "/session", gin.H{"password": "secret-password"})
	require.Equal(t, http.StatusOK, login.Code)
	cookie := adminSessionCookie(t, login)

	rec := performJSONRequest(router, http.MethodGet, "/panel/tables/api_tokens", nil, cookie)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminPanelRowsHideUserSecrets(t *testing.T) {
	_, router, _, _ := setupAdminPanelRouter(t)

	login := performJSONRequest(router, http.MethodPost, "/session", gin.H{"password": "secret-password"})
	require.Equal(t, http.StatusOK, login.Code)
	cookie := adminSessionCookie(t, login)

	rec := performJSONRequest(router, http.MethodGet, "/panel/tables/users", nil, cookie)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), "password")
	assert.Contains(t, rec.Body.String(), "admin@example.com")
}
