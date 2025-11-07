package main

import (
	"context"
	"log"
	"time"
)

func main() {
	log.Println("Starting database load service...")

	ctx := context.Background()
	log.Println("Loading data into database...")

	// Simulate data loading work
	loadData(ctx)

	log.Println("Database load completed successfully")
}

func loadData(ctx context.Context) {
	// Example data loading steps
	log.Println("Loading products...")
	time.Sleep(150 * time.Millisecond) // Simulate work
	log.Println("✓ Loaded 100 products")

	log.Println("Loading users...")
	time.Sleep(100 * time.Millisecond) // Simulate work
	log.Println("✓ Loaded 50 users")

	log.Println("Loading orders...")
	time.Sleep(120 * time.Millisecond) // Simulate work
	log.Println("✓ Loaded 200 orders")

	log.Println("All data loaded successfully")
}
