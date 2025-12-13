.PHONY: build run clean test docker fast

build:
	@echo "Building anti-nuke engine..."
	@go mod tidy
	@go build -o bin/antinuke ./cmd/main.go
	@echo "Build complete: bin/antinuke"

fast:
	@echo "Building ULTRA-OPTIMIZED anti-nuke engine..."
	@go mod tidy
	@go build -ldflags="-s -w" -gcflags="all=-l -B -C" -o bin/antinuke ./cmd/main.go
	@echo "Ultra-fast build complete: bin/antinuke"

run: build
	@echo "Starting engine..."
	@./bin/antinuke

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f *.log
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

docker:
	@echo "Building Docker image..."
	@docker build -t antinuke:latest .

docker-run:
	@echo "Running in Docker..."
	@docker-compose up -d

docker-stop:
	@docker-compose down

fmt:
	@go fmt ./...

vet:
	@go vet ./...

dev: build
	@./scripts/run_dev.sh

prod:
	@./scripts/run_prod.sh
