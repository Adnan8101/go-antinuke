#!/bin/bash

echo "Building Anti-Nuke Engine..."

if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

go mod tidy
go build -o bin/antinuke ./cmd/main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Starting engine..."
    ./bin/antinuke
else
    echo "Build failed!"
    exit 1
fi
