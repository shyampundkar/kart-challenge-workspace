package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler()

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	// Execute
	handler.Health(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestHealthHandler_Ready(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler()

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ready", nil)

	// Execute
	handler.Ready(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
}

func TestHealthHandler_Health_ResponseFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler()

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	// Execute
	handler.Health(c)

	// Assert
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestHealthHandler_Ready_ResponseFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler()

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/ready", nil)

	// Execute
	handler.Ready(c)

	// Assert
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "ready")
}
