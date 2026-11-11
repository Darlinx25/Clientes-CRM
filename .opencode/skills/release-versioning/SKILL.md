---
name: release-versioning
description: Use when bumping the app version, building the Windows/Linux release, changing VERSION, winres, the build script, or release checks. Covers the single source of version truth and the pre-release gate.
---

# Release versioning

`VERSION` at the repository root is the single source of truth. Runtime code
reads it through `meerkat/version`; the build script reads it to name artifacts
and inject `-X meerkat/version.Version`.

## Bump checklist

1. Edit `VERSION` (for example `0.5.1`).
2. Update `backend/winres/winres.json` to the same version.
3. Keep `frontend/package.json` and the two top-level version fields in
   `frontend/package-lock.json` aligned; do not mass-replace dependency
   versions.
4. Run the gate:

   ```bash
   make release-check
   ```

5. Build artifacts:

   ```bash
   scripts/build-windows-exe.sh
   ```

6. Verify `/health` reports `"app_version"` and the latest
   `"expected_schema_version"`.

## Rules

- Do not hardcode a version in Go, React, Docker, or docs.
- A version bump does not require a migration, but a schema change does require
  a new migration and an upgrade rehearsal.
- Never commit `dist/`, `backend/web/dist/`, `data/`, `photos/`, or `.env`.
