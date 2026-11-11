# Changelog

## 0.5.0 - 2026-11-11

### Upgrade safety

- `InitDB` now creates a consistent pre-upgrade SQLite snapshot with
  `VACUUM INTO` before changing an existing database.
- Dirty migration states are refused with the snapshot path instead of being
  forced automatically.
- Database migrations are protected by an advisory lock and followed by
  `PRAGMA quick_check`.
- Added SHA256 checksums (`backend/database/migrations.lock`) and a migration
  linter for gaps and destructive SQL.
- Added committed upgrade fixtures and rehearsals that verify old databases
  reach the latest schema without losing rows.
- `SeedAdminUser` no longer resets an existing admin password on every start.

### Diagnostics and tooling

- Added the root `VERSION` file and `meerkat/version` package as the single
  runtime version source.
- `/health` reports app version, current/expected schema version, and whether
  the schema is dirty or up to date.
- Added `migrationtool` and `mkfixture` commands plus root/backend Make targets.
- Added GitHub Actions for backend tests, upgrade rehearsals, frontend
  type-checking, tests and build.
- Added OpenCode configuration, instructions, skills, commands and read-only
  review subagents.
