.PHONY: help build run test docker-build docker-up docker-down docker-logs docker-restart verify-migration clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/dem ./cmd/dem
	@echo "✅ Build complete: bin/dem"

run: build ## Build and run the application locally
	@echo "Starting application..."
	@./bin/dem

test: ## Run all tests
	@echo "Running tests..."
	@go test ./...

test-property: ## Run property-based tests only
	@echo "Running property-based tests..."
	@go test ./... -run TestProperty

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	@go test -v ./...

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker-compose build

docker-up: ## Start Docker services
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "✅ Services started"
	@echo "Application: http://localhost:8080"

docker-up-logs: ## Start Docker services with logs
	@echo "Starting Docker services..."
	@docker-compose up

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	@docker-compose down

docker-down-volumes: ## Stop Docker services and remove volumes (⚠️ deletes data!)
	@echo "⚠️  This will delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "✅ Services stopped and volumes removed"; \
	fi

docker-logs: ## View Docker logs
	@docker-compose logs -f

docker-logs-app: ## View application logs only
	@docker-compose logs -f app

docker-logs-mysql: ## View MySQL logs only
	@docker-compose logs -f mysql

docker-restart: ## Restart Docker services
	@echo "Restarting Docker services..."
	@docker-compose restart
	@echo "✅ Services restarted"

docker-restart-app: ## Restart application only
	@echo "Restarting application..."
	@docker-compose restart app
	@echo "✅ Application restarted"

docker-ps: ## Show Docker service status
	@docker-compose ps

docker-exec-app: ## Execute shell in application container
	@docker-compose exec app sh

docker-exec-mysql: ## Execute MySQL client
	@docker-compose exec mysql mysql -u demuser -p dem

verify-migration: ## Verify database migrations
	@./verify-migration.sh

test-alert: ## Run alert test script
	@./test_alert.sh
	@echo ""
	@echo "Now restart the application to trigger the alert check"

backup-db: ## Backup MySQL database
	@echo "Creating database backup..."
	@docker-compose exec mysql mysqldump -u demuser -p dem > backup-$$(date +%Y%m%d-%H%M%S).sql
	@echo "✅ Backup created"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f dem.db dem.db-shm dem.db-wal
	@echo "✅ Clean complete"

clean-docker: ## Remove Docker images and volumes
	@echo "Removing Docker images and volumes..."
	@docker-compose down --rmi all -v
	@echo "✅ Docker cleanup complete"

install-deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@golangci-lint run ./...

format: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

dev: ## Start development environment (Docker)
	@echo "Starting development environment..."
	@docker-compose up --build

prod: ## Start production environment (Docker)
	@echo "Starting production environment..."
	@docker-compose -f docker-compose.yml up -d --build
	@echo "✅ Production environment started"

health: ## Check application health
	@echo "Checking application health..."
	@curl -s http://localhost:8080/health | jq .
