# Backlog

Open items not yet scoped into a release. Closed items move out of this file once shipped.

## Recently shipped

- **gRPC transport** — Implemented in `internal/api/grpc/`, schema in `api/proto/chronos/v1/chronos.proto`. Wired into `cmd/chronos/serve.go` via `--grpc-port` / `CHRONOS_GRPC_PORT`. Provides `Ingest` (unary), `ListSignals`, and `GetSignal` with the same filter shape as `/v1/signals`. Bearer auth shared with HTTP.

## Open

### Detector explainability is wire-only

`Signal.Explanation` is persisted and exposed on HTTP/gRPC, but no detector currently populates it. Downstream narrators that expected feature-evolution / threshold / detector-version after #21 still see an omitted field. Filling this in per detector (without adding prose) is the remaining half of that issue.

### Scheduler re-emits on every tick

`pipeline.Scheduler` generates a new UUID per detect run, so the same stall/trend on unchanged data is saved again each `CHRONOS_DETECTION_INTERVAL`. Operators currently have to de-duplicate by `(scope, series, pattern, window)` downstream, or shorten retention. A content-addressed signal id (or a "same window, skip" check) would make the ingest+scheduler path production-safe.

### gRPC is a subset of HTTP

HTTP has batch ingest, config validate, federation export, SSE, and opaque cursor pagination. gRPC has unary `Ingest`, `ListSignals`, and `GetSignal`. ROADMAP previously claimed client-streaming ingest; the proto is unary. Close the gap or document the subset as intentional.

### Public Go client lags the HTTP contract

`client.Signal` now unmarshals `explanation` and `confidence_class`. Still missing: `since_cursor` / `next_cursor` on List, `POST /v1/ingest/batch`, `scope_in` / `scope_ids` allowlists, and `GET /v1/federation/export`.

### Default `CHRONOS_MAX_SIGNALS=10` vs eleven detectors

A single detect run with several series can emit more than ten signals (correlation is O(N²) in series count). The cap silently truncates after sort. Operators with more than a handful of series should raise it; a higher default or `0` (unlimited) may be more honest now that the detector set is larger.

### Capability ports unused

`ports.TextSearcher` and `ports.VectorSearcher` are declared for forward compatibility and have no implementations or callers.
