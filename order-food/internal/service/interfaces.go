package service

import "github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"

// ProductServiceInterface defines the interface for product operations
type ProductServiceInterface interface {
	ListProducts() []models.Product
	ListProductsPaginated(limit, offset int) ([]models.Product, int, error)
	GetProduct(id string) (models.Product, error)
}

// OrderServiceInterface defines the interface for order operations
type OrderServiceInterface interface {
	CreateOrder(req models.OrderReq) (models.Order, error)
	GetOrder(id string) (models.Order, error)
	ListOrdersPaginated(limit, offset int) ([]models.Order, int, error)
}

// PromoCodeServiceInterface defines the interface for promo code operations
type PromoCodeServiceInterface interface {
	ValidatePromoCode(code string) (bool, error)
}
