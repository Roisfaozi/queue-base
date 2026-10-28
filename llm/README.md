# Project memory

## Read by task

Start with [AGENTS.md](../AGENTS.md), then one relevant cache and workflow. Skills live in [.agents/skills](../.agents/skills/); Claude discovers the same files through `.claude/skills`. Do not load the entire archive or skill collection.

- Queue lifecycle, reset/configuration, device identity: [qms-domain](../.agents/skills/qms-domain/SKILL.md) and [domain rules](cache/domain-rules.md).
- Backend ownership and route wiring: [architecture](cache/architecture.md), [go-service](workflows/go-service.md), or [API endpoint](workflows/api-endpoint.md).
- Frontend ownership and contracts: [frontend-surface](../.agents/skills/frontend-surface/SKILL.md), then [cross-stack workflow](workflows/cross-stack-change.md). QMS currently belongs to `apps/web`; inspect `apps/client` consumers before changing shared contracts.
- Feature work: [feature workflow](workflows/feature.md). Regressions: [bugfix workflow](workflows/bugfix.md).
- Reviews and delegation: [review-findings](../.agents/skills/review-findings/SKILL.md), [dispatch-prompt](../.agents/skills/dispatch-prompt/SKILL.md).
- Completion: [verification anti-patterns](conventions/verification-anti-patterns.md). Harness validation: `pnpm harness:check`.

[Harness checker](../scripts/check-harness.mjs) validates metadata across all canonical skills and inline local Markdown links in its explicit active-document list. Add newly maintained entrypoints to that list. The checker intentionally does not parse full YAML/Markdown, scan historical/vendor links, or prove runtime skill selection. Its [regression tests](../scripts/check-harness.test.mjs) use disposable fixtures and clean them afterward.

## Ownership and freshness

- `cache/`: stable facts checked against live code/config/tests, not unverified task outcomes. Correct contradicted text instead of appending another incompatible addendum.
- `conventions/`, `workflows/`, `test-playbooks/`: operational rules and task-specific procedures.
- [tasks/todo.md](tasks/todo.md): existing task-state entrypoint. Preserve recorded history; a completed label is not fresh verification. Do not create a competing backlog.
- [plans/](plans/README.md): durable staged plans and approved execution scope. A historical plan does not authorize destructive actions now.
- `research/`, `recommendations/`: evidence, proposed directions, and deferred ideas. Validate snapshots before acting.

Runtime truth comes from live code, wiring/config, then tests. Approved requirements describe intended behavior; a discrepancy is a gap to report, not permission to silently weaken the requirement. Legacy `code_context.txt` and rebuild notes are historical evidence, not a second implementation to copy.

Keep owner decisions and unfinished specifications intact. Promote only verified stable facts into cache. Report blockers and skipped checks separately from missing implementation. Use absolute dates for new evidence. No mass archival, deletion, or new task status inferred from filenames or commit titles.
