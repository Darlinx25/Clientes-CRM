---
title: Upgrades
parent: Development
nav_order: 5
---

# Database upgrade contract

Meerkat CRM ships as a single replaceable executable. A user may run version
0.3 for years and then replace the file with today's build. The new binary must
open the old `data/meerkat.db`, migrate it, and keep every row. This document is
the contract every schema change must satisfy.

## Startup sequence

`database.InitDB(dbPath, backupDir)` always does the following:

1. Open SQLite with WAL, a 10-second busy timeout and foreign keys enabled.
2. Take an advisory lock on `<dbPath>.lock`. A second instance cannot migrate
   the same database concurrently.
3. Read `schema_migrations` without creating it. A missing table means a new
   database (`current = -1`).
4. If the database needs an upgrade, or is dirty, copy it with `VACUUM INTO`
   into `backupDir` as `pre-upgrade-vNNN-to-vMMM-<timestamp>.db`.
5. Refuse to continue if the schema is dirty. The error includes the snapshot
   path. `migrate.Force` is never called automatically.
6. Refuse to open a database whose schema version is newer than the running
   binary; using an old executable on a new database is a downgrade, not an
   upgrade.
7. Apply embedded migrations.
8. Run `PRAGMA quick_check` before serving requests.

If the snapshot cannot be created, startup fails. Upgrading without a recovery
point is not allowed.

The automatic snapshot contains only the SQLite database. Profile photos are
outside the upgrade contract and are never copied, migrated or required by a
migration.

## Where snapshots live

- `BACKUP_DIR` defaults to `backups/` next to the executable
  (`C:\Historial\backups\` in the Windows build).
- The snapshot is a normal `.db` file. It can be opened with any SQLite tool.
- Automatic weekly/manual zip backups are a separate feature and must not be
  confused with the pre-upgrade snapshot.

## Recovering a failed upgrade

1. Stop the app.
2. Locate the snapshot named in the error.
3. Preserve the broken database for diagnosis:
   `mv data/meerkat.db data/meerkat.db.broken`.
4. Copy the snapshot back to `data/meerkat.db`.
5. Start the previous version, or fix the failing migration in a new build.
6. Only after the data is safe, remove the `.broken` file.

Never run `migrate force` to make a dirty database start. A dirty state means a
migration failed halfway; forcing the version can skip the missing half.

## Schema versions

- The embedded migration files are the only schema source of truth.
- The latest version is the highest numeric prefix. Fixture tests assert it; a
  new migration must update that assertion.
- `/health` returns `app_version`, `schema_version`,
  `expected_schema_version`, `schema_dirty` and `schema_up_to_date`.
- A database with no `schema_migrations` table reports `schema_version: -1`
  rather than failing health checks.

## Adding a migration

1. Add `0000NN_description.up.sql` and `.down.sql`; never edit an old pair.
2. Make the up migration safe for existing rows: add with defaults/nullable,
   backfill explicitly, then tighten if necessary.
3. Run `make -C backend migrations-checksum`. This appends only new files to
   `migrations.lock`; it refuses to rewrite a historical hash unless `-force`
   is used after manual review.
4. Add or update a fixture assertion when data shape or values change.
5. Run `make test-upgrade`, `make migrations-lint`, and `make backend-test`.

From migration `000040` onward the linter rejects `DROP TABLE`, `DROP COLUMN`,
`RENAME COLUMN`, `RENAME TO`, unqualified `DELETE FROM`, and `UPDATE ... SET`
unless the file contains:

```sql
-- migration:allow-destructive reason=why this preserves user data
```

The annotation is a prompt for review, not a bypass. A destructive migration
still needs a test proving the data survives or an explicit product decision.

## Fixtures

Committed fixtures in `backend/database/testdata/fixtures/` represent old
releases with real rows. Tests copy each fixture to a temporary directory, call
the production `InitDB`, and verify:

- the schema reaches the latest version;
- the schema is not dirty;
- row counts do not decrease for user content tables;
- spot-checked values are unchanged;
- an upgrade creates a pre-upgrade snapshot;
- a dirty database is refused and its snapshot preserves the dirty state.

Regenerate fixtures only from embedded migrations:

```bash
make -C backend fixtures
```

Never generate a fixture from `data/meerkat.db`, and never point a test at that
file.
