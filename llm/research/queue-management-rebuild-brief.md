# QMS Rebuild Brief Addendum

This note augments existing starter AI-native context for QMS rebuild work. It does not replace repo-standard workflow docs.

## QMS target rules

- tenant-first queue platform
- branch under tenant
- one queue master row per ticket/day
- forward appends `queue_journeys`
- `visit_journeys` is readable internal history
- settings inherit tenant -> branch -> service -> counter
- no external integration scope unless explicitly requested later

## Starter compatibility

Keep existing starter architecture, module boundaries, verification habits, and live-code-first rules intact. Use QMS architecture only to add domain detail.

## QMS Testing Rule

Implementation should follow TDD whenever feasible. Each feature slice should explicitly plan tests for:

- positive case
- negative case
- edge case
- vulnerability/security case
