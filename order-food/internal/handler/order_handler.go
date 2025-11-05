package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/service"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	service *service.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// PlaceOrder handles POST /order
// @Summary Place an order
// @Description Place a new order in the store
// @Tags order
// @Accept json
// @Produce json
// @Param order body models.OrderReq true "Order request"
// @Success 200 {object} models.Order
// @Failure 400 {object} models.ApiResponse "Invalid input"
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Failure 422 {object} models.ApiResponse "Validation exception"
// @Security ApiKeyAuth
// @Router /order [post]
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req models.OrderReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse(http.StatusUnprocessableEntity, err.Error()))
		return
	}

	// Validate that items array is not empty
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Items array cannot be empty"))
		return
	}

	order, err := h.service.PlaceOrder(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, order)
}
