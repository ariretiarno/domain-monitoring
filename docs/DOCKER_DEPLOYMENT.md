# Docker Deployment Guide

This guide explains how to deploy the Domain Expiration Monitor using Docker and Docker Compose with MySQL.

## Quick Start

### 1. Clone and Configure

```bash
# Clone the repository
git clone <your-repo-url>
cd whois-monitoring

# Create environment file
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` file:

```bash
# MySQL Configuration
MYSQL_ROOT_PASSWORD=your-secure-root-password
MYSQL_DATABASE=dem
MYSQL_USER=demuser
MYSQL_PASSWORD=your-secure-password
MYSQL_PORT=3306

# Application Configuration
PORT=8080
MONITORING_INTERVAL=24h
ALERT_THRESHOLDS=90d,60d,30d,7d
GOOGLE_CHAT_WEBHOOK=https://chat.googleapis.com/v1/spaces/YOUR-WEBHOOK-URL
RETENTION_PERIOD=90d
```

### 3. Start the Application

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Check status
docker-compose ps
```

### 4. Access the Application

Open your browser to: http://localhost:8080

## Architecture

The Docker Compose setup includes:

- **MySQL 8.0**: Database server with persistent storage
- **Application**: Go application with automatic migrations
- **Network**: Isolated bridge network for service communication
- **Volumes**: Persistent MySQL data storage

## Database Migrations

Migrations run automatically when the application starts:

1. Application waits for MySQL to be healthy
2. Connects to MySQL database
3. Creates tables if they don't exist:
   - `domains` - Domain information and monitoring status
   - `config` - Application configuration
   - `alerts` - Alert history

The migration is idempotent - safe to run multiple times.

## Configuration Details

### MySQL Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MYSQL_ROOT_PASSWORD` | `rootpassword` | MySQL root password |
| `MYSQL_DATABASE` | `dem` | Database name |
| `MYSQL_USER` | `demuser` | Application database user |
| `MYSQL_PASSWORD` | `dempassword` | Application user password |
| `MYSQL_PORT` | `3306` | MySQL port (host mapping) |

### Application Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DRIVER` | `mysql` | Database driver (mysql or sqlite3) |
| `DB_HOST` | `mysql` | MySQL hostname |
| `DB_PORT` | `3306` | MySQL port |
| `DB_NAME` | `dem` | Database name |
| `DB_USER` | `demuser` | Database user |
| `DB_PASSWORD` | `dempassword` | Database password |
| `PORT` | `8080` | HTTP server port |
| `MONITORING_INTERVAL` | `24h` | How often to check domains |
| `ALERT_THRESHOLDS` | `90d,60d,30d,7d` | Alert thresholds |
| `GOOGLE_CHAT_WEBHOOK` | - | Google Chat webhook URL |
| `RETENTION_PERIOD` | `90d` | How long to keep old data |

## Management Commands

### Start Services

```bash
# Start in background
docker-compose up -d

# Start with logs
docker-compose up
```

### Stop Services

```bash
# Stop services
docker-compose stop

# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes (deletes data!)
docker-compose down -v
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f mysql

# Last 100 lines
docker-compose logs --tail=100 app
```

### Restart Services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart app
```

### Rebuild Application

```bash
# Rebuild and restart
docker-compose up -d --build

# Force rebuild
docker-compose build --no-cache
docker-compose up -d
```

## Database Access

### Connect to MySQL

```bash
# Using docker-compose
docker-compose exec mysql mysql -u demuser -p dem

# Direct docker command
docker exec -it dem-mysql mysql -u demuser -p dem
```

### Backup Database

```bash
# Create backup
docker-compose exec mysql mysqldump -u demuser -p dem > backup.sql

# With timestamp
docker-compose exec mysql mysqldump -u demuser -p dem > backup-$(date +%Y%m%d-%H%M%S).sql
```

### Restore Database

```bash
# Restore from backup
docker-compose exec -T mysql mysql -u demuser -p dem < backup.sql
```

## Monitoring

### Health Checks

The application includes health checks:

```bash
# Check application health
curl http://localhost:8080/health

# Check MySQL health
docker-compose exec mysql mysqladmin ping -h localhost -u root -p
```

### Container Status

```bash
# View container status
docker-compose ps

# View resource usage
docker stats dem-app dem-mysql
```

## Troubleshooting

### Application Won't Start

**Check logs:**
```bash
docker-compose logs app
```

**Common issues:**
- MySQL not ready: Wait for MySQL health check to pass
- Connection refused: Check DB_HOST matches service name
- Authentication failed: Verify DB_USER and DB_PASSWORD

### MySQL Connection Issues

**Verify MySQL is running:**
```bash
docker-compose ps mysql
```

**Check MySQL logs:**
```bash
docker-compose logs mysql
```

**Test connection:**
```bash
docker-compose exec mysql mysql -u demuser -p -e "SELECT 1;"
```

### Port Already in Use

If port 8080 or 3306 is already in use:

```bash
# Change ports in .env
PORT=8081
MYSQL_PORT=3307

# Restart
docker-compose down
docker-compose up -d
```

### Reset Everything

```bash
# Stop and remove everything
docker-compose down -v

# Remove images
docker-compose down --rmi all -v

# Start fresh
docker-compose up -d
```

## Production Deployment

### Security Recommendations

1. **Change default passwords:**
   ```bash
   # Generate strong passwords
   openssl rand -base64 32
   ```

2. **Use secrets management:**
   - Docker secrets
   - Kubernetes secrets
   - HashiCorp Vault

3. **Restrict network access:**
   - Don't expose MySQL port publicly
   - Use firewall rules
   - Enable SSL/TLS

4. **Regular backups:**
   ```bash
   # Automated backup script
   #!/bin/bash
   docker-compose exec mysql mysqldump -u demuser -p$DB_PASSWORD dem | \
     gzip > /backups/dem-$(date +%Y%m%d-%H%M%S).sql.gz
   ```

### Performance Tuning

**MySQL Configuration:**

Create `mysql.cnf`:
```ini
[mysqld]
max_connections = 100
innodb_buffer_pool_size = 256M
innodb_log_file_size = 64M
```

Mount in docker-compose.yml:
```yaml
mysql:
  volumes:
    - ./mysql.cnf:/etc/mysql/conf.d/custom.cnf
```

**Application Scaling:**

```yaml
app:
  deploy:
    replicas: 3
    resources:
      limits:
        cpus: '1'
        memory: 512M
```

### Monitoring Setup

**Add Prometheus metrics:**
```yaml
prometheus:
  image: prom/prometheus
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"
```

**Add Grafana:**
```yaml
grafana:
  image: grafana/grafana
  ports:
    - "3000:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=admin
```

## Kubernetes Deployment

For Kubernetes deployment, see `k8s/` directory (if available) or convert using:

```bash
# Install kompose
curl -L https://github.com/kubernetes/kompose/releases/download/v1.31.2/kompose-linux-amd64 -o kompose
chmod +x kompose

# Convert docker-compose to k8s
./kompose convert
```

## Updates and Maintenance

### Update Application

```bash
# Pull latest code
git pull

# Rebuild and restart
docker-compose up -d --build
```

### Update MySQL

```bash
# Backup first!
docker-compose exec mysql mysqldump -u demuser -p dem > backup.sql

# Update image version in docker-compose.yml
# mysql:8.0 -> mysql:8.1

# Pull new image
docker-compose pull mysql

# Restart
docker-compose up -d mysql
```

## Support

For issues or questions:
- Check logs: `docker-compose logs -f`
- Review documentation: README.md, MYSQL_SETUP.md
- Open an issue on GitHub
