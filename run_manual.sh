#!/bin/bash

# B3 Trade Aggregator - Manual Run Script
# This script runs both the web application and CLI without Docker
# It automatically handles .env file creation and loads environment variables

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to start PostgreSQL if not running
start_postgres() {
    print_status "Checking if PostgreSQL is running..."
    
    # Check if PostgreSQL container is running
    if docker compose ps postgres | grep -q "Up"; then
        print_success "PostgreSQL is already running"
        return 0
    fi
    
    print_status "Starting PostgreSQL with Docker Compose..."
    
    # Start PostgreSQL
    if docker compose up -d postgres; then
        print_success "PostgreSQL started successfully"
        
        # Wait for PostgreSQL to be ready
        print_status "Waiting for PostgreSQL to be ready..."
        local max_attempts=30
        local attempt=1
        
        while [ $attempt -le $max_attempts ]; do
            if docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
                print_success "PostgreSQL is ready to accept connections"
                return 0
            fi
            
            print_status "Waiting for PostgreSQL... (attempt $attempt/$max_attempts)"
            sleep 2
            attempt=$((attempt + 1))
        done
        
        print_warning "PostgreSQL may not be fully ready, but continuing..."
        return 0
    else
        print_error "Failed to start PostgreSQL"
        return 1
    fi
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if Go is installed
check_go() {
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.24+ first."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go version detected: $GO_VERSION"
}

# Function to handle .env file
setup_env_file() {
    if [ ! -f .env ]; then
        if [ -f local-env.txt ]; then
            print_warning ".env file not found, renaming local-env.txt to .env"
            cp local-env.txt .env
            print_success "Created .env file from local-env.txt"
        else
            print_error "Neither .env nor local-env.txt file found!"
            print_error "Please create a .env file with your environment variables."
            exit 1
        fi
    else
        print_success ".env file found"
    fi
    
    # Load environment variables
    print_status "Loading environment variables from .env file..."
    
    # Read and export each line from .env file
    while IFS= read -r line || [ -n "$line" ]; do
        # Skip empty lines and comments
        if [[ -n "$line" && ! "$line" =~ ^[[:space:]]*# ]]; then
            # Export the variable
            export "$line"
            print_status "Exported: $line"
        fi
    done < .env
    
    # Verify essential variables
    if [ -z "$DATABASE_URL" ] && [ -z "$SRV_DB_HOST" ]; then
        print_warning "DATABASE_URL not set, will use individual database variables"
    fi
    
    if [ -z "$API_PORT" ]; then
        export API_PORT=8080
        print_warning "API_PORT not set, using default: 8080"
    fi
    
    # Debug: Show key environment variables
    print_status "Environment variables loaded:"
    print_status "  SRV_DB_HOST: ${SRV_DB_HOST:-'not set'}"
    print_status "  SRV_DB_PORT: ${SRV_DB_PORT:-'not set'}"
    print_status "  SRV_DB_NAME: ${SRV_DB_NAME:-'not set'}"
    print_status "  SRV_DB_USER: ${SRV_DB_USER:-'not set'}"
    print_status "  SRV_DB_SSL_MODE: ${SRV_DB_SSL_MODE:-'not set'}"
    print_status "  API_PORT: ${API_PORT:-'not set'}"
    print_status "  FILE_PATH: ${FILE_PATH:-'not set'}"
}

# Function to build applications
build_applications() {
    print_status "Building applications..."
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Build web application
    print_status "Building web application..."
    go build -o bin/app cmd/app/main.go
    if [ $? -eq 0 ]; then
        print_success "Web application built successfully"
    else
        print_error "Failed to build web application"
        exit 1
    fi
    
    # Build CLI application
    print_status "Building CLI application..."
    go build -o bin/ingest cmd/ingest/main.go
    if [ $? -eq 0 ]; then
        print_success "CLI application built successfully"
    else
        print_error "Failed to build CLI application"
        exit 1
    fi
}

# Function to check PostgreSQL connection
check_postgres() {
    print_status "Checking PostgreSQL connection..."
    
    # Extract database connection details
    if [ -n "$DATABASE_URL" ]; then
        DB_HOST=$(echo "$DATABASE_URL" | sed -n 's/.*@\([^:]*\):.*/\1/p')
        DB_PORT=$(echo "$DATABASE_URL" | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
        DB_NAME=$(echo "$DATABASE_URL" | sed -n 's/.*\/\([^?]*\).*/\1/p')
    else
        DB_HOST=${SRV_DB_HOST:-localhost}
        DB_PORT=${SRV_DB_PORT:-5432}
        DB_NAME=${SRV_DB_NAME:-b3_trade_aggregator}
    fi
    
    # Check if PostgreSQL is running
    if command_exists pg_isready; then
        if pg_isready -h "$DB_HOST" -p "$DB_PORT" >/dev/null 2>&1; then
            print_success "PostgreSQL is accessible at $DB_HOST:$DB_PORT"
        else
            print_warning "PostgreSQL is not accessible at $DB_HOST:$DB_PORT"
            print_warning "Make sure PostgreSQL is running before starting the applications"
        fi
    else
        print_warning "pg_isready not available, skipping PostgreSQL connection check"
    fi
}

# Function to run CLI application
run_cli() {
    print_status "Starting CLI application for data ingestion..."
    
    # Ensure SSL mode is set for local development
    if [ "$SRV_DB_SSL_MODE" = "require" ]; then
        export SRV_DB_SSL_MODE=disable
        print_warning "Changed SRV_DB_SSL_MODE from 'require' to 'disable' for local development"
    fi
    
    # Debug: Show environment variables before running CLI
    print_status "CLI environment variables:"
    print_status "  SRV_DB_HOST: $SRV_DB_HOST"
    print_status "  SRV_DB_PORT: $SRV_DB_PORT"
    print_status "  SRV_DB_NAME: $SRV_DB_NAME"
    print_status "  SRV_DB_USER: $SRV_DB_USER"
    print_status "  SRV_DB_SSL_MODE: $SRV_DB_SSL_MODE"
    print_status "  FILE_PATH: $FILE_PATH"
    
    # Run CLI with progress tracking
    if [ -n "$FILE_PATH" ]; then
        print_status "Processing file: $FILE_PATH"
        ./bin/ingest -file "$FILE_PATH"
    else
        print_status "No FILE_PATH specified, CLI will use default from .env"
        ./bin/ingest
    fi
    
    CLI_EXIT_CODE=$?
    if [ $CLI_EXIT_CODE -eq 0 ]; then
        print_success "CLI application completed successfully"
    else
        print_warning "CLI application exited with code $CLI_EXIT_CODE"
    fi
}

# Function to run web application
run_web() {
    print_status "Starting web application on port $API_PORT..."
    
    # Ensure SSL mode is set for local development
    if [ "$SRV_DB_SSL_MODE" = "require" ]; then
        export SRV_DB_SSL_MODE=disable
        print_warning "Changed SRV_DB_SSL_MODE from 'require' to 'disable' for local development"
    fi
    
    # Debug: Show environment variables before running web app
    print_status "Web app environment variables:"
    print_status "  API_PORT: $API_PORT"
    print_status "  SRV_MODE: $SRV_MODE"
    print_status "  SRV_DB_HOST: $SRV_DB_HOST"
    print_status "  SRV_DB_PORT: $SRV_DB_PORT"
    print_status "  SRV_DB_NAME: $SRV_DB_NAME"
    print_status "  SRV_DB_USER: $SRV_DB_USER"
    print_status "  SRV_DB_SSL_MODE: $SRV_DB_SSL_MODE"
    
    # Run web application in background
    ./bin/app &
    WEB_PID=$!
    
    # Wait a moment for the application to start
    sleep 2
    
    # Check if the application is running
    if kill -0 $WEB_PID 2>/dev/null; then
        print_success "Web application started successfully (PID: $WEB_PID)"
        print_success "Web application is running on http://localhost:$API_PORT"
        print_status "Press Ctrl+C to stop the web application"
        
        # Wait for the web application
        wait $WEB_PID
    else
        print_error "Failed to start web application"
        exit 1
    fi
}

# Function to cleanup on exit
cleanup() {
    print_status "Cleaning up..."
    
    # Kill web application if it's running
    if [ -n "$WEB_PID" ] && kill -0 $WEB_PID 2>/dev/null; then
        print_status "Stopping web application (PID: $WEB_PID)..."
        kill $WEB_PID 2>/dev/null || true
    fi
    
    print_success "Cleanup completed"
}

# Set trap to cleanup on script exit
trap cleanup EXIT INT TERM

# Main execution
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  B3 Trade Aggregator Runner${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    # Check prerequisites
    check_go
    
    # Setup environment
    setup_env_file
    
    # Start PostgreSQL if not running
    start_postgres
    
    # Check database connection
    check_postgres
    
    # Build applications
    build_applications
    
    echo ""
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Starting Applications${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    
    # Ask user which application to run
    echo "Which application would you like to run?"
    echo "1) Web Application (API Server)"
    echo "2) CLI Application (Data Ingestion)"
    echo "3) Both (CLI first, then Web)"
    echo "4) Exit"
    echo ""
    read -p "Enter your choice (1-4): " choice
    
    case $choice in
        1)
            print_status "Running web application only..."
            run_web
            ;;
        2)
            print_status "Running CLI application only..."
            run_cli
            ;;
        3)
            print_status "Running both applications..."
            print_status "First: CLI application for data ingestion"
            run_cli
            echo ""
            print_status "Second: Web application for API access"
            run_web
            ;;
        4)
            print_status "Exiting..."
            exit 0
            ;;
        *)
            print_error "Invalid choice. Please run the script again."
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
