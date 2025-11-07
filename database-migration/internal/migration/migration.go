package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Config holds database connection configuration
type Config struct {
	Host           string
	Port           string
	User           string
	Password       string
	DBName         string
	SSLMode        string
	MigrationsPath string // Path to migration files
}

// Migrator handles database migrations using golang-migrate
type Migrator struct {
	db      *sql.DB
	migrate *migrate.Migrate
	config  Config
}

// NewMigrator creates a new Migrator instance with golang-migrate
func NewMigrator(config Config) (*Migrator, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to PostgreSQL database: %s", config.DBName)

	// Create postgres driver instance for golang-migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Set default migrations path if not provided
	if config.MigrationsPath == "" {
		config.MigrationsPath = "file://migrations"
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		config.MigrationsPath,
		config.DBName,
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	log.Printf("Migrations loaded from: %s", config.MigrationsPath)

	return &Migrator{
		db:      db,
		migrate: m,
		config:  config,
	}, nil
}

// Close closes the database connection and migrate instance
func (m *Migrator) Close() error {
	if m.migrate != nil {
		srcErr, dbErr := m.migrate.Close()
		if srcErr != nil {
			return fmt.Errorf("failed to close migrate source: %w", srcErr)
		}
		if dbErr != nil {
			return fmt.Errorf("failed to close migrate database: %w", dbErr)
		}
	}
	return nil
}

// Run executes all pending migrations (up)
func (m *Migrator) Run(ctx context.Context) error {
	log.Println("Starting database migrations...")

	// Get current version
	version, dirty, err := m.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("No migrations have been applied yet")
	} else {
		log.Printf("Current migration version: %d (dirty: %v)", version, dirty)
	}

	// Run all pending migrations
	err = m.migrate.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("✓ Database is already up to date")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	// Get new version
	newVersion, _, err := m.migrate.Version()
	if err != nil {
		return fmt.Errorf("failed to get new version: %w", err)
	}

	log.Printf("✓ All migrations completed successfully. Current version: %d", newVersion)
	return nil
}

// Down rolls back one migration
func (m *Migrator) Down(ctx context.Context) error {
	log.Println("Rolling back last migration...")

	version, dirty, err := m.migrate.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			log.Println("No migrations to roll back")
			return nil
		}
		return fmt.Errorf("failed to get current version: %w", err)
	}

	log.Printf("Current version: %d (dirty: %v)", version, dirty)

	err = m.migrate.Steps(-1)
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to roll back")
			return nil
		}
		return fmt.Errorf("rollback failed: %w", err)
	}

	newVersion, _, err := m.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get new version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("✓ Rolled back to initial state (no migrations applied)")
	} else {
		log.Printf("✓ Rolled back to version: %d", newVersion)
	}

	return nil
}

// MigrateToVersion migrates to a specific version
func (m *Migrator) MigrateToVersion(ctx context.Context, targetVersion uint) error {
	log.Printf("Migrating to version: %d", targetVersion)

	err := m.migrate.Migrate(targetVersion)
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Printf("Already at version %d", targetVersion)
			return nil
		}
		return fmt.Errorf("failed to migrate to version %d: %w", targetVersion, err)
	}

	log.Printf("✓ Successfully migrated to version: %d", targetVersion)
	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (version uint, dirty bool, err error) {
	return m.migrate.Version()
}

// Force forces the migration version (use with caution!)
func (m *Migrator) Force(version int) error {
	log.Printf("⚠️  Forcing migration version to: %d", version)
	return m.migrate.Force(version)
}
