---
title: Testing
parent: Development
nav_order: 4
---

# Testing

## Backend Tests

### Running Tests

```sh
cd backend
go test ./...
```

### Test Database

Most tests use an in-memory SQLite database and GORM `AutoMigrate`. Tests that
exercise startup or upgrades use a temporary file from `t.TempDir()`. Never
point a test at `data/meerkat.db`.

```go
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

### Upgrade tests

`backend/database/upgrade_test.go` embeds the committed fixtures in
`backend/database/testdata/fixtures/`, upgrades each one through the production
`InitDB`, and checks row counts, spot values, snapshots, dirty-schema refusal and
newer-schema refusal.

```sh
make test-upgrade
```

### Writing Tests

Use the test helpers in the `_test` files alongside each package. Create a Gin
test context and call controller functions directly. Assert on the
`httptest.ResponseRecorder`.

## Frontend Tests

Frontend unit tests use Vitest and React Testing Library:

```sh
cd frontend
corepack yarn test
```

Playwright covers integrated E2E flows against the Docker test stack.

### Setup

```sh
cd frontend
yarn playwright install  # install browsers (first time only)
```

The E2E tests expect the full application running on `http://localhost:7300`. `global-setup.ts` seeds a test user before the suite runs.

### Running E2E Tests

```sh
yarn test:e2e           # headless
yarn test:e2e:headed    # with browser visible
yarn test:e2e:ui        # Playwright UI mode
yarn test:e2e:debug     # debug mode
```

### Writing E2E Tests

Tests live in `frontend/e2e/`. Use the shared fixtures from `fixtures.ts` for login/logout helpers and the `TEST_USER` credentials from `global-setup.ts`. Group tests with `test.describe` and keep each test independent.
