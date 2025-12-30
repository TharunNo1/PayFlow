# Variables
BINARY_NAME=payflow
DB_URL=postgres://user:password@localhost:5432/payflow?sslmode=disable
REDIS_URL=localhost:6379

.PHONY: all build run docker-up docker-down test race-test cover audit clean

all: build

## docker-up: Start PostgreSQL and Redis
docker-up:
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 3

## build: Build the API and Audit binaries
build:
	go build -o bin/$(BINARY_NAME) cmd/api/main.go
	go build -o bin/audit cmd/audit/main.go

## run: Run the API server
run: docker-up
	DATABASE_URL=$(DB_URL) REDIS_ADDR=$(REDIS_URL) go run cmd/api/main.go

## audit: Run the reconciliation script
audit:
	go run cmd/audit/main.go

## test: Run all Go tests
test:
	go test -v ./...

# Run all tests with the race detector
race-test:
	@echo "Running all tests with race detector..."
	CGO_ENABLED=1 go test -v -race ./...

# Run tests and show coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## docker-down: Stop all services
docker-down:
	docker-compose down

test-clean:
	@echo "Cleaning test environment..."
	docker exec -it $$(docker ps -qf "name=redis") redis-cli flushall
	docker exec -it $$(docker ps -qf "name=db") psql -U user -d payflow -c "TRUNCATE accounts, entries, payout_tasks CASCADE;"

## clean: Remove binaries
clean:
	rm -rf bin/