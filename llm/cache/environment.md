# Environment and Execution Context

Verified against the working tree on 2026-10-28. Read live scripts before running commands; local credentials and personal tool permissions do not belong in this cache.

## Configuration ownership

[internal/config/config.go](../../internal/config/config.go) loads environment-backed server, database, Redis, JWT, Casbin, CORS, rate-limit, SMTP, realtime, telemetry, storage/TUS, and SSO settings. [app.go](../../internal/config/app.go) wires their consumers. Use `.env.example` as the key template; do not print or commit secret values.

Root `.env` and `.env.local` are ignored. `make run` regenerates Swagger before starting the API. Worktree development uses the existing Makefile targets; do not guess ports or copy another project's database/search_path configuration.

## Backend commands

Source: [Makefile](../../Makefile), [root package scripts](../../package.json).

- Unit: `make test` / `make test-unit`, serial execution with `-p 1 -parallel 1 -count=1`.
- Focused unit: `make test-custom RUN='<TestName>' PKG=./internal/modules/<module>/...`.
- Verbose focused diagnostic: `make test-local TEST_PKG=./internal/modules/<module>/...`; output is unfiltered, so prefer focused tests.
- Integration: `make test-integration` or `make test-integration-custom RUN='<TestName>' PKG=./tests/integration/...`.
- E2E: `make test-e2e` or `make test-e2e-custom RUN='<TestName>' PKG=./tests/e2e/...`.
- `make test-all` prints a summary but currently always exits 0. Run constituent targets separately as acceptance gates.
- `make lint`, `make mocks`, `make docs`, `make build`, `make doctor` remain the existing lint, generation, build, and readiness surfaces. Generation changes files; do not treat it as read-only validation.

Integration/E2E use Docker-backed MySQL and Redis through Testcontainers. Docker unavailability can skip tests; exit 0 alone does not prove scenarios executed. Shared cleanup explains serial defaults. See [integration setup](../../tests/integration/setup/test_container.go).

## Frontend commands and side effects

Sources: [web package](../../apps/web/package.json), [client package](../../apps/client/package.json), [workspace package](../../package.json).

- Dev: `pnpm dev:web`, `pnpm dev:client`.
- Web typecheck: `pnpm --filter casbin-web typecheck` (`tsc --noEmit`).
- Client typecheck: `pnpm --filter casbin-client typecheck` (`react-router typegen && tsc`), which generates router types.
- Both apps run `biome lint --write .`. `pnpm lint` therefore modifies files and is not a read-only check or a substitute for typecheck.
- Root `pnpm format` writes formatting. For a read-only focused formatting check, use the installed Biome CLI without `--write`.
- Build: `pnpm --filter casbin-web build` / `pnpm --filter casbin-client build`, or `pnpm build`.
- Browser E2E: `pnpm --filter casbin-client test:e2e` when that app owns the flow. Root `pnpm test` only runs workspace scripts that exist; it is not proof of tests for every app.

QMS currently belongs to `apps/web`; inspect actual consumers before choosing the app to run. App environment settings remain owned by their respective app/proxy code, not interchangeable with backend environment variables.

## Harness and existing automation

`pnpm harness:check` runs the dependency-free Node integrity checker and its regression tests. The dedicated harness workflow needs Node 24, no package install, database, Redis, or Docker.

[lefthook.yml](../../lefthook.yml) already installs Git-hook checks through `pnpm prepare`. Some hooks format and restage files; some suppress failures. Hook execution is not proof that every quality check passed. This harness setup does not change those policies or install Claude hooks, plugins, MCP servers, or shared permissions.

Claude reads [CLAUDE.md](../../CLAUDE.md), which imports [AGENTS.md](../../AGENTS.md). `.claude/skills` links to the canonical `.agents/skills`; settings, worktrees, and session artifacts remain local. Newly discovered skills may require a fresh session; static link validation does not itself prove runtime discovery.
