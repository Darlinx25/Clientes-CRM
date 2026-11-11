---
description: Add a new SQLite migration with checksums and an upgrade rehearsal
agent: build
---
Create the next database migration for this repository.

Request: $ARGUMENTS

Follow `db-upgrade-safety` and `docs/upgrades.md`:

1. Inspect the current highest migration number and the relevant models.
2. Create the next `.up.sql`/`.down.sql` pair in
   `backend/database/migrations/` without touching existing files.
3. Write SQL that preserves existing rows and works on SQLite.
4. Run `make -C backend migrations-checksum`.
5. Add or update an upgrade assertion if data shape changes.
6. Run `make migrations-lint`, `make test-upgrade`, and `make backend-test`.
7. Report the new migration number, the data-preservation reasoning, and the
   exact test results.
