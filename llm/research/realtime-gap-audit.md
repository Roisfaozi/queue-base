# Realtime QMS Event Gap

## Status
- Frontend apps/web: WebSocket provider and hook ready (`useWebSocket`). Presence and audit consumers exist.
- Frontend apps/web QMS UI: No QMS consumers exist.
- Backend WS Infra: `wsManager` and Redis pub/sub ready.
- Backend QMS (queue, caller, signage): **No event producers exist.** 

## Evidence
- `rg -n "BroadcastToChannel" internal/modules/queue` -> 0 results
- `rg -n "BroadcastToChannel" internal/modules/caller` -> 0 results

## Next Step
Realtime QMS (dashboard auto-refresh, signage auto-refresh) is blocked until backend emits WS events for `QUEUE_REGISTER`, `QUEUE_CALL`, `QUEUE_FORWARD`, etc.
