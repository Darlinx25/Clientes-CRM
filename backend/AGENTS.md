# Backend agent notes

Read the root `AGENTS.md` first. This file adds backend-specific rules.

## Database and upgrades

- `database.InitDB(dbPath, backupDir)` is the production entry point. It takes
  an advisory lock, snapshots the database before an upgrade, refuses dirty
  schemas, runs migrations, and verifies integrity with `PRAGMA quick_check`.
- `RunMigrations` must never call `m.Force(...)`.
- Historical migrations are immutable. Add `0000NN_description.up.sql` and
  `.down.sql`, then run `make -C backend migrations-checksum`.
- Destructive SQL is rejected from migration `000040` onward unless the file
  contains `-- migration:allow-destructive reason=...` and a fixture proves
  existing data survives.
- `backend/database/migrations.lock` stores SHA256 hashes. Do not hand-edit it.
- `SeedAdminUser` may create the first admin. It must never reset an existing
  admin password on startup.

## Tests

- `make backend-test` is the baseline.
- `make test-upgrade` runs the upgrade rehearsal.
- Fixtures live in `backend/database/testdata/fixtures/` and are embedded into
  the tests. Regenerate only with `make -C backend fixtures`.
- Controller tests use `:memory:` SQLite and a router built in the test. Follow
  the existing pattern instead of connecting to a real file.

## HTTP and services

- Handlers stay thin: bind/validate input, call a service, map the error.
- Follow the existing JSON error envelope and middleware conventions in
  `backend/routes/routes.go`.
- Never log secrets, passwords, reset tokens, or full request bodies that may
  contain credentials.
- Health responses must keep working when the database has no migration table.
