# Variables
BINARY_NAME=mockcraft
MAIN_PATH=main.go
BUILD_DIR=build
COVERAGE_DIR=coverage
LINT_TIMEOUT=5m

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint
GOFMT=gofmt

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test coverage lint fmt tidy help run debug

all: clean build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run:
	@echo "Running $(BINARY_NAME)..."
	@if [ "$(ARGS)" = "" ]; then \
		./$(BUILD_DIR)/$(BINARY_NAME); \
	else \
		./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS); \
	fi

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

lint:
	@echo "Running linter..."
	$(GOLINT) run --timeout=$(LINT_TIMEOUT) ./...

fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .

tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

deps:
	@echo "Installing dependencies..."
	$(GOGET) -v ./...

# Cross-compilation targets
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Development workflow
dev: fmt lint test build

# Help target
help:
	@echo "Available targets:"
	@echo "  all            - Clean and build the project"
	@echo "  build          - Build the project"
	@echo "  run            - Build and run the project"
	@echo "  run ARGS=\"...\" - Run with arguments (e.g., make run ARGS=\"generate allergy\")"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Tidy dependencies"
	@echo "  deps           - Install dependencies"
	@echo "  build-all      - Build for all platforms"
	@echo "  dev            - Run fmt, lint, test, and build"

debug:
	dlv debug main.go -- $(ARGS) 