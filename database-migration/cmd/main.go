package main

import (
	"context"
	"log"
	"os"

	"github.com/shyampundkar/kart-challenge-workspace/database-migration/internal/migration"
)

func main() {
	log.Println("Starting database migration service...")

	// Get database configuration from environment variables
	dbConfig := migration.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "orderfood"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	log.Printf("Connecting to database: %s@%s:%s/%s", dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	// Create migrator
	migrator, err := migration.NewMigrator(dbConfig)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer migrator.Close()

	// Run migrations
	log.Println("Running database migrations...")
	ctx := context.Background()
	if err := migrator.Run(ctx); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migration completed successfully")
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
