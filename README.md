# RESTful API for User Management

## Project Structure

This is the skeleton of the RESTful application. The tests demonstrate the testing approach, but they do not exhaust the topic.

```
├── bin/                     # Build binaries
├── cmd/
│   ├── api/                 # API server entry point
│   │   └── main.go
│   └── migrate/             # Database migration tool
│       └── main.go
├── migrations/              # Database migration files
│   ├── 000001_*.up.sql
│   └── 000001_*.down.sql
├── internal/
│   ├── config/              # Configuration handling
│   ├── domain/              # Domain models
│   ├── handler/             # HTTP handlers
│   ├── logger/              # Logging utilities
│   ├── middleware/          # HTTP middleware
│   └── storage/             # Data storage layer
├── .golangci.yml            # Linter configuration
├── Dockerfile               # Docker build definition
├── docker-compose.yml       # Docker Compose services
├── Makefile                 # Build and run commands
├── README.md
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL database
- Make (for Makefile commands)
- Docker and Docker Compose (optional, for containerization)

The project includes a Makefile with helpful commands:
```bash
make help
```

## API Examples

### Create a User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Get a User

```bash
curl -X GET http://localhost:8080/api/users/1
```

### Update a User

```bash
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json"
  -d '{
    "first_name": "Johnny",
    "last_name": "Doe"
  }'
```

### Delete a User

```bash
curl -X DELETE http://localhost:8080/api/users/1
```

### List Users with Pagination

```bash
curl -X GET "http://localhost:8080/api/users?page=1&page_size=10"
```
