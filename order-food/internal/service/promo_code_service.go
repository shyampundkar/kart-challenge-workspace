package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PromoCodeService handles promo code validation
type PromoCodeService struct {
	db *sql.DB
}

// NewPromoCodeService creates a new promo code service
func NewPromoCodeService(db *sql.DB) *PromoCodeService {
	return &PromoCodeService{db: db}
}

// ValidatePromoCode checks if a promo code is valid
// Rules:
// 1. Must be 8-10 characters long
// 2. Must appear in at least 2 different files in the coupons table
func (s *PromoCodeService) ValidatePromoCode(code string) (bool, error) {
	// Rule 1: Check length
	if len(code) < 8 || len(code) > 10 {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Rule 2: Check if code appears in at least 2 files
	query := `
		SELECT COUNT(DISTINCT file_name)
		FROM coupons
		WHERE coupon = $1
	`

	var fileCount int
	err := s.db.QueryRowContext(ctx, query, code).Scan(&fileCount)
	if err != nil {
		return false, fmt.Errorf("failed to validate promo code: %w", err)
	}

	return fileCount >= 2, nil
}
