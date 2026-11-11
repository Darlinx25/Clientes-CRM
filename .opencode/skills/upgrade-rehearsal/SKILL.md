---
name: upgrade-rehearsal
description: Use when changing database schema, seeds, startup, backups, or release code and you need to prove an old Meerkat database still upgrades without row loss. Covers fixtures, test-upgrade, dirty databases, and restoring snapshots.
---

# Upgrade rehearsal

An upgrade is only proven against a database that already contains rows.

## Fast loop

```bash
make -C backend test-upgrade
```

This runs:

- every committed fixture in `backend/database/testdata/fixtures/` through
  `InitDB` to the latest version;
- row-count and spot-value checks for contacts, notes, reminders, activities,
  relationships and companies;
- the dirty-database refusal and snapshot test;
- the new-database-no-snapshot test;
- checksum and lint tests.

## Adding a fixture

Only add a fixture version that represents a real historical release shape.
Generate it from the embedded migrations, never from a developer's live DB:

```bash
cd backend
go run ./cmd/mkfixture -version 23 -out database/testdata/fixtures/v000023.db -force
go run ./cmd/mkfixture -version 39 -out database/testdata/fixtures/v000039.db -force
```

Then update `TestLatestMigrationVersion` only when you intentionally add a new
migration. Keep the fixture binary committed so CI can run without generating
it.

## Reading failures

- "dirty at version N": do not force. The snapshot path is in the error. Inspect
  the failed migration, fix forward with a new migration, and restore the
  snapshot if the live database must be recovered.
- "table X lost rows": the migration deleted data. Redesign it to preserve the
  rows or document a deliberate, user-approved destructive change.
- "checksum mismatch": a historical migration was edited. Revert it; add a new
  migration instead.

Snapshots contain only the SQLite database, never photos.
