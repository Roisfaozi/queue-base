# Project Overview

Verified against the working tree on 2026-10-28. This is a QMS product built on a multi-tenant starter, not an empty starter awaiting queue modules.

## Runtime topology

- One product API entrypoint: [cmd/api/main.go](../../cmd/api/main.go). [cmd/gen/main.go](../../cmd/gen/main.go) is a development scaffolder, not another product API.
- Go module: `github.com/Roisfaozi/queue-base`. Gin, GORM, Casbin, Redis sessions, Asynq workers, SSE/WebSocket, TUS/storage, and OpenTelemetry are wired in [app.go](../../internal/config/app.go).
- [package.json](../../package.json) defines `casbin-monorepo`, pnpm 10.8.0, Turbo, Biome, and Lefthook. Shared frontend packages retain `@casbin/*` names.
- [apps/web](../../apps/web/) (`casbin-web`) uses Next.js App Router and currently owns QMS dashboard/operational pages.
- [apps/client](../../apps/client/) (`casbin-client`) uses React Router v7 and remains an active IAM/admin/showcase surface. Its [route registry](../../apps/client/app/routes.ts) does not currently expose QMS routes. Do not add a second QMS frontend without a requirement.

## Backend ownership

Platform modules include auth, organization/branches, user, role, permission, access, API keys, audit, webhooks, and stats. QMS modules include queue, queue_config, service/branch-service, counter, caller, scanner, signage, operator_assignment, and qms_client. Actual constructor and route registration, not a historical module list, determine what is active.

[Queue domain skill](../../.agents/skills/qms-domain/SKILL.md) covers master ticket/journey semantics, atomic numbering, typed inherited settings, scoped relations, device identity, and required tests. [Architecture](architecture.md) records runtime boundaries; [domain rules](domain-rules.md) records QMS invariants.

## Scope and historical evidence

[QMS architecture document](../../documentation/QMS_Rebuild_Multi_Tenant_Queue_Architecture_Document.md) supplies product intent. [Rebuild brief](../research/queue-management-rebuild-brief.md) and [roadmap](../plans/roadmap/queue-management-rebuild.md) retain design history. Their existence or completed labels do not prove current integration coverage.

`code_context.txt` is legacy evidence only. Do not reintroduce legacy destination-ticket duplication, menu/station naming, or external integration assumptions without comparing current entities and contracts.

For command details and side effects, read [environment](environment.md) and the live package scripts/Makefile. For task routing, read [project memory](../README.md).
