# Backlog

Open items not yet scoped into a release. Closed items move out of this file once shipped.

## Recently shipped

- **gRPC transport** — Implemented in `internal/api/grpc/`, schema in `api/proto/chronos/v1/chronos.proto`. Wired into `cmd/chronos/serve.go` via `--grpc-port` / `CHRONOS_GRPC_PORT`. Provides `Ingest` (client-streaming) and `ListSignals` with the same filter shape as `/v1/signals`. Bearer auth shared with HTTP.

## Open

_(none currently)_
