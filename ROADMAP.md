# Chronos Roadmap

Chronos is the **time / pattern perception** layer of the cognitive stack (Mnemos → Chronos → Nous → Praxis). The engine is feature-complete for the v1 contract; this roadmap covers what comes next.

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

### 1. Stream-replay safety

- [x] **SSE `Last-Event-ID` replay** — server consults the persistence layer on reconnect and replays signals detected at or after the cursor (excluding the cursor itself). `?last_event_id=<uuid>` query fallback for environments that strip the header.
- [x] **At-least-once delivery (durable outbox)** — `internal/notify.Outbox` wraps an `AckingNotifier` with exponential-backoff retry. Optional `PersistencePath` snapshots pending deliveries to JSON-on-disk; restart re-loads pending rows. SSE replay covers the SSE side via `Last-Event-ID`.

### 2. New detectors

- [x] **Change-point detection** — best-split mean-shift test (`PatternTypeChangePoint`). Distinct from Spike/Drop; emits `regime_before` / `regime_after` evidence.
- [x] **Outlier-cluster detection** (`PatternTypeOutlierCluster`) — cohort-level anomaly clusters. Uses each series's own rolling baseline; groups outlier events into time buckets; emits a single signal per qualifying bucket with `member_count` ≥ `CHRONOS_OUTLIER_CLUSTER_MIN_SERIES`.
- [x] **Cross-scope correlation** (`PatternTypeCrossScopeCorrelation`) — pairwise Pearson correlation across (scope, series) pairs in different scopes. New `CrossScopeDetector` interface; engine runs cross-scope detectors after per-scope detectors.

Open question: how much per-detector observability do we expose? Current design is "fire and forget"; operators may want detector latency / drop counters per type.

### 3. Performance

- [x] **Write-coalescing decorator** (`internal/store/batching/Repo`) — opt-in, wraps any `EntityStateRepository` with buffered Ingest. Trades small per-write latency for one fsync per batch; passes through Save / List / Count / DeleteOlderThan unchanged.
- [x] **Postgres bulk-load path** — chunked multi-row `INSERT…ON CONFLICT` for batches above `BulkSaveThreshold` (200 rows). Round-trips drop ~1000× on backfills.
- [x] **Detector parallelism** — `Engine.WithParallelDetectors(true)` runs every (scope, detector) pair in its own goroutine. Wired through `pipeline.NewEngine` via `CHRONOS_DETECTOR_PARALLELISM`.

### 4. Operations

- [x] Deployment runbook — `docs/DEPLOYMENT.md` (k8s shape, Prometheus scrape config, ops procedures, known limitations).
- [x] `CHANGELOG.md` (Keep-a-Changelog format alongside GoReleaser-driven GitHub Releases).
- [x] Reference Grafana dashboard JSON (`deploy/grafana/dashboards/chronos-overview.json`).
- [x] SLO doc with error-budget burn-rate alert tiers (`docs/SLOs.md`).

### 5. Adapter ecosystem (community-driven)

The point of the no-adapters-in-Chronos rule is that adapters live close to their domain. Anticipated near-term integrations from neighbouring projects:

- **Mnemos action+outcome stream** — feed Mnemos's recorded outcomes into Chronos as a metric series; pattern detection on the outcome.
- **Nous risk score over time** — feed Nous's risk-score time series; detect "risk piling up" patterns.

Both are out-of-tree adapters. This roadmap tracks them only as expected use cases — implementation belongs to the consuming repo.

## Non-goals

- Becoming a TSDB. Chronos persists what it must to detect; it is not a Prometheus / VictoriaMetrics replacement.
- Becoming an alerting system. Chronos emits signals; alerting is downstream.
- Adding domain-specific detectors. The engine is domain-agnostic by design — detectors must work generically over `EntityState` features.

## Versioning policy

Chronos follows [Semantic Versioning](https://semver.org/). The wire contract documented in `docs/wire-contract.md` is the stability boundary; renaming any documented Pattern, Evidence.Kind, or metric key is a major-version change.
