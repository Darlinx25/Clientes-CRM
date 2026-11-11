---
title: Agent Workflow
parent: Development
nav_order: 6
---

# Agent workflow

This repository is configured for OpenCode. `opencode.json` loads the
repository instructions and restricts edits to the paths that can destroy data.

## After changing OpenCode files

OpenCode loads configuration once at startup. After editing `opencode.json`,
`.opencode/`, `AGENTS.md`, or a skill, quit and restart OpenCode. The current
session keeps using the old configuration.

## Instruction files

- `AGENTS.md`: product rule, repository map, commands, definition of done.
- `backend/AGENTS.md`: migration, startup, seed and backend test rules.
- `frontend/AGENTS.md`: API client, i18n, MUI and frontend test rules.
- `docs/upgrades.md`: the database contract.

## Commands

| Command | Purpose |
|---|---|
| `/new-migration <description>` | Add the next migration, refresh checksums and rehearse the upgrade |
| `/validate-upgrade` | Run lint, upgrade fixtures and all backend tests |
| `/release-check` | Run the full local release gate |

These are thin wrappers around `make` targets, so the same checks work in CI.

## Skills

Skills live under `.opencode/skills/` and are loaded when a task matches their
description:

- `db-migration-safety`: schema changes, migrations, seeds, backups.
- `upgrade-rehearsal`: proving old databases still upgrade without row loss.
- `release-versioning`: `VERSION`, winres, build scripts, release gate.

## Subagents

- `migration-reviewer`: read-only review of migrations, seeds, backups and
  upgrade tests.
- `ui-verifier`: read-only review of React/MUI changes, API usage, i18n and
  tests.

Use them before finishing a risky change:

```text
@migration-reviewer review the current diff for data-loss risk
@ui-verifier review the frontend changes
```

## Permission model

- `data/`, `photos/`, `.env`, databases and fixtures are protected.
- Any edit under `backend/database/migrations/`, `migrations.lock`,
  `VERSION`, `winres.json` or the CI workflows asks for confirmation.
- `git commit`, `git push`, `git tag`, `rm -rf`, `git reset --hard` and
  `git clean` are restricted.

The agent must never use `data/meerkat.db` for tests. All test databases come
from `t.TempDir()` or `:memory:`.

## Recommended loop

```bash
# 1. Understand
make -C backend test-upgrade

# 2. Change code / add a migration
make -C backend migrations-checksum   # only for NEW migration files

# 3. Verify
make migrations-lint
make backend-test
make frontend-typecheck   # when TypeScript changed
make frontend-test        # when frontend behavior changed

# 4. Before a release
make release-check
```

If a safety check fails, fix the cause. Do not relax the fixture, delete the
snapshot, force a dirty migration, or edit a historical migration.
