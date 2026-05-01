# Changelog

All notable changes to Chronos are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The wire contract documented in [`docs/wire-contract.md`](docs/wire-contract.md) is the stability boundary. Renaming any documented Pattern, Evidence.Kind, or metric key is a major-version change.

## [Unreleased]

### Added
- **Postgres bulk-load path** — `EntityStateRepository.Save` switches to chunked multi-row `INSERT…ON CONFLICT` above `BulkSaveThreshold` (200 rows). Chunks of 1 000 rows keep us under the 65 535 placeholder limit. Backfills now batch-commit; per-row UPSERT path retained for small batches where round-trip overhead amortises poorly.
- **Durable webhook outbox** — `notify.OutboxConfig.PersistencePath` opts the in-memory outbox into JSON-on-disk persistence. Enqueue / flush snapshot to atomic-rename file; startup re-loads pending deliveries. Survives process restart.
- **Detector parallelism** — `Engine.WithParallelDetectors(true)` runs every (scope, detector) pair in its own goroutine. Off by default; flip on via `CHRONOS_DETECTOR_PARALLELISM=1`. Race tests pass.
- **At-least-once webhook outbox** (`internal/notify/outbox.go`) — `Outbox` wraps an `AckingNotifier` with retry+exponential-backoff (default 5 attempts, 1s→30s). Failed deliveries reschedule until success or max-attempts. In-memory only; restart loses pending. Operators wanting durable delivery wire their own persistent notifier.
- **OutlierCluster detector** (`PatternTypeOutlierCluster`) — detects time windows where multiple series in a scope go anomalous together. Cohort-level, not per-series. Tunable via `CHRONOS_OUTLIER_CLUSTER_MIN_SERIES` (default 3), `CHRONOS_OUTLIER_CLUSTER_Z` (default 2.5), `CHRONOS_OUTLIER_CLUSTER_WINDOW` (default 5m). Signals carry `member_count` and one `outlier_member` evidence row per participating series.
- **CrossScopeCorrelation detector** (`PatternTypeCrossScopeCorrelation`) — pairwise Pearson correlation across (scope, series) pairs in DIFFERENT scopes. Tunable via `CHRONOS_CROSS_SCOPE_MIN` (default 0.8) and `CHRONOS_CROSS_SCOPE_MIN_POINTS` (default 5). Drops same-scope pairs (handled by within-scope `Correlation`).
- **`CrossScopeDetector` interface** — parallel to `Detector`; runs once over the full state list rather than per-scope. Engine picks up `DefaultCrossScopeDetectors` automatically.
- **Write-coalescing repository decorator** (`internal/store/batching`) — opt-in `EntityStateRepository` wrapper that batches Ingest calls into a single Save transaction. Tunable via `MaxBatch` and `MaxWait`. `Close` drains the buffer on shutdown.
- **ChangePoint detector** (`PatternTypeChangePoint`) — best-split mean-shift test detects step changes in the outcome metric. Distinct from Spike/Drop; emits `regime_before` / `regime_after` evidence with `mean`, `stddev`, `n` per regime; `shift` and `split_index` in signal metrics. Tunable via `CHRONOS_CHANGEPOINT_MIN_SHIFT` (default 1.5) and `CHRONOS_CHANGEPOINT_MIN_POINTS` (default 8). Wired into `DefaultDetectors`.
- **SSE `Last-Event-ID` replay** — clients reconnect with the standard `Last-Event-ID` HTTP header (or `?last_event_id=<uuid>` query fallback) and the server replays signals detected at or after the cursor before continuing the live stream. Each SSE frame now carries an `id:` line.
- `ROADMAP.md` — public roadmap with status, next steps, non-goals, semver policy.
- `docs/DEPLOYMENT.md` — operator runbook (k8s shape, Prometheus scrape, Grafana import, ops procedures, known limitations).
- `docs/SLOs.md` — formal availability / latency / error-rate / freshness SLOs with error-budget burn-rate tiers.
- `deploy/grafana/dashboards/chronos-overview.json` — reference Grafana dashboard (signals, observations, HTTP, webhooks, adapters).
- `docs/wire-contract.md` — "Transport parity" section explaining HTTP-string ↔ gRPC enum mapping and metric-key parity; `change_point` Pattern entry; SSE replay protocol.
- gRPC: new `PATTERN_TYPE_CHANGE_POINT = 9` enum value in `api/proto/chronos/v1/chronos.proto`; `client.PatternTypeChangePoint` constant.
- Config: `CHRONOS_CHANGEPOINT_MIN_SHIFT`, `CHRONOS_CHANGEPOINT_MIN_POINTS`.
- `CHANGELOG.md`.

### Changed
- `docs/configuration.md` — added `CHRONOS_GRPC_PORT` / `CHRONOS_GRPC_HOST` rows; clarified that `CHRONOS_API_TOKEN` gates both transports; documented `--grpc-port` flag override.
- `docs/backlog.md` — gRPC moved to "Recently shipped"; stale work-in-progress stub removed.
- `README.md` — API section split into HTTP and gRPC subsections; quickstart shows `--grpc-port`.

## [Released via GoReleaser]

Prior releases are tracked as GitHub Releases. Notable feature waves:

- **Eight detectors GA**: Recurrence, Trend, Spike, Drop, Stall, Anomaly, Seasonality, Correlation.
- **Multi-backend storage** per [Mnemos ADR-0001](https://github.com/felixgeelhaar/Mnemos/blob/main/docs/adr/0001-multi-backend-storage.md): `memory`, `sqlite`, `postgres`, `mysql`/`mariadb`, `libsql`. Wire-compatible engines (CockroachDB, YugabyteDB, Neon, Crunchy Bridge, TimescaleDB, AlloyDB Omni; PlanetScale, TiDB, Vitess; Turso) work through the native providers unchanged.
- **gRPC transport** alongside HTTP, with `Ingest` (client-streaming) and `ListSignals`. Schema in `api/proto/chronos/v1/chronos.proto`. Bearer auth shared with HTTP.
- **Push notifications**: webhooks (HMAC-SHA256-signed) and Server-Sent Events.
- **In-process detection scheduler** (`CHRONOS_DETECTION_INTERVAL`) — required for SSE to receive events.
- **VEX security audit** complete; nox baseline tracked in-tree.
