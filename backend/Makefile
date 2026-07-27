.PHONY: dev build test lint docker-up docker-down

## Run the backend in development mode (requires .env)
dev:
	go run ./cmd/api/...

## Build the binary
build:
	go build -o bin/api ./cmd/api/...

## Run all tests
test:
	go test ./...

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## Start MongoDB via Docker Compose
docker-up:
	docker compose up -d

## Stop and remove Docker containers
docker-down:
	docker compose down
