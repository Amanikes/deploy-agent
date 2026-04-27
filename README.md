# Deploy Agent (Go)

A lightweight remote deployment agent that connects to a backend over native WebSocket, receives deployment commands, executes them on the host, and reports progress/acknowledgements.

## Current Status

- Core WebSocket client implemented with auto-reconnect.
- Header-based agent authentication implemented (`X-Agent-ID`, `X-Agent-Token`).
- Non-blocking command listener implemented (each command handled in its own goroutine).
- Command dispatch implemented for `deploy`, `restart`, and `status`.
- Execution logic moved into `internal/pipeline/orchestrator.go`.
- Progress and lifecycle acknowledgements implemented.
- Build compiles successfully with `go build ./...`.

## What The Agent Can Do

- Connect to backend WebSocket endpoint and keep retrying on disconnect.
- Parse inbound JSON commands in this shape:
  - `id`: string
  - `action`: string
  - `project`: string
  - `payload`: `map[string]string`
- Normalize legacy action aliases:
  - `build` -> `deploy`
  - `get_status` -> `status`
- Execute deployment workflows from the pipeline orchestrator:
  - Clone repository from `repo_url` when `repo_dir` is not provided.
  - Pull latest changes when `repo_dir` is provided.
  - Run tests via `test_cmd` or default `go test ./...` when `run_tests=true`.
  - Deploy using `deploy_cmd` or `docker compose -f <compose_file> up -d [service]`.
- Execute restart workflows:
  - `restart_cmd`, or
  - `docker restart <container>`, or
  - `docker compose -f <compose_file> restart <service>`.
- Execute status workflows:
  - `status_cmd`, or
  - `docker ps --filter name=<container>`.
- Send ack messages back to backend:
  - `received`, `started`, `progress`, `completed`, `failed`, `invalid`.

## What The Agent Cannot Do Yet

- Socket.IO protocol support (native WebSocket JSON only).
- Persistent command queue/replay in the agent while disconnected.
- Command cancellation once started.
- Per-command timeout enforcement.
- Concurrency limits (worker pool) beyond goroutine per message.
- Rich structured payload types beyond `map[string]string`.

## Runtime Requirements

- Go 1.24+
- Git installed on target host
- Docker and Docker Compose (if using compose-based deploy/restart/status)
- A backend endpoint that speaks native WebSocket (not Socket.IO framing)

## Configuration

The agent currently requires a `.env` file and exits if missing.

Required keys:

```env
BACKEND_URL=ws://localhost:8080/ws
AGENT_ID=your-agent-id
AGENT_TOKEN=your-agent-token
```

Notes:

- `AGENT_ID` and `AGENT_TOKEN` are mandatory.
- `BACKEND_URL` defaults to `http://localhost:8080` if not set, but should be a websocket URL (`ws://` or `wss://`) for actual runtime.

## Run

From repo root:

```bash
go run ./cmd/agent
```

Why not `go run .`?

- The root module directory does not contain a `main` package file.
- The executable entrypoint is `cmd/agent/main.go`.

## Command Contract (Backend -> Agent)

Example:

```json
{
  "id": "cmd_123",
  "action": "deploy",
  "project": "my-service",
  "payload": {
    "repo_dir": "/srv/my-service",
    "run_tests": "true",
    "deploy_cmd": "docker compose up -d --build"
  }
}
```

Common `payload` keys:

- Deploy:
  - `repo_dir` or `repo_url`
  - `run_tests` (`true`/`false`)
  - `test_cmd`
  - `deploy_cmd`
  - `compose_file`
  - `service`
- Restart:
  - `repo_dir`
  - `restart_cmd`
  - `container`
  - `compose_file`
  - `service`
- Status:
  - `repo_dir`
  - `status_cmd`
  - `container`

## Ack Contract (Agent -> Backend)

Ack message shape:

```json
{
  "command_id": "cmd_123",
  "action": "deploy",
  "status": "progress",
  "message": "running test command",
  "project": "my-service",
  "meta": {
    "step": "run_tests"
  },
  "at": "2026-04-22T11:00:00Z"
}
```

## Task Flow

1. Agent starts (`cmd/agent/main.go`).
2. Config loads from `.env`.
3. Agent opens WebSocket connection with auth headers.
4. Agent enters listen loop.
5. Backend sends command JSON.
6. Agent parses command and spawns goroutine for dispatch.
7. Agent sends `received` and `started` acks.
8. API dispatch delegates to `pipeline.Execute(...)`.
9. Orchestrator performs steps (`clone/pull`, optional tests, deploy/restart/status commands).
10. Pipeline emits progress callbacks; API forwards them as `progress` acks.
11. On success, API sends `completed`.
12. On error, API sends `failed` with message.
13. If websocket drops, agent reconnects and resumes listening.

## Testing

Unit and integration tests are provided for the orchestrator and API handler logic:

- All task execution now uses context-aware functions, supporting cancellation and progress reporting.
- To run tests:

```bash
go test ./internal/pipeline
go test ./internal/api
```

### Example Tests

- Cancelling a running deploy task returns a context.Canceled error and triggers progress callbacks.
- Unsupported actions return a clear error.
- Status commands are executed and their output is validated.

See `internal/pipeline/orchestrator_test.go` and `internal/api/handler_test.go` for details.

## Recent Changes

- Refactored orchestrator to use only context-aware task functions (deploy, restart, status).
- Added support for command cancellation and rollback.
- Added unit and integration tests for cancellation, unsupported actions, and status command execution.
