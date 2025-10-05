package server

import (
	"asocial/internal/auth"
	"asocial/internal/config"
	"asocial/internal/db"
	"asocial/internal/handler"
	"asocial/internal/middleware"
	"asocial/internal/pubsub"
	"asocial/internal/repository"
	"asocial/internal/service"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

// Run starts the chat server
func Run() {
	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("Starting asocial chat server")

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Configuration loaded",
		"server_port", cfg.Server.Port,
		"redis_addr", cfg.Redis.Addr,
		"redis_channel", cfg.Redis.Channel,
	)

	// Initialize database connection
	dbCfg := db.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
	database, err := db.New(dbCfg, logger)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.DB)

	// Initialize Firebase
	firebaseClient, err := auth.InitializeFirebase(context.Background(), cfg.Auth.FirebaseCredentialsPath)
	if err != nil {
		logger.Error("Failed to initialize Firebase", "error", err)
		os.Exit(1)
	}

	// Initialize Firebase auth service
	firebaseService := auth.NewFirebaseService(firebaseClient, userRepo, logger)

	// Initialize Melody (WebSocket manager)
	m := melody.New()
	m.Config.MaxMessageSize = int64(cfg.Server.MaxMessageSize)

	// Initialize Redis pub/sub
	redisPubSub, err := pubsub.NewRedisPubSub(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.Channel,
		cfg.Redis.DB,
		logger,
	)
	if err != nil {
		logger.Error("Failed to initialize Redis pub/sub", "error", err)
		os.Exit(1)
	}
	defer redisPubSub.Close()

	// Initialize message service
	msgService := service.NewMessageService(redisPubSub, m, logger)

	// Start subscriber in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := msgService.StartSubscriber(ctx); err != nil && err != context.Canceled {
			logger.Error("Subscriber error", "error", err)
		}
	}()

	// Initialize handlers
	wsHandler := handler.NewWebSocketHandler(m, msgService, logger)
	healthHandler := handler.NewHealthHandler(msgService, logger)
	isDev := os.Getenv("ENVIRONMENT") != "production"
	authHandler := handler.NewAuthHandler(firebaseService, logger, cfg.Auth.AppURL, isDev)

	// Setup Gin router
	router := gin.Default()

	// Register health and debug routes
	router.GET("/health", healthHandler.HandleLiveness)
	router.GET("/ready", healthHandler.HandleReadiness)
	router.GET("/api/debug/goroutines", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"goroutines": runtime.NumGoroutine(),
		})
	})

	// Register auth routes
	authGroup := router.Group("/api/auth")
	{
		// Public routes
		authGroup.POST("/check-username", authHandler.HandleCheckUsername)

		// Protected auth routes (middleware auto-creates user on first call)
		authGroup.GET("/me", middleware.AuthMiddleware(firebaseService, logger), authHandler.HandleMe)
		authGroup.POST("/logout", middleware.AuthMiddleware(firebaseService, logger), authHandler.HandleLogout)
	}

	// Register WebSocket route (optionally authenticated)
	router.GET("/api/chat", middleware.OptionalAuthMiddleware(firebaseService, logger), wsHandler.HandleUpgrade)

	// Create HTTP server
	addr := ":" + cfg.Server.Port
	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("HTTP server listening", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Cancel subscriber context
	cancel()

	// Graceful shutdown with 5 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Close all WebSocket connections
	if err := m.Close(); err != nil {
		logger.Error("Error closing WebSocket connections", "error", err)
	}

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error shutting down HTTP server", "error", err)
	}

	logger.Info("Server stopped gracefully")
}
