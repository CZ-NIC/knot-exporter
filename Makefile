# Makefile for Knot DNS Exporter

# Build variables
VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GO_VERSION ?= $(shell go version | awk '{print $3}')

# Go build flags
LDFLAGS = -X main.version=$(VERSION) \
          -X main.buildTime=$(BUILD_TIME) \
          -X main.gitCommit=$(GIT_COMMIT) \
          -X main.goVersion=$(GO_VERSION)

# Build flags for CGO
CGO_ENABLED = 1
CGO_CFLAGS = -std=c99
CGO_LDFLAGS = -L/usr/lib64 -lknot

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) \
	CGO_CFLAGS="$(CGO_CFLAGS)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	go build -ldflags "$(LDFLAGS)" -o knot-exporter .

# Build with race detector
.PHONY: build-race
build-race:
	CGO_ENABLED=$(CGO_ENABLED) \
	CGO_CFLAGS="$(CGO_CFLAGS)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	go build -race -ldflags "$(LDFLAGS)" -o knot-exporter .

# Static build (if supported)
.PHONY: build-static
build-static:
	CGO_ENABLED=$(CGO_ENABLED) \
	CGO_CFLAGS="$(CGO_CFLAGS)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS) -static" \
	go build -ldflags "$(LDFLAGS) -extldflags '-static'" -o knot-exporter .

# Test
.PHONY: test
test:
	go test -v ./...

# Test with race detector
.PHONY: test-race
test-race:
	go test -race -v ./...

# Clean
.PHONY: clean
clean:
	rm -f knot-exporter

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Update dependencies
.PHONY: update-deps
update-deps:
	go get -u ./...
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Vet code
.PHONY: vet
vet:
	go vet ./...

# Check dependencies
.PHONY: check-deps
check-deps:
	@command -v pkg-config >/dev/null 2>&1 || { echo >&2 "pkg-config is required but not installed. Aborting."; exit 1; }
	@pkg-config --exists libknot || { echo >&2 "libknot development files are required but not found. Aborting."; exit 1; }
	@echo "Dependencies check passed"

# Install (requires root for systemd service)
.PHONY: install
install: build
	install -D -m 755 knot-exporter /usr/local/bin/knot-exporter

# Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"

# Development build with debug symbols
.PHONY: dev
dev:
	CGO_ENABLED=$(CGO_ENABLED) \
	CGO_CFLAGS="$(CGO_CFLAGS) -g" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	go build -gcflags="all=-N -l" -ldflags "$(LDFLAGS)" -o knot-exporter .

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  build-race  - Build with race detector"
	@echo "  build-static- Build static binary"
	@echo "  test        - Run tests"
	@echo "  test-race   - Run tests with race detector"
	@echo "  clean       - Remove built binary"
	@echo "  deps        - Download dependencies"
	@echo "  update-deps - Update dependencies"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code (requires golangci-lint)"
	@echo "  vet         - Vet code"
	@echo "  check-deps  - Check if required dependencies are installed"
	@echo "  install     - Install binary to /usr/local/bin/"
	@echo "  version     - Show version information"
	@echo "  dev         - Build development version with debug symbols"
	@echo "  help        - Show this help"
