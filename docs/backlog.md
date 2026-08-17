# Backlog

Open items not yet scoped into a release. Closed items move out of this file once shipped.

## Recently shipped

- **gRPC transport parity** — Additive RPCs: `IngestBatch`, `StreamSignals`, `ValidateConfig`, `ExportFederation`, plus `since_cursor` / `next_cursor` on `ListSignals`. Unary `Ingest` unchanged. Schema in `api/proto/chronos/v1/chronos.proto`.
- **Detector explainability** — all eleven detectors populate `Signal.Explanation`.
- **Scheduler same-window skip** plus **content-addressed `PerceptionID`** (UUID v5). Unchanged windows upsert; growing `window.End` still emits a new row.
- **Public Go client HTTP and gRPC parity** — `ListPage`, `Scopes`, `IngestBatch`, `FederationExport`, `ValidateConfig`, `Stream`.
- **Default `CHRONOS_MAX_SIGNALS=100`**.
- **Per-detector observability** — latency, emit, skip, and truncation counters labelled by pattern.
- **OutlierCluster persist** and MySQL explanation / `ScopeIDs` parity.

## Open

### Capability ports unused

`ports.TextSearcher` and `ports.VectorSearcher` are declared for forward compatibility and have no implementations or callers. Leave them until a detector actually needs full-text or embedding search; do not add unused store implementations.

### Adapter ecosystem (out of tree)

Mnemos action+outcome stream and Nous risk-score series are expected consumers. Implementation belongs to those repos.
