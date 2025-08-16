# Makefile for Olane Go

.PHONY: build test clean install deps example fmt lint vet check run

# Variables
BINARY_NAME=olane
BUILD_DIR=build
CMD_DIR=cmd/example
PKG_DIRS=./pkg/...

# Default target
all: check build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt $(PKG_DIRS)
	go fmt ./cmd/...

# Lint code
lint:
	golangci-lint run $(PKG_DIRS)
	golangci-lint run ./cmd/...

# Vet code
vet:
	go vet $(PKG_DIRS)
	go vet ./cmd/...

# Run all checks
check: fmt vet test

# Build the example binary
build: deps
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Run tests
test:
	go test -v $(PKG_DIRS)

# Run tests with coverage
test-coverage:
	go test -v -cover $(PKG_DIRS)
	go test -v -coverprofile=coverage.out $(PKG_DIRS)
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test -bench=. $(PKG_DIRS)

# Run the example
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run the example directly without building
run-dev:
	go run ./$(CMD_DIR)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install binary to $GOPATH/bin
install: deps
	go install ./$(CMD_DIR)

# Generate documentation
docs:
	godoc -http=:6060

# Update dependencies
update:
	go get -u ./...
	go mod tidy

# Initialize a new project using this as a template
init-project:
	@echo "Initializing new Olane Go project..."
	@read -p "Enter module name (e.g., github.com/user/project): " MODULE_NAME; \
	find . -name "*.go" -exec sed -i.bak "s|github.com/olane-labs/olane-go|$$MODULE_NAME|g" {} \; && \
	find . -name "*.bak" -delete && \
	echo "module $$MODULE_NAME" > go.mod && \
	echo "" >> go.mod && \
	cat go.mod.template >> go.mod 2>/dev/null || true && \
	echo "Project initialized with module name: $$MODULE_NAME"

# Show help
help:
	@echo "Available targets:"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo "  vet            - Vet code"
	@echo "  check          - Run fmt, vet, and test"
	@echo "  build          - Build the example binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  bench          - Run benchmarks"
	@echo "  run            - Build and run the example"
	@echo "  run-dev        - Run the example without building"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  docs           - Start documentation server"
	@echo "  update         - Update dependencies"
	@echo "  init-project   - Initialize a new project using this as template"
	@echo "  help           - Show this help message"
