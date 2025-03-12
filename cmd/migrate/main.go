package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/huberts90/restful-api/internal/config"
)

func main() {
	var migrationsPath string
	var direction string
	var steps int

	// Parse command-line arguments
	flag.StringVar(&migrationsPath, "path", "migrations", "Path to migration files")
	flag.StringVar(&direction, "direction", "up", "Migration direction: up, down")
	flag.IntVar(&steps, "steps", 0, "Number of migration steps (0 means all)")
	flag.Parse()

	// Load database configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create DSN string for PostgreSQL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	// Create migration instance
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("Failed to close migration source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("Failed to close migration database: %v", dbErr)
		}
	}()

	// Enable verbose logging
	m.Log = &MigrateLogger{}

	// Execute migration based on command
	switch direction {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
	default:
		log.Fatalf("Invalid command: %s", direction)
	}

	// Handle completion
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully")
}

// MigrateLogger is a simple logger for the migrate package
type MigrateLogger struct{}

// Printf implements the migrate.Logger interface
func (l *MigrateLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Verbose implements the migrate.Logger interface
func (l *MigrateLogger) Verbose() bool {
	return true
}
