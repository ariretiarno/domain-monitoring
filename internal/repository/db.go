package repository

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection and provides migration functionality
type DB struct {
	*sqlx.DB
	driver string
}

// NewDB creates a new database connection and runs migrations
// driver: "sqlite3" or "mysql"
// dbPath: for SQLite "dem.db", for MySQL "user:password@tcp(host:port)/dbname?parseTime=true"
func NewDB(dbPath string, driver string) (*DB, error) {
	if driver == "" {
		driver = "sqlite3"
	}
	
	db, err := sqlx.Connect(driver, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	wrapper := &DB{DB: db, driver: driver}

	// Run migrations
	if err := wrapper.Migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return wrapper, nil
}

// Driver returns the database driver name
func (db *DB) Driver() string {
	return db.driver
}

// Migrate runs the database schema migrations
func (db *DB) Migrate() error {
	schemaSQL := schema
	if db.driver == "mysql" {
		schemaSQL = schemaMySQL
	}
	
	_, err := db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping checks if the database connection is alive
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// WithTransaction executes a function within a database transaction
func (db *DB) WithTransaction(fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReconnectWithBackoff attempts to reconnect to the database with exponential backoff
func (db *DB) ReconnectWithBackoff(maxRetries int) error {
	backoff := time.Second
	
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		
		if i < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	
	return fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
}

// IsConstraintError checks if an error is a constraint violation
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite constraint errors contain "UNIQUE constraint failed" or "constraint failed"
	errStr := err.Error()
	return contains(errStr, "UNIQUE constraint failed") || 
	       contains(errStr, "constraint failed")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
