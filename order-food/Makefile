.PHONY: help build run test clean docker-build docker-run deps

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o bin/order-food ./cmd/main.go

run: ## Run the application
	go run ./cmd/main.go

test: ## Run tests
	go test -v ./...

test-api: ## Test API endpoints (requires running server)
	./test-api.sh

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

docker-build: ## Build Docker image
	docker build -t order-food:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 order-food:latest

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: fmt vet ## Run linters

dev: ## Run in development mode with hot reload (requires air)
	air
