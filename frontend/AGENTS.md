# Frontend agent notes

Read the root `AGENTS.md` first. This file only adds frontend-specific rules.

- Stack: React 19, TypeScript, MUI, React Router, i18next, Vitest, Playwright.
- All HTTP calls go through `src/api/client.ts`; it already handles the API base
  URL and auth. Do not call `fetch` directly from components.
- Do not hardcode versions or URLs. Release version comes from the backend
  health endpoint; API base comes from `REACT_APP_API_URL`.
- Keep translations in `src/i18n/`. Add the Spanish and English keys together;
  never leave a literal user-facing string in a component.
- Reuse `ConfirmDialog`, `EmptyState`, and other existing shared components
  before creating new ones.
- Prefer the existing MUI theme tokens and spacing over ad-hoc colors.
- Run `yarn tsc --noEmit`, `yarn test`, and `yarn build` after changes.
- E2E tests are in `e2e/` and use `docker-compose.test.yml`; do not point them
  at `data/meerkat.db` or a developer's real server.
- The app is single-user in practice (shared admin), but code must still handle
  missing/empty data without crashing.
