package repository

import (
	"errors"
	"sync"

	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
)

// ProductRepository handles product data operations
type ProductRepository struct {
	products map[string]models.Product
	mu       sync.RWMutex
}

// NewProductRepository creates a new product repository with sample data
func NewProductRepository() *ProductRepository {
	repo := &ProductRepository{
		products: make(map[string]models.Product),
	}
	repo.seedData()
	return repo
}

// seedData populates the repository with sample products
func (r *ProductRepository) seedData() {
	sampleProducts := []models.Product{
		{ID: "1", Name: "Chicken Waffle", Price: 12.99, Category: "Waffle"},
		{ID: "2", Name: "Belgian Waffle", Price: 10.99, Category: "Waffle"},
		{ID: "3", Name: "Blueberry Pancakes", Price: 9.99, Category: "Pancakes"},
		{ID: "4", Name: "Chocolate Pancakes", Price: 11.99, Category: "Pancakes"},
		{ID: "5", Name: "Caesar Salad", Price: 8.99, Category: "Salad"},
		{ID: "6", Name: "Greek Salad", Price: 9.49, Category: "Salad"},
		{ID: "7", Name: "Margherita Pizza", Price: 13.99, Category: "Pizza"},
		{ID: "8", Name: "Pepperoni Pizza", Price: 15.99, Category: "Pizza"},
		{ID: "9", Name: "Cheeseburger", Price: 11.49, Category: "Burger"},
		{ID: "10", Name: "Veggie Burger", Price: 10.49, Category: "Burger"},
	}

	for _, product := range sampleProducts {
		r.products[product.ID] = product
	}
}

// GetAll returns all products
func (r *ProductRepository) GetAll() []models.Product {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]models.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}
	return products
}

// GetByID returns a product by ID
func (r *ProductRepository) GetByID(id string) (models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return models.Product{}, errors.New("product not found")
	}
	return product, nil
}

// GetByIDs returns multiple products by their IDs
func (r *ProductRepository) GetByIDs(ids []string) ([]models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]models.Product, 0, len(ids))
	for _, id := range ids {
		product, exists := r.products[id]
		if !exists {
			return nil, errors.New("product not found: " + id)
		}
		products = append(products, product)
	}
	return products, nil
}
