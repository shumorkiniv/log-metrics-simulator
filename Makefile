.PHONY: help build build-backend build-frontend run run-backend run-frontend test test-backend test-frontend \
        clean clean-backend clean-frontend docker-build docker-build-backend docker-build-frontend \
        docker-up docker-down docker-logs docker-restart deploy install install-backend install-frontend \
        logs-app logs-backend logs-frontend ssh health generate-test fmt lint vet security-check \
        dev dev-backend dev-frontend

# Default goal
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := log-metrics-simulator
BACKEND_DIR := ./backend
FRONTEND_DIR := ./frontend
DOCKER_COMPOSE := docker-compose
GO_VERSION := 1.24

## help: Show this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## install: Install all dependencies
install: install-backend install-frontend

## install-backend: Install Go dependencies
install-backend:
	@echo "Installing Go dependencies..."
	cd $(BACKEND_DIR) && go mod download && go mod verify

## install-frontend: Install Node.js dependencies
install-frontend:
	@echo "Installing Node.js dependencies..."
	cd $(FRONTEND_DIR) && npm install

## build: Build both backend and frontend
build: build-backend build-frontend

## build-backend: Build backend application
build-backend:
	@echo "Building backend..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -ldflags="-w -s" -o ../$(BINARY_NAME) .

## build-frontend: Build frontend application
build-frontend:
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && npm run build

## run: Run both backend and frontend in development mode
run:
	@echo "Starting development servers..."
	$(MAKE) -j2 run-backend run-frontend

## run-backend: Run backend in development mode
run-backend:
	@echo "Starting backend server..."
	cd $(BACKEND_DIR) && go run .

## run-frontend: Run frontend in development mode
run-frontend:
	@echo "Starting frontend development server..."
	cd $(FRONTEND_DIR) && npm run dev

## dev: Start development environment with hot reload
dev:
	@echo "Starting development environment..."
	$(DOCKER_COMPOSE) -f docker-compose.dev.yml up --build

## dev-backend: Run backend with hot reload
dev-backend:
	@echo "Starting backend with hot reload..."
	cd $(BACKEND_DIR) && air

## dev-frontend: Run frontend with hot reload
dev-frontend:
	@echo "Starting frontend with hot reload..."
	cd $(FRONTEND_DIR) && npm run dev

## test: Run all tests
test: test-backend test-frontend

## test-backend: Run backend tests
test-backend:
	@echo "Running backend tests..."
	cd $(BACKEND_DIR) && go test -v -race -timeout 30s ./...

## test-frontend: Run frontend tests
test-frontend:
	@echo "Running frontend tests..."
	cd $(FRONTEND_DIR) && npm run test

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running backend tests with coverage..."
	cd $(BACKEND_DIR) && go test -v -race -coverprofile=coverage.out ./...
	cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: backend/coverage.html"

## fmt: Format Go code
fmt:
	@echo "Formatting Go code..."
	cd $(BACKEND_DIR) && go fmt ./...

## lint: Run linters
lint: lint-backend lint-frontend

## lint-backend: Lint backend code
lint-backend:
	@echo "Linting backend code..."
	cd $(BACKEND_DIR) && golangci-lint run

## lint-frontend: Lint frontend code
lint-frontend:
	@echo "Linting frontend code..."
	cd $(FRONTEND_DIR) && npm run lint

## vet: Run Go vet
vet:
	@echo "Running go vet..."
	cd $(BACKEND_DIR) && go vet ./...

## security-check: Run security checks
security-check:
	@echo "Running security checks..."
	cd $(BACKEND_DIR) && gosec ./...
	cd $(FRONTEND_DIR) && npm audit

## clean: Clean all build artifacts
clean: clean-backend clean-frontend

## clean-backend: Clean backend build artifacts
clean-backend:
	@echo "Cleaning backend artifacts..."
	rm -f $(BINARY_NAME)
	cd $(BACKEND_DIR) && go clean -cache -modcache -i -r
	rm -f $(BACKEND_DIR)/coverage.out $(BACKEND_DIR)/coverage.html

## clean-frontend: Clean frontend build artifacts
clean-frontend:
	@echo "Cleaning frontend artifacts..."
	cd $(FRONTEND_DIR) && rm -rf dist node_modules/.cache

## docker-build: Build all Docker images
docker-build: docker-build-backend docker-build-frontend

## docker-build-backend: Build backend Docker image
docker-build-backend:
	@echo "Building backend Docker image..."
	docker build -t log-metrics-simulator-backend:latest $(BACKEND_DIR)

## docker-build-frontend: Build frontend Docker image (if needed)
docker-build-frontend:
	@echo "Building frontend Docker image..."
	docker build -t log-metrics-simulator-frontend:latest $(FRONTEND_DIR)

## docker-up: Start all services
docker-up:
	@echo "Starting Docker services..."
	$(DOCKER_COMPOSE) up -d

## docker-up-build: Build and start all services
docker-up-build:
	@echo "Building and starting Docker services..."
	$(DOCKER_COMPOSE) up -d --build

## docker-down: Stop all services
docker-down:
	@echo "Stopping Docker services..."
	$(DOCKER_COMPOSE) down

## docker-down-clean: Stop services and remove volumes
docker-down-clean:
	@echo "Stopping services and cleaning up..."
	$(DOCKER_COMPOSE) down -v --remove-orphans

## docker-logs: Show logs from all services
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## docker-restart: Restart all services
docker-restart:
	@echo "Restarting Docker services..."
	$(DOCKER_COMPOSE) restart

## deploy: Deploy to production
deploy: clean docker-build docker-up-build
	@echo "Deployment completed!"

## logs-app: Show application logs
logs-app:
	$(DOCKER_COMPOSE) logs -f log-metrics-simulator

## logs-backend: Show backend logs
logs-backend:
	$(DOCKER_COMPOSE) logs -f backend

## logs-frontend: Show frontend logs (if applicable)
logs-frontend:
	$(DOCKER_COMPOSE) logs -f frontend

## ssh: Connect to backend container
ssh:
	$(DOCKER_COMPOSE) exec backend /bin/sh

## ssh-backend: Connect to backend container
ssh-backend:
	$(DOCKER_COMPOSE) exec backend /bin/sh

## health: Check application health
health:
	@echo "Checking application health..."
	@curl -f http://localhost:8080/health || echo "Health check failed"

## metrics: Get application metrics
metrics:
	@echo "Getting application metrics..."
	@curl -s http://localhost:8080/metrics

## generate-test: Generate test data
generate-test:
	@echo "Generating test data..."
	@curl -X POST http://localhost:8080/api/v1/generate \
		-H "Content-Type: application/json" \
		-d '{"log_count": 100, "scenario": "normal_load"}' || echo "Failed to generate test data"

## generate-load: Generate high load test data
generate-load:
	@echo "Generating high load test data..."
	@curl -X POST http://localhost:8080/api/v1/generate \
		-H "Content-Type: application/json" \
		-d '{"log_count": 1000, "scenario": "high_load"}' || echo "Failed to generate load test data"

## generate-black-friday: Generate Black Friday scenario
generate-black-friday:
	@echo "Generating Black Friday scenario..."
	@curl -X POST http://localhost:8080/api/v1/generate \
		-H "Content-Type: application/json" \
		-d '{"log_count": 5000, "scenario": "black_friday"}' || echo "Failed to generate Black Friday scenario"

## prometheus: Open Prometheus UI
prometheus:
	@echo "Opening Prometheus UI..."
	@open http://localhost:9090 2>/dev/null || echo "Open http://localhost:9090 manually"

## grafana: Open Grafana UI
grafana:
	@echo "Opening Grafana UI..."
	@open http://localhost:3001 2>/dev/null || echo "Open http://localhost:3001 manually"

## setup: Initial project setup
setup: install
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@command -v gosec >/dev/null 2>&1 || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	@command -v air >/dev/null 2>&1 || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	@echo "Setup completed!"

## check: Run all checks (format, lint, vet, test)
check: fmt vet lint test
	@echo "All checks completed!"

## ci: Run CI pipeline locally
ci: clean install check test-coverage
	@echo "CI pipeline completed!"

## update-deps: Update all dependencies
update-deps:
	@echo "Updating Go dependencies..."
	cd $(BACKEND_DIR) && go get -u ./...
	cd $(BACKEND_DIR) && go mod tidy
	@echo "Updating Node.js dependencies..."
	cd $(FRONTEND_DIR) && npm update

## status: Show project status
status:
	@echo "=== Project Status ==="
	@echo "Go version: $(shell go version)"
	@echo "Node version: $(shell node --version 2>/dev/null || echo 'Not installed')"
	@echo "Docker version: $(shell docker --version 2>/dev/null || echo 'Not installed')"
	@echo "Docker Compose version: $(shell docker-compose --version 2>/dev/null || echo 'Not installed')"
	@echo ""
	@echo "=== Service Status ==="
	@$(DOCKER_COMPOSE) ps 2>/dev/null || echo "Docker services not running"