.PHONY: help test build clean install lint fmt vet coverage run-examples

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run tests
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

test-verbose: ## Run tests with verbose output
	go test -v ./...

test-coverage: ## Run tests and show coverage
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: ## Build the project
	go build -v ./...

build-cmd: ## Build conform-validate command
	go build -v -o bin/conform-validate ./cmd/conform-validate

install: ## Install conform-validate command
	go install ./cmd/conform-validate

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

coverage: test-coverage ## Alias for test-coverage

run-examples: ## Run all examples
	@echo "Running examples..."
	@for dir in examples/*/; do \
		echo "\n=== Running $$dir ==="; \
		cd $$dir && go run main.go && cd ../..; \
	done

run-basic: ## Run basic example
	cd examples/basic && go run main.go

run-generic: ## Run generic example
	cd examples/generic && go run main.go

run-fileconfig: ## Run fileconfig example
	cd examples/fileconfig && go run main.go

run-environment: ## Run environment example
	cd examples/environment && go run main.go

run-hotreload: ## Run hotreload example
	cd examples/hotreload && go run main.go

run-advanced: ## Run advanced example
	cd examples/advanced && go run main.go

run-realworld: ## Run realworld example
	cd examples/realworld && go run main.go

run-custom: ## Run custom example
	cd examples/custom && go run main.go

run-customvalidator: ## Run customvalidator example
	cd examples/customvalidator && go run main.go

run-errors: ## Run errors example
	cd examples/errors && go run main.go

check: fmt vet test ## Run all checks (fmt, vet, test)

.DEFAULT_GOAL := help

