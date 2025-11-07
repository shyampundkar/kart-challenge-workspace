package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/service"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/utils"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	service          *service.OrderService
	promoCodeService *service.PromoCodeService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *service.OrderService, promoCodeService *service.PromoCodeService) *OrderHandler {
	return &OrderHandler{
		service:          service,
		promoCodeService: promoCodeService,
	}
}

// CreateOrder handles POST /order with promo code validation and HATEOAS
// @Summary Place an order
// @Description Place a new order in the store
// @Tags order
// @Accept json
// @Produce json
// @Param order body models.OrderReq true "Order request"
// @Success 200 {object} models.Order
// @Failure 400 {object} models.APIResponse "Invalid input"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 403 {object} models.APIResponse "Forbidden"
// @Failure 422 {object} models.APIResponse "Validation exception"
// @Security ApiKeyAuth
// @Router /order [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.OrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// Validate promo code if provided
	if req.CouponCode != "" {
		valid, err := h.promoCodeService.ValidatePromoCode(req.CouponCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to validate promo code"))
			return
		}
		if !valid {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Invalid promo code. Code must be 8-10 characters and exist in at least 2 files."))
			return
		}
	}

	order, err := h.service.CreateOrder(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	response := models.HATEOASResponse{
		Data: order,
		Links: []models.Link{
			{Href: fmt.Sprintf("/api/order/%s", order.ID), Rel: "self", Method: "GET"},
			{Href: "/api/order", Rel: "collection", Method: "GET"},
			{Href: "/api/product", Rel: "products", Method: "GET"},
		},
	}

	c.JSON(http.StatusCreated, response)
}

// GetOrder handles GET /order/:orderId with HATEOAS
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("orderId")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Invalid ID supplied"))
		return
	}

	order, err := h.service.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(http.StatusNotFound, "Order not found"))
		return
	}

	response := models.HATEOASResponse{
		Data: order,
		Links: []models.Link{
			{Href: fmt.Sprintf("/api/order/%s", orderID), Rel: "self", Method: "GET"},
			{Href: "/api/order", Rel: "collection", Method: "GET"},
			{Href: "/api/product", Rel: "products", Method: "GET"},
		},
	}

	c.JSON(http.StatusOK, response)
}

// ListOrders handles GET /order with pagination and HATEOAS
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Parse pagination parameters
	page := utils.ParseInt(c.Query("page"), 1)
	perPage := utils.ParseInt(c.Query("perPage"), 10)

	// Calculate offset
	offset := (page - 1) * perPage

	// Get paginated orders
	orders, total, err := h.service.ListOrdersPaginated(perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to fetch orders"))
		return
	}

	// Add HATEOAS links to each order
	ordersWithLinks := make([]models.OrderWithLinks, len(orders))
	for i, order := range orders {
		ordersWithLinks[i] = models.OrderWithLinks{
			Order: order,
			Links: []models.Link{
				{Href: fmt.Sprintf("/api/order/%s", order.ID), Rel: "self", Method: "GET"},
				{Href: "/api/order", Rel: "collection", Method: "GET"},
			},
		}
	}

	// Build pagination response
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	response := models.PaginatedResponse{
		Data: ordersWithLinks,
		Pagination: models.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
			TotalItems: total,
		},
		Links: utils.BuildPaginationLinks(page, totalPages, "/api/order", perPage),
	}

	c.JSON(http.StatusOK, response)
}
