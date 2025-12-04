# MySQL Setup Guide

## Prerequisites

- MySQL 5.7+ or MariaDB 10.2+
- Go 1.21+

## Quick Setup

### 1. Create MySQL Database

```sql
-- Connect to MySQL
mysql -u root -p

-- Create database
CREATE DATABASE dem CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER 'dem_user'@'localhost' IDENTIFIED BY 'your_secure_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON dem.* TO 'dem_user'@'localhost';
FLUSH PRIVILEGES;

-- Exit MySQL
EXIT;
```

### 2. Configure Environment

Edit `.env` file:

```bash
# Switch to MySQL
DB_DRIVER=mysql
DB_PATH=dem_user:your_secure_password@tcp(localhost:3306)/dem?parseTime=true

# HTTP Server
HTTP_ADDR=:8080
```

### 3. Run Application

```bash
./bin/dem
```

The application will automatically create the required tables on first run.

## Remote MySQL Server

For remote MySQL server:

```bash
DB_DRIVER=mysql
DB_PATH=dem_user:password@tcp(mysql.example.com:3306)/dem?parseTime=true
```

## Docker MySQL

Quick MySQL setup with Docker:

```bash
docker run -d \
  --name dem-mysql \
  -e MYSQL_ROOT_PASSWORD=rootpass \
  -e MYSQL_DATABASE=dem \
  -e MYSQL_USER=dem_user \
  -e MYSQL_PASSWORD=your_password \
  -p 3306:3306 \
  mysql:8.0

# Wait for MySQL to start
sleep 10

# Update .env
DB_DRIVER=mysql
DB_PATH=dem_user:your_password@tcp(localhost:3306)/dem?parseTime=true

# Run application
./bin/dem
```

## Connection String Format

```
username:password@tcp(host:port)/database?parseTime=true
```

Parameters:
- `parseTime=true` - Required for proper datetime handling
- `charset=utf8mb4` - Optional, for emoji support
- `loc=Local` - Optional, for timezone handling

Example with all parameters:
```
dem_user:pass@tcp(localhost:3306)/dem?parseTime=true&charset=utf8mb4&loc=Local
```

## Troubleshooting

### Connection Refused

```bash
# Check if MySQL is running
sudo systemctl status mysql

# Or for macOS
brew services list
```

### Access Denied

```sql
-- Reset user password
ALTER USER 'dem_user'@'localhost' IDENTIFIED BY 'new_password';
FLUSH PRIVILEGES;
```

### Table Already Exists

The application uses `CREATE TABLE IF NOT EXISTS`, so it's safe to run multiple times.

## Migration from SQLite

```bash
# Export from SQLite
sqlite3 dem.db .dump > backup.sql

# Import to MySQL (requires manual conversion)
# Or use a tool like: https://github.com/techouse/sqlite3-to-mysql

# Alternatively, start fresh with MySQL and re-add domains
```

## Performance Tuning

For production MySQL:

```sql
-- Increase connection limits
SET GLOBAL max_connections = 200;

-- Enable query cache (MySQL 5.7)
SET GLOBAL query_cache_size = 67108864;
SET GLOBAL query_cache_type = 1;
```

## Backup

```bash
# Backup database
mysqldump -u dem_user -p dem > dem_backup_$(date +%Y%m%d).sql

# Restore database
mysql -u dem_user -p dem < dem_backup_20231204.sql
```
