# Hotdesk Booking API - Backend

Go backend API for the Hotdesk Booking System.

## Tech Stack

- **Language**: Go 1.25+
- **Framework**: Fiber v2
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Migrations**: Goose v3.26+
- **Logger**: Zap (structured logging)
- **Authentication**: JWT

## Prerequisites

- Go 1.22 or higher
- PostgreSQL 15 (via Docker or local)
- Redis 7 (via Docker or local)
- Goose CLI (`brew install goose`)

## Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Environment Variables

Copy the example environment file and update as needed:

```bash
cp .env.example .env
```

Key environment variables:
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `JWT_SECRET`: Secret key for JWT tokens
- `BACKEND_PORT`: API server port (default: 8080)

### 3. Start Dependencies (Docker)

From the project root:

```bash
docker-compose up -d
```

This starts PostgreSQL and Redis containers.

### 4. Run Database Migrations

```bash
make migrate-up
```

## Database Migrations

This project uses [Goose](https://github.com/pressly/goose) for database migrations.

### Migration Commands

All migration commands are available via the Makefile:

```bash
# Show all available commands
make help

# Run all pending migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Show migration status
make migrate-status

# Show current migration version
make migrate-version

# Create a new migration
make migrate-create NAME=create_users_table

# Reset database (rollback all, then run all)
make migrate-reset
```

### Creating a New Migration

1. Create a new migration file:
```bash
make migrate-create NAME=create_users_table
```

2. Edit the generated file in `migrations/`:
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

3. Run the migration:
```bash
make migrate-up
```

### Migration Best Practices

- **Always test rollbacks**: Ensure your `-- +goose Down` section properly reverses the `-- +goose Up` changes
- **Use transactions**: Wrap complex migrations in `StatementBegin` and `StatementEnd`
- **Keep migrations small**: One logical change per migration
- **Never edit applied migrations**: Create a new migration to modify existing schema
- **Version control**: Commit migration files to git

### Troubleshooting

**Issue**: `database "hotdesk_booking" does not exist`

**Solution**: Ensure Docker containers are running and create the database:
```bash
docker exec hotdesk-postgres psql -U postgres -c "CREATE DATABASE hotdesk_booking;"
```

**Issue**: Connection refused to localhost:5432

**Solution**:
1. Check if Docker containers are running: `docker ps`
2. Ensure no local PostgreSQL is conflicting: `lsof -i :5432`
3. If a local PostgreSQL is running, stop it or use a different port

## Running the Server

### Development

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`.

### Health Check

Test the API is running:

```bash
curl http://localhost:8080/api/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2026-01-26T12:00:00Z",
  "services": {
    "database": "not_connected",
    "redis": "not_connected"
  }
}
```

## Code Quality

### Linting

Run golangci-lint:

```bash
golangci-lint run ./...
```

Configuration is in `.golangci.yml`.

### Formatting

```bash
# Format code
gofmt -w .

# Format imports
goimports -w .
```

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── handlers/                # HTTP handlers
│   ├── middleware/              # HTTP middleware
│   ├── models/                  # Data models
│   └── database/                # Database connection
├── migrations/                  # Database migrations
├── .env.example                 # Example environment variables
├── .golangci.yml                # Linter configuration
├── Makefile                     # Common commands
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## API Documentation

### Health Check
- **GET** `/api/health` - Returns API health status

More endpoints will be documented as they are implemented.

## License

Private - Not licensed for public use.
