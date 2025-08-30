.PHONY: build run test clean docker-build docker-run deps migrate

# Build the application
build:
	go build -o bin/app cmd/app/main.go

# Run the application
run:
	go run cmd/app/main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Build Docker image
docker-build:
	docker build -t b3-trade-aggregator .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	docker-compose down

# Run Docker Compose with logs
docker-logs:
	docker-compose logs -f

# Run migrations
migrate:
	# Add migration commands here
	# Example: migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up

# Install dependencies
deps:
	go mod download
	go mod tidy

# Create data directory
setup:
	mkdir -p data
	mkdir -p bin

# Full setup: dependencies, build, and run
setup-full: setup deps build

# Development with hot reload (requires air or similar tool)
dev:
	# Install air if not present: go install github.com/cosmtrek/air@latest
	air

# Database operations
db-reset:
	docker-compose down -v
	docker-compose up -d postgres
	sleep 5
	# Add migration commands here

# Performance test
perf-test:
	# Add performance testing commands here
	# Example: ab -n 1000 -c 10 http://localhost:8080/trades/aggregated?ticker=PETR4
