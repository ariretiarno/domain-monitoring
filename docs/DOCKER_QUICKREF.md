# Docker Quick Reference

Quick reference for common Docker commands for the Domain Expiration Monitor.

## Starting & Stopping

```bash
# Start everything
docker-compose up -d

# Start with logs visible
docker-compose up

# Stop services
docker-compose stop

# Stop and remove containers
docker-compose down

# Stop and remove everything including volumes (⚠️ deletes data!)
docker-compose down -v
```

## Viewing Logs

```bash
# All services
docker-compose logs -f

# Just the app
docker-compose logs -f app

# Just MySQL
docker-compose logs -f mysql

# Last 100 lines
docker-compose logs --tail=100 app
```

## Rebuilding

```bash
# Rebuild and restart
docker-compose up -d --build

# Force complete rebuild
docker-compose build --no-cache
docker-compose up -d
```

## Database Access

```bash
# Connect to MySQL
docker-compose exec mysql mysql -u demuser -p dem

# Run SQL query
docker-compose exec mysql mysql -u demuser -p dem -e "SELECT * FROM domains;"

# Backup database
docker-compose exec mysql mysqldump -u demuser -p dem > backup.sql

# Restore database
docker-compose exec -T mysql mysql -u demuser -p dem < backup.sql
```

## Monitoring

```bash
# Check status
docker-compose ps

# Check resource usage
docker stats dem-app dem-mysql

# Check health
curl http://localhost:8080/health

# Verify migrations
./verify-migration.sh
```

## Troubleshooting

```bash
# Restart a service
docker-compose restart app

# View container details
docker inspect dem-app

# Execute command in container
docker-compose exec app sh

# Check MySQL connection
docker-compose exec mysql mysqladmin ping -h localhost -u root -p
```

## Cleanup

```bash
# Remove stopped containers
docker-compose rm

# Remove unused images
docker image prune

# Remove unused volumes
docker volume prune

# Remove everything unused
docker system prune -a
```

## Environment

```bash
# View environment variables
docker-compose exec app env

# Edit .env file
nano .env

# Reload after .env changes
docker-compose down
docker-compose up -d
```

## Scaling (if needed)

```bash
# Run multiple app instances
docker-compose up -d --scale app=3

# Check running instances
docker-compose ps
```

## Quick Start Script

```bash
# Use the helper script
./docker-start.sh
```

## Common Issues

### Port already in use
```bash
# Change port in .env
PORT=8081

# Restart
docker-compose down
docker-compose up -d
```

### MySQL not ready
```bash
# Wait for health check
docker-compose logs mysql | grep "ready for connections"

# Or restart
docker-compose restart mysql
```

### Application won't connect to MySQL
```bash
# Check network
docker network inspect whois-monitoring_dem-network

# Check MySQL is accessible
docker-compose exec app ping mysql
```

### Reset everything
```bash
# Nuclear option - removes all data!
docker-compose down -v
docker-compose up -d
```
