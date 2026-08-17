# Backlog

Open items not yet scoped into a release. Closed items move out of this file once shipped.

## Recently shipped

- **gRPC transport** — Implemented in `internal/api/grpc/`, schema in `api/proto/chronos/v1/chronos.proto`. Wired into `cmd/chronos/serve.go` via `--grpc-port` / `CHRONOS_GRPC_PORT`. Provides `Ingest` (unary), `ListSignals`, and `GetSignal` with the same filter shape as `/v1/signals`. Bearer auth shared with HTTP.
- **Detector explainability** — all eleven detectors populate `Signal.Explanation` (`feature_evolution`, `comparable_peers`, `threshold_used`, `detector_version`). Numeric/structured only; no prose.
- **Scheduler same-window skip** — `pipeline.Scheduler` does not re-save when `(scope, series, pattern, window start/end)` already exists. Unchanged data no longer appends a new UUID every tick.
- **Public Go client HTTP parity** — `ListPage` / `since_cursor`, `Scopes` (`scope_in`), `IngestBatch`, `FederationExport`, plus `Explanation` / `ConfidenceClass`.
- **Default `CHRONOS_MAX_SIGNALS=100`** — was 10, which silently truncated multi-detector runs. `0` remains unlimited.
- **OutlierCluster persist** — `Signal.Validate` allows `Series == uuid.Nil` for that pattern; MySQL stores explanation / confidence class and honours `ScopeIDs`.

## Open

### gRPC is a subset of HTTP (intentional for now)

HTTP has batch ingest, config validate, federation export, SSE, and opaque cursor pagination. gRPC has unary `Ingest`, `ListSignals`, and `GetSignal`. Earlier roadmap text claimed client-streaming ingest; the proto is unary. Closing the gap is a proto/buf-breaking design choice — do it as a versioned RPC addition, or keep documenting the subset.

### Scheduler skip is window-identity, not content-addressed

A later tick that **grows** `window.End` (new observations) still emits a new row. That is correct for trend/spike as the perception window moves, but operators who want a single live row per `(scope, series, pattern)` still need downstream de-dupe or a content-addressed signal id. The skip only stops duplicates when the window is unchanged.

### Capability ports unused

`ports.TextSearcher` and `ports.VectorSearcher` are declared for forward compatibility and have no implementations or callers. Leave them until a detector actually needs full-text or embedding search; do not add unused store implementations.

### Detector observability

Detectors are fire-and-forget. Operators may want per-pattern latency, skip, and drop counters (especially when `CHRONOS_MAX_SIGNALS` truncates). Metrics exist at the HTTP/process level; they are not broken out per detector today.
