package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB holds the connection pool
var DB *pgxpool.Pool

// Config holds database connection configuration
type Config struct {
	URL             string
	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultConfig returns a Config with default values
func DefaultConfig(databaseURL string) *Config {
	return &Config{
		URL:             databaseURL,
		MaxConnections:  20,
		MinConnections:  2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}

// Connect initializes the connection pool with the given configuration
func Connect(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = pool
	return pool, nil
}

// Close closes the connection pool
func Close() {
	if DB != nil {
		DB.Close()
	}
}

// HealthCheck performs a health check on the database connection
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database connection pool not initialized")
	}

	// Execute a simple query to check connection
	var result int
	err := DB.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetPool returns the current connection pool
func GetPool() *pgxpool.Pool {
	return DB
}
