# Backend Map

## Current starter entrypoints

- API binary: `cmd/api/main.go`
- Generator binary: `cmd/gen/main.go`
- Composition root: `internal/config/app.go`
- Router: `internal/router/router.go`

## Current shared infrastructure

- config loading: `internal/config/config.go`
- DB: `internal/config/gorm.go`
- Redis: `internal/config/redis.go`
- Casbin: `internal/config/casbin.go`, `internal/config/casbin_model.conf`
- storage: `internal/config/storage.go`, `pkg/storage/*`
- transaction manager: `pkg/tx`
- JWT: `pkg/jwt`
- SSE: `pkg/sse`
- WebSocket: `pkg/ws`
- TUS: `pkg/tus`
- telemetry: `pkg/telemetry`
- query builder: `pkg/querybuilder`

## Current modules in live code

- `access`
- `api_key`
- `audit`
- `auth`
- `organization`
- `permission`
- `project`
- `role`
- `stats`
- `user`
- `webhook`

These modules are real today. Queue-specific modules are not.

## Queue rebuild backend target

Based on rebuild docs, likely new backend ownership will need these slices:

- `internal/modules/queue`
- `internal/modules/queue_forward`
- `internal/modules/scanner`
- `internal/modules/menu` or `internal/modules/service_point`
- `internal/modules/patient` or `internal/modules/customer`
- 
Treat those names as design candidates until implemented.

## Runtime extension points safest for rebuild

- add new module constructors in `internal/modules/<name>/module.go`
- wire them in `internal/config/app.go`
- expose routes through `internal/router/router.go` or module route helpers
- keep business logic in usecase/service layer
- keep transaction orchestration on server side, not in handlers

## Queue-domain backend rules from docs

- duplicate queue detection must happen before queue creation
- queue counter increment must happen under row lock inside transaction
- forward flow must lock source queue before state decision
- forward flow must check destination duplicate before creating new queue
- notification should happen after commit or safe async enqueue
- scanner handler should delegate to queue/forward services, not own business branching

## Verification focus when queue code starts

Start narrow:

- usecase tests for registration/forward/scanner branches
- repository tests for counter lock and duplicate queries
- integration tests for concurrency and rollback behavior

Escalate when boundary crosses:

- middleware/auth if scanner uses headers or JWT paths
- worker/outbox if notifications or integrations become async
