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
	api := router.Group("/api")
	{
		// Product routes (no auth required)
		api.GET("/product", productHandler.ListProducts)
		api.GET("/product/:productId", productHandler.GetProduct)

		// Order routes (auth required)
		orderRoutes := api.Group("")
		orderRoutes.Use(middleware.AuthMiddleware())
		orderRoutes.POST("/order", orderHandler.PlaceOrder)
	}

	return router
}
