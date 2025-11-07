package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

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

	// Serialize items and products to JSON
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	productsJSON, err := json.Marshal(order.Products)
	if err != nil {
		return fmt.Errorf("failed to marshal products: %w", err)
	}

	query := `INSERT INTO orders (id, items, products, created_at, updated_at)
	          VALUES ($1, $2, $3, NOW(), NOW())`

	_, err = r.db.ExecContext(ctx, query, order.ID, itemsJSON, productsJSON)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}

// GetByID returns an order by ID
func (r *OrderRepository) GetByID(id string) (models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, items, products FROM orders WHERE id = $1`

	var order models.Order
	var itemsJSON, productsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(&order.ID, &itemsJSON, &productsJSON)
	if err == sql.ErrNoRows {
		return models.Order{}, errors.New("order not found")
	}
	if err != nil {
		return models.Order{}, fmt.Errorf("error querying order: %w", err)
	}

	// Deserialize JSON fields
	if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
		return models.Order{}, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	if err := json.Unmarshal(productsJSON, &order.Products); err != nil {
		return models.Order{}, fmt.Errorf("failed to unmarshal products: %w", err)
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

	// Get paginated results
	query := `SELECT id, items, products FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying orders: %w", err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var order models.Order
		var itemsJSON, productsJSON []byte

		if err := rows.Scan(&order.ID, &itemsJSON, &productsJSON); err != nil {
			log.Printf("Error scanning order: %v", err)
			continue
		}

		// Deserialize JSON fields
		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			log.Printf("Error unmarshaling items: %v", err)
			continue
		}

		if err := json.Unmarshal(productsJSON, &order.Products); err != nil {
			log.Printf("Error unmarshaling products: %v", err)
			continue
		}

		orders = append(orders, order)
	}

	return orders, total, nil
}
