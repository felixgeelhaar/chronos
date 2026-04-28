# Configuration

Chronos is configured exclusively through `CHRONOS_*` environment variables. The CLI accepts a small set of flags that override the corresponding env vars at runtime.

## Reference

| Variable | Default | Used by | Description |
|---|---|---|---|
| `CHRONOS_DB_TYPE` | `sqlite` | both | Backend: `sqlite`, `postgres`, or `memory`. |
| `CHRONOS_DB_CONN` | `chronos.db` | both | SQLite path (or `:memory:`) or Postgres connection string. |
| `CHRONOS_MAX_SIGNALS` | `10` | `compute` | Cap on signals produced per detect run. |
| `CHRONOS_JOB_TIMEOUT` | `10m` | `compute` | Overall compute timeout (Go duration syntax). |
| `CHRONOS_SIM_THRESHOLD` | `0.85` | Recurrence | Minimum cosine similarity for a peer to count. |
| `CHRONOS_MIN_SAMPLE` | `2` | Recurrence | Minimum peer cases required to emit. |
| `CHRONOS_TREND_MIN_SLOPE` | `0.05` | Trend (Tier B) | Minimum absolute regression slope. |
| `CHRONOS_TREND_MIN_POINTS` | `4` | Trend (Tier B) | Minimum observations to consider a trend. |
| `CHRONOS_SPIKE_Z` | `2.5` | Spike (Tier B) | Absolute z-score threshold for a spike. |
| `CHRONOS_DROP_Z` | `2.5` | Drop (Tier B) | Absolute z-score threshold for a drop. |
| `CHRONOS_SPIKE_WINDOW` | `5` | Spike/Drop (Tier B) | Rolling baseline size in points. |
| `CHRONOS_STALL_MAX_STDDEV` | `0.05` | Stall | Max stddev of normalised outcome to qualify. |
| `CHRONOS_STALL_MIN_POINTS` | `4` | Stall | Minimum observations to qualify. |
| `CHRONOS_ANOMALY_MAX_SIM` | `0.5` | Anomaly | Max cosine similarity to nearest peer to count as isolated. |
| `CHRONOS_ANOMALY_MIN_PEERS` | `2` | Anomaly | Minimum peers required for cross-entity comparison. |
| `CHRONOS_SEASONALITY_MIN_AUTOCORR` | `0.5` | Seasonality | Minimum autocorrelation at any lag. |
| `CHRONOS_SEASONALITY_MIN_POINTS` | `12` | Seasonality | Minimum observations to consider. |
| `CHRONOS_SEASONALITY_MIN_PERIOD` | `2` | Seasonality | Minimum lag (period) considered. |
| `CHRONOS_CORRELATION_MIN` | `0.7` | Correlation | Minimum |Pearson r| between two series to emit. |
| `CHRONOS_CORRELATION_MIN_POINTS` | `5` | Correlation | Minimum aligned observations between two series. |
| `CHRONOS_HTTP_PORT` | `7778` | `serve` | HTTP listen port. |
| `CHRONOS_HTTP_HOST` | `127.0.0.1` | `serve` | HTTP listen host. Use `0.0.0.0` to bind all interfaces. |
| `CHRONOS_VERBOSE` | unset | CLI | When set to any non-empty value, prints the cause chain on errors. |

`serve` flags `--port` and `--host` override their env counterparts. `compute` accepts `--scope-id` (preferred) or `--coach-id` (legacy alias).

## Precedence

CLI flags > environment variables > defaults baked into `internal/config/config.Default()`.

## Tuning detectors

Each detector has its own knob namespace (`CHRONOS_<DETECTOR>_*`) so you can tune one without affecting others. Recurrence is the only detector enabled in Tier A; the rest become live as Tier B detectors land in `internal/detect/`.

- **Recurrence** (`SIM_THRESHOLD`, `MIN_SAMPLE`) — raise threshold for fewer, more specific peers; lower it for more candidates. Below ~0.7 admits noise. `MIN_SAMPLE` of 2 is the smallest defensible value; five is the saturation point of the confidence sample-factor.
- **Trend / Spike / Drop / Stall** thresholds influence trigger sensitivity; smaller windows react faster but produce more noise.
- **Anomaly** (`MAX_SIM`, `MIN_PEERS`) — lower `MAX_SIM` makes anomalies harder to qualify (only truly isolated entities trigger); higher `MIN_PEERS` requires more cohort coverage.
- **Seasonality** (`MIN_AUTOCORR`, `MIN_POINTS`, `MIN_PERIOD`) — `MIN_POINTS` should be at least two full periods; raising `MIN_PERIOD` past 2 avoids labelling noisy near-flat series as periodic.
- **Correlation** (`MIN`, `MIN_POINTS`) — pairwise correlation cost is O(N²) in series count per scope. Tighten `MIN` and `MIN_POINTS` for noisy data.

## Backend choice

- **`memory`** — tests and one-off exploration. State is lost on process exit.
- **`sqlite`** — single-process or embedded deployments. Pure Go (`modernc.org/sqlite`), no CGO. Use `:memory:` for ephemeral, a path for persistent.
- **`postgres`** — multi-process and production. Set `CHRONOS_DB_CONN` to a libpq connection string (`postgres://user:pass@host:5432/chronos?sslmode=disable`). Tables are created on first connect.

## Operations

- The HTTP server installs a `SIGINT`/`SIGTERM` handler that drains in-flight requests for up to 10 seconds before shutdown.
- The compute pipeline applies `CHRONOS_JOB_TIMEOUT` as a `context.WithTimeout` covering fetch, save, detect, and persist. Long-running upstreams should respect the supplied `context.Context`.
- `POST /v1/ingest` is the streaming entry point. Compute does not run automatically on ingest; run `chronos compute` (or call detection from a scheduler) at the cadence appropriate for your patterns.
- Logs are emitted via `log/slog` to stderr with a text handler at info level.

## Authentication (HTTP API)

The reference server does not authenticate today. Deployments that need auth wrap `internal/api.Server` in their own middleware (token check, mTLS, JWT) before mounting the routes. The `client.WithToken(...)` option is wire-ready for the standard `Authorization: Bearer <token>` header.
