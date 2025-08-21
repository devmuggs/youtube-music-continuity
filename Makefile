# YouTube Music Continuity Makefile

.PHONY: help dev backend-dev backend-build extension-build clean test

# Default target
help:
	@echo "YouTube Music Continuity - Available commands:"
	@echo ""
	@echo "  dev              - Start development environment"
	@echo "  backend-dev      - Run backend in development mode"
	@echo "  backend-build    - Build backend binary"
	@echo "  extension-build  - Package extension for distribution"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo ""

# Development environment
dev: backend-dev

# Run backend in development mode
backend-dev:
	@echo "Starting backend server..."
	cd backend && go run cmd/server/main.go

# Build backend binary
backend-build:
	@echo "Building backend..."
	cd backend && go build -o bin/server cmd/server/main.go
	@echo "Backend built: backend/bin/server"

# Package extension for distribution
extension-build:
	@echo "Packaging extension..."
	cd extension && zip -r ../youtube-music-continuity-extension.zip . -x "*.DS_Store" "*.git*"
	@echo "Extension packaged: youtube-music-continuity-extension.zip"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f backend/bin/server
	rm -f youtube-music-continuity-extension.zip
	rm -f backend/continuity.db

# Run tests
test:
	@echo "Running backend tests..."
	cd backend && go test ./...

# Install dependencies
deps:
	@echo "Installing backend dependencies..."
	cd backend && go mod tidy

# Run backend with hot reload (requires air)
backend-watch:
	@echo "Starting backend with hot reload..."
	cd backend && air

# Database setup
db-setup:
	@echo "Setting up database..."
	cd backend && touch continuity.db

# Check if required tools are installed
check-tools:
	@echo "Checking required tools..."
	@which go > /dev/null || (echo "Go is not installed" && exit 1)
	@echo "All required tools are available"