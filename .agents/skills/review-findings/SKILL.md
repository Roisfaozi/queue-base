---
name: review-findings
description: Use when writing code-review findings, audit recommendations, or fix lists in queue-base, especially when a recommendation would hide a feature, close a route, delete implementation, or change owner decisions.
---

# Review Findings

## Boundary

This skill shapes verified findings; it does not authorize implementation. For applying approved findings use the existing execution/review-response workflow. Read [AGENTS.md](../../../AGENTS.md), the actual diff and affected production path; use [qms-domain](../qms-domain/SKILL.md) when ticket, settings, or device behavior is involved.

## Procedure

1. Check working-tree ownership and the originating requirement. Separate current behavior, intended behavior, and reviewer preference.
2. Trace a reachable input/state through actual wiring to a concrete wrong result. Cite `file:line`; distinguish executed evidence from inspection. If evidence is incomplete, state what would confirm it rather than asserting a defect.
3. Write each finding with severity, defect, reproduction/failure scenario, impact, and preservation constraint. Suggest the narrow owning-layer fix and verification needed.
4. Mark decisions requiring owner input separately. A recommendation column, historical plan, or unanswered question is not approval.

## Preservation gate

- Hide/disable means stop rendering or gate behavior; retain useful implementation unless deletion was explicitly requested.
- Close an endpoint means unregister/gate the route; it does not automatically authorize deleting controller, usecase, tests, or reusable components.
- Deletion, schema/data changes, and changed product ownership require explicit scope and owner approval. Preserve user changes, tenant isolation, Redis sessions, Casbin policy semantics, and public compatibility.
- For QMS, preserve one master ticket, journey forwarding, and branch-scoped device/settings relations. Do not call deliberate design debt without evidence.

Stop if source intent, affected consumers, or preservation constraints cannot be established. Do not silently execute a finding while preparing the review.

## Output

Return verified findings ranked by severity, with evidence and preserved behavior. List unresolved owner decisions and unrun verification separately. If none survive verification, say so; do not invent cleanup findings to fill a report.
