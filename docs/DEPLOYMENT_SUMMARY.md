# Deployment Summary

This document provides a quick overview of the deployment options for the Domain Expiration Monitor.

## What's Included

### Docker Files
- ✅ `Dockerfile` - Multi-stage build for Go application
- ✅ `docker-compose.yml` - Complete stack with MySQL
- ✅ `.dockerignore` - Optimized build context
- ✅ `docker-start.sh` - Easy startup script
- ✅ `verify-migration.sh` - Database migration verification

### Documentation
- ✅ `DOCKER_DEPLOYMENT.md` - Complete Docker deployment guide
- ✅ `DOCKER_QUICKREF.md` - Quick reference for Docker commands
- ✅ `README.md` - Updated with Docker instructions
- ✅ `.env.example` - Updated with all environment variables

### Features
- ✅ Automatic database migrations on startup
- ✅ MySQL 8.0 with persistent storage
- ✅ Health checks for both services
- ✅ Isolated network for security
- ✅ Environment-based configuration
- ✅ Production-ready setup

## Quick Start

### 1. Docker Deployment (Recommended)

```bash
# Clone repository
git clone <your-repo>
cd whois-monitoring

# Start services
./docker-start.sh

# Access application
open http://localhost:8080
```

### 2. Local Development

```bash
# Build
go build -o bin/dem ./cmd/dem

# Run
./bin/dem
```

## Architecture

```
┌─────────────────────────────────────────┐
│         Docker Compose Stack            │
├─────────────────────────────────────────┤
│                                         │
│  ┌──────────────┐    ┌──────────────┐  │
│  │              │    │              │  │
│  │  Application │───▶│    MySQL     │  │
│  │   (Port 8080)│    │  (Port 3306) │  │
│  │              │    │              │  │
│  └──────────────┘    └──────────────┘  │
│         │                    │          │
│         │                    │          │
│    ┌────▼────┐         ┌────▼────┐     │
│    │ Health  │         │ Volume  │     │
│    │ Check   │         │  Data   │     │
│    └─────────┘         └─────────┘     │
│                                         │
└─────────────────────────────────────────┘
```

## Database Migrations

Migrations are **automatic** and **idempotent**:

1. Application starts
2. Connects to database (waits for MySQL to be ready)
3. Runs migration SQL
4. Creates tables if they don't exist
5. Application ready

### Tables Created

- `domains` - Domain information and monitoring status
- `config` - Application configuration
- `alerts` - Alert history

### Verify Migrations

```bash
./verify-migration.sh
```

## Environment Variables

### Required for Docker

```bash
# MySQL
MYSQL_ROOT_PASSWORD=your-secure-password
MYSQL_PASSWORD=your-secure-password

# Application
GOOGLE_CHAT_WEBHOOK=https://chat.googleapis.com/...
```

### Optional

```bash
# Monitoring
MONITORING_INTERVAL=24h
ALERT_THRESHOLDS=90d,60d,30d,7d
RETENTION_PERIOD=90d

# Ports
PORT=8080
MYSQL_PORT=3306
```

## Common Commands

### Start
```bash
docker-compose up -d
```

### View Logs
```bash
docker-compose logs -f
```

### Stop
```bash
docker-compose stop
```

### Restart
```bash
docker-compose restart
```

### Backup Database
```bash
docker-compose exec mysql mysqldump -u demuser -p dem > backup.sql
```

### Access MySQL
```bash
docker-compose exec mysql mysql -u demuser -p dem
```

## Testing Alerts

1. Configure webhook at http://localhost:8080/config
2. Run test script:
   ```bash
   ./test_alert.sh
   ```
3. Restart application:
   ```bash
   docker-compose restart app
   ```
4. Check alert history in UI

## Production Checklist

- [ ] Change default passwords in `.env`
- [ ] Set strong `MYSQL_ROOT_PASSWORD`
- [ ] Set strong `MYSQL_PASSWORD`
- [ ] Configure `GOOGLE_CHAT_WEBHOOK`
- [ ] Set appropriate `ALERT_THRESHOLDS`
- [ ] Configure backup schedule
- [ ] Set up monitoring/alerting
- [ ] Review security settings
- [ ] Test disaster recovery
- [ ] Document runbook procedures

## Troubleshooting

### Application won't start
```bash
# Check logs
docker-compose logs app

# Common issues:
# - MySQL not ready (wait 10-30 seconds)
# - Wrong credentials (check .env)
# - Port in use (change PORT in .env)
```

### MySQL connection failed
```bash
# Check MySQL is running
docker-compose ps mysql

# Check MySQL logs
docker-compose logs mysql

# Test connection
docker-compose exec mysql mysqladmin ping -h localhost -u root -p
```

### Migrations didn't run
```bash
# Check app logs for migration errors
docker-compose logs app | grep -i migration

# Manually verify
./verify-migration.sh

# Force restart
docker-compose restart app
```

## File Structure

```
.
├── Dockerfile                  # Application container
├── docker-compose.yml          # Complete stack definition
├── .dockerignore              # Build optimization
├── docker-start.sh            # Easy startup script
├── verify-migration.sh        # Migration verification
├── .env.example               # Environment template
├── DOCKER_DEPLOYMENT.md       # Full Docker guide
├── DOCKER_QUICKREF.md         # Quick reference
└── DEPLOYMENT_SUMMARY.md      # This file
```

## Next Steps

1. **Configure**: Edit `.env` with your settings
2. **Start**: Run `./docker-start.sh`
3. **Configure Webhook**: Visit http://localhost:8080/config
4. **Add Domains**: Add domains to monitor
5. **Test Alerts**: Run `./test_alert.sh`

## Support

- Full Docker guide: [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md)
- Quick reference: [DOCKER_QUICKREF.md](DOCKER_QUICKREF.md)
- MySQL setup: [MYSQL_SETUP.md](MYSQL_SETUP.md)
- Alert testing: [ALERT_TESTING_GUIDE.md](ALERT_TESTING_GUIDE.md)
- General usage: [README.md](README.md)

## Migration from SQLite to MySQL

If you're currently using SQLite and want to migrate to MySQL:

1. Export SQLite data:
   ```bash
   sqlite3 dem.db .dump > sqlite_backup.sql
   ```

2. Start MySQL with Docker:
   ```bash
   docker-compose up -d mysql
   ```

3. Convert and import (manual process - schema differences)

4. Update `.env` to use MySQL

5. Restart application

See [MYSQL_MIGRATION.md](MYSQL_MIGRATION.md) for detailed steps.
