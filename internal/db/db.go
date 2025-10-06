package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
	logger *slog.Logger
}

// New creates a new database connection
func New(cfg Config, logger *slog.Logger) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
	)

	return &DB{
		DB:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info("closing database connection")
	return db.DB.Close()
}

// Ping checks if the database is reachable
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// HealthCheck performs a health check on the database
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
