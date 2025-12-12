#!/bin/bash

echo "=== Building Anti-Nuke Engine ==="

echo "Checking Go version..."
go version

echo "Tidying dependencies..."
go mod tidy

echo "Running tests..."
go test ./... -v

if [ $? -ne 0 ]; then
    echo "Tests failed!"
    exit 1
fi

echo "Building binary..."
mkdir -p bin
go build -o bin/antinuke ./cmd/main.go

if [ $? -eq 0 ]; then
    echo "✓ Build successful: bin/antinuke"
    echo "File size: $(du -h bin/antinuke | cut -f1)"
else
    echo "✗ Build failed"
    exit 1
fi
