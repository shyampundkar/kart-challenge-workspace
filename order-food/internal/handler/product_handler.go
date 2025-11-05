package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/service"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service *service.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// ListProducts handles GET /product
// @Summary List products
// @Description Get all products available for order
// @Tags product
// @Produce json
// @Success 200 {array} models.Product
// @Router /product [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	products := h.service.ListProducts()
	c.JSON(http.StatusOK, products)
}

// GetProduct handles GET /product/:productId
// @Summary Find product by ID
// @Description Returns a single product
// @Tags product
// @Produce json
// @Param productId path int true "ID of product to return"
// @Success 200 {object} models.Product
// @Failure 400 {object} models.ApiResponse "Invalid ID supplied"
// @Failure 404 {object} models.ApiResponse "Product not found"
// @Router /product/{productId} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("productId")

	if productID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Invalid ID supplied"))
		return
	}

	product, err := h.service.GetProduct(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(http.StatusNotFound, "Product not found"))
		return
	}

	c.JSON(http.StatusOK, product)
}
