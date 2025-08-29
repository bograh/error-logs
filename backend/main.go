package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"error-logs/internal/config"
	"error-logs/internal/database"
	"error-logs/internal/handlers"
	"error-logs/internal/redis"
	"error-logs/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	redisClient.FlushAll(context.Background())

	// Initialize services
	errorService := services.NewErrorService(db, redisClient)
	analyticsService := services.NewAnalyticsService(db, redisClient)
	monitoringService := services.NewMonitoringService(db, redisClient)
	alertsService := services.NewAlertsService(db, redisClient)
	settingsService := services.NewSettingsService(db, redisClient)

	// Initialize handlers
	errorHandler := handlers.NewErrorHandler(errorService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService)
	alertsHandler := handlers.NewAlertsHandler(alertsService)
	settingsHandler := handlers.NewSettingsHandler(settingsService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key", "*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// API Key authentication middleware
		r.Use(handlers.APIKeyMiddleware(db))

		// Error endpoints
		r.Post("/errors", errorHandler.CreateError)
		r.Get("/errors", errorHandler.GetErrors)
		r.Get("/errors/{id}", errorHandler.GetError)
		r.Put("/errors/{id}/resolve", errorHandler.ResolveError)
		r.Delete("/errors/{id}", errorHandler.DeleteError)

		// Stats endpoint
		r.Get("/stats", errorHandler.GetStats)

		// Analytics endpoints
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/trends", analyticsHandler.GetTrends)
			r.Get("/performance", analyticsHandler.GetPerformanceMetrics)
		})

		// Monitoring endpoints
		r.Route("/monitoring", func(r chi.Router) {
			r.Get("/services", monitoringHandler.GetServiceHealth)
			r.Get("/metrics", monitoringHandler.GetSystemMetrics)
			r.Get("/uptime", monitoringHandler.GetUptime)
		})

		// Alert endpoints
		r.Route("/alerts", func(r chi.Router) {
			r.Route("/rules", func(r chi.Router) {
				r.Get("/", alertsHandler.GetAlertRules)
				r.Post("/", alertsHandler.CreateAlertRule)
				r.Put("/{id}", alertsHandler.UpdateAlertRule)
				r.Delete("/{id}", alertsHandler.DeleteAlertRule)
			})
			r.Route("/incidents", func(r chi.Router) {
				r.Get("/", alertsHandler.GetIncidents)
				r.Post("/", alertsHandler.CreateIncident)
				r.Put("/{id}", alertsHandler.UpdateIncident)
			})
		})

		// Settings endpoints
		r.Route("/settings", func(r chi.Router) {
			r.Route("/api-keys", func(r chi.Router) {
				r.Get("/", settingsHandler.GetAPIKeys)
				r.Post("/", settingsHandler.CreateAPIKey)
				r.Delete("/{id}", settingsHandler.DeleteAPIKey)
			})
			r.Route("/team", func(r chi.Router) {
				r.Get("/", settingsHandler.GetTeamMembers)
				r.Post("/invite", settingsHandler.InviteTeamMember)
			})
			r.Get("/integrations", settingsHandler.GetIntegrations)
		})
	})

	// Start background worker for processing Redis queue
	go errorService.StartQueueProcessor(context.Background())

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	log.Printf("Server starting on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
