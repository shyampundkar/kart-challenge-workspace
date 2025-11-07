package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

const (
	batchSize = 1000 // Insert in batches for better performance
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

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Load coupon data
	dataDir := getEnv("DATA_DIR", "/data")

	if err := loadCoupons(ctx, db, dataDir); err != nil {
		log.Fatalf("Failed to load coupons: %v", err)
	}

	log.Println("Database load completed successfully")
}

func loadCoupons(ctx context.Context, db *sql.DB, dataDir string) error {
	log.Println("Loading coupons from text files...")

	// Find all .txt files in the data directory
	files, err := filepath.Glob(filepath.Join(dataDir, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("No .txt files found in %s, skipping coupon load", dataDir)
		return nil
	}

	log.Printf("Found %d files to process concurrently", len(files))

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	var totalCoupons atomic.Int64
	errChan := make(chan error, len(files))

	// Process files concurrently
	for _, filePath := range files {
		wg.Add(1)
		go func(fp string) {
			defer wg.Done()

			fileName := filepath.Base(fp)
			log.Printf("Processing file: %s", fileName)

			count, err := loadCouponsFromFile(ctx, db, fp, fileName)
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

func loadCouponsFromFile(ctx context.Context, db *sql.DB, filePath, fileName string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Set a larger buffer for scanner (default is 64KB, increase to 1MB)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	var coupons []string
	count := 0

	for scanner.Scan() {
		coupon := strings.TrimSpace(scanner.Text())
		if coupon == "" {
			continue // Skip empty lines
		}

		coupons = append(coupons, coupon)

		// Insert in batches
		if len(coupons) >= batchSize {
			if err := insertCouponsBatch(ctx, db, coupons, fileName); err != nil {
				return count, err
			}
			count += len(coupons)
			coupons = coupons[:0] // Reset slice

			// Log progress every 10k coupons
			if count%10000 == 0 {
				log.Printf("  Progress: %d coupons inserted from %s", count, fileName)
			}
		}
	}

	// Insert remaining coupons
	if len(coupons) > 0 {
		if err := insertCouponsBatch(ctx, db, coupons, fileName); err != nil {
			return count, err
		}
		count += len(coupons)
	}

	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("error reading file: %w", err)
	}

	return count, nil
}

func insertCouponsBatch(ctx context.Context, db *sql.DB, coupons []string, fileName string) error {
	if len(coupons) == 0 {
		return nil
	}

	// Build bulk insert query
	valueStrings := make([]string, 0, len(coupons))
	valueArgs := make([]interface{}, 0, len(coupons)*2)
	argPos := 1

	for _, coupon := range coupons {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", argPos, argPos+1))
		valueArgs = append(valueArgs, coupon, fileName)
		argPos += 2
	}

	query := fmt.Sprintf("INSERT INTO coupons (coupon, file_name) VALUES %s ON CONFLICT DO NOTHING",
		strings.Join(valueStrings, ","))

	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctxTimeout, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert coupons batch: %w", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
