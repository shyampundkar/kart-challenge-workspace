package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/service"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/utils"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service service.ProductServiceInterface
}

// NewProductHandler creates a new product handler
func NewProductHandler(service service.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{service: service}
}

// ListProducts handles GET /product with pagination and HATEOAS
// @Summary List products
// @Description Get all products available for order
// @Tags product
// @Produce json
// @Success 200 {array} models.Product
// @Router /product [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// Parse pagination parameters
	page := utils.ParseInt(c.Query("page"), 1)
	perPage := utils.ParseInt(c.Query("perPage"), 10)

	// Calculate offset
	offset := (page - 1) * perPage

	// Get paginated products
	products, total, err := h.service.ListProductsPaginated(perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to fetch products"))
		return
	}

	// Add HATEOAS links to each product
	productsWithLinks := make([]models.ProductWithLinks, len(products))
	for i, product := range products {
		productsWithLinks[i] = models.ProductWithLinks{
			Product: product,
			Links: []models.Link{
				{Href: fmt.Sprintf("/api/v1/products/%s", product.ID), Rel: "self", Method: "GET"},
				{Href: "/api/v1/products", Rel: "collection", Method: "GET"},
			},
		}
	}

	// Build pagination response
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	response := models.PaginatedResponse{
		Data: productsWithLinks,
		Pagination: models.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
			TotalItems: total,
		},
		Links: utils.BuildPaginationLinks(page, totalPages, "/api/v1/products", perPage),
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct handles GET /product/:productId with HATEOAS
// @Summary Find product by ID
// @Description Returns a single product
// @Tags product
// @Produce json
// @Param productId path int true "ID of product to return"
// @Success 200 {object} models.Product
// @Failure 400 {object} models.APIResponse "Invalid ID supplied"
// @Failure 404 {object} models.APIResponse "Product not found"
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

	response := models.HATEOASResponse{
		Data: product,
		Links: []models.Link{
			{Href: fmt.Sprintf("/api/v1/products/%s", productID), Rel: "self", Method: "GET"},
			{Href: "/api/v1/products", Rel: "collection", Method: "GET"},
		},
	}

	c.JSON(http.StatusOK, response)
}
