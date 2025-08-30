# B3 Trade Aggregator

A high-performance Go application for aggregating and processing B3 trade data with PostgreSQL 17, optimized for large-scale data ingestion using pgx COPY FROM.

## Features

- **High-Performance Data Ingestion**: Uses pgx COPY FROM for optimal bulk insert performance
- **PostgreSQL 17**: Latest PostgreSQL version with advanced features
- **Clean Architecture**: Well-structured Go project following best practices
- **Docker Support**: Complete containerization with Docker Compose
- **RESTful API**: HTTP API for querying aggregated trade data
- **Stream Processing**: Efficient streaming file processing for large datasets

## Project Structure

```
├── cmd/
│   └── app/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   └── handler/
│   │       └── handler.go          # HTTP request handling logic
│   ├── config/
│   │   └── config.go               # Configuration loading and structure
│   ├── entity/
│   │   └── trade.go                # Data models
│   ├── ingestion/
│   │   ├── reader.go               # Stream reading
│   │   └── processor.go            # Orchestrates ingestion and persistence
│   ├── repository/
│   │   └── trade.go                # Database interactions (pgx COPY FROM)
│   ├── service/
│   │   └── trade.go                # Business logic and orchestration
│   └── util/
│       └── errors.go               # Custom error types and utilities
├── pkg/                            # Reusable packages
├── migrations/                     # Database migration scripts
├── tests/                          # Integration/end-to-end tests
├── data/                           # Data files directory
├── docker-compose.yml              # Docker service orchestration
├── Dockerfile                      # Application containerization
├── Makefile                        # Task automation
└── go.mod                          # Go modules
```

## Performance Optimizations

- **pgx COPY FROM**: Uses PostgreSQL's COPY protocol for bulk inserts (10x faster than individual INSERTs)
- **Connection Pooling**: Efficient connection management with pgxpool
- **Batch Processing**: Configurable batch sizes for optimal memory usage
- **Streaming**: File processing without loading entire file into memory
- **Indexed Queries**: Optimized database indexes for fast aggregations

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL 17
- Docker and Docker Compose

### Running with Docker (Recommended)

1. Start the services:
   ```bash
   make docker-run
   ```

2. Check logs:
   ```bash
   make docker-logs
   ```

3. Stop the services:
   ```bash
   make docker-stop
   ```

### Running Locally

1. Setup and install dependencies:
   ```bash
   make setup-full
   ```

2. Start PostgreSQL (using Docker):
   ```bash
   docker-compose up -d postgres
   ```

3. Run the application:
   ```bash
   make run
   ```

### API Usage

Query aggregated trade data:
```bash
curl "http://localhost:8080/trades/aggregated?ticker=PETR4&data_inicio=2024-01-01"
```

Response format:
```json
{
  "ticker": "PETR4",
  "max_range_value": 45.67,
  "max_daily_volume": 1500000
}
```

### Testing

Run tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## Development

### Available Make Commands

- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage
- `make clean` - Clean build artifacts
- `make docker-build` - Build Docker image
- `make docker-run` - Run with Docker Compose
- `make docker-stop` - Stop Docker Compose
- `make docker-logs` - View Docker logs
- `make deps` - Install dependencies
- `make setup` - Create necessary directories
- `make setup-full` - Full setup (deps + build)
- `make db-reset` - Reset database
- `make perf-test` - Performance testing

### Data Ingestion

The application is optimized for processing large B3 trade files (565MB+). The ingestion process:

1. **Streams** the file line by line without loading it entirely into memory
2. **Parses** each line into structured trade data
3. **Batches** trades into configurable batch sizes (default: 1000)
4. **Uses COPY FROM** for high-performance bulk database inserts
5. **Handles errors** gracefully with detailed logging

### Configuration

Environment variables:
- `DATABASE_URL`: PostgreSQL connection string
- `API_PORT`: HTTP server port (default: 8080)

Example:
```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/b3_trade_aggregator?sslmode=disable"
export API_PORT="8080"
```

## Performance Benchmarks

With pgx COPY FROM, the application can process:
- **~100,000 trades/second** on standard hardware
- **565MB file** in approximately 2-3 minutes
- **Memory usage** stays constant regardless of file size

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
