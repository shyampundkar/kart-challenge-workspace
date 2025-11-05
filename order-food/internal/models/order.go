package models

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// OrderReq represents a request to create a new order
type OrderReq struct {
	CouponCode string      `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items" binding:"required,min=1,dive"`
}

// Order represents a completed order
type Order struct {
	ID       string     `json:"id"`
	Items    []OrderItem `json:"items"`
	Products []Product   `json:"products"`
}
