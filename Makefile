.PHONY: help build run test clean deps seed docker-build docker-up docker-down

# Variables
APP_NAME=handsoff
SERVER_BINARY=bin/$(APP_NAME)-server
GO=go
GOFLAGS=-v

API_PORT=$(shell grep API_PORT .env | cut -d'=' -f2)


help: ## Show this help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

build: ## Build all binaries
	@echo "Building unified server ..."
	$(GO) build $(GOFLAGS) -o $(SERVER_BINARY) ./cmd/server
	@echo "✅ Server build completed"

run: ## Run server
	@echo "Found port $(API_PORT), attempting to kill existing process..."
	@fuser -k "$(API_PORT)/tcp" || true
	@echo "Starting unified server on port $(API_PORT)..."
	@$(GO) run ./cmd/server/main.go



lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -rf data/*.db
	rm -rf temp/

# Database operations
db-reset: ## Reset database (WARNING: deletes all data)
	rm -f data/*.db
	@echo "✅ Database reset. Run 'make seed' to create admin user."

web-dev: ## Run frontend dev server
	cd web && npm run dev

web-build: ## Build frontend for production
	cd web && npm run build

# All-in-one commands
all: clean deps build ## Clean, install deps, and build


.DEFAULT_GOAL := help
