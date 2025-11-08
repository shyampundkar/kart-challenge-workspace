package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/handler"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/repository"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/router"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/service"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Starting Order Food API server...")

	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Initialize services
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, productRepo)
	promoCodeService := service.NewPromoCodeService(db)

	// Initialize handlers
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService, promoCodeService)
	healthHandler := handler.NewHealthHandler()

	// Setup router
	r := router.SetupRouter(productHandler, orderHandler, healthHandler)

	// Start server
	log.Printf("Server is running on port %s", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("API endpoint: http://localhost:%s/api/v1", port)
	log.Printf("Products: http://localhost:%s/api/v1/products", port)
	log.Printf("Create Order: POST http://localhost:%s/api/v1/orders (requires api_key: apitest)", port)

	// Graceful shutdown
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Cleanup
	log.Println("Server stopped")
	_ = ctx // Use context if needed for cleanup
}

func connectDB() (*sql.DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "orderfood")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection with retries
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		if err := db.PingContext(ctx); err == nil {
			log.Println("Successfully connected to database")
			return db, nil
		}
		log.Printf("Waiting for database connection... (attempt %d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after retries")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
