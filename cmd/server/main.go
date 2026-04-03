package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rajneesh180/finance-backend/internal/config"
	"github.com/Rajneesh180/finance-backend/internal/database"
	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/handler"
	"github.com/Rajneesh180/finance-backend/internal/middleware"
	"github.com/Rajneesh180/finance-backend/internal/repository"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	pool, err := database.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	recordRepo := repository.NewRecordRepository(pool)
	dashboardRepo := repository.NewDashboardRepository(pool)

	// Services
	authSvc := service.NewAuthService(cfg.JWTSecret, cfg.JWTExpiry)
	userSvc := service.NewUserService(userRepo, authSvc)
	recordSvc := service.NewRecordService(recordRepo, dashboardRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(userSvc)
	userHandler := handler.NewUserHandler(userSvc)
	recordHandler := handler.NewRecordHandler(recordSvc)
	dashHandler := handler.NewDashboardHandler(recordSvc)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(logger))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc))

		// User profile
		r.Get("/users/me", userHandler.GetProfile)
		r.Put("/users/me", userHandler.UpdateProfile)

		// Financial records (analyst + admin can create/update)
		r.Route("/records", func(r chi.Router) {
			r.Get("/", recordHandler.List)
			r.Get("/{id}", recordHandler.GetByID)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(domain.RoleAdmin, domain.RoleAnalyst))
				r.Post("/", recordHandler.Create)
				r.Put("/{id}", recordHandler.Update)
				r.Delete("/{id}", recordHandler.Delete)
			})
		})

		// Dashboard
		r.Get("/dashboard/summary", dashHandler.Summary)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleAdmin))
			r.Get("/users", userHandler.ListUsers)
			r.Get("/users/{id}", userHandler.GetUser)
			r.Put("/users/{id}", userHandler.AdminUpdateUser)
			r.Delete("/users/{id}", userHandler.DeleteUser)
		})
	})

	// Server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	logger.Info("server stopped")
}
