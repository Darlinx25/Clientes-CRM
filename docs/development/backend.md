---
title: Backend Development
parent: Development
nav_order: 2
---

# Backend Development

## Setup

Requires current [Go version](https://go.dev/doc/install).

```sh
cd backend
cp .env.example .env   # then edit with your values
go mod tidy
```

## Running Locally

```sh
source .env
go run main.go
```

The server starts on `PORT` (default `8080`). Migrations run automatically.

## Adding a New Endpoint

1. Define an input DTO in `models/` with validation tags.
2. Add a controller function in `controllers/`.
3. Register the route in `routes/routes.go`.
4. Add the validation schema to the middleware registration if needed.

## Database Migrations

### Creating a Migration

```sh
make migrate-create NAME=add_foo_column
```

This creates two files in `database/migrations/`: `NNNNNN_add_foo_column.up.sql` and `.down.sql`.

### Running Migrations

Migrations run automatically on startup through `database.InitDB`, which
snapshots an existing database first and refuses a dirty schema. For local
development you can inspect status or apply pending migrations:

```sh
make migrate-status
make migrate-up
```

`make migrate-down` is for development only. Never force a dirty schema on a
user's database.

### Upgrade safety checks

After adding a migration:

```sh
make migrations-checksum   # append the new file to migrations.lock
make migrations-lint       # checksums, gaps, destructive SQL
make test-upgrade          # old fixtures -> latest schema, no row loss
```

Historical migrations and `migrations.lock` are immutable. See
[Upgrades](../upgrades.md) for the full contract.

## Models and DTOs

GORM models (`Contact`, `Activity`, etc.) live alongside input DTOs (`ContactInput`, etc.) in `models/`. DTOs are what controllers receive after validation while models are what GORM persists.

All models include `UserID uint` for tenant isolation.

## Services

Complex business logic lives in `services/`. Controllers call services; services own multi-step operations (e.g. sending emails). Services receive a `*gorm.DB` and any needed config, they do not access `*gin.Context`.

## Testing

```sh
go test ./...
```

Tests use an in-memory SQLite database (auto-migrated). See [Testing](testing.md).
