# Docker Setup Complete! 🎉

Your Domain Expiration Monitor now has complete Docker support with automatic MySQL migrations.

## What Was Created

### Docker Configuration
- ✅ **Dockerfile** - Multi-stage build optimized for Go
- ✅ **docker-compose.yml** - Complete stack with MySQL 8.0
- ✅ **.dockerignore** - Optimized build context
- ✅ **Makefile** - Convenient commands for common tasks

### Scripts
- ✅ **docker-start.sh** - Easy startup with validation
- ✅ **verify-migration.sh** - Database migration verification
- ✅ **test_alert.sh** - Updated to work immediately

### Documentation
- ✅ **DOCKER_DEPLOYMENT.md** - Complete deployment guide
- ✅ **DOCKER_QUICKREF.md** - Quick command reference
- ✅ **DEPLOYMENT_SUMMARY.md** - Overview and checklist
- ✅ **README.md** - Updated with Docker instructions
- ✅ **.env.example** - All environment variables

### Code Updates
- ✅ **cmd/dem/main.go** - MySQL connection from environment variables
- ✅ **internal/scheduler/scheduler.go** - Alert evaluation even when WHOIS fails
- ✅ **internal/repository/schema.go** - Already had migrations (no changes needed)
- ✅ **internal/repository/db.go** - Already had auto-migration (no changes needed)

## How It Works

### Automatic Migrations

When the application starts:

1. **Waits for MySQL** - Health check ensures MySQL is ready
2. **Connects to database** - Uses environment variables
3. **Runs migrations** - Creates tables if they don't exist
4. **Starts application** - Ready to use!

The migration code in `internal/repository/db.go` and `schema.go` handles:
- Creating tables with proper schema
- Creating indexes for performance
- Supporting both SQLite and MySQL
- Idempotent operations (safe to run multiple times)

### Docker Compose Stack

```yaml
services:
  mysql:
    - MySQL 8.0 database
    - Persistent volume for data
    - Health checks
    - Exposed on port 3306
  
  app:
    - Go application
    - Waits for MySQL health check
    - Auto-runs migrations
    - Exposed on port 8080
    - Health checks
```

## Quick Start

### 1. Configure Environment

```bash
# Copy example
cp .env.example .env

# Edit with your settings
nano .env
```

Minimum required:
```bash
MYSQL_ROOT_PASSWORD=your-secure-password
MYSQL_PASSWORD=your-secure-password
GOOGLE_CHAT_WEBHOOK=https://chat.googleapis.com/...
```

### 2. Start Services

```bash
# Easy way
./docker-start.sh

# Or with make
make docker-up

# Or directly
docker-compose up -d
```

### 3. Verify

```bash
# Check services
docker-compose ps

# Check migrations
./verify-migration.sh

# Check health
curl http://localhost:8080/health
```

### 4. Use Application

Open http://localhost:8080

## Common Tasks

### Using Make (Recommended)

```bash
# See all commands
make help

# Start Docker
make docker-up

# View logs
make docker-logs

# Restart
make docker-restart

# Test alerts
make test-alert

# Backup database
make backup-db

# Stop everything
make docker-down
```

### Using Docker Compose

```bash
# Start
docker-compose up -d

# Logs
docker-compose logs -f

# Restart
docker-compose restart

# Stop
docker-compose down
```

### Using Scripts

```bash
# Start everything
./docker-start.sh

# Verify migrations
./verify-migration.sh

# Test alerts
./test_alert.sh
```

## Testing Alerts

The alert system now works correctly with test domains:

```bash
# 1. Run test script
./test_alert.sh

# 2. Restart app (triggers immediate check)
docker-compose restart app

# 3. Check UI for alert history
open http://localhost:8080
```

The fix: Scheduler now evaluates alerts even when WHOIS queries fail, which is perfect for test domains that don't exist.

## Database Access

### MySQL Client

```bash
# Connect to MySQL
docker-compose exec mysql mysql -u demuser -p dem

# Run query
docker-compose exec mysql mysql -u demuser -p dem -e "SELECT * FROM domains;"
```

### Backup & Restore

```bash
# Backup
docker-compose exec mysql mysqldump -u demuser -p dem > backup.sql

# Restore
docker-compose exec -T mysql mysql -u demuser -p dem < backup.sql
```

## Migration Details

### Tables Created

**domains**
- Stores domain information
- Tracks monitoring status
- Records WHOIS data

**config**
- Application settings
- Alert thresholds
- Webhook configuration

**alerts**
- Alert history
- Success/failure tracking
- Error messages

### Schema Differences

The code automatically uses the correct schema:
- **SQLite**: Uses TEXT and INTEGER types
- **MySQL**: Uses VARCHAR, JSON, DATETIME types

Both schemas are in `internal/repository/schema.go`.

## Production Deployment

### Security Checklist

- [ ] Change all default passwords
- [ ] Use strong passwords (32+ characters)
- [ ] Don't expose MySQL port publicly
- [ ] Use HTTPS for webhook URLs
- [ ] Set up firewall rules
- [ ] Enable MySQL SSL/TLS
- [ ] Regular security updates

### Monitoring Checklist

- [ ] Set up log aggregation
- [ ] Configure alerting for failures
- [ ] Monitor disk space
- [ ] Monitor MySQL performance
- [ ] Track alert success rate
- [ ] Set up backup automation

### Backup Strategy

```bash
# Daily backup cron job
0 2 * * * cd /path/to/app && docker-compose exec mysql mysqldump -u demuser -p$DB_PASSWORD dem | gzip > /backups/dem-$(date +\%Y\%m\%d).sql.gz

# Keep 30 days
find /backups -name "dem-*.sql.gz" -mtime +30 -delete
```

## Documentation

- **DOCKER_DEPLOYMENT.md** - Complete guide with troubleshooting
- **DOCKER_QUICKREF.md** - Quick command reference
- **DEPLOYMENT_SUMMARY.md** - Overview and architecture
- **MYSQL_SETUP.md** - MySQL-specific setup
- **ALERT_TESTING_GUIDE.md** - How to test alerts
- **README.md** - General usage

## Next Steps

1. **Configure**: Edit `.env` with your settings
2. **Start**: Run `make docker-up` or `./docker-start.sh`
3. **Verify**: Run `./verify-migration.sh`
4. **Configure Webhook**: Visit http://localhost:8080/config
5. **Add Domains**: Start monitoring domains
6. **Test Alerts**: Run `./test_alert.sh`

## Support

If you encounter issues:

1. Check logs: `docker-compose logs -f`
2. Verify migrations: `./verify-migration.sh`
3. Check health: `curl http://localhost:8080/health`
4. Review documentation in the files above

## Summary

You now have:
- ✅ Complete Docker setup with MySQL
- ✅ Automatic database migrations
- ✅ Production-ready configuration
- ✅ Comprehensive documentation
- ✅ Easy-to-use scripts and Makefile
- ✅ Working alert system with test support

Everything is ready to deploy! 🚀
