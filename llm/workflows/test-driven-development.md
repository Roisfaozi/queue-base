# TDD-First Workflow

## Purpose

Primary workflow for development that must start from tests and use table-driven coverage for every feature, flow, endpoint, or module scenario set in this repo.

Use this workflow to force failing-first development, keep scenario coverage explicit, and prevent one-off tests from hiding missing negative or security cases.

## Use When

- adding new backend behavior with a real test seam
- fixing regressions that can be reproduced with package, route, or integration tests
- adding or changing feature, flow, endpoint, or usecase tests
- writing tests for auth, tenant, access, permission, API-key, upload, worker, or Casbin boundaries

## Do Not Use When

- only realistic proof is manual or E2E and unit TDD would be fake confidence
- behavior is not yet clear enough to define a stable failing case
- the change is docs-only or no code/test surface changes are expected

## Required Read Order

1. `AGENTS.md`
2. `llm/conventions/testing.md`
3. `llm/workflows/feature.md` or narrower workflow for the owning area
4. relevant domain cache under `llm/cache/*`
5. target package tests and adjacent live code

## Workflow

### Step 1 — Define Scenario Table

Before code, write the scenario map as a table-driven test matrix with rows for:

- positive
- negative
- edge
- security
- vulnerability

Only drop a category if it is truly meaningless for the seam and the omission is obvious from the behavior itself.

### Step 2 — Write Failing Test First

- encode the smallest failing table row set that proves the behavior
- confirm failure is behavior failure, not compile or setup noise
- keep one test file or one table per coherent behavior slice

### Step 3 — Implement Minimal Code

- make smallest change that turns the table green
- preserve existing contracts for auth, tenant, permission, access, API-key, and transaction boundaries
- do not add unrelated cleanup while test is red

### Step 4 — Expand Table If Needed

- add rows for newly discovered edge/security paths
- keep assertions boundary-relevant, not implementation-detail heavy
- avoid splitting one feature flow into multiple redundant non-table tests

### Step 5 — Refactor Only After Green

- remove duplication only after all required table rows pass
- keep category visibility intact in names or row data

## Table-Driven Test Rules

- use a struct slice with named fields for scenario inputs, expectations, and category
- keep each row self-contained
- include category in row data or test name so review can inspect coverage fast
- for feature/flow/endpoint tests, table should usually include:
  - success path
  - invalid input or rejection path
  - boundary or edge path
  - auth/permission/security path
  - vulnerability or abuse path
- when one category is not applicable, say so implicitly by omitting the row only if the seam cannot express it

## Stop Conditions

Stop and choose broader seam if:

- unit TDD would be misleading for the real runtime boundary
- route, auth, tenant, or permission behavior only proves correctly at integration/E2E layer
- the scenario table would be fake or speculative

## Completion Output

Report:

- scenario table added or updated
- categories covered
- files changed
- commands run and exact result
- seam or category limitations if any
