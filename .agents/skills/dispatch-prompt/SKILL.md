---
name: dispatch-prompt
description: Use when preparing a delegation prompt, parallel task split, or fix dispatch in queue-base where file ownership, owner decisions, preserved behavior, and acceptance evidence must be explicit.
---

# Dispatch Prompt

## Boundary

Prepare a bounded task for an agent only when delegation is appropriate. This skill does not opt the user into multi-agent orchestration, worktree creation, commits, or execution of review recommendations.

Read [AGENTS.md](../../../AGENTS.md), the approved scope and current working-tree ownership. Review-derived work also uses [review-findings](../review-findings/SKILL.md). Load [qms-domain](../qms-domain/SKILL.md) for queue/configuration/device work and the existing backend/API/frontend skill for the actual owner.

## Procedure

1. Separate confirmed defects from suggestions and unresolved owner decisions. Resolve material public/security/data decisions before dispatch; omit unapproved fixes.
2. Give the task goal, known evidence, exact allowed files, read-only context sources, and non-goals. Include the relevant code and test paths rather than assuming shared conversation context.
3. State preserved behavior: tenant/branch scope, session and Casbin checks, compatibility, QMS master/journey identity, transaction/side-effect semantics, and user-owned changes as applicable.
4. Specify acceptance cases and exact verification commands from the repository. Require actual results, skipped checks, and blockers, not merely a summary that tests passed.
5. Split independent work by disjoint writable paths. Serialize tasks sharing an interface, router, schema, generated mock, or contract. Use dependency order rather than concurrent edits to the same files.

## Required dispatch fields

- Goal and approved requirement.
- Allowed writable paths and owner; read-only dependencies.
- Evidence, relevant skills, invariants, and behavior to preserve.
- Acceptance cases and runnable checks.
- STOP conditions and required handoff.

STOP if live code contradicts the brief, another task owns a needed file, an owner decision is unanswered, or the task requires broader permissions/destructive actions. A recommendation is not an owner decision. Never ask another agent to bypass a tool denial or permission boundary.

## Completion

The receiving agent reports files changed, behavior proved, commands and results, skipped verification, and unresolved decisions. The coordinator reviews the actual diff and integrates only approved scope. No fabricated pending results; no commit/push unless requested.
