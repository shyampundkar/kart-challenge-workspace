package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware_LogsRequest(t *testing.T) {
	// Setup - capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "200")
}

func TestLoggerMiddleware_LogsDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		// Setup
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(LoggerMiddleware())
		router.Handle(method, "/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Create request
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/test", nil)

		// Execute
		router.ServeHTTP(w, req)

		// Assert
		logOutput := buf.String()
		assert.Contains(t, logOutput, method, "Failed to log method: "+method)
	}
}

func TestLoggerMiddleware_LogsStatusCodes(t *testing.T) {
	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, statusCode := range statusCodes {
		// Setup
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(LoggerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(statusCode, gin.H{"message": "test"})
		})

		// Create request
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		// Execute
		router.ServeHTTP(w, req)

		// Assert - check that status code is logged
		logOutput := buf.String()
		statusCodeStr := fmt.Sprintf("%d", statusCode)
		assert.Contains(t, logOutput, statusCodeStr, "Failed for status code: %d", statusCode)
	}
}

func TestLoggerMiddleware_LogsLatency(t *testing.T) {
	// Setup
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Execute
	router.ServeHTTP(w, req)

	// Assert - should contain some latency information
	logOutput := buf.String()
	assert.NotEmpty(t, logOutput)
}

func TestLoggerMiddleware_NextCalled(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handlerCalled := false
	router := gin.New()
	router.Use(LoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Execute
	router.ServeHTTP(w, req)

	// Assert - handler should have been called
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}
