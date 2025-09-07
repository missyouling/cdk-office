# CDK-Office Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt

# Binary name
BINARY_NAME=cdk-office
BINARY_UNIX=$(BINARY_NAME)_unix

# Directories
CMD_DIR=cmd/server
INTERNAL_DIR=internal
PKG_DIR=pkg
FRONTEND_DIR=frontend

# Test coverage
COVERAGE_REPORT=coverage.out
COVERAGE_HTML=coverage.html

# Default target
all: build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(CMD_DIR)/main.go

# Build for Unix/Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(CMD_DIR)/main.go

# Install dependencies
deps:
	$(GOMOD) tidy
	cd $(FRONTEND_DIR) && pnpm install

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(CMD_DIR)/main.go
	./$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -coverprofile=$(COVERAGE_REPORT) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_REPORT) -o $(COVERAGE_HTML)

# Run tests with coverage and show in browser
test-coverage-html:
	$(GOTEST) -coverprofile=$(COVERAGE_REPORT) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_REPORT)

# Vet the code
vet:
	$(GOVET) ./...

# Format the code
fmt:
	$(GOFMT) ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(COVERAGE_REPORT)
	rm -f $(COVERAGE_HTML)

# Lint the code
lint:
	golangci-lint run

# Install linter
install-lint:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "Available targets:"
	@echo "  all               - Build the application (default)"
	@echo "  build             - Build the application"
	@echo "  build-linux       - Build for Unix/Linux"
	@echo "  deps              - Install dependencies"
	@echo "  run               - Run the application"
	@echo "  test              - Run tests"
	@echo "  test-coverage     - Run tests with coverage"
	@echo "  test-coverage-html - Run tests with coverage and show in browser"
	@echo "  vet               - Vet the code"
	@echo "  fmt               - Format the code"
	@echo "  clean             - Clean build files"
	@echo "  lint              - Lint the code"
	@echo "  install-lint      - Install linter"
	@echo "  help              - Show this help"

.PHONY: all build build-linux deps run test test-coverage test-coverage-html vet fmt clean lint install-lint help