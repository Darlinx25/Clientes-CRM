---
title: Agent Workflow
parent: Development
nav_order: 6
---

# Agent workflow

OpenCode loads `AGENTS.md`, `backend/AGENTS.md`, and `frontend/AGENTS.md` at
startup. Keep those files short and put detailed procedures in this directory.

## After changing OpenCode files

OpenCode loads configuration once at startup. After editing `opencode.json`,
`.opencode/`, `AGENTS.md`, or a skill, quit and restart OpenCode. The current
session keeps using the old configuration.

## Instruction files

- `AGENTS.md`: product rule, commands, protected files and definition of done.
- `backend/AGENTS.md`: migrations, startup, seeds and backend tests.
- `frontend/AGENTS.md`: API client, i18n, MUI and frontend tests.
- `docs/upgrades.md`: database upgrade contract, as reference.

## Commands

| Command | Purpose |
|---|---|
| `/new-migration <description>` | Add the next migration, refresh checksums and rehearse the upgrade |
| `/release-check` | Run the full local release gate |

`make test-upgrade` remains available for a direct upgrade rehearsal.

## Skills

Skills live under `.opencode/skills/` and load only when a task matches:

- `db-upgrade-safety`: schema changes, migrations, seeds, backups and upgrade
  rehearsals.
- `release-versioning`: `VERSION`, winres, build scripts and release checks.

## Subagent

- `migration-reviewer`: read-only review of migrations, seeds, backups and
  upgrade tests.

Use it before finishing a risky database change:

```text
@migration-reviewer review the current diff for data-loss risk
```

## Permissions

- `data/`, `photos/`, `.env`, databases and fixtures are protected.
- Edits under `backend/database/migrations/`, `migrations.lock`, `VERSION`,
  `winres.json` and CI workflows ask for confirmation.
- `git commit`, `git push`, `git tag`, `rm -rf`, `git reset --hard` and
  `git clean` are restricted.

Never use `data/meerkat.db` for tests. Use `t.TempDir()` or `:memory:`.

## Recommended loop

```bash
make test-upgrade
make migrations-lint
make backend-test
make frontend-typecheck   # when TypeScript changed
make frontend-test        # when frontend behavior changed
make release-check        # before a release
```

If a safety check fails, fix the cause. Do not relax fixtures, delete
snapshots, force a dirty migration, or edit a historical migration.
