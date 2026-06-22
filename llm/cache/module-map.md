# Module Map

## Live starter modules

Current live modules under `internal/modules` are:

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

These remain source of truth for current runtime only.

## Rebuild-target module map

Queue-management rebuild docs imply future module ownership closer to:

### Queue core

Likely responsibilities:

- register queue
- allocate next queue number
- detect active duplicate queue
- list/detail queue state
- compute business date window

### Queue forward

Likely responsibilities:

- lock and validate source queue
- validate destination station/menu/device
- enforce pharmacy transition rule
- detect existing destination queue
- create destination queue and update source state atomically

### Scanner

Likely responsibilities:

- validate `x-client-id` + `x-api-key`
- validate branch/station/menu/device relation
- choose existing queue vs forward vs new registration path
- delegate into queue or forward usecases

### Service-point metadata

Likely responsibilities:

- branch
- locket/counter/station
- menu/submenu or service code
- device/client registration
- explicit pharmacy classification fields

### Patient or customer

Likely responsibilities:

- patient identity lookup
- appointment lookup
- active queue search inputs

### Integration layer

Likely responsibilities:

- internal notification dispatch after forward commit when needed
- external adapters separated from core queue business rules

## Mapping rule for future coding

Do not force old legacy folder names from `code_context.txt` directly into this starter if they fight current repo layering.

Preferred mapping:

- old handler/controller logic -> `delivery/http`
- old business/service logic -> `usecase`
- old repository/data logic -> `repository`
- old cross-cutting helper logic -> `pkg/*` or shared infra only when truly reusable

## Smell list

If future queue implementation does any of these, stop and refactor:

- handler directly increments counters
- handler directly contains forward/pharmacy branching
- regex on menu names decides pharmacy semantics
- notification happens before transaction commit
- scanner middleware only checks header presence
