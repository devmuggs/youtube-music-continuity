#!/bin/bash

# YouTube Music Continuity Setup Script

set -e

echo "🎵 YouTube Music Continuity Setup"
echo "=================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher from https://golang.org/dl/"
    exit 1
fi

echo "✅ Go found: $(go version)"

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please upgrade to Go 1.21 or higher."
    exit 1
fi

echo "✅ Go version check passed"

# Navigate to project directory
cd "$(dirname "$0")"
PROJECT_DIR=$(pwd)

echo "📁 Project directory: $PROJECT_DIR"

# Setup backend
echo ""
echo "🔧 Setting up backend..."
cd backend

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Build the server
echo "🏗️  Building backend server..."
go build -o bin/server cmd/server/main.go

echo "✅ Backend setup complete"

# Go back to project root
cd "$PROJECT_DIR"

# Create extension package
echo ""
echo "📦 Packaging Chrome extension..."
make extension-build

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Start the backend server:"
echo "   make backend-dev"
echo ""
echo "2. Load the Chrome extension:"
echo "   - Open chrome://extensions/"
echo "   - Enable Developer mode"
echo "   - Click 'Load unpacked'"
echo "   - Select the 'extension/' directory"
echo ""
echo "3. Visit https://music.youtube.com and enjoy!"
echo ""
echo "📚 For more information, see:"
echo "   - docs/API.md - API documentation"
echo "   - docs/EXTENSION.md - Extension setup guide"
echo "   - README.md - Project overview"