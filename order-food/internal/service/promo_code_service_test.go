package service

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPromoCodeService_ValidatePromoCode_ValidCode(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: code exists in 2 files
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("HAPPYHRS").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Test
	valid, err := service.ValidatePromoCode("HAPPYHRS")

	// Assert
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPromoCodeService_ValidatePromoCode_InvalidCode_TooShort(t *testing.T) {
	// Setup mock database
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Test with code that's too short (less than 8 characters)
	valid, err := service.ValidatePromoCode("SHORT")

	// Assert
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestPromoCodeService_ValidatePromoCode_InvalidCode_TooLong(t *testing.T) {
	// Setup mock database
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Test with code that's too long (more than 10 characters)
	valid, err := service.ValidatePromoCode("VERYLONGCODE")

	// Assert
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestPromoCodeService_ValidatePromoCode_InvalidCode_OnlyOneFile(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: code exists in only 1 file
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("ONLYONCE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Test
	valid, err := service.ValidatePromoCode("ONLYONCE")

	// Assert
	assert.NoError(t, err)
	assert.False(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPromoCodeService_ValidatePromoCode_InvalidCode_NotFound(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: code doesn't exist
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("NOTFOUND").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Test
	valid, err := service.ValidatePromoCode("NOTFOUND")

	// Assert
	assert.NoError(t, err)
	assert.False(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPromoCodeService_ValidatePromoCode_DatabaseError(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: database error
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("TESTCODE").
		WillReturnError(sql.ErrConnDone)

	// Test
	valid, err := service.ValidatePromoCode("TESTCODE")

	// Assert
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "failed to validate promo code")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPromoCodeService_ValidatePromoCode_ExactlyTwoFiles(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: code exists in exactly 2 files
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("TWOFILES").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Test
	valid, err := service.ValidatePromoCode("TWOFILES")

	// Assert
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPromoCodeService_ValidatePromoCode_MoreThanTwoFiles(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: code exists in 3 files (8 characters)
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("POPULAR1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Test
	valid, err := service.ValidatePromoCode("POPULAR1")

	// Assert
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPromoCodeService_ValidatePromoCode_MinimumLength(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: code with exactly 8 characters exists in 2 files
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("EIGHTCHR").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Test
	valid, err := service.ValidatePromoCode("EIGHTCHR")

	// Assert
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPromoCodeService_ValidatePromoCode_MaximumLength(t *testing.T) {
	// Setup mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewPromoCodeService(db)

	// Mock expectation: code with exactly 10 characters exists in 2 files
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT file_name\\)").
		WithArgs("TENCHARS10").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Test
	valid, err := service.ValidatePromoCode("TENCHARS10")

	// Assert
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.NoError(t, mock.ExpectationsWereMet())
}
