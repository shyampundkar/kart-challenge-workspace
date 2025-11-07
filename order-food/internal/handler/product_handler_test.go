package handler

import (
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

// MockProductService is a mock implementation of ProductServiceInterface
type MockProductService struct {
	mock.Mock
}

// Verify interface compliance
var _ service.ProductServiceInterface = (*MockProductService)(nil)

func (m *MockProductService) ListProducts() []models.Product {
	args := m.Called()
	return args.Get(0).([]models.Product)
}

func (m *MockProductService) ListProductsPaginated(limit, offset int) ([]models.Product, int, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.Product), args.Int(1), args.Error(2)
}

func (m *MockProductService) GetProduct(id string) (models.Product, error) {
	args := m.Called(id)
	return args.Get(0).(models.Product), args.Error(1)
}

func TestProductHandler_ListProducts_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	// Mock data
	products := []models.Product{
		{ID: "1", Name: "Chicken Waffle", Price: 12.99, Category: "Waffle"},
		{ID: "2", Name: "Beef Waffle", Price: 14.99, Category: "Waffle"},
	}

	mockService.On("ListProductsPaginated", 10, 0).Return(products, 2, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/product?page=1&perPage=10", nil)

	// Execute
	handler.ListProducts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.PerPage)
	assert.Equal(t, 1, response.Pagination.TotalPages)
	assert.Equal(t, 2, response.Pagination.TotalItems)

	mockService.AssertExpectations(t)
}

func TestProductHandler_ListProducts_WithCustomPagination(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	// Mock data - page 2 with 5 items per page
	products := []models.Product{
		{ID: "6", Name: "Product 6", Price: 10.99, Category: "Category"},
	}

	mockService.On("ListProductsPaginated", 5, 5).Return(products, 11, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/product?page=2&perPage=5", nil)

	// Execute
	handler.ListProducts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 2, response.Pagination.Page)
	assert.Equal(t, 5, response.Pagination.PerPage)
	assert.Equal(t, 3, response.Pagination.TotalPages) // 11 items / 5 per page = 3 pages

	mockService.AssertExpectations(t)
}

func TestProductHandler_ListProducts_DatabaseError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	mockService.On("ListProductsPaginated", 10, 0).Return([]models.Product{}, 0, errors.New("database error"))

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/product", nil)

	// Execute
	handler.ListProducts(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "Failed to fetch products", response.Message)

	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProduct_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	// Mock data
	product := models.Product{
		ID:       "1",
		Name:     "Chicken Waffle",
		Price:    12.99,
		Category: "Waffle",
	}

	mockService.On("GetProduct", "1").Return(product, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "productId", Value: "1"}}
	c.Request = httptest.NewRequest("GET", "/api/product/1", nil)

	// Execute
	handler.GetProduct(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HATEOASResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify HATEOAS links
	assert.NotNil(t, response.Links)
	assert.Len(t, response.Links, 2)
	assert.Equal(t, "/api/product/1", response.Links[0].Href)
	assert.Equal(t, "self", response.Links[0].Rel)

	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProduct_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	mockService.On("GetProduct", "999").Return(models.Product{}, errors.New("not found"))

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "productId", Value: "999"}}
	c.Request = httptest.NewRequest("GET", "/api/product/999", nil)

	// Execute
	handler.GetProduct(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "Product not found", response.Message)

	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProduct_EmptyID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	// Create request with empty ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "productId", Value: ""}}
	c.Request = httptest.NewRequest("GET", "/api/product/", nil)

	// Execute
	handler.GetProduct(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Invalid ID supplied", response.Message)
}

func TestProductHandler_ListProducts_HATEOASLinksPresent(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockProductService)
	handler := NewProductHandler(mockService)

	products := []models.Product{
		{ID: "1", Name: "Product 1", Price: 10.99, Category: "Category"},
	}

	mockService.On("ListProductsPaginated", 10, 0).Return(products, 1, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/product", nil)

	// Execute
	handler.ListProducts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that data has HATEOAS links
	data := response.Data.([]interface{})
	assert.NotEmpty(t, data)

	// Check pagination links
	assert.NotNil(t, response.Links)
	assert.NotEmpty(t, response.Links)

	mockService.AssertExpectations(t)
}
