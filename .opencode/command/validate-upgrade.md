---
description: Rehearse old-database upgrades and report row-loss or dirty-schema failures
agent: build
---
Validate the repository's upgrade contract without touching `data/`.

1. Run `make migrations-lint`.
2. Run `make test-upgrade`.
3. Run `make backend-test`.
4. Inspect `backend/database/testdata/fixtures/` and confirm each fixture
   reaches the latest schema.
5. If a fixture fails, explain whether the cause is a destructive migration, a
   missing backfill, a dirty-schema path, or a test expectation. Do not edit
   historical migrations to make the test pass.

Report commands, failures, and the smallest safe fix.
