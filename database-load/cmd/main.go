package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

const (
	batchSize      = 50000 // Optimized batch size for maximum throughput
	maxConcurrency = 8     // Increased concurrency for parallel processing
)

func main() {
	log.Println("Starting database load service...")

	ctx := context.Background()

	// Get database connection string from environment
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "orderfood")

	// Connection string for sql.DB (used for products - keeping backward compatibility)
	sqlConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connection string for pgx (used for coupons with CopyFrom)
	pgxConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Connect to database using sql.DB for products
	db, err := sql.Open("postgres", sqlConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Load data
	dataDir := getEnv("DATA_DIR", "/data")

	// Load products first
	if err := loadProducts(ctx, db, filepath.Join(dataDir, "products")); err != nil {
		log.Fatalf("Failed to load products: %v", err)
	}

	// Load coupons using pgx CopyFrom
	if err := loadCouponsWithPgx(ctx, pgxConnStr, dataDir); err != nil {
		log.Fatalf("Failed to load coupons: %v", err)
	}

	// Convert coupons table to LOGGED for crash safety
	if err := convertToLoggedTable(ctx, pgxConnStr); err != nil {
		log.Printf("Warning: Failed to convert table to LOGGED: %v", err)
	}

	log.Println("Database load completed successfully")
}

func loadProducts(ctx context.Context, db *sql.DB, productsDir string) error {
	log.Println("Loading products from CSV files...")

	// Find all .csv files in the products directory
	files, err := filepath.Glob(filepath.Join(productsDir, "*.csv"))
	if err != nil {
		return fmt.Errorf("failed to list product files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("No .csv files found in %s, skipping product load", productsDir)
		return nil
	}

	totalProducts := 0

	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		log.Printf("Processing product file: %s", fileName)

		count, err := loadProductsFromFile(ctx, db, filePath)
		if err != nil {
			return fmt.Errorf("failed to load products from %s: %w", fileName, err)
		}

		totalProducts += count
		log.Printf("✓ Loaded %d products from %s", count, fileName)
	}

	log.Printf("✓ Total products loaded: %d", totalProducts)
	return nil
}

func loadProductsFromFile(ctx context.Context, db *sql.DB, filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	_, err = reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV records: %w", err)
	}

	count := 0
	for _, record := range records {
		if len(record) < 4 {
			log.Printf("Warning: Skipping invalid product record: %v", record)
			continue
		}

		id := strings.TrimSpace(record[0])
		name := strings.TrimSpace(record[1])
		priceStr := strings.TrimSpace(record[2])
		category := strings.TrimSpace(record[3])

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Printf("Warning: Invalid price '%s' for product '%s': %v", priceStr, name, err)
			continue
		}

		// Insert product
		query := `INSERT INTO products (id, name, price, category, created_at, updated_at)
		          VALUES ($1, $2, $3, $4, NOW(), NOW())
		          ON CONFLICT (id) DO UPDATE
		          SET name = EXCLUDED.name,
		              price = EXCLUDED.price,
		              category = EXCLUDED.category,
		              updated_at = NOW()`

		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err = db.ExecContext(ctxTimeout, query, id, name, price, category)
		cancel()

		if err != nil {
			return count, fmt.Errorf("failed to insert product '%s': %w", name, err)
		}

		count++
	}

	return count, nil
}

// Coupon represents a coupon record for batch processing
type Coupon struct {
	Code     string
	FileName string
}

func loadCouponsWithPgx(ctx context.Context, connStr, dataDir string) error {
	log.Println("Loading coupons from text files using pgx CopyFrom...")

	// Find all .txt files in the data directory
	files, err := filepath.Glob(filepath.Join(dataDir, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("No .txt files found in %s, skipping coupon load", dataDir)
		return nil
	}

	log.Printf("Found %d files to process", len(files))

	// Optimize PostgreSQL for bulk loading
	if err := optimizePostgresForBulkLoad(ctx, connStr); err != nil {
		log.Printf("Warning: Failed to optimize PostgreSQL settings: %v", err)
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var totalCoupons atomic.Int64
	errChan := make(chan error, len(files))

	// Process files concurrently with limited concurrency
	for _, filePath := range files {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(fp string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			fileName := filepath.Base(fp)
			log.Printf("Processing file: %s", fileName)

			count, err := loadCouponsFromFileWithPgx(ctx, connStr, fp, fileName)
			if err != nil {
				errChan <- fmt.Errorf("failed to load coupons from %s: %w", fileName, err)
				return
			}

			totalCoupons.Add(int64(count))
			log.Printf("✓ Loaded %d coupons from %s", count, fileName)
		}(filePath)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		return <-errChan
	}

	log.Printf("✓ Total coupons loaded: %d", totalCoupons.Load())
	return nil
}

func loadCouponsFromFileWithPgx(ctx context.Context, connStr, filePath, fileName string) (int, error) {
	// Connect to database using pgx
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Set a larger buffer for scanner (default is 64KB, increase to 1MB)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	var batch []Coupon
	totalCount := 0

	for scanner.Scan() {
		coupon := strings.TrimSpace(scanner.Text())
		if coupon == "" {
			continue // Skip empty lines
		}

		batch = append(batch, Coupon{
			Code:     coupon,
			FileName: fileName,
		})

		// Insert batch when it reaches batchSize
		if len(batch) >= batchSize {
			count, err := insertCouponsBatchWithCopyFrom(ctx, conn, batch)
			if err != nil {
				return totalCount, fmt.Errorf("failed to insert batch: %w", err)
			}
			totalCount += count
			batch = batch[:0] // Reset slice

			// Log progress every 50k coupons
			if totalCount%50000 == 0 {
				log.Printf("  Progress: %d coupons inserted from %s", totalCount, fileName)
			}
		}
	}

	// Insert remaining coupons
	if len(batch) > 0 {
		count, err := insertCouponsBatchWithCopyFrom(ctx, conn, batch)
		if err != nil {
			return totalCount, fmt.Errorf("failed to insert final batch: %w", err)
		}
		totalCount += count
	}

	if err := scanner.Err(); err != nil {
		return totalCount, fmt.Errorf("error reading file: %w", err)
	}

	return totalCount, nil
}

func insertCouponsBatchWithCopyFrom(ctx context.Context, conn *pgx.Conn, coupons []Coupon) (int, error) {
	if len(coupons) == 0 {
		return 0, nil
	}

	// Use CopyFrom directly to the coupons table for maximum performance
	// This is much faster than using a temp table
	rows := make([][]interface{}, len(coupons))
	for i, c := range coupons {
		rows[i] = []interface{}{c.Code, c.FileName}
	}

	copyCount, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"coupons"},
		[]string{"coupon", "file_name"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		// If error is due to duplicate key, that's expected - log and continue
		if strings.Contains(err.Error(), "duplicate key") {
			log.Printf("Warning: Duplicate keys found in batch, some rows skipped")
			return int(copyCount), nil
		}
		return 0, fmt.Errorf("failed to copy data: %w", err)
	}

	return int(copyCount), nil
}

// optimizePostgresForBulkLoad sets PostgreSQL parameters for optimal bulk loading performance
func optimizePostgresForBulkLoad(ctx context.Context, connStr string) error {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	optimizations := []string{
		"SET synchronous_commit = OFF",           // Faster commits, acceptable for bulk load
		"SET maintenance_work_mem = '1GB'",       // More memory for index maintenance
		"SET checkpoint_timeout = '30min'",       // Less frequent checkpoints
		"SET max_wal_size = '4GB'",               // Allow more WAL before checkpoint
		"SET wal_buffers = '16MB'",               // Larger WAL buffers
		"SET effective_cache_size = '2GB'",       // Hint about available cache
	}

	for _, sql := range optimizations {
		if _, err := conn.Exec(ctx, sql); err != nil {
			log.Printf("Warning: Failed to set optimization '%s': %v", sql, err)
		}
	}

	log.Println("PostgreSQL optimized for bulk loading")
	return nil
}

// convertToLoggedTable converts the UNLOGGED coupons table to a regular logged table
// This should be called after bulk loading is complete
func convertToLoggedTable(ctx context.Context, connStr string) error {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	log.Println("Converting coupons table from UNLOGGED to LOGGED for crash safety...")
	_, err = conn.Exec(ctx, "ALTER TABLE coupons SET LOGGED")
	if err != nil {
		return fmt.Errorf("failed to convert table to logged: %w", err)
	}

	log.Println("✓ Coupons table converted to LOGGED (crash-safe)")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
