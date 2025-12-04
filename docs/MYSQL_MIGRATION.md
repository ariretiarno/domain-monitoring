# MySQL Migration Guide

## Adding MySQL Support

The application currently uses SQLite, but you can add MySQL support with minimal changes.

### Step 1: Add MySQL Driver

```bash
go get github.com/go-sql-driver/mysql
```

### Step 2: Update `internal/repository/db.go`

Add MySQL support alongside SQLite:

```go
import (
    _ "github.com/go-sql-driver/mysql"  // Add this
    _ "github.com/mattn/go-sqlite3"
)

// NewDB creates a new database connection
// For SQLite: dbPath = "dem.db"
// For MySQL: dbPath = "user:password@tcp(localhost:3306)/dbname"
func NewDB(dbPath string, driver string) (*DB, error) {
    if driver == "" {
        driver = "sqlite3"  // Default to SQLite
    }
    
    db, err := sqlx.Connect(driver, dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    wrapper := &DB{DB: db}

    // Run migrations
    if err := wrapper.Migrate(); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }

    return wrapper, nil
}
```

### Step 3: Update Schema for MySQL Compatibility

Modify `internal/repository/schema.go`:

```go
const schemaSQLite = `
-- Current SQLite schema
`

const schemaMySQL = `
CREATE TABLE IF NOT EXISTS domains (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    expiration_date DATETIME NOT NULL,
    nameservers JSON NOT NULL,
    registrant TEXT NOT NULL,
    registrar VARCHAR(255) NOT NULL,
    last_checked DATETIME NOT NULL,
    next_check DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_domains_name (name),
    INDEX idx_domains_expiration_date (expiration_date),
    INDEX idx_domains_next_check (next_check)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY,
    monitoring_interval BIGINT NOT NULL,
    alert_thresholds JSON NOT NULL,
    google_chat_webhook TEXT NOT NULL,
    retention_period BIGINT NOT NULL,
    updated_at DATETIME NOT NULL,
    CHECK (id = 1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(255) PRIMARY KEY,
    domain_id VARCHAR(255) NOT NULL,
    domain_name VARCHAR(255) NOT NULL,
    threshold BIGINT NOT NULL,
    expiration_date DATETIME NOT NULL,
    sent_at DATETIME NOT NULL,
    success TINYINT(1) NOT NULL,
    error_message TEXT NOT NULL,
    INDEX idx_alerts_domain_id (domain_id),
    INDEX idx_alerts_sent_at (sent_at),
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`

func (db *DB) Migrate() error {
    schema := schemaSQLite
    if db.DriverName() == "mysql" {
        schema = schemaMySQL
    }
    
    _, err := db.Exec(schema)
    if err != nil {
        return fmt.Errorf("failed to execute schema: %w", err)
    }
    return nil
}
```

### Step 4: Update main.go

```go
func main() {
    log.Println("Domain Expiration Monitor starting...")

    // Get database configuration from environment
    dbDriver := getEnv("DB_DRIVER", "sqlite3")  // "sqlite3" or "mysql"
    dbPath := getEnv("DB_PATH", "dem.db")       // For SQLite
    // For MySQL: DB_PATH="user:password@tcp(localhost:3306)/dem?parseTime=true"
    
    db, err := repository.NewDB(dbPath, dbDriver)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()
    
    // ... rest of the code
}
```

### Step 5: Environment Variables

Create a `.env` file or set environment variables:

```bash
# For SQLite (default)
DB_DRIVER=sqlite3
DB_PATH=dem.db

# For MySQL
DB_DRIVER=mysql
DB_PATH="user:password@tcp(localhost:3306)/dem?parseTime=true"
HTTP_ADDR=:8080
```

### Step 6: MySQL Setup

```sql
-- Create database
CREATE DATABASE dem CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER 'dem_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON dem.* TO 'dem_user'@'localhost';
FLUSH PRIVILEGES;
```

### Step 7: Run with MySQL

```bash
# Set environment variables
export DB_DRIVER=mysql
export DB_PATH="dem_user:your_password@tcp(localhost:3306)/dem?parseTime=true"

# Run the application
./bin/dem
```

## Benefits of MySQL

- **Scalability**: Better for high-traffic scenarios
- **Replication**: Built-in master-slave replication
- **Concurrent Access**: Better handling of concurrent writes
- **Remote Access**: Can run database on separate server
- **Backup Tools**: More mature backup/restore ecosystem

## When to Use Each

**Use SQLite when:**
- Single server deployment
- Low to medium traffic
- Simple deployment (single binary)
- No remote database access needed

**Use MySQL when:**
- Multiple application instances
- High traffic/concurrent users
- Need database replication
- Separate database server
- Enterprise environment
