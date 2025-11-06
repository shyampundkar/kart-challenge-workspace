package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
)

const (
	// ValidAPIKey is the expected API key for authentication
	ValidAPIKey = "apitest"
	// APIKeyHeader is the header name for the API key
	APIKeyHeader = "api_key"
)

// AuthMiddleware validates the API key from the request header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(APIKeyHeader)

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Unauthorized: API key is required"))
			c.Abort()
			return
		}

		if apiKey != ValidAPIKey {
			c.JSON(http.StatusForbidden, models.ErrorResponse(http.StatusForbidden, "Forbidden: Invalid API key"))
			c.Abort()
			return
		}

		c.Next()
	}
}
