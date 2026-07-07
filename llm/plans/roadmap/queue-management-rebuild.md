# QMS Rebuild Roadmap Addendum

Latest design priority:

- `documentation/New Design Document — QMS MVP Operatio.md`

This roadmap adds QMS-specific implementation ordering without removing starter repo planning habits.

## QMS rule order

1. lock tenant/branch semantics
2. model queues vs queue_journeys vs visit_journeys
3. model typed settings inheritance via branch_service scope
4. add qms client and credential binding
5. add operator assignment domain
6. implement queue creation
7. implement forward via queue_journeys
8. implement caller/signage/scanner orchestration
9. wire routes and tests

## QMS TDD Requirement

Each phase must include failing-first tests where feasible and must not be marked complete without coverage for:

- positive cases
- negative cases
- edge cases
- vulnerability/security cases
