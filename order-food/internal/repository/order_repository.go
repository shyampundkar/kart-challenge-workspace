package repository

import (
	"errors"
	"sync"

	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
)

// OrderRepository handles order data operations
type OrderRepository struct {
	orders map[string]models.Order
	mu     sync.RWMutex
}

// NewOrderRepository creates a new order repository
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]models.Order),
	}
}

// Create stores a new order
func (r *OrderRepository) Create(order models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return errors.New("order already exists")
	}

	r.orders[order.ID] = order
	return nil
}

// GetByID returns an order by ID
func (r *OrderRepository) GetByID(id string) (models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return models.Order{}, errors.New("order not found")
	}
	return order, nil
}

// GetAll returns all orders
func (r *OrderRepository) GetAll() []models.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := make([]models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	return orders
}
