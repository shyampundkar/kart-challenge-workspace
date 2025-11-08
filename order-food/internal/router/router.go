package router

import (
	"github.com/gin-gonic/gin"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/handler"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/middleware"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	healthHandler *handler.HealthHandler,
) *gin.Engine {
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Product routes (no auth required)
		v1.GET("/products", productHandler.ListProducts)
		v1.GET("/products/:productId", productHandler.GetProduct)

		// Order routes (auth required)
		orderRoutes := v1.Group("")
		orderRoutes.Use(middleware.AuthMiddleware())
		orderRoutes.GET("/orders", orderHandler.ListOrders)
		orderRoutes.GET("/orders/:orderId", orderHandler.GetOrder)
		orderRoutes.POST("/orders", orderHandler.CreateOrder)
	}

	return router
}
