#!/bin/bash

echo "=== Anti-Nuke Production Mode ==="

if [ ! -f .env ]; then
    echo "Error: .env file not found"
    exit 1
fi

export $(cat .env | xargs)

echo "Building optimized binary..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/antinuke ./cmd/main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Setting CPU affinity and permissions..."
sudo setcap 'cap_sys_nice,cap_ipc_lock=+ep' bin/antinuke

echo "Starting production engine..."
mkdir -p logs snapshots

nohup ./bin/antinuke > logs/stdout.log 2>&1 &
echo $! > antinuke.pid

echo "Started with PID: $(cat antinuke.pid)"
echo "Logs: tail -f logs/antinuke.log"
