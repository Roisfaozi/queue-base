# Queue Management Domain Map

## Purpose

Map every feature from `documentation/task-overview.md`, `documentation/project-diagram.md`, and legacy `code_context.txt` into repo-native modules, routes, data ownership, and build order.

This is the design bridge between current starter runtime and final queue-management product.

## Current runtime truth

Current repo already has starter infrastructure for:

- auth and session
- organization and tenant management
- Casbin authorization
- API keys
- audit, webhook, worker, stats
- realtime and upload support
- generic project/user/role/permission/access modules

Current repo does **not** yet have queue-domain modules for registration, forward, scanner, pharmacy, or queue-specific integrations.

## Feature inventory from provided docs

### Core queue features

- customer/patient queue registration
- duplicate prevention for active queue
- queue numbering per branch/menu/business day
- queue counter locking under concurrency
- queue state transitions
- queue listing and queue detail read paths
- queue estimate calculation based on real data rather than constant placeholder
- queue reset / business-day boundary at `04:00 Asia/Jakarta`

### Forwarding features

- forward queue between station/menu/device destinations
- source queue locking before mutation
- destination duplicate check before create
- source done/continued updates in same transaction
- notification after durable commit
- special validation for pharmacy flow
- branch/menu/device relation validation

### Scanner features

- scanner login with `x-client-id` and `x-api-key`
- scanner credential validation against repository truth
- scanner check-in middleware and service flow
- branch/station/menu/device relation validation
- existing active queue lookup
- delegation to queue registration or forward service

### Pharmacy features

- explicit source code/flag for `PENERIMAAN_RESEP`
- no regex-only truth for pharmacy detection
- history-based allowance when patient already passed reception
- validation of allowed forward chains involving pharmacy destination/source

### Platform and admin features from legacy/system docs

- users
- roles
- permissions
- menus/submenus
- branches
- companies
- lockets/counters
- devices
- assets
- dashboard data
- access-right registry
- authentication and token handling
- Casbin RBAC policy management
- organization/tenant handling if multi-tenant behavior remains

### Integration features

- scanner login/check-in route surface
- internal notification dispatch after forward commit
- optional internal appointment/patient adapter only if later proved necessary

### Quality and safety issues from docs

- Casbin matcher/path mismatch
- wildcard credentialed CORS
- scanner auth that only checks header presence
- duplicated Casbin policy mutation
- ignored Casbin error in registration flow
- unsafe order parameter handling
- protected debug DB stats route

## Proposed module placement

### Queue core

`internal/modules/queue`

Responsibilities:

- register queue
- assign queue number
- duplicate search
- queue list/detail
- queue estimate
- business-day helper integration
- queue reset logic

### Forward orchestration

`internal/modules/queue_forward`

Responsibilities:

- forward queue transaction
- source queue lock
- destination validation
- pharmacy rule enforcement
- post-commit notification trigger
- source state mutation

### Scanner

`internal/modules/scanner`

Responsibilities:

- scanner credential validation
- scanner login or session flow
- station/menu/branch relation checks
- active queue detection
- delegation into queue/forward usecases

### Service point metadata

`internal/modules/service_point` or equivalent split

Responsibilities:

- branch
- locket/counter
- menu/submenu
- device
- station relation
- explicit pharmacy classification

### Patient/customer data

`internal/modules/customer` or `internal/modules/patient`

Responsibilities:

- patient identity lookup
- appointment lookup
- queue history inputs
- active queue search filters

### Existing platform modules to reuse

- `auth`
- `organization`
- `permission`
- `role`
- `access`
- `api_key`
- `audit`
- `webhook`
- `stats`
- `user`
- `project`

These should stay intact unless a queue requirement explicitly consumes them.

## Data ownership map

### Queue tables/entities

Likely entities needed:

- queue/customer queue record
- queue counter row
- queue transition history
- queue estimate inputs if cached
- queue status enum / state machine

### Scanner tables/entities

Likely entities needed:

- scanner client or device credential
- scanner branch mapping
- scanner station/menu mapping
- scanner session or login token if separate from standard auth

### Service-point tables/entities

Likely entities needed:

- branch
- menu
- submenu
- locket/counter
- device
- service point relation to pharmacy flag/code

### External linkage entities

Likely entities needed:

- appointment mapping
- patient lookup cache or relation table
- visit journey / source tracking if required by integration layer

## Route design map

### Queue routes

Likely under `/api/v1` with tenant or authenticated protection depending final product decision.

Expected route types:

- register queue
- get queue by id
- list queues
- queue estimate
- queue status transition / call / cancel / done / missed

### Forward routes

- forward existing queue to new target
- forward by scan/check-in path
- history endpoints if needed

### Scanner routes

- scanner login / credential exchange
- scanner check-in
- scanner queue lookup
- scanner branch/station/menu validation

### Admin/platform routes

Keep current admin-style platform routes for permission/role/access/user/org if still relevant to product.

## Build order

### Step 1

Lock domain names and schema.

### Step 2

Implement queue registration and queue counter transaction.

### Step 3

Implement forward service and pharmacy validator.

### Step 4

Implement scanner auth and check-in orchestration.

### Step 5

Wire external integration adapters and consumer sync.

### Step 6

Refine frontend consumers if queue UI is part of starter branch.

## Design constraints

- handlers stay thin
- repository stays data-only
- business rules stay in usecase/service layer
- side effects run after commit or via safe outbox/worker path
- explicit code/flag beats regex for pharmacy semantics
- business-day logic belongs in shared domain helper
- queue state machine centralizes status transitions

## Open questions

- needs confirmation: final module naming between `queue` vs `customer` as primary bounded context
- needs confirmation: whether scanner login is separate JWT, API-key exchange, or both
- needs confirmation: whether organization/tenant stays in final queue product or only platform admin layer
