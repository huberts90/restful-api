package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/huberts90/restful-api/internal/config"
	"github.com/huberts90/restful-api/internal/handler"
	"github.com/huberts90/restful-api/internal/logger"
	"github.com/huberts90/restful-api/internal/middleware"
	"github.com/huberts90/restful-api/internal/storage"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

func main() {
	// @MENTION_ME
	// runtime.GOMAXPROCS()

	// Load configuration from
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up the logger
	zapLogger, err := logger.NewLogger(cfg.IsProd)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync() // nolint: errcheck

	// Set up the database
	store, err := storage.NewPostgresStore(cfg.Postgres, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	// @MENTION_ME: always try to close resources
	_ = store.Close()

	zapLogger.Info("Successfully connected to database")

	// Create router
	router := mux.NewRouter()

	// Apply logging middleware
	router.Use(middleware.LoggingMiddleware(zapLogger))

	// Public routes (no authentication required)
	router.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	// Create API subrouter with authentication
	// TODO: apiRouter.Use(authMiddleware.Middleware())
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Register handlers
	userHandler := handler.NewUserHandler(store, zapLogger)
	userHandler.RegisterRoutes(apiRouter)

	// Create and configure the server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second, // TODO: make the parameters adjustable
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		zapLogger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Set up graceful shutdown
	// @MENTION_ME: buffered channel does not block the sender until receiver reads from the channels
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLogger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited gracefully")
}
