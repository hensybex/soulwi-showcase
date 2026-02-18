# Variables
BINARY_NAME=soulwi-api
BINARY_PATH=./cmd/server
BUILD_DIR=./bin
COMPOSE_FILE=compose.yaml

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
.PHONY: run
run: ## Run the API server in development mode
	go run $(BINARY_PATH)

.PHONY: build
build: ## Build the API binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	go clean
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Code quality commands
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: fmt vet ## Run formatting and static analysis
	@echo "Linting complete"

# Dependency management
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go mod tidy
	go mod download

# Database commands
.PHONY: db-up
db-up: ## Start database service only
	@echo "Starting database service..."
	docker compose -f $(COMPOSE_FILE) up -d db

.PHONY: db-down
db-down: ## Stop database service
	@echo "Stopping database service..."
	docker compose -f $(COMPOSE_FILE) stop db

.PHONY: db-logs
db-logs: ## Show database logs
	@echo "Showing database logs..."
	docker compose -f $(COMPOSE_FILE) logs -f db

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "Resetting database..."
	docker compose -f $(COMPOSE_FILE) down -v
	docker compose -f $(COMPOSE_FILE) up -d db

# Development workflow
.PHONY: dev-setup
dev-setup: deps db-up ## Setup development environment
	@echo "Development environment setup complete"
	@echo "Database: localhost:5444"
	@echo "Run 'make run' to start the API server"

.PHONY: dev
dev: dev-setup run ## Complete development workflow (setup + run)

.PHONY: dev-clean
dev-clean: db-down clean ## Clean development environment
	@echo "Development environment cleaned"
