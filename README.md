# Domain Expiration Monitor

A Go-based application that monitors domain expiration dates via WHOIS queries and sends configurable alerts to Google Chat.
<img width="1187" height="563" alt="image" src="https://github.com/user-attachments/assets/0c681ba4-1985-4115-a4d1-a6b83564a63c" />


## Features

- 🔍 Automatic WHOIS monitoring with configurable intervals
- 📊 Web UI for domain management and configuration
- 🔔 Google Chat webhook integration for alerts
- ⏰ Configurable alert thresholds via UI
- 💾 SQLite or MySQL database support
- 🧪 Comprehensive property-based testing
- 🌍 Multi-TLD support (.com, .ar, .org, etc.)

## Quick Start

### Option 1: Docker (Recommended)

The easiest way to get started with MySQL included:

```bash
# Start with Docker Compose
./docker-start.sh

# Or manually
docker-compose up -d
```

Access the application at http://localhost:8080

See [docs/DOCKER_DEPLOYMENT.md](docs/DOCKER_DEPLOYMENT.md) for detailed Docker documentation.

### Option 2: Local Build

#### Build

```bash
go build -o bin/dem ./cmd/dem
```

#### Configure

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
# Edit .env with your settings
```

#### Run

```bash
./bin/dem
```

The application will start on `http://localhost:8080` by default.

## Configuration

### Environment Variables (.env file)

```bash
# Database (SQLite - default)
DB_DRIVER=sqlite3
DB_PATH=dem.db

# Database (MySQL - for local)
# DB_DRIVER=mysql
# DB_HOST=localhost
# DB_PORT=3306
# DB_NAME=dem
# DB_USER=demuser
# DB_PASSWORD=your-password

# HTTP Server
HTTP_ADDR=:8080
PORT=8080

# Application Settings
MONITORING_INTERVAL=24h
ALERT_THRESHOLDS=90d,60d,30d,7d
GOOGLE_CHAT_WEBHOOK=https://chat.googleapis.com/v1/spaces/...
RETENTION_PERIOD=90d
```

For Docker deployment, see [docs/DOCKER_DEPLOYMENT.md](docs/DOCKER_DEPLOYMENT.md).

### Web UI Configuration

Access http://localhost:8080/config to configure:
- Monitoring interval (how often to check domains)
- Google Chat webhook URL
- Alert thresholds (when to send alerts)
- Data retention period

## Usage

1. **Add a domain**: Navigate to the dashboard and enter a domain name
2. **Configure alerts**: Go to `/config` to set up Google Chat webhook and monitoring intervals
3. **View details**: Click on any domain to see detailed WHOIS information and alert history

## Architecture

- **Domain Layer**: Core business models and logic
- **Repository Layer**: SQLite database access with connection pooling
- **WHOIS Service**: Domain information retrieval with retry logic
- **Alert Service**: Threshold evaluation and Google Chat notifications
- **Scheduler**: Periodic monitoring with worker pool
- **Web UI**: HTTP server with HTML templates

## Testing

Run all tests:
```bash
go test ./...
```

Run property-based tests only:
```bash
go test ./... -run TestProperty
```

## API Endpoints

- `GET /` - Dashboard
- `GET /health` - Health check
- `GET /domains/:id` - Domain details
- `POST /domains` - Add domain
- `DELETE /domains?id=:id` - Delete domain
- `GET /config` - Configuration page
- `POST /config` - Update configuration

## Database Support

### SQLite (Default)
- Zero configuration
- Single file database
- Perfect for single-server deployments

### MySQL
- Better for high-traffic scenarios
- Supports replication
- See [docs/MYSQL_SETUP.md](docs/MYSQL_SETUP.md) for setup instructions

### Database Migrations

Migrations run automatically when the application starts:
- Creates tables if they don't exist
- Safe to run multiple times (idempotent)
- No manual migration steps required

Verify migrations:
```bash
./verify-migration.sh
```

## Testing

See [docs/TESTING.md](docs/TESTING.md) for:
- How to test alerts
- Using webhook.site for testing
- Running property-based tests

## Production Checklist

Before deploying to production:

- [ ] Change default passwords in `.env`
- [ ] Set strong `MYSQL_ROOT_PASSWORD` and `MYSQL_PASSWORD`
- [ ] Configure `GOOGLE_CHAT_WEBHOOK`
- [ ] Set appropriate `ALERT_THRESHOLDS`
- [ ] Configure automated backups
- [ ] Set up monitoring/alerting
- [ ] Review security settings
- [ ] Test disaster recovery
- [ ] Use HTTPS with reverse proxy
- [ ] Set up log rotation

## Documentation

- **QUICKSTART.md** - Get started in 5 minutes
- **docs/DOCKER_DEPLOYMENT.md** - Complete Docker guide
- **docs/MYSQL_SETUP.md** - MySQL setup instructions
- **docs/ALERT_TESTING_GUIDE.md** - How to test alerts
- **docs/TESTING.md** - Testing guide
- **CHANGELOG.md** - Version history

## Requirements

- Go 1.21 or higher
- SQLite3 (default) or MySQL 8.0+ (optional)
- Docker & Docker Compose (for Docker deployment)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## License

MIT
