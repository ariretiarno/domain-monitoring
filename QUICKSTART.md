# Quick Start Guide

Get the Domain Expiration Monitor running in 5 minutes.

## Option 1: Docker (Recommended)

```bash
# 1. Configure
cp .env.example .env
nano .env  # Set MYSQL_ROOT_PASSWORD, MYSQL_PASSWORD, GOOGLE_CHAT_WEBHOOK

# 2. Start
./docker-start.sh
# or: docker-compose up -d

# 3. Access
open http://localhost:8080
```

## Option 2: Local Build

```bash
# 1. Build
go build -o bin/dem ./cmd/dem

# 2. Run
./bin/dem

# 3. Access
open http://localhost:8080
```

## First Steps

### 1. Configure Alerts (Optional)

Visit http://localhost:8080/config
- Add Google Chat webhook URL
- Set alert thresholds: `90,60,30,7` (days)
- Click "Save"

### 2. Add a Domain

On the dashboard:
- Enter domain name (e.g., `example.com`)
- Click "Add Domain"
- View WHOIS info and expiration date

### 3. Test Alerts

```bash
./test_alert.sh
docker-compose restart app  # if using Docker
```

Check domain details to see alert history.

## Common Commands

### Docker
```bash
docker-compose up -d        # Start
docker-compose logs -f      # View logs
docker-compose restart      # Restart
docker-compose down         # Stop
```

### Make
```bash
make docker-up              # Start Docker
make docker-logs            # View logs
make test-alert             # Test alerts
make help                   # See all commands
```

### Local
```bash
./bin/dem                   # Run application
curl http://localhost:8080/health  # Check health
```

## Troubleshooting

**Port in use?**
```bash
# Change port in .env
PORT=8081
```

**No alerts?**
- Check webhook URL is configured
- Verify domain expires within threshold
- Check alert history in UI

**Docker issues?**
```bash
docker-compose logs app     # Check logs
docker-compose ps           # Check status
```

## Documentation

- **README.md** - Full documentation
- **docs/DOCKER_DEPLOYMENT.md** - Docker guide
- **docs/MYSQL_SETUP.md** - MySQL setup
- **docs/ALERT_TESTING_GUIDE.md** - Alert testing

## Next Steps

1. Add your domains
2. Configure monitoring interval
3. Set up backup schedule
4. Review production checklist in README.md
