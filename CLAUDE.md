# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**sgame** is a multiplayer quiz game (Jeopardy-style) with a Go backend and Next.js frontend. Players join rooms, a host selects questions from packs, and players answer/bet in real time over WebSocket.

## Commands

### Backend (`backend/`)

```bash
# Hot-reload dev server (uses air)
air

# Run directly
go run ./cmd/server/main.go

# Build
CGO_ENABLED=0 go build -o ./tmp/main ./cmd/server/main.go

# Unit tests
go test ./... -v

# Single test
go test ./internal/domain/... -run TestFunctionName -v

# E2E tests (starts testcontainers; requires Docker)
go test ./test/e2e/... -v -race -timeout 120s

# E2E subset
go test ./test/e2e/ -run TestPubSub -v
go test ./test/e2e/ -run TestStream -v
go test ./test/e2e/ -run TestRoom -v

# Regenerate Wire DI (after changing wire.go)
go generate ./internal/app/...

# Regenerate Swagger docs (after changing handler annotations)
swag init -g cmd/server/main.go -o docs
```

### Frontend (`frontend/`)

```bash
npm run dev        # dev server on :3000
npm run build
npm run lint
npm run lint:fix
```

### Full stack

```bash
cp .env.template .env   # fill in values first
docker-compose -f docker-compose.dev.yaml up --build
```

## Backend Architecture

### Layered structure

```
internal/
  domain/         Pure domain types and state-machine logic (no I/O)
  service/        Business logic; orchestrates domain + infrastructure
  interface/      Go interfaces for repository, cache, realtime, storage
  infrastructure/ Concrete implementations:
    database/mongo/      MongoDB repositories
    cache/redis/         Redis session & room cache
    realtime/pubsub/     Redis Pub/Sub multiplexer
    realtime/streams/    Redis Streams multiplexer
    realtime/ws/         WebSocket channel wrapper
    storage/             GCS and Minio adapters
  transport/
    http/          Gin HTTP controllers (auth, user, pack, pack_draft, room)
    ws/            WebSocket handlers (lobby, room)
  eventsprocessor/
    server/        Internal server-side event processors (one per room event type)
    client/        Per-connection event processors (RoomEventsProcessor, LobbyEventsProcessor)
  app/             Wire DI setup (wire.go → wire_gen.go) and app startup (app.go)
  config/          Env-based config loaded at startup
  message/         Shared Message struct {Id, Event, Payload}
pkg/
  custerr/         Typed HTTP errors (400/401/403/404/500)
  custvalid/       Custom gin/validator tags
  envvar/          Env var helpers
  sets/            Generic set type
```

### Dependency injection

Google Wire is used. The canonical wiring is in `internal/app/wire.go`. After any change to provider signatures, run `go generate ./internal/app/...` to regenerate `wire_gen.go`. Never edit `wire_gen.go` by hand.

### Real-time multiplexing

The server maintains two Redis multiplexers injected as three typed channel getters:

| Type | Implementation | Use |
|------|---------------|-----|
| `PubSubChannelGetter` | `pubsub.Manager` | Lobby broadcast; room server channel for notifying clients |
| `StreamsChannelGetter` | `streams.StreamManager` (non-persistent) | Room broadcast stream; spectators start from `$` |
| `PersistentStreamsChannelGetter` | `streams.StreamManager` (persistent) | Room internal events; late joiners resume from stored `lastId` |

Both managers multiplex a single Redis connection across N Go subscribers. When the last subscriber on a channel leaves, the manager unsubscribes from Redis automatically.

### Room state machine

The `domain.Room` struct is the core entity. All state transitions are pure functions in `internal/domain/room.go`. The 12 states are:

`WaitingForStart → SelectingQuestion → RevealingQuestion → ShowingQuestion → Answering → SelectingQuestion` (loop) then `→ SelectingFinalRoundCategory → FinalRoundBetting → ShowingFinalRoundQuestion → ValidatingFinalRoundAnswers → GameOver`

Special question types alter the normal flow: **CatInBag** goes to `Passing`; **Auction** goes to `Betting` — both then converge to `Answering`.

Timer-driven transitions (reveal timer, betting timer, pass timer) are triggered by `RoomInternalEventsProcessor`, which runs per room as a goroutine. On server restart, `app.Start()` recovers ownership of existing rooms from the cache and re-spawns these goroutines.

### WebSocket connection flow

1. HTTP `GET /api/ws/room/:id` — authenticated via session cookie
2. `RoomHandler.connect` calls `roomService.Connect`, sends initial `room_updated` state snapshot to the client
3. A `RoomEventsProcessor` is created for the connection and runs in a goroutine
4. The processor reads from three channels concurrently: client WS messages, room pub/sub (server notifications), room persistent stream (replayed events)

### Config

All configuration comes from environment variables. Required vars: `MONGO_*`, `REDIS_HOST/PORT`, `STORAGE_PROVIDER`, `BUCKET_NAME`, `FRONTEND_URL`, `HOST`, `PORT`. Optional: `TIME_TO_BET` (default 60s), `TIME_TO_PASS` (default 60s), `QUESTION_DEMO_DURATION` (default 5s), `IDLE_ROOM_TTL` (default 600s).

### Swagger

Docs are generated from annotations on HTTP handlers and served at `/api/swagger/*`. After changing handler docs, re-run `swag init`. The generated files (`docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`) are committed.

## Frontend Architecture

Next.js 14 App Router. TypeScript + Tailwind CSS.

- **Data fetching**: `@tanstack/react-query`
- **Drag-and-drop**: `@dnd-kit` (used in pack draft editor)
- **Virtualization**: `@tanstack/react-virtual` (long lists)
- **API types**: auto-generated from Swagger via `openapi-typescript`
- **WebSocket**: custom hooks in `hooks/` for lobby and room connections
- **State**: React Context in `contexts/` for auth and room state

Pages follow the App Router convention in `app/`. Shared UI components live in `components/`.

## E2E Tests

Tests in `backend/test/e2e/` use `testcontainers-go` to spin up real MongoDB and Redis containers. `TestMain` starts containers once per binary. The `testhelper/` package provides `TestApp` (wraps `httptest.Server` built from the Wire graph), `WSClient`, and a `noopStorage` stub.

The `TestFullGame` test drives all 12 room states end-to-end with 1 host + 2 players. Config uses short timers (`TIME_TO_BET=2`, `TIME_TO_PASS=2`, etc.) to keep timer-driven transitions fast.
