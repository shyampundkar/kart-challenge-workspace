package service

import (
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/repository"
)

// ProductService handles product business logic
type ProductService struct {
	repo *repository.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// ListProducts returns all available products
func (s *ProductService) ListProducts() []models.Product {
	return s.repo.GetAll()
}

// GetProduct returns a single product by ID
func (s *ProductService) GetProduct(id string) (models.Product, error) {
	return s.repo.GetByID(id)
}
