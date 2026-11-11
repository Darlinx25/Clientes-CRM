# AGENTS.md

Meerkat CRM is an all-in-one personal CRM: Go + Gin + GORM/SQLite backend,
React 19 + MUI frontend, embedded in the same binary for Windows/Linux users.
The OpenCode configuration for this repository lives in `opencode.json` and
`.opencode/`.

## The one non-negotiable rule

**Every update must open an existing database without losing data.**

Users install one `.exe`, keep using it for years, and replace the file to
update. They never wipe `data/`. Therefore:

- Never edit a migration that has already shipped (`000001` to the latest).
  Add a new migration instead.
- Never force a dirty migration state. `InitDB` takes a pre-upgrade snapshot,
  refuses dirty databases, and refuses to start rather than guess.
- Never drop/rename a column or table that may contain user data without an
  explicit data-preserving migration and a new fixture test.
- New migrations must work on SQLite with existing rows, not only on an empty
  database.
- Update `backend/database/migrations.lock` only with
  `make migrations-checksum`, and only for newly added files.

If a change cannot preserve existing data, stop and ask the user before writing
code. Do not hide the trade-off in a migration.

## Repository map

- `backend/`: Go API, migrations, background jobs, embedded UI package.
  - `backend/database/`: connection, migrations, snapshots, seeds.
  - `backend/controllers/`: HTTP handlers.
  - `backend/services/`: business logic.
  - `backend/models/`: GORM models.
  - `backend/cmd/migrationtool/`: checksum/lint tool.
  - `backend/cmd/mkfixture/`: upgrade fixture generator.
- `frontend/`: React 19 + TypeScript + MUI + i18next.
- `docs/`: user and developer documentation.
- `scripts/`: release and Windows build helpers.
- `data/`, `photos/`, `.env`: local runtime data. **Never edit, commit, delete,
  or use these for tests.**

## Commands

From the repository root:

```bash
make backend-test        # all Go tests, including fixture upgrades
make test-upgrade        # upgrade rehearsals only
make frontend-test       # Vitest
make frontend-typecheck  # tsc --noEmit
make frontend-build      # production React build
make migrations-lint     # checksums + gaps + destructive SQL
make migrations-checksum # record a NEW migration in migrations.lock
make release-check       # full local gate before a release
```

Backend-only equivalents live in `backend/Makefile`. Run `make -C backend help`
from the root or `make help` inside `backend/`.

## Working agreements

- Inspect the relevant code before editing; follow existing patterns and
  keep changes focused.
- Do not add comments unless they explain a non-obvious safety decision.
- Do not add a dependency if the repository already has an equivalent.
- Backend tests must use temporary databases (`t.TempDir()` or `:memory:`).
  Never point a test at `data/meerkat.db`.
- Frontend API calls go through `frontend/src/api/client.ts`; do not hardcode
  API URLs or versions in components.
- `VERSION` at the repository root is the release version. Runtime code reads
  it through `meerkat/version`. Do not hardcode a version string.
- Health is `/health`; it reports app version and database schema state.
- Photos are out of scope for upgrades and agent work. Do not add photo
  backup, migration, or test behavior unless the user explicitly asks.
- Git: never commit, amend, push, or tag unless the user explicitly asks.
  Never commit `.env`, `data/`, `photos/`, databases, or build artifacts.

## Definition of done

A change is complete only when:

1. The relevant tests pass (`make backend-test` / `make frontend-test`).
2. The frontend type-checks when TypeScript changed.
3. Migration safety checks pass when the schema changed.
4. Upgrade fixtures still reach the latest schema without losing rows.
5. The user-facing behavior and docs match the code.

See `docs/upgrades.md` for the database contract and
`docs/agent-workflow.md` for OpenCode commands, skills, and subagents.
