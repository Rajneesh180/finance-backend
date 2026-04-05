# finance-backend

Backend for a finance dashboard — handles users, financial records, and role-based access control.

Built with Go, PostgreSQL, and chi router. Deployed on Railway.

**Live API:** https://finance-backend-production-b3d0.up.railway.app

## What it does

- User registration/login with JWT tokens
- Three roles (viewer, analyst, admin) with middleware-enforced permissions
- CRUD for financial records with ownership checks
- Dashboard endpoints that return aggregated data (totals, category breakdown, monthly trends)
- Filtering, pagination, and text search on records
- Soft deletes so nothing is permanently lost

## Tech

- **Go 1.25** — chi for routing, pgx for Postgres, shopspring/decimal for money (no floats)
- **PostgreSQL 16** — NUMERIC(15,2) for amounts, partial indexes, UUID primary keys
- **JWT (HS256)** with bcrypt password hashing
- Auto-migrations on startup using embedded SQL

## Project layout

```
cmd/server/          Entry point, router setup
internal/
  api/               Response helpers, validation
  config/            Env-based config
  database/          Postgres pool + auto-migrations
  domain/            Structs, DTOs, enums
  handler/           HTTP handlers (one per resource)
  middleware/        Auth, RBAC, CORS, rate limiting, logging
  repository/        SQL queries (raw pgx, no ORM)
  service/           Business logic layer
  testutil/          Mocks for testing
docs/                OpenAPI spec
```

## Running locally

### With Docker (easiest)

```bash
cp .env.example .env   # edit DB creds and JWT_SECRET
docker compose up --build -d
```

### Without Docker

Needs Go 1.25+ and a running Postgres instance.

```bash
cp .env.example .env
# set DATABASE_URL or individual DB_* vars in .env
make run
```

Migrations run automatically on startup — no separate step needed.

### Tests

```bash
make test
```

All tests use in-memory mock repos, no database required.

## Config

Set these in `.env` or as environment variables:

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | yes* | Full Postgres connection string |
| `JWT_SECRET` | yes | Secret for signing tokens |
| `PORT` | no | Server port (default: 8080) |
| `JWT_EXPIRY_MINUTES` | no | Token lifetime (default: 15) |

*If `DATABASE_URL` isn't set, the server builds it from `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`. See `.env.example`.

## API

All responses use `{"success": true, "data": ...}` or `{"success": false, "error": {"code": "...", "message": "..."}}`.

Full OpenAPI 3.0 spec in [docs/openapi.yaml](docs/openapi.yaml) — paste it into [Swagger Editor](https://editor.swagger.io) to browse interactively.

### Auth

| Method | Path | Description |
|---|---|---|
| POST | `/auth/register` | Create account (email, password, name) |
| POST | `/auth/login` | Returns JWT token + user object |

Register takes `{"email": "...", "password": "...", "name": "..."}`. Password must be at least 8 chars.

Login returns `{"token": "eyJ...", "user": {...}}`. Pass the token as `Authorization: Bearer <token>` on all other endpoints.

### Profile

`GET /users/me` and `PUT /users/me` — get or update the current user's name/email.

### Records

| Method | Path | Role |
|---|---|---|
| GET | `/records` | any |
| GET | `/records/{id}` | any |
| POST | `/records` | analyst, admin |
| PUT | `/records/{id}` | analyst, admin |
| DELETE | `/records/{id}` | analyst, admin |

Users only see their own records (admins see everything).

Body for create/update:

```json
{
  "amount": "1500.50",
  "type": "income",
  "category": "salary",
  "date": "2025-04-01",
  "description": "monthly salary"
}
```

Listing supports pagination (`page`, `per_page`), filtering (`type`, `category`, `date_from`, `date_to`), text search (`search`), and sorting (`sort_by`, `sort_order`).

### Dashboard

`GET /dashboard/summary` — totals, category breakdown, monthly trends. Accepts `date_from`/`date_to` filters.
`GET /dashboard/recent` — last N records (default 10, max 50).

Admins see aggregated data across all users.

### Admin

`GET /admin/users`, `GET /admin/users/{id}`, `PUT /admin/users/{id}`, `DELETE /admin/users/{id}` — user management, admin-only. Can change roles and deactivate accounts.

## Roles

- **viewer** — read-only access to own records and dashboard
- **analyst** — can also create/update/delete own records
- **admin** — full access, including user management

New accounts start as `viewer`.

## Seed data

Migration 003 creates three test users (password: `password123`): `admin@zorvyn.com`, `analyst@zorvyn.com`, `viewer@zorvyn.com`. Also seeds a few sample records for the analyst.

## Makefile

`make run`, `make test`, `make lint`, `make docker-up`, `make docker-down` — the usual. See the [Makefile](Makefile) for the full list.

## Design decisions

Using `shopspring/decimal` instead of `float64` for money. Floats accumulate rounding errors on aggregation and that's a non-starter for anything financial. Amounts go in as `NUMERIC(15,2)` in Postgres and come out as strings in JSON so nothing gets silently truncated.

JWT is HS256 with a configurable expiry (default 15 min). No refresh tokens — would need a token store and revocation logic that felt out of scope. If I were extending this I'd add refresh rotation, but for now a short-lived access token is enough.

Soft deletes on everything. Records and users get a `deleted_at` timestamp instead of being removed. Slightly annoying because every query needs a `WHERE deleted_at IS NULL`, but you keep full audit history and don't break foreign keys. There's a partial index to keep it fast.

Three roles: viewer, analyst, admin. Simple additive model — viewer can read, analyst can write, admin can do everything including manage users. No fine-grained permissions. Works for this scope.

Rate limiting is in-memory per IP (token bucket in a `sync.Map`). Resets on restart, doesn't work across instances. Redis would be the real answer for horizontal scaling, but adding an external dependency for a single-instance demo didn't feel worth it.

Raw SQL via pgx, no ORM. Mostly because the record listing query builds filters dynamically and I wanted full control over the SQL. Means more boilerplate for simple CRUD but fewer surprises.

UUIDs as primary keys to avoid enumeration. Slightly larger indexes but the dataset is small enough that it doesn't matter.
