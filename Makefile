.PHONY: help build run test clean deps seed docker-build docker-up docker-down

# Variables
APP_NAME=handsoff
API_BINARY=bin/$(APP_NAME)-api
WORKER_BINARY=bin/$(APP_NAME)-worker
GO=go
GOFLAGS=-v

help: ## Show this help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

build: ## Build all binaries
	@echo "Building API server..."
	$(GO) build $(GOFLAGS) -o $(API_BINARY) ./cmd/api
	@echo "Building Worker..."
	$(GO) build $(GOFLAGS) -o $(WORKER_BINARY) ./cmd/worker
	@echo "✅ Build completed"

run-api: ## Run API server
	$(GO) run ./cmd/api/main.go

run-worker: ## Run worker
	$(GO) run ./cmd/worker/main.go

run: ## Run both API and worker (requires separate terminals)
	@echo "Run 'make run-api' in one terminal and 'make run-worker' in another"


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
