package main

import (
	"context"
	"log"
	"time"

	"github.com/shyampundkar/kart-challenge-workspace/database-load/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	log.Println("Starting database load service...")

	// Initialize telemetry
	config := telemetry.GetConfig("database-load")
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
	tracer := otel.Tracer("database-load")

	// Start root span
	ctx, span := tracer.Start(context.Background(), "dataload.execute")
	defer span.End()

	log.Println("Loading data into database...")

	// Simulate data loading work with tracing
	loadData(ctx, tracer)

	log.Println("Database load completed successfully")
}

func loadData(ctx context.Context, tracer trace.Tracer) {
	// Example data loading steps with tracing
	_, span := tracer.Start(ctx, "dataload.loadProducts")
	defer span.End()

	log.Println("Loading products...")
	time.Sleep(150 * time.Millisecond) // Simulate work
	span.AddEvent("Loaded 100 products")

	_, span2 := tracer.Start(ctx, "dataload.loadUsers")
	defer span2.End()

	log.Println("Loading users...")
	time.Sleep(100 * time.Millisecond) // Simulate work
	span2.AddEvent("Loaded 50 users")

	_, span3 := tracer.Start(ctx, "dataload.loadOrders")
	defer span3.End()

	log.Println("Loading orders...")
	time.Sleep(120 * time.Millisecond) // Simulate work
	span3.AddEvent("Loaded 200 orders")

	log.Println("All data loaded successfully")
}
