# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

```bash
make build         # Build ./bin/chronos with version/commit/buildDate ldflags
make test          # go test -race -count=1 ./...
make check         # fmt + vet + test
make lint          # golangci-lint run ./...
make sqlc          # Regenerate internal/store/sqlite/sqlcgen from sql/sqlite/queries.sql
make clean         # Remove ./bin and *.db

# Run a single test
go test -race -run TestRecurrence ./internal/detect
go test -race -run '^TestSQLite_' ./internal/store/sqlite
```

CI (`.github/workflows/ci.yml`) runs the same `go test -race -count=1` plus `golangci-lint` and `make build` on Go 1.22 and 1.23. The codebase targets Go 1.23 (`go.mod`); match locally before pushing.

## Project intent

Chronos is the **Time / Pattern Perception** layer of the cognitive stack (Mnemos → Chronos → Nous → Praxis). It ingests time-series observations and emits structured **signals** — `Recurrence`, `Trend`, `Spike`, `Drop`, `Stall`, … Each signal carries Pattern, Strength (intensity), Confidence (sureness), Window, Evidence, and Metrics.

Two non-negotiable rules:

1. **Signals, not opinions.** Chronos perceives; Nous interprets. No Title/Summary/Suggestion in domain or wire types. No dismissal, no feedback, no IsActive. Those are Nous and Mnemos concerns.
2. **The core engine knows nothing about the domain.** Domain knowledge enters only through adapters that produce `chronos.EntityState`. If domain-specific code starts leaking into `internal/` or the top-level `chronos` package, that is a design break.

See [`docs/cognitive-stack.md`](docs/cognitive-stack.md) for how the four systems compose.

## Architecture

Public surface (stable):

- **Top-level `chronos` package** — `EntityState`, `Source` interface, registry (`Register`, `Get`, `Adapters`). Adapter authors import this.
- **`client/`** — HTTP-API SDK for consumers of a running Chronos server. Functional options (`WithToken`, `WithLogger`, `WithTimeout`, `WithUserAgent`) and a fluent builder (`c.Signals().Scope(id).Pattern(...).List(ctx)`).

Private (everything under `internal/`, may change without notice):

```
(out-of-tree      →  internal/pipeline   →  internal/store/...   →  internal/api
 chronos.Source)     orchestrator wired       per-aggregate           HTTP transport
 from cmd/chronos         repositories             + DTO conversion
       ↓
 internal/detect.Engine
       ↓
 detectors (one per PatternType)
```

Layering, in dependency order (inner → outer):

1. `internal/domain` — pure types: `Signal`, `Evidence`, `TimeWindow`, `PatternType`, plus `Validate` / `Normalise`. No I/O. No prose.
2. `internal/ports` — outbound interfaces in one file: `EntityStateRepository` (with both `Ingest` and batch `Save`), `SignalRepository` (`Save`/`List(filter)`/`Get`/`Count`). No `Dismiss`, no `Active`, no `Feedback`.
3. `internal/similarity` — pure math (cosine, weighted, Euclidean).
4. `internal/detect` — detectors + Engine. One file per pattern: `recurrence.go` (and Tier-B `trend.go`, `spike.go`, `drop.go`, `stall.go`).
5. `internal/store/{memory,sqlite,postgres}` — port implementations, one `Conn` per backend wiring two repositories sharing a `*sql.DB`.
6. `internal/store` (top of subtree) — `Open(ctx, dbType, connStr)` factory returns a uniform port-typed `*Conn`.
7. `internal/pipeline` — orchestration: `Compute(ctx, ComputeInput)` runs fetch → save observations → detect → save signals.
8. `internal/api` — HTTP transport + `SignalDTO` / `IngestRequest`. Strictly transport — no rendering of human-readable copy.
9. `cmd/chronos` — flat CLI: `main.go` + one file per subcommand + `errors.go` (`ChronosError{Code,Message,Cause,Hint}`).

### Invariants worth knowing

- **Last feature is the outcome metric. Higher is better.** Recurrence emits an `outcome_diff = peer.outcome - subject.outcome` per evidence row; Tier-B detectors will reuse this convention for slope and z-score sign.
- **Scope is the grouping primitive.** Detectors operate within a `ScopeID`. The Engine groups states by scope before fan-out.
- **Self-comparison and forward-looking comparisons are excluded** by the Recurrence detector.
- **Adapter registration is `init()`-driven.** `chronos.Register(src)` panics on a nil source or empty name, so wiring mistakes surface at program start. The CLI imports adapter packages with `_` so their `init()` runs.

### Persistence layout

- **SQLite**: pure-Go `modernc.org/sqlite` driver. PRAGMAs (`foreign_keys`, `journal_mode=wal`, `busy_timeout`) encoded in the DSN. Migrations are embedded via `go:embed migrations/001_initial.sql`. sqlc-generated code lives in `internal/store/sqlite/sqlcgen/`. Two aggregates: `entity_states` and `signals` + `signal_evidence`.
- **Postgres**: hand-written queries (Postgres is a secondary backend). Schema embedded via `go:embed` from `internal/store/postgres/migrations/001_initial.sql`. Same two-aggregate shape.
- **Memory**: thread-safe in-memory backend used for tests.

To change SQL: edit `sql/sqlite/queries.sql` and/or the relevant migration file, then `make sqlc`.

### Configuration

All env-var driven. Defaults in `config.Default()` (`internal/config/config.go`):

| Variable | Default | Notes |
|---|---|---|
| `CHRONOS_DB_TYPE` | `sqlite` | `sqlite` / `postgres` / `memory` |
| `CHRONOS_DB_CONN` | `chronos.db` | Path or DSN |
| `CHRONOS_MAX_SIGNALS` | `10` | Cap per detect run |
| `CHRONOS_JOB_TIMEOUT` | `10m` | Compute timeout |
| `CHRONOS_SIM_THRESHOLD` | `0.85` | Recurrence: min cosine similarity |
| `CHRONOS_MIN_SAMPLE` | `2` | Recurrence: min peer cases |
| `CHRONOS_TREND_MIN_SLOPE` | `0.05` | Trend: minimum |slope| (Tier B) |
| `CHRONOS_TREND_MIN_POINTS` | `4` | Trend: min observations (Tier B) |
| `CHRONOS_SPIKE_Z` | `2.5` | Spike: |z-score| threshold (Tier B) |
| `CHRONOS_DROP_Z` | `2.5` | Drop: |z-score| threshold (Tier B) |
| `CHRONOS_SPIKE_WINDOW` | `5` | Spike/Drop rolling baseline size (Tier B) |
| `CHRONOS_STALL_MAX_STDDEV` | `0.05` | Stall: max stddev of normalised outcome (Tier B) |
| `CHRONOS_STALL_MIN_POINTS` | `4` | Stall: min observations (Tier B) |
| `CHRONOS_ANOMALY_MAX_SIM` | `0.5` | Anomaly: max similarity to nearest peer to count as isolated |
| `CHRONOS_ANOMALY_MIN_PEERS` | `2` | Anomaly: minimum peers for cross-entity comparison |
| `CHRONOS_SEASONALITY_MIN_AUTOCORR` | `0.5` | Seasonality: minimum autocorrelation at any lag |
| `CHRONOS_SEASONALITY_MIN_POINTS` | `12` | Seasonality: minimum observations |
| `CHRONOS_SEASONALITY_MIN_PERIOD` | `2` | Seasonality: minimum lag (period) considered |
| `CHRONOS_CORRELATION_MIN` | `0.7` | Correlation: minimum |Pearson r| to emit |
| `CHRONOS_CORRELATION_MIN_POINTS` | `5` | Correlation: minimum aligned observations |
| `CHRONOS_HTTP_PORT` | `7778` | Serve port |
| `CHRONOS_HTTP_HOST` | `127.0.0.1` | Serve host |
| `CHRONOS_VERBOSE` | unset | When set, the CLI prints error causes |

`serve` flags `--port` / `--host` override env. `compute` accepts `--scope-id` (preferred) or `--coach-id` (alias retained for the original Ascend wiring).

## Adding a detector

1. New file in `internal/detect/<pattern>.go`. Implement `Detector`.
2. Add the `PatternType` constant if not already reserved in `internal/domain/types.go`.
3. Register in `detect.DefaultDetectors`.
4. Add config knobs to `internal/config/config.go`.
5. Document evidence `Kind` and `Metrics` keys in [`docs/cognitive-stack.md`](docs/cognitive-stack.md).
6. Tests in `internal/detect/<pattern>_test.go` covering trigger + no-trigger.

## Conventions

- Go 1.23+. Deps: `google/uuid`, `lib/pq`, `modernc.org/sqlite`.
- Conventional Commits.
- Tests are stdlib-only and table-driven where useful.
- Wrap errors with `%w`. The CLI uses `*ChronosError` for exit-code routing.
- Logging is `log/slog`. Libraries return errors; only `cmd/` and `internal/api` log.
