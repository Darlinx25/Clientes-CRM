---
name: db-migration-safety
description: Use when adding or editing any SQLite schema, GORM model, migration, seed, backup, or upgrade code. Covers immutable migrations, pre-upgrade snapshots, dirty-schema refusal, checksums, destructive SQL, and fixture rehearsals.
---

# Database migration safety

The product rule is absolute: replacing the binary must open the previous
database without losing data. A migration is not "done" because it works on an
empty database.

## Workflow

1. Read `docs/upgrades.md` and the latest migrations before writing SQL.
2. Never edit an existing file in `backend/database/migrations/`. Create the
   next numbered pair:

   ```bash
   make -C backend migrations-checksum   # after adding the files
   ```

   The next number is the highest existing number + 1. Keep `.up.sql` and
   `.down.sql` symmetrical.
3. Write SQL that works with existing rows. Add columns with defaults or
   nullable values, backfill explicitly, then tighten constraints if needed.
4. Do not drop or rename user data casually. If it is unavoidable, add
   `-- migration:allow-destructive reason=...` and explain the data-preserving
   path in the PR/commit notes.
5. Add or update a fixture test. A new migration that changes data or shape
   must be exercised from the legacy fixture:

   ```bash
   make -C backend test-upgrade
   ```

6. Run the whole gate:

   ```bash
   make migrations-lint
   make backend-test
   ```

## What the code already does

- `InitDB` takes a lock, snapshots the DB with `VACUUM INTO` before an upgrade,
  refuses a dirty schema, runs migrations, then runs `PRAGMA quick_check`.
- The snapshot lives next to the configured backup directory as
  `pre-upgrade-vNNN-to-vMMM-*.db`.
- `migrations.lock` stores SHA256 for every migration file. Adding a new file
  is expected; changing an old hash is not.
- Destructive SQL is rejected from migration `000040` onward unless it carries
  the allow comment above.

## Anti-patterns

- Calling `migrate.Force` to get past a dirty state.
- Editing `000001`...`000039` or the lock file by hand.
- Using `data/meerkat.db` for tests.
- Backing up or migrating `photos/`; it is out of scope.
- Making a migration depend on application code that may change later.
