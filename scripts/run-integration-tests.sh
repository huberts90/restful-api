#!/bin/bash
set -e

# This script runs integration tests for the RESTful API with Docker Compose
# It's intended to be used both locally and in CI environments

# Ensure we clean up containers on exit
function cleanup {
  echo "Cleaning up containers..."
  docker-compose down
}
trap cleanup EXIT

# Start PostgreSQL container
echo "Starting PostgreSQL container..."
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
  if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo "PostgreSQL is ready!"
    break
  fi
  echo "Waiting for PostgreSQL to start... ($i/30)"
  sleep 1
  
  if [ $i -eq 30 ]; then
    echo "PostgreSQL failed to start in time."
    exit 1
  fi
done

# Run the migrations
echo "Running migrations..."
docker-compose up migrate

# Run the integration tests
echo "Running integration tests..."
POSTGRES_HOST=localhost \
POSTGRES_PORT=5432 \
POSTGRES_USER=postgres \
POSTGRES_PASSWORD=postgres \
POSTGRES_DB=users_db \
POSTGRES_SSLMODE=disable \
go test -v ./internal/handler -run TestUserHandlerIntegrationSuite

echo "Integration tests completed successfully!" 