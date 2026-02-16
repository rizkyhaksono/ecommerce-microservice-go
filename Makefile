# Microservices Makefile

.PHONY: build up down logs restart clean

# Build all services
build:
	@echo "Building all services..."
	docker compose build

# Start all services
up:
	@echo "Starting services..."
	docker compose up -d

# Stop all services
down:
	@echo "Stopping services..."
	docker compose down

# View logs
logs:
	docker compose logs -f

# Restart all services
restart: down up

# Clean up docker resources
clean:
	docker compose down -v --rmi local --remove-orphans

# Run tests (requires services to be running for integration tests, or use go test locally)
test:
	@echo "Running tests in all services..."
	go test ./pkg/... ./services/...

# Sync shared packages (workspace)
sync:
	go work sync

# Generate swagger docs (requires swaggo installed)
swagger:
	@echo "Generating swagger for User Service..."
	cd services/user && swag init --parseDependency --parseInternal
	@echo "Generating swagger for Catalog Service..."
	cd services/catalog && swag init --parseDependency --parseInternal
	@echo "Generating swagger for Order Service..."
	cd services/order && swag init --parseDependency --parseInternal
