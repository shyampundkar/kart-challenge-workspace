package service

import (
	"github.com/google/uuid"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/repository"
)

// OrderService handles order business logic
type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// PlaceOrder creates a new order
func (s *OrderService) PlaceOrder(req models.OrderReq) (models.Order, error) {
	// Extract product IDs from order items
	productIDs := make([]string, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	// Fetch products
	products, err := s.productRepo.GetByIDs(productIDs)
	if err != nil {
		return models.Order{}, err
	}

	// Create order
	order := models.Order{
		ID:         uuid.New().String(),
		CouponCode: req.CouponCode,
		Items:      req.Items,
		Products:   products,
	}

	// Store order
	if err := s.orderRepo.Create(order); err != nil {
		return models.Order{}, err
	}

	return order, nil
}

// GetOrder returns an order by ID
func (s *OrderService) GetOrder(id string) (models.Order, error) {
	return s.orderRepo.GetByID(id)
}

// CreateOrder creates a new order (alias for PlaceOrder)
func (s *OrderService) CreateOrder(req models.OrderReq) (models.Order, error) {
	return s.PlaceOrder(req)
}

// ListOrdersPaginated returns paginated orders with total count
func (s *OrderService) ListOrdersPaginated(limit, offset int) ([]models.Order, int, error) {
	return s.orderRepo.GetAll(limit, offset)
}
