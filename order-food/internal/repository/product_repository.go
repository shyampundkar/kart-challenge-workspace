package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
)

// ProductRepository handles product data operations
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository with an existing database connection
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// NewProductRepositoryWithConnection creates a new product repository and establishes a connection
func NewProductRepositoryWithConnection() *ProductRepository {
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &ProductRepository{
		db: db,
	}
}

// connectDB establishes a connection to PostgreSQL
func connectDB() (*sql.DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "orderfood")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection with retries
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		if err := db.PingContext(ctx); err == nil {
			log.Println("Successfully connected to products database")
			return db, nil
		}
		log.Printf("Waiting for database connection... (attempt %d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after retries")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetAll returns all products
func (r *ProductRepository) GetAll() []models.Product {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, price, category FROM products ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error querying products: %v", err)
		return []models.Product{}
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Category); err != nil {
			log.Printf("Error scanning product: %v", err)
			continue
		}
		products = append(products, product)
	}

	return products
}

// GetAllPaginated returns paginated products with total count
func (r *ProductRepository) GetAllPaginated(limit, offset int) ([]models.Product, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM products`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	// Get paginated results
	query := `SELECT id, name, price, category FROM products ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Category); err != nil {
			log.Printf("Error scanning product: %v", err)
			continue
		}
		products = append(products, product)
	}

	return products, total, nil
}

// GetByID returns a product by ID
func (r *ProductRepository) GetByID(id string) (models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, price, category FROM products WHERE id = $1`
	var product models.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Category,
	)

	if err == sql.ErrNoRows {
		return models.Product{}, errors.New("product not found")
	}
	if err != nil {
		return models.Product{}, fmt.Errorf("error querying product: %w", err)
	}

	return product, nil
}

// GetByIDs returns multiple products by their IDs
func (r *ProductRepository) GetByIDs(ids []string) ([]models.Product, error) {
	if len(ids) == 0 {
		return []models.Product{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Build query with placeholders
	query := `SELECT id, name, price, category FROM products WHERE id = ANY($1)`

	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0, len(ids))
	foundIDs := make(map[string]bool)

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Category); err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, product)
		foundIDs[product.ID] = true
	}

	// Check if all requested IDs were found
	for _, id := range ids {
		if !foundIDs[id] {
			return nil, errors.New("product not found: " + id)
		}
	}

	return products, nil
}
