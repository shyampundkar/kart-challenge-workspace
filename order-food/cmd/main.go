package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/handler"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/repository"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/router"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/service"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/telemetry"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Starting Order Food API server...")

	// Initialize telemetry
	config := telemetry.GetConfig("order-food")
	shutdownTelemetry, err := telemetry.InitTelemetry(config)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		telemetry.GracefulShutdown(shutdownTelemetry, 5*time.Second)
	}()

	// Initialize repositories
	productRepo := repository.NewProductRepository()
	orderRepo := repository.NewOrderRepository()

	// Initialize services
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, productRepo)

	// Initialize handlers
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)
	healthHandler := handler.NewHealthHandler()

	// Setup router with telemetry
	r := router.SetupRouter(productHandler, orderHandler, healthHandler)

	// Start server
	log.Printf("Server is running on port %s", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("API endpoint: http://localhost:%s/api", port)
	log.Printf("Products: http://localhost:%s/api/product", port)
	log.Printf("Create Order: POST http://localhost:%s/api/order (requires api_key: apitest)", port)
	log.Printf("Metrics: http://localhost:%s/metrics", port)
	log.Printf("Jaeger: %s", config.JaegerEndpoint)

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
