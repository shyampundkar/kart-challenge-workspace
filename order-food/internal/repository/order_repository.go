package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
)

// OrderRepository handles order data operations
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new order repository connected to PostgreSQL
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// Create stores a new order
func (r *OrderRepository) Create(order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `INSERT INTO orders (id, coupon_code, created_at, updated_at)
	               VALUES ($1, $2, NOW(), NOW())`
	_, err = tx.ExecContext(ctx, orderQuery, order.ID, order.CouponCode)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, created_at)
	              VALUES ($1, $2, $3, NOW())`
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID returns an order by ID
func (r *OrderRepository) GetByID(id string) (models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get order details
	orderQuery := `SELECT id, coupon_code FROM orders WHERE id = $1`
	var order models.Order
	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(&order.ID, &order.CouponCode)
	if err == sql.ErrNoRows {
		return models.Order{}, errors.New("order not found")
	}
	if err != nil {
		return models.Order{}, fmt.Errorf("error querying order: %w", err)
	}

	// Get order items with product details
	itemsQuery := `
		SELECT oi.product_id, oi.quantity, p.id, p.name, p.price, p.category
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
		ORDER BY oi.id`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return models.Order{}, fmt.Errorf("error querying order items: %w", err)
	}
	defer rows.Close()

	order.Items = make([]models.OrderItem, 0)
	order.Products = make([]models.Product, 0)

	for rows.Next() {
		var item models.OrderItem
		var product models.Product

		err := rows.Scan(
			&item.ProductID, &item.Quantity,
			&product.ID, &product.Name, &product.Price, &product.Category,
		)
		if err != nil {
			return models.Order{}, fmt.Errorf("error scanning order item: %w", err)
		}

		order.Items = append(order.Items, item)
		order.Products = append(order.Products, product)
	}

	return order, nil
}

// GetAll returns all orders with pagination
func (r *OrderRepository) GetAll(limit, offset int) ([]models.Order, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM orders`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		log.Printf("Error counting orders: %v", err)
		return nil, 0, fmt.Errorf("error counting orders: %w", err)
	}

	// Get paginated orders
	ordersQuery := `SELECT id, coupon_code FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, ordersQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying orders: %w", err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	orderIDs := make([]string, 0)

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CouponCode); err != nil {
			log.Printf("Error scanning order: %v", err)
			continue
		}
		orders = append(orders, order)
		orderIDs = append(orderIDs, order.ID)
	}

	// If no orders found, return empty list
	if len(orders) == 0 {
		return orders, total, nil
	}

	// Get all order items and products for these orders with a single query
	itemsQuery := `
		SELECT oi.order_id, oi.product_id, oi.quantity, p.id, p.name, p.price, p.category
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = ANY($1)
		ORDER BY oi.order_id, oi.id`

	itemRows, err := r.db.QueryContext(ctx, itemsQuery, pq.Array(orderIDs))
	if err != nil {
		log.Printf("Error querying order items: %v", err)
		return orders, total, nil
	}
	defer itemRows.Close()

	// Map to store items and products for each order
	orderItemsMap := make(map[string][]models.OrderItem)
	orderProductsMap := make(map[string][]models.Product)

	for itemRows.Next() {
		var orderID string
		var item models.OrderItem
		var product models.Product

		err := itemRows.Scan(
			&orderID, &item.ProductID, &item.Quantity,
			&product.ID, &product.Name, &product.Price, &product.Category,
		)
		if err != nil {
			log.Printf("Error scanning order item: %v", err)
			continue
		}

		orderItemsMap[orderID] = append(orderItemsMap[orderID], item)
		orderProductsMap[orderID] = append(orderProductsMap[orderID], product)
	}

	// Populate items and products for each order
	for i := range orders {
		orders[i].Items = orderItemsMap[orders[i].ID]
		orders[i].Products = orderProductsMap[orders[i].ID]
	}

	return orders, total, nil
}
