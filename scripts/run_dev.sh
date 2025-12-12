#!/bin/bash

echo "=== Anti-Nuke Development Mode ==="

export DISCORD_TOKEN=${DISCORD_TOKEN:-""}

if [ -f .env ]; then
    export $(cat .env | xargs)
fi

echo "Building..."
go build -o bin/antinuke ./cmd/main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Starting in development mode..."
echo "Logs: tail -f logs/antinuke.log"

mkdir -p logs
./bin/antinuke
