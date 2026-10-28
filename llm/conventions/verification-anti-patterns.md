# Verification anti-patterns

Use before claiming completion, especially for tenant, device, transaction, or cross-stack work. Pair with [verification-before-completion](../../.agents/skills/verification-before-completion/SKILL.md).

## Prove the changed seam

- A mocked unit test does not prove a row lock, SQL predicate, rollback, or worker delivery. Use real persistence/integration evidence for those claims.
- A constructor or handler tested in isolation may not be the one mounted in production. Trace [app wiring](../../internal/config/app.go) and [router registration](../../internal/router/router.go) to the changed code.
- An authenticated success does not prove tenant isolation. Exercise a foreign tenant and mismatched branch/service/counter; for device work also test credential/type mismatch.
- A successful JWT parse is not live session verification. Inspect the actual middleware path; independently built route groups do not inherit each other's checks.
- A correct frontend type does not prove runtime payloads, cookies, proxy headers, empty/error states, or another app's consumer. Check the real producer and affected consumers.
- A forwarding test must preserve the master ticket and inspect journeys, not merely assert HTTP 200. Concurrency/reset changes need relevant duplicate/reset/parallel cases.

## Report execution, not optimism

- `make test-all` currently always exits 0; use individual suite targets. Docker-dependent tests can skip when Docker is unavailable. Record whether the target scenario ran.
- Both app lint scripts write files. Client typecheck generates router types. Existing Lefthook commands may suppress failures; a successful commit hook does not prove every check passed.
- Root `pnpm test` does not imply tests exist in every workspace. Name the app/package and scenario actually exercised.
- Static harness checks prove metadata, links, and discovery wiring, not automatic skill selection or runtime security. Confirm newly loaded skills in a fresh session when needed.
- Reading code, adding a test, running a build, and exercising a browser flow are different evidence. State each accurately, with command, result, and blocker for skipped checks.

Use the smallest sufficient check, then expand only for crossed boundaries. Preserve unrelated failing tests and report them instead of claiming this task fixed them. Never replace a required critical check with a weaker one without disclosing the remaining gap.
