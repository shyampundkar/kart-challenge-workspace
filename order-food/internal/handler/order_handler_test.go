package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderService is a mock implementation of OrderServiceInterface
type MockOrderService struct {
	mock.Mock
}

// Verify interface compliance
var _ service.OrderServiceInterface = (*MockOrderService)(nil)

func (m *MockOrderService) CreateOrder(req models.OrderReq) (models.Order, error) {
	args := m.Called(req)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *MockOrderService) GetOrder(id string) (models.Order, error) {
	args := m.Called(id)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *MockOrderService) ListOrdersPaginated(limit, offset int) ([]models.Order, int, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.Order), args.Int(1), args.Error(2)
}

// MockPromoCodeService is a mock implementation of PromoCodeServiceInterface
type MockPromoCodeService struct {
	mock.Mock
}

// Verify interface compliance
var _ service.PromoCodeServiceInterface = (*MockPromoCodeService)(nil)

func (m *MockPromoCodeService) ValidatePromoCode(code string) (bool, error) {
	args := m.Called(code)
	return args.Bool(0), args.Error(1)
}

func TestOrderHandler_CreateOrder_Success_WithValidPromoCode(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	// Mock data
	orderReq := models.OrderReq{
		CouponCode: "HAPPYHRS",
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 2},
		},
	}

	order := models.Order{
		ID:    "order-123",
		Items: orderReq.Items,
		Products: []models.Product{
			{ID: "1", Name: "Product 1", Price: 10.99, Category: "Category"},
		},
	}

	mockPromoService.On("ValidatePromoCode", "HAPPYHRS").Return(true, nil)
	mockOrderService.On("CreateOrder", orderReq).Return(order, nil)

	// Create request
	body, _ := json.Marshal(orderReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateOrder(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.HATEOASResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify HATEOAS links
	assert.NotNil(t, response.Links)
	assert.Len(t, response.Links, 3)

	mockPromoService.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_Success_WithoutPromoCode(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	// Mock data
	orderReq := models.OrderReq{
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 1},
		},
	}

	order := models.Order{
		ID:    "order-456",
		Items: orderReq.Items,
	}

	mockOrderService.On("CreateOrder", orderReq).Return(order, nil)

	// Create request
	body, _ := json.Marshal(orderReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateOrder(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockOrderService.AssertExpectations(t)
	// Promo service should not be called
	mockPromoService.AssertNotCalled(t, "ValidatePromoCode")
}

func TestOrderHandler_CreateOrder_InvalidPromoCode(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	// Mock data
	orderReq := models.OrderReq{
		CouponCode: "INVALID",
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 1},
		},
	}

	mockPromoService.On("ValidatePromoCode", "INVALID").Return(false, nil)

	// Create request
	body, _ := json.Marshal(orderReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateOrder(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "Invalid promo code")

	mockPromoService.AssertExpectations(t)
	// Order service should not be called
	mockOrderService.AssertNotCalled(t, "CreateOrder")
}

func TestOrderHandler_CreateOrder_PromoCodeValidationError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	// Mock data
	orderReq := models.OrderReq{
		CouponCode: "TESTCODE",
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 1},
		},
	}

	mockPromoService.On("ValidatePromoCode", "TESTCODE").Return(false, errors.New("database error"))

	// Create request
	body, _ := json.Marshal(orderReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateOrder(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "Failed to validate promo code")

	mockPromoService.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "CreateOrder")
}

func TestOrderHandler_CreateOrder_InvalidJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	// Create request with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateOrder(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_GetOrder_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	// Mock data
	order := models.Order{
		ID: "order-123",
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 2},
		},
	}

	mockOrderService.On("GetOrder", "order-123").Return(order, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "orderId", Value: "order-123"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/orders/order-123", nil)

	// Execute
	handler.GetOrder(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HATEOASResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify HATEOAS links
	assert.NotNil(t, response.Links)
	assert.Len(t, response.Links, 3)

	mockOrderService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	mockOrderService.On("GetOrder", "nonexistent").Return(models.Order{}, errors.New("not found"))

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "orderId", Value: "nonexistent"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/orders/nonexistent", nil)

	// Execute
	handler.GetOrder(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order not found", response.Message)

	mockOrderService.AssertExpectations(t)
}

func TestOrderHandler_ListOrders_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	// Mock data
	orders := []models.Order{
		{ID: "order-1", Items: []models.OrderItem{{ProductID: "1", Quantity: 1}}},
		{ID: "order-2", Items: []models.OrderItem{{ProductID: "2", Quantity: 2}}},
	}

	mockOrderService.On("ListOrdersPaginated", 10, 0).Return(orders, 2, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/orders?page=1&perPage=10", nil)

	// Execute
	handler.ListOrders(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.PerPage)
	assert.Equal(t, 1, response.Pagination.TotalPages)
	assert.Equal(t, 2, response.Pagination.TotalItems)

	mockOrderService.AssertExpectations(t)
}

func TestOrderHandler_ListOrders_DatabaseError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockOrderService := new(MockOrderService)
	mockPromoService := new(MockPromoCodeService)
	handler := NewOrderHandler(mockOrderService, mockPromoService)

	mockOrderService.On("ListOrdersPaginated", 10, 0).Return([]models.Order{}, 0, errors.New("database error"))

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/orders", nil)

	// Execute
	handler.ListOrders(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to fetch orders", response.Message)

	mockOrderService.AssertExpectations(t)
}
