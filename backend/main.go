package main

import (
	"context"
	"fmt"
	"meerkat/config"
	"meerkat/database"
	apperrors "meerkat/errors"
	"meerkat/i18n"
	"meerkat/logger"
	"meerkat/middleware"
	"meerkat/routes"
	"meerkat/services"
	"meerkat/web"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

func main() {
	// Run gin in release mode to keep logging and CPU usage light on small
	// machines unless GIN_MODE is explicitly set.
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize logger first
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	isPretty := os.Getenv("LOG_PRETTY")
	prettyLog := isPretty == "true" || isPretty == "1"

	// In development, use pretty logs by default
	if os.Getenv("GIN_MODE") != "release" {
		prettyLog = true
	}

	logger.InitLogger(logger.Config{
		Level:  logLevel,
		Pretty: prettyLog,
	})

	logger.Info().Msg("Loading server...")

	logger.Info().Msg("Loading configuration...")
	cfg := config.LoadConfig()

	logger.Info().Msg("Validating configuration...")
	cfg.ValidateOrPanic()

	// `historial backup [dest]`: create a consistent snapshot archive without
	// starting the server. Useful from scripts or a double-click shortcut.
	// Default destination is the portable "backups" folder next to the exe.
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		dest := cfg.BackupDir
		if len(os.Args) > 2 {
			dest = os.Args[2]
		}
		result, err := services.CreateBackup(cfg.DBPath, cfg.ProfilePhotoDir, dest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Backup created: %s (%d bytes)\n", result.Path, result.SizeBytes)
		os.Exit(0)
	}

	logger.Info().Msg("Loading database and running migrations...")
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}

	logger.Info().Msg("Seeding admin user and unifying data...")
	if err := database.SeedAdminUser(db); err != nil {
		logger.Error().Err(err).Msg("Failed to seed admin user")
	} else if err := database.ReassignContentToAdmin(db); err != nil {
		logger.Error().Err(err).Msg("Failed to unify data under admin user")
	}

	logger.Info().Msg("Initializing i18n translations...")
	if err := i18n.Init(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize i18n")
	}

	logger.Info().Msg("Running scheduler...")
	// Schedule the reminder task daily
	if !cfg.UseResend {
		logger.Warn().Msg("No Mails to be sent since Resend configuration is not set")
	}
	s := gocron.NewScheduler(cfg.GetReminderLocation())
	task := func() {
		// Use rate-limited version to prevent duplicate emails during rapid restarts
		if err := services.SendRemindersWithRateLimit(db, *cfg); err != nil {
			logger.Error().Err(err).Msg("Error sending reminders")
		}
	}
	s.Every(1).Day().At(cfg.ReminderTime).Do(task)
	go task() // Run initially once on startup (rate-limited to prevent duplicates)
	s.Every(5).Minutes().Do(func() {
		services.ProcessWebhookRetries(db, *cfg)
	})
	// Sync calendar subscriptions regularly (rate-limited via job lock)
	calendarSyncTask := func() {
		services.SyncCalendarsWithRateLimit(db, *cfg)
	}
	s.Every(cfg.CalDAVSyncIntervalHours).Hours().Do(calendarSyncTask)
	go calendarSyncTask() // Run initially once on startup (rate-limited to prevent duplicates)
	// Sync CardDAV contact connections regularly (rate-limited via job lock)
	contactSyncTask := func() {
		services.SyncContactsWithRateLimit(db, *cfg)
	}
	s.Every(cfg.CardDAVSyncIntervalHours).Hours().Do(contactSyncTask)
	go contactSyncTask() // Run initially once on startup (rate-limited to prevent duplicates)
	go s.StartBlocking()

	r := gin.Default()

	// Limit multipart form memory to 10MB to prevent DoS via large request bodies
	r.MaxMultipartMemory = 10 << 20 // 10 MB

	// For production, set FRONTEND_URL to specific origin(s) like "https://yourdomain.com"
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "PROPFIND", "REPORT", "MKCOL", "COPY", "MOVE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Depth", "If-Match", "If-None-Match"},
		ExposeHeaders:    []string{"Content-Length", "ETag"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Cache preflight for 12 hours
	}

	// Handle wildcard "*" for development: allow any origin
	if cfg.FrontendURL == "*" {
		corsConfig.AllowOriginFunc = func(origin string) bool {
			return true // Allow all origins in development
		}
	} else {
		// Production: allow specific origin(s)
		corsConfig.AllowOrigins = []string{cfg.FrontendURL}
	}

	r.Use(cors.New(corsConfig))

	// Add security headers (HSTS only when serving over HTTPS, signaled by COOKIE_SECURE)
	r.Use(middleware.SecurityHeadersMiddleware(cfg.CookieSecure))

	// Add request body size limit middleware (10MB default) to prevent DoS
	r.Use(middleware.DefaultBodySizeLimitMiddleware())

	// Add request ID middleware for tracing
	r.Use(middleware.RequestIDMiddleware())

	// Add logging middleware (after request ID)
	r.Use(middleware.LoggingMiddleware())

	// Add error handling middleware
	r.Use(apperrors.ErrorHandlerMiddleware())

	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		logger.Fatal().Err(err).Strs("proxies", cfg.TrustedProxies).Msg("Failed to set trusted proxies")
	}

	// Inject db and cfg into context
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("cfg", *cfg)
		c.Next()
	})

	// Initialize OIDC provider if configured
	var oidcProvider *services.OIDCProvider
	if cfg.OIDC.Enabled {
		logger.Info().Str("provider", cfg.OIDC.ProviderURL).Msg("Initializing OIDC provider...")
		oidcCtx, oidcCancel := context.WithTimeout(context.Background(), 30*time.Second)
		var oidcErr error
		oidcProvider, oidcErr = services.InitOIDCProvider(oidcCtx, cfg)
		oidcCancel()
		if oidcErr != nil {
			logger.Fatal().Err(oidcErr).Msg("Failed to initialize OIDC provider")
		}
		logger.Info().Msg("OIDC provider initialized successfully")
	}

	// Register all routes from routes.go
	routes.RegisterRoutes(r, cfg, db, oidcProvider)

	// When the frontend is embedded in the binary (web_embed build), fall back
	// to the SPA. API/CardDAV paths keep their JSON 404 so the client never
	// receives HTML for a bad API call.
	if web.Available() {
		r.NoRoute(func(c *gin.Context) {
			p := c.Request.URL.Path
			if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/health") ||
				strings.HasPrefix(p, "/carddav") || strings.HasPrefix(p, "/.well-known/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
			web.Handler().ServeHTTP(c.Writer, c.Request)
		})
	}

	// Create HTTP server with timeout configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	logger.Info().
		Str("port", cfg.Port).
		Int("read_timeout", cfg.ReadTimeout).
		Int("write_timeout", cfg.WriteTimeout).
		Int("idle_timeout", cfg.IdleTimeout).
		Msg("Starting server")

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to run server")
		}
	}()

	logger.Info().Msg("Server is ready to handle requests")

	if host := lanIPv4(); host != "" {
		logger.Info().Msgf("Open in your browser: http://localhost:%s  (other computers on the network: http://%s:%s)", cfg.Port, host, cfg.Port)
	} else {
		logger.Info().Msgf("Open in your browser: http://localhost:%s", cfg.Port)
	}

	// Block until we receive a shutdown signal
	<-quit
	logger.Info().Msg("Shutting down server...")

	// Stop the scheduler first to prevent new jobs from starting
	logger.Info().Msg("Stopping scheduler...")
	s.Stop()

	// Create a deadline to wait for active requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown of HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close database connection
	logger.Info().Msg("Closing database connection...")
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Error().Err(err).Msg("Error closing database connection")
		}
	}

	logger.Info().Msg("Server exited gracefully")
}

// lanIPv4 returns the first non-loopback IPv4 address of this machine, used to
// print the URL that other computers on the same network can open.
func lanIPv4() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil {
					return ip4.String()
				}
			}
		}
	}
	return ""
}
