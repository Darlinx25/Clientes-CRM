---
name: db-upgrade-safety
description: Use when adding or editing any SQLite schema, GORM model, migration, seed, backup, or upgrade code. Covers immutable migrations, pre-upgrade snapshots, dirty-schema refusal, checksums, destructive SQL, fixtures, and proving old databases upgrade without row loss.
---

# Database upgrade safety

The product rule is absolute: replacing the binary must open the previous
database without losing data. A migration is not done because it works on an
empty database.

## Workflow

1. Read `docs/upgrades.md` and the latest migrations before writing SQL.
2. Never edit an existing file in `backend/database/migrations/`. Create the
   next numbered pair (`0000NN_description.up.sql` and `.down.sql`).
3. Write SQL that works with existing rows: add columns with defaults or
   nullable values, backfill explicitly, then tighten constraints if needed.
4. Do not drop or rename user data casually. If unavoidable, add
   `-- migration:allow-destructive reason=...` and prove the data-preserving
   path with a fixture.
5. Record the new files and rehearse the upgrade:

   ```bash
   make -C backend migrations-checksum
   make test-upgrade
   ```

6. Run the gate:

   ```bash
   make migrations-lint
   make backend-test
   ```

## What the code already does

- `InitDB` takes a lock, snapshots the database with `VACUUM INTO` before an
  upgrade, refuses a dirty schema, runs migrations, then runs
  `PRAGMA quick_check`.
- The snapshot lives next to the configured backup directory as
  `pre-upgrade-vNNN-to-vMMM-*.db`.
- `migrations.lock` stores SHA256 for every migration file. Adding a new file
  is expected; changing an old hash is not.
- Destructive SQL is rejected from migration `000040` onward unless it carries
  the allow comment above.

## Fixtures

`backend/database/testdata/fixtures/` represents old releases with real rows.
Tests copy each fixture to a temporary directory, call the production `InitDB`,
and verify schema version, row counts, spot values, snapshots and dirty-schema
refusal. Regenerate fixtures only from embedded migrations:

```bash
make -C backend fixtures
```

Never generate a fixture from `data/meerkat.db`.

## Reading failures

- "dirty at version N": do not force. Inspect the failed migration and fix
  forward with a new migration.
- "table X lost rows": redesign the migration to preserve the rows or get an
  explicit product decision.
- "checksum mismatch": a historical migration was edited. Revert it and add a
  new migration instead.

## Anti-patterns

- Calling `migrate.Force` to get past a dirty state.
- Editing a shipped migration or the lock file by hand.
- Using `data/meerkat.db` for tests.
- Backing up or migrating `photos/`; it is out of scope.
