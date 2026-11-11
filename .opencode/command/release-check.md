---
description: Run the full local release gate and summarize blockers
agent: build
---
Run the repository release gate:

```bash
make release-check
```

If a step fails, stop at the first real blocker, diagnose it, and do not bump a
version or rewrite a migration to hide it. When everything passes, confirm:

- `VERSION`, `backend/winres/winres.json`, and frontend package versions agree;
- `/health` exposes app and schema versions;
- no `.env`, `data/`, `photos/`, database, or build artifact is staged.

Report the exact commands run and their outcomes.
