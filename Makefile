.PHONY: help build test clean frontend docker vercel workers lint fmt

# Variables
BINARY_NAME=who-dat
VERSION?=2.0.0
GO_FILES=$(shell find . -name '*.go' -not -path './dist/*' -not -path './node_modules/*')

# Help target
help: ## Show this help message
	@echo "Who-Dat Build System"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: frontend ## Run development server
	@echo "Starting development server..."
	go run cmd/server/main.go

# Build targets
build: frontend build-server ## Build for all platforms

build-server: ## Build standalone server binary
	@echo "Building server..."
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$(BINARY_NAME) ./cmd/server

build-docker: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) -f Dockerfile .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

build-workers: frontend ## Build for Cloudflare Workers (requires TinyGo)
	@echo "Building for Cloudflare Workers..."
	@echo "Note: This requires TinyGo to be installed"
	@if command -v tinygo >/dev/null 2>&1; then \
		GOOS=js GOARCH=wasm tinygo build -o cmd/workers/who-dat.wasm -target wasm ./cmd/workers/main.go; \
		echo "WASM binary created at cmd/workers/who-dat.wasm"; \
	else \
		echo "Error: TinyGo is not installed. Install from https://tinygo.org/"; \
		exit 1; \
	fi

# Frontend
frontend: ## Build frontend assets
	@echo "Building frontend..."
	npm install
	npm run build

# Testing
test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./tests/unit/...
	go test -v -race ./tests/integration/...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage.out ./tests/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

coverage: test ## Generate test coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "Warning: golangci-lint not installed. Skipping..."; \
	fi

fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w $(GO_FILES)

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Deployment
deploy-vercel: ## Deploy to Vercel
	@echo "Deploying to Vercel..."
	@if [ ! -f deployments/vercel.json ]; then \
		cp deployments/vercel.json vercel.json; \
	fi
	vercel --prod

deploy-docker: build-docker ## Deploy Docker container
	@echo "Docker image built. Push to your registry:"
	@echo "  docker tag $(BINARY_NAME):$(VERSION) your-registry/$(BINARY_NAME):$(VERSION)"
	@echo "  docker push your-registry/$(BINARY_NAME):$(VERSION)"

deploy-workers: build-workers ## Deploy to Cloudflare Workers
	@echo "Deploying to Cloudflare Workers..."
	cd deployments && wrangler deploy

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -f cmd/workers/who-dat.wasm
	go clean

clean-all: clean ## Clean everything including node_modules
	rm -rf node_modules/

# Run
run: build-server ## Build and run the server
	@echo "Starting server..."
	./bin/$(BINARY_NAME)

# Docker targets
docker-run: build-docker ## Build and run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):latest

docker-shell: build-docker ## Open shell in Docker container
	docker run -it --entrypoint /bin/sh $(BINARY_NAME):latest

# Utilities
mod-tidy: ## Tidy Go modules
	go mod tidy

mod-update: ## Update Go dependencies
	go get -u ./...
	go mod tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# CI/CD
ci: lint vet test ## Run CI checks (lint, vet, test)
	@echo "All CI checks passed!"

.DEFAULT_GOAL := help
