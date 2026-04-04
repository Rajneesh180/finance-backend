# Finance Data Processing and Access Control Backend

REST API for managing financial records with role-based access control, built with Go and PostgreSQL.

## Features

- JWT authentication with bcrypt password hashing
- Role-based access control (admin, analyst, viewer)
- Financial record CRUD with ownership enforcement
- Dashboard aggregations (totals, category breakdown, monthly trends)
- Paginated listing with filtering and sorting
- Soft deletes across all entities
- IP-based rate limiting (token bucket)
- CORS support
- Structured JSON logging
- Docker and docker-compose setup

## Tech Stack

- **Go 1.25** with [chi](https://github.com/go-chi/chi) router
- **PostgreSQL 16** with [pgx](https://github.com/jackc/pgx) driver
- **JWT** via [golang-jwt](https://github.com/golang-jwt/jwt)
- **Decimal** via [shopspring/decimal](https://github.com/shopspring/decimal) (no floating point for money)
- **UUID** primary keys via [google/uuid](https://github.com/google/uuid)

## Project Structure

```
cmd/server/          - Application entry point
internal/
  api/               - JSON response helpers, shared validator
  config/            - Environment-based configuration
  database/          - PostgreSQL connection pool
  domain/            - Types, DTOs, enums
  handler/           - HTTP handlers
  middleware/         - Auth, RBAC, CORS, rate limiting, logging
  repository/        - Database queries
  service/           - Business logic
  testutil/          - Mock repositories for tests
migrations/          - SQL migration files
```

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 16+ (or Docker)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for manual migrations)

### Quick Start with Docker

```bash
docker compose up --build -d
```

This starts PostgreSQL and the API server on port 8080. Run migrations manually after the containers are ready:

```bash
make migrate-up
```

### Local Development

1. Copy environment file and edit as needed:

```bash
cp .env.example .env
```

2. Start PostgreSQL (or use docker-compose for just the database):

```bash
docker compose up db -d
```

3. Run migrations:

```bash
make migrate-up
```

4. Build and run:

```bash
make run
```

The server starts on `http://localhost:8080`.

### Running Tests

```bash
make test
```

Tests use mock repositories and run without a database.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `DATABASE_URL` | — | Full connection string (overrides individual DB_ vars) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `finance` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | — | **Required.** Secret for signing JWTs |
| `JWT_EXPIRY_MINUTES` | `15` | Token expiry in minutes |

## API Endpoints

All responses follow the format:

```json
{
  "success": true,
  "data": { ... },
  "meta": { ... }
}
```

Error responses:

```json
{
  "success": false,
  "error": { "code": "not_found", "message": "record not found" }
}
```

### Authentication

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | No | Create account |
| POST | `/auth/login` | No | Get JWT token |

**Register:**

```json
{
  "email": "user@example.com",
  "password": "minimum8chars",
  "name": "Jane Doe"
}
```

**Login:**

```json
{
  "email": "user@example.com",
  "password": "minimum8chars"
}
```

Returns `{ "token": "eyJ...", "user": { ... } }`.

### User Profile

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/users/me` | Yes | Get current user |
| PUT | `/users/me` | Yes | Update name/email |

### Financial Records

| Method | Path | Auth | Role | Description |
|---|---|---|---|---|
| GET | `/records` | Yes | Any | List records (paginated) |
| GET | `/records/{id}` | Yes | Any | Get single record |
| POST | `/records` | Yes | analyst, admin | Create record |
| PUT | `/records/{id}` | Yes | analyst, admin | Update record |
| DELETE | `/records/{id}` | Yes | analyst, admin | Delete record |

Users can only access their own records. Admins bypass ownership checks.

**Create/Update body:**

```json
{
  "amount": "1500.50",
  "type": "income",
  "category": "salary",
  "date": "2025-04-01",
  "description": "monthly salary"
}
```

**Query parameters for listing:**

| Param | Description |
|---|---|
| `page` | Page number (default: 1) |
| `per_page` | Items per page (default: 20, max: 100) |
| `type` | Filter by `income` or `expense` |
| `category` | Filter by category |
| `date_from` | Start date (YYYY-MM-DD) |
| `date_to` | End date (YYYY-MM-DD) |
| `sort_by` | Sort field: `date`, `amount`, `created_at` |
| `sort_order` | `asc` or `desc` |

### Dashboard

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/dashboard/summary` | Yes | Aggregated financial summary |

Query parameters: `date_from`, `date_to`. Admins see all users' data.

Returns total income, total expenses, net balance, category breakdown, and monthly trends.

### Admin

| Method | Path | Auth | Role | Description |
|---|---|---|---|---|
| GET | `/admin/users` | Yes | admin | List all users |
| GET | `/admin/users/{id}` | Yes | admin | Get user by ID |
| PUT | `/admin/users/{id}` | Yes | admin | Update role/active status |
| DELETE | `/admin/users/{id}` | Yes | admin | Soft delete user |

**Admin update body:**

```json
{
  "role": "analyst",
  "is_active": true
}
```

## Roles

| Role | Permissions |
|---|---|
| `viewer` | Read own records, view dashboard |
| `analyst` | All viewer permissions + create/update/delete own records |
| `admin` | Full access to all records and user management |

New accounts default to `viewer`. Only admins can change roles.

## Seed Data

Migration `003_seed_data` creates three test users (password: `password123`):

| Email | Role |
|---|---|
| admin@finance.local | admin |
| analyst@finance.local | analyst |
| viewer@finance.local | viewer |

It also creates 8 sample financial records for the analyst user.

## Makefile Targets

```
make build        # Build binary
make run          # Build and run
make test         # Run tests with race detector
make lint         # Run golangci-lint
make migrate-up   # Apply all migrations
make migrate-down # Roll back last migration
make docker-up    # Start all services
make docker-down  # Stop all services
```
