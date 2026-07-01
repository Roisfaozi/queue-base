# Worktree Flow

This guide explains the recommended branch and worktree strategy for parallel development streams in this repo.

## 1. Branch Model

Use this promotion path:

- `main` = production-ready
- `staging` = release candidate and pre-release integration
- `dev` = daily integration branch

Daily feature work should start from `dev`.

Default rule:

- if current branch is `dev`, `make wt-new feat/name` creates worktree from `dev`
- if current branch is something else, that branch becomes the default base
- use second positional arg only when you want to override current branch, for example `make wt-new feat/name staging`

## 2. Why Worktrees

Use git worktrees when:

- frontend and backend move in parallel
- caller logic needs isolated testing from dashboard work
- multiple queue-management streams run at the same time
- one branch must stay open while another needs urgent work

Each worktree should have:

- its own git branch
- its own `.env.local`
- its own docker compose project name
- its own exposed local ports

Default root behavior:

- default worktree root is `.worktrees/` inside repo
- this is safer for restricted or sandboxed environments
- if you prefer sibling folders, pass `WORKTREE_ROOT=../queue-base-worktrees` explicitly

## 3. Recommended Stream Split

Example parallel streams:

- `feat/frontend-dashboard`
  - focus: `apps/web`, `apps/client`, `packages/*`
- `feat/caller-runtime`
  - focus: queue serving flow, station/counter runtime, realtime, SSE/WS
- `feat/queue-core`
  - focus: queue lifecycle, queue journeys, routing and domain rules

This split reduces conflicts and keeps review scope smaller.

## 4. Create New Worktree

From main repo checkout on `dev`:

```bash
make wt-new feat/frontend-dashboard
make wt-new feat/caller-runtime
```

Then move into each worktree and initialize local env:

```bash
cd .worktrees/feat-frontend-dashboard
make dev-up
```

```bash
cd .worktrees/feat-caller-runtime
make dev-up
```

`make wt-new` now bootstraps `.env.local` automatically.

Override base branch example:

```bash
make wt-new feat/caller-runtime staging
cd .worktrees/feat-caller-runtime
```

For existing worktrees, ensure env and local file state with:

```bash
make wt-enter feat/frontend-dashboard
```

To jump into worktree directly from current shell:

```bash
cd "$(make wt-path feat/frontend-dashboard)"
```

## 5. Daily Commands

Inspect current state:

```bash
make wt-list
make dev-status
make doctor
```

Keep env template aligned:

```bash
make env-sync
```

Run local migrations:

```bash
make migrate-up-local
```

Run narrow tests:

```bash
make test-local
make test-local TEST_PKG=./internal/modules/queue/...
```

Stop local stack:

```bash
make dev-down
```

## 6. Remove Worktree Safely

When feature work is merged or no longer needed:

```bash
make wt-rm feat/frontend-dashboard
```

Current behavior:

- stops local compose stack first when `.env.local` exists
- removes git worktree path
- does not auto-delete remote branch
- refuses to remove the currently active branch checkout

Clean stale git metadata:

```bash
make wt-prune
```

## 7. Merge Strategy

Recommended order:

1. feature branch merges into `dev`
2. integration and regression checks happen on `dev`
3. promote `dev` into `staging`
4. final release checks happen on `staging`
5. promote `staging` into `main`

## 8. Conflict Rules

If parallel streams touch the same API contract:

- settle contract shape first
- merge backend contract branch first if frontend depends on it
- rebase dependent frontend branch after contract merge
- do not let UI-only assumptions redefine backend rules

If streams touch queue journeys or tenant/branch boundaries:

- verify tenant scope in repository and usecase layers
- verify branch ownership checks
- verify realtime payload consumers still match producer changes

## 9. Verification Expectations

At minimum per stream:

- positive case
- negative case
- edge case
- vulnerability or boundary case

Examples:

- positive: valid branch-scoped queue serving succeeds
- negative: invalid counter assignment is rejected
- edge: business date reset boundary behaves correctly
- vulnerability: cross-tenant access is rejected

## 10. Frontend Worktree Flow

Worktree env sync now also generates frontend-local env files:

- `apps/web/.env.local`
- `apps/client/.env.local`

These files follow the current worktree backend port automatically.

Inspect generated frontend env:

```bash
make front-status
```

Show frontend run commands and ports:

```bash
make front-dev
```

Run Next.js frontend:

```bash
make web-dev
```

Run React Router frontend:

```bash
make client-dev
```

Expected behavior:

- `apps/web` uses `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_WS_URL` from current worktree
- `apps/client` uses `VITE_API_PROXY_TARGET` and `VITE_DEV_PORT` from current worktree
- frontend branch no longer needs manual API-port rewiring when backend worktree port changes
