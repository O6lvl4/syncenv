.PHONY: build install test clean run help

BINARY_NAME=syncenv
MAIN_PATH=./cmd/syncenv
VERSION?=dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	go build -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)

install: ## Install the binary to GOPATH/bin
	go install -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)

test: ## Run tests
	go test -v ./internal/...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total

test-race: ## Run tests with race detection
	go test -race -v ./internal/...

test-short: ## Run short tests only
	go test -short -v ./internal/...

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean

run: build ## Build and run the binary
	./$(BINARY_NAME)

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

vet: ## Run go vet
	go vet ./...

all: clean deps fmt vet test build ## Run all checks and build
