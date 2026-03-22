# Go Real-Time Chat

A real-time chat application with a Go backend and React frontend, built to teach the parts of Go that show up in systems interviews and production services: goroutines, channels, `select`, `sync.RWMutex`, explicit error handling, `context.Context`, and single-binary deployment with `embed.FS`.

## What this project teaches

- `internal/chat/hub.go`: a hub goroutine that owns message routing and keeps the hot path lock-free
- `internal/chat/client.go`: a read pump and write pump per WebSocket connection
- `internal/chat/room.go`: shared room membership protected by `sync.RWMutex`
- `internal/middleware/ratelimit.go`: per-IP token bucket rate limiting with `golang.org/x/time/rate`
- `cmd/server/main.go`: graceful shutdown with `signal.NotifyContext`
- `embed.go`: Go's single-binary deployment model
- `internal/chat/hub_test.go`: table-driven tests around room listing, broadcasts, and isolation

## Stack

- Go + `net/http`
- `github.com/go-chi/chi/v5`
- `github.com/gorilla/websocket`
- `golang.org/x/time/rate`
- React + Vite + TypeScript + Tailwind CSS
- Railway-ready Docker deployment

## Local development

### 1. Install dependencies

```bash
go mod tidy
cd web
npm install
```

### 2. Run the frontend in development

```bash
cd web
npm run dev
```

The Vite dev server proxies `/api` and `/ws` to `http://localhost:8080`.

### 3. Run the Go server

```bash
go run ./cmd/server
```

The app uses `PORT` if it is set, otherwise it listens on `8080`.

## Tests and verification

Run the frontend build first so `web/dist` exists for `embed.FS`.

```bash
cd web
npm run build
cd ..
go test ./...
go build ./cmd/server
```

## API surface

- `GET /api/health`
- `GET /api/rooms`
- `GET /ws?room=general&username=alice`

Client messages:

```json
{
  "type": "message",
  "content": "hello"
}
```

Server events:

```json
{
  "type": "message",
  "room": "general",
  "username": "alice",
  "content": "hello",
  "users": ["alice", "bob"],
  "timestamp": "2026-03-21T18:12:00Z"
}
```

The server also emits `join` and `leave` events so the frontend can update room activity and presence.

## Deploying to Railway

1. Push the project to a Git provider supported by Railway.
2. Create a new Railway project and select the repo.
3. Railway will detect the `Dockerfile` automatically.
4. Set `DEMO_ACCESS_PASSWORD` so the live chat stays gated behind a password-protected session.
5. Set a hard spend limit in Railway cost controls before sharing the URL.
6. Set `PORT=8080` if Railway does not inject it for you.
7. Deploy. Railway supports long-lived WebSocket connections, so the chat transport works without switching to polling or serverless fallbacks.

## Architecture notes

- The hub is the only component that coordinates registration, unregistration, and broadcasts.
- Each client has two goroutines: one reading frames from the socket and one writing outbound frames.
- Room membership uses `sync.RWMutex` because multiple goroutines need read access, while writes are relatively rare.
- Invalid payloads and rate-limit violations close the connection explicitly instead of failing silently.

## Module path note

The module is initialized locally as `go-chat`. If you publish the project to GitHub later, you can update the module path to match the repository URL.
