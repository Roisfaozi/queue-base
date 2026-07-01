# Worktree Verification Playbook

Reusable checklist for worktree DX verification.

## 1. New Worktree Creation

```bash
make wt-new feat/test-worktree
cd .worktrees/feat-test-worktree
make env-init
make env-sync
make doctor
```

Verify:

- worktree path exists
- `COMPOSE_PROJECT_NAME` unique per branch
- `APP_PORT`, `MYSQL_PORT`, `REDIS_PORT` not colliding with default root stack
- `MYSQL_DBNAME` includes worktree slug

## 2. Existing Branch Reuse

```bash
make wt-new feat/test-worktree
```

Expected:

- if branch exists but not attached, worktree creation succeeds
- if branch already attached to another worktree, command fails with clear message

## 3. Backend Stack

```bash
make dev-up
make dev-status
make migrate-up-local
make test-local
make dev-down
```

Verify:

- backend starts on worktree-specific ports
- MySQL and Redis ports do not fall back to global defaults
- migration waits until MySQL is reachable

## 4. Frontend Stack

```bash
make front-status
make front-dev
make web-dev
make client-dev
```

Verify:

- `apps/web/.env.local` exists and points to current worktree backend URL
- `apps/client/.env.local` exists and points to current worktree backend URL
- Next.js and React Router dev ports are generated per worktree

## 5. Cleanup

```bash
make dev-down
cd ../..
make wt-rm feat/test-worktree
make wt-prune
```

Verify:

- stack stops before worktree removal
- branch is not deleted automatically
- stale git metadata is pruned
