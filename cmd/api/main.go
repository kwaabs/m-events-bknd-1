package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/kwaabs/m-events/internal/auth"
	"github.com/kwaabs/m-events/internal/config"
	"github.com/kwaabs/m-events/internal/database"
	"github.com/kwaabs/m-events/internal/handlers"
	"github.com/kwaabs/m-events/internal/middleware"
	"github.com/kwaabs/m-events/internal/repository"
)

func main() {
	// ── Logger ────────────────────────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := database.New(cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected", "host", cfg.Database.Host, "db", cfg.Database.DBName)

	// ── Valkey ────────────────────────────────────────────────────────────────
	valkeySvc, err := auth.NewValkeyService(cfg.Valkey.URL)
	if err != nil {
		slog.Error("failed to connect to Valkey", "error", err)
		os.Exit(1)
	}
	defer valkeySvc.Close()
	slog.Info("valkey connected", "url", cfg.Valkey.URL)

	// ── Auth services ─────────────────────────────────────────────────────────
	ldapSvc := auth.NewLDAPService(cfg.LDAP)
	jwtSvc := auth.NewJWTService(cfg.JWT)

	// Dev account service — only populated in development.
	// Map config.DevAccount → auth.DevAccount to avoid an import cycle.
	var devSvc *auth.DevAuthService
	if cfg.Server.IsDevelopment() {
		var devAccounts []auth.DevAccount
		for _, a := range cfg.DevAccounts.Accounts {
			devAccounts = append(devAccounts, auth.DevAccount{
				Username:     a.Username,
				DisplayName:  a.DisplayName,
				Email:        a.Email,
				PasswordHash: a.PasswordHash,
			})
		}
		devSvc = auth.NewDevAuthService(devAccounts)
		if len(devAccounts) > 0 {
			slog.Warn("dev auth enabled",
				"accounts", len(devAccounts),
				"note", "disable in production by setting APP_ENV=production",
			)
		}
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	customerRepo := repository.NewCustomerRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(ldapSvc, devSvc, jwtSvc, valkeySvc, cfg.Valkey, cfg.Server.IsDevelopment())
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardRepo)
	tileHandler := handlers.NewTileHandler(jwtSvc, valkeySvc)

	// ── Router ────────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// ── Public routes ─────────────────────────────────────────────────────────
	r.Get("/health", handlers.HealthCheck)

	r.Route("/api/v1", func(r chi.Router) {

		// Public — no token required
		r.Post("/auth/login", authHandler.Login)

		// ── Tiles — auth handled inside handler (supports ?token= for MapLibre) ──
		// GET /api/v1/tiles/{source}/{z}/{x}/{y}?token=<jwt>
		r.Get("/tiles/{source}/{z}/{x}/{y}", tileHandler.ProxyTile)

		// Protected — all routes below require a valid Bearer JWT
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(jwtSvc, valkeySvc))

			// ── Auth ──────────────────────────────────────────────────────────
			// GET  /api/v1/auth/me
			r.Get("/auth/me", authHandler.Me)
			// POST /api/v1/auth/logout
			r.Post("/auth/logout", authHandler.Logout)

			// ── Customers ─────────────────────────────────────────────────────
			r.Route("/customers", func(r chi.Router) {
				// GET /api/v1/customers
				r.Get("/", customerHandler.List)

				// GET /api/v1/customers/no-coordinates
				r.Get("/no-coordinates", customerHandler.NoCoordinates)

				// GET /api/v1/customers/account/{accountNumber}
				r.Get("/account/{accountNumber}", customerHandler.GetByAccount)

				// GET /api/v1/customers/account/{accountNumber}/events
				r.Get("/account/{accountNumber}/events", customerHandler.EventsByAccount)

				// GET /api/v1/customers/meter/{meterNumber}
				r.Get("/meter/{meterNumber}", customerHandler.GetByMeter)

				// GET /api/v1/customers/meter/{meterNumber}/events
				r.Get("/meter/{meterNumber}/events", customerHandler.EventsByMeter)
			})

			// ── Martin health (protected) ───────────────────────────────────────
			r.Get("/martin/health", tileHandler.MartinHealth)

			// ── Dashboard ─────────────────────────────────────────────────────
			r.Route("/dashboard", func(r chi.Router) {
				// GET /api/v1/dashboard/summary
				r.Get("/summary", dashboardHandler.Summary)

				// GET /api/v1/dashboard/regions
				r.Get("/regions", dashboardHandler.Regions)

				// GET /api/v1/dashboard/districts?region_code=
				r.Get("/districts", dashboardHandler.Districts)
			})
		})
	})

	// ── Server ────────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Server.Port, "env", cfg.Server.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
