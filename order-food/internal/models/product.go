package models

// Product represents a product available for order
type Product struct {
	ID       string  `json:"id" binding:"required"`
	Name     string  `json:"name" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
	Category string  `json:"category" binding:"required"`
}
