# QMS Rebuild Roadmap Addendum

This roadmap adds QMS-specific implementation ordering without removing starter repo planning habits.

## QMS rule order

1. lock tenant/branch semantics
2. model queues vs queue_journeys vs visit_journeys
3. model settings inheritance
4. implement queue creation
5. implement forward via queue_journeys
6. implement scanner orchestration
7. wire routes and tests

## QMS TDD Requirement

Each phase must include failing-first tests where feasible and must not be marked complete without coverage for:

- positive cases
- negative cases
- edge cases
- vulnerability/security cases
