# AGENTS.md

Meerkat CRM is an all-in-one personal CRM: Go + Gin + GORM/SQLite backend,
React 19 + MUI frontend, embedded in one binary for Windows/Linux users.
OpenCode configuration lives in `opencode.json` and `.opencode/`.

## Non-negotiable rule

**Every update must open an existing database without losing data.**

Users replace the executable and keep `data/`. Therefore:

- Never edit a migration that has already shipped. Add a new numbered pair.
- Never force a dirty migration. `InitDB` snapshots before upgrades, refuses
  dirty schemas, and refuses a database newer than the binary.
- Never drop or rename user data without a data-preserving migration and
  fixture coverage.
- New migrations must work on SQLite with existing rows, not only empty ones.
- Update `backend/database/migrations.lock` only with
  `make migrations-checksum`, and only for new files.
- If a change cannot preserve existing data, stop and ask before writing code.

## Repository map

- `backend/`: API, migrations, jobs, embedded UI package.
  - `backend/database/`: connection, migrations, snapshots, seeds.
  - `backend/controllers/`: HTTP handlers.
  - `backend/services/`: business logic.
  - `backend/models/`: GORM models.
  - `backend/cmd/`: migration and fixture tools.
- `frontend/`: React 19 + TypeScript + MUI + i18next.
- `docs/`: user and developer documentation.
- `scripts/`: release and Windows build helpers.
- `data/`, `photos/`, `.env`: runtime data. Never edit, commit, delete, or use
  them for tests.

## Commands

From the repository root:

```bash
make backend-test        # all Go tests, including upgrade tests
make test-upgrade        # upgrade rehearsals only
make frontend-test       # Vitest
make frontend-typecheck  # tsc --noEmit
make frontend-build      # production React build
make migrations-lint     # checksums + gaps + destructive SQL
make migrations-checksum # record a NEW migration in migrations.lock
make release-check       # full local gate before a release
```

## Working agreements

- Inspect the relevant code before editing; follow existing patterns and keep
  changes focused.
- Do not add comments unless they explain a non-obvious safety decision.
- Do not add a dependency if the repository already has an equivalent.
- Backend tests must use temporary databases (`t.TempDir()` or `:memory:`).
  Never point a test at `data/meerkat.db`.
- Frontend HTTP calls go through `frontend/src/api/client.ts`; do not hardcode
  API URLs or versions in components.
- `VERSION` is the release version. Runtime code reads it through
  `meerkat/version`; do not hardcode version strings.
- Health is `/health`; it reports app version and schema state.
- Photos are out of scope for upgrades and agent work unless explicitly
  requested.
- Git: never commit, amend, push, or tag unless explicitly asked. Never commit
  `.env`, `data/`, `photos/`, databases, or build artifacts.

## Definition of done

A change is complete only when:

1. Relevant tests pass (`make backend-test` / `make frontend-test`).
2. The frontend type-checks when TypeScript changed.
3. Migration safety checks pass when the schema changed.
4. Upgrade tests still reach the latest schema without losing rows.
5. User-facing behavior and docs match the code.

See `docs/upgrades.md` for the database contract and
`docs/agent-workflow.md` for OpenCode commands, skills, and permissions.
Those documents are references; they are not loaded into every session.
