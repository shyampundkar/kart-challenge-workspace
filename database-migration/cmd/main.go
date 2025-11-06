package main

import (
	"context"
	"log"
	"time"

	"github.com/shyampundkar/kart-challenge-workspace/database-migration/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	log.Println("Starting database migration service...")

	// Initialize telemetry
	config := telemetry.GetConfig("database-migration")
	shutdown, err := telemetry.InitTracer(config)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down telemetry: %v", err)
		}
	}()

	// Get tracer
	tracer := otel.Tracer("database-migration")

	// Start root span
	ctx, span := tracer.Start(context.Background(), "migration.execute")
	defer span.End()

	log.Println("Running database migrations...")

	// Simulate migration work with tracing
	runMigrations(ctx, tracer)

	log.Println("Database migration completed successfully")
}

func runMigrations(ctx context.Context, tracer trace.Tracer) {
	// Example migration steps with tracing
	_, span := tracer.Start(ctx, "migration.createTables")
	defer span.End()

	log.Println("Creating tables...")
	time.Sleep(100 * time.Millisecond) // Simulate work
	span.AddEvent("Tables created successfully")

	_, span2 := tracer.Start(ctx, "migration.createIndexes")
	defer span2.End()

	log.Println("Creating indexes...")
	time.Sleep(50 * time.Millisecond) // Simulate work
	span2.AddEvent("Indexes created successfully")

	log.Println("All migrations completed")
}
