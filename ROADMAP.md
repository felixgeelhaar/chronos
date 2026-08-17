# Chronos Roadmap

Chronos is the **time / pattern perception** layer of the cognitive stack (Mnemos → Chronos → agent runtimes). The engine is feature-complete for the v1 contract; this roadmap covers what comes next.

## Status (May 2026)

### ✅ Shipped

**Core engine.**
- `internal/domain` — `Signal`, `Evidence`, `TimeWindow`, `PatternType`, validation, normalization.
- `internal/detect` Engine + eleven detectors: Recurrence, Trend, Spike, Drop, Stall, Anomaly, Seasonality, Correlation, ChangePoint, OutlierCluster, plus CrossScopeCorrelation.
- `internal/pipeline.Compute` — orchestration: fetch → save → detect → save signals.
- `internal/similarity` — cosine, weighted, Euclidean.

**Storage (per [Mnemos ADR-0001](https://github.com/felixgeelhaar/Mnemos/blob/main/docs/adr/0001-multi-backend-storage.md)).**
- `memory://` — in-process backend for tests.
- `sqlite://` — pure-Go (`modernc.org/sqlite`), default for single-process deployments.
- `postgres://` — production multi-process backend; verified compat with CockroachDB, YugabyteDB, Neon, Crunchy Bridge, TimescaleDB, AlloyDB Omni.
- `mysql://` / `mariadb://` — verified compat with PlanetScale, TiDB, MariaDB, Vitess.
- `libsql://` — Turso remote and local libSQL files.

**Transports.**
- HTTP REST (`internal/api`) — `/v1/ingest`, `/v1/signals`, `/v1/signals/{id}`, `/v1/signals/stream` (SSE), `/health`, `/metrics`.
- gRPC (`internal/api/grpc`) — `Ingest` (unary) + `ListSignals` + `GetSignal`. Schema in `api/proto/chronos/v1/chronos.proto`. Runs alongside HTTP on a separate port.
- Webhook push (`CHRONOS_WEBHOOK_URLS`) with HMAC-SHA256 signing.

**Infra.**
- Bearer-token auth on HTTP and gRPC (shares `CHRONOS_API_TOKEN`).
- Conventional Commits, golangci-lint clean, race-tests green on Go 1.25 and 1.26.
- GoReleaser — Homebrew formula updates, Docker images.
- coverctl per-domain coverage gating; nox security baseline.

**Adapters.**
- The engine is domain-agnostic. `chronos.Source` is the seam; adapters live in the repo that owns the domain. Verified integration: [`felixgeelhaar/ascend`](https://github.com/felixgeelhaar/ascend) (athlete training weeks).

## Next

Shipped items from earlier roadmap slices (SSE replay, outbox, ChangePoint / OutlierCluster / CrossScopeCorrelation, write-coalescing, postgres bulk-load, detector parallelism, ops docs) stay checked in git history; they are no longer open work. Remaining scope:

### 1. Transport parity (gRPC stays a subset)

HTTP is the full surface: batch ingest, config validate, federation export, SSE, cursor pagination. gRPC is unary `Ingest` + `ListSignals` + `GetSignal`. That subset is **intentional until someone needs the rest on proto** — adding client-streaming ingest or SSE-equivalent RPCs is a versioned schema change, not a drive-by. Document the subset in README / wire-contract; do not silently claim streaming ingest.

### 2. Signal identity over time

The scheduler now skips re-saves when `(scope, series, pattern, window)` already exists. It does **not** collapse successive windows as new points arrive (trend/spike `window.End` grows → new row). Optional later work:

- Content-addressed signal IDs (hash of perception identity) as an alternative to the List-then-skip check.
- An upsert / "current window" mode for operators who want one live row per series+pattern.

### 3. Detector observability

Per-pattern latency, skip, and drop counters (including truncation by `CHRONOS_MAX_SIGNALS`). Process-level Prometheus metrics exist; they are not broken out per detector.

### 4. Capability ports

`ports.TextSearcher` and `ports.VectorSearcher` remain unused. Implement them only when a detector actually needs FTS or embeddings.

### 5. Adapter ecosystem (community-driven)

The point of the no-adapters-in-Chronos rule is that adapters live close to their domain. Anticipated near-term integrations from neighbouring projects:

- **Mnemos action+outcome stream** — feed Mnemos's recorded outcomes into Chronos as a metric series; pattern detection on the outcome.
- **decisionkit risk score over time** — feed [decisionkit](https://github.com/felixgeelhaar/decisionkit) risk-score time series; detect "risk piling up" patterns. (Nous owned this when it was a live service; it is archived.)

Both are out-of-tree adapters. This roadmap tracks them only as expected use cases — implementation belongs to the consuming repo.

## Non-goals

- Becoming a TSDB. Chronos persists what it must to detect; it is not a Prometheus / VictoriaMetrics replacement.
- Becoming an alerting system. Chronos emits signals; alerting is downstream.
- Adding domain-specific detectors. The engine is domain-agnostic by design — detectors must work generically over `EntityState` features.

## Versioning policy

Chronos follows [Semantic Versioning](https://semver.org/). The wire contract documented in `docs/wire-contract.md` is the stability boundary; renaming any documented Pattern, Evidence.Kind, or metric key is a major-version change.
