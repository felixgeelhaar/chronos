# Chronos

<p align="center">
  <strong>Generic pattern detection engine for time-series data</strong>
</p>

<p align="center">
  <a href="https://github.com/felixgeelhaar/chronos/actions"><img src="https://github.com/felixgeelhaar/chronos/workflows/CI/badge.svg" alt="CI Status"></a>
  <a href="https://codecov.io/gh/felixgeelhaar/chronos"><img src="https://codecov.io/gh/felixgeelhaar/chronos/branch/main/graph/badge.svg" alt="Coverage"></a>
  <a href="https://goreportcard.com/report/github.com/felixgeelhaar/chronos"><img src="https://goreportcard.com/badge/github.com/felixgeelhaar/chronos" alt="Go Report Card"></a>
  <a href="https://github.com/felixgeelhaar/chronos/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

---

Chronos ingests structured time-series data from any source, extracts feature vectors, detects similarity patterns across entities, and surfaces actionable insights with statistical confidence.

Think of it as **"Mnemos for numbers"** — where [Mnemos](https://github.com/felixgeelhaar/Mnemos) reasons about claims and contradictions in text, Chronos reasons about trajectories and patterns in structured data.

## Features

- **Generic engine** — Works with any time-series data: athletes, servers, sensors, stocks
- **Multi-database** — SQLite for local/embedded, PostgreSQL for production
- **Pluggable adapters** — Bring your own data source via a simple interface
- **Statistical insights** — Cosine similarity, confidence scoring, sample-size awareness
- **Coach-in-the-loop** — Every insight is suggest-review-approve, never autonomous
- **Lightweight** — Single Go binary, no ML frameworks, no GPU required
- **Well-tested** — Table-driven tests, race detection, coverage reporting

## Quick Start

### Installation

```bash
# From source
go install github.com/felixgeelhaar/chronos/cmd/chronos@latest

# Or clone and build
git clone https://github.com/felixgeelhaar/chronos.git
cd chronos
make build
```

### Basic Usage

```bash
# Set database (sqlite, postgres, or memory)
export CHRONOS_DB_TYPE=sqlite
export CHRONOS_DB_CONN=chronos.db

# Compute patterns from an adapter
./bin/chronos compute --adapter=ascend --coach-id=your-coach-id

# Start HTTP API
./bin/chronos serve --port=7778

# Query insights
curl http://localhost:7778/v1/insights?scope_id=your-coach-id
```

## Adapters

Chronos is data-source agnostic. Adapters map external data into Chronos' generic model:

| Adapter | Status | Description |
|---------|--------|-------------|
| `ascend` | Proof of concept | Weightlifting coaching platform (PostgreSQL) |
| `memory` | Built-in | In-memory for testing |
| `prometheus` | Planned | Metrics monitoring |
| `influxdb` | Planned | IoT time-series |

**Writing an adapter** is simple — implement the `adapter.Source` interface:

```go
type Source interface {
    Name() string
    Fetch(ctx context.Context, cfg map[string]string) ([]vector.EntityState, error)
}
```

See `adapters/ascend/` for a complete example.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Adapter    │     │  Chronos    │     │  Chronicle  │
│  (external) │ --> │  (engine)   │ --> │  (insights) │
│             │     │             │     │             │
│ • Ascend    │     │ • Extract   │     │ • Similarity│
│ • Prometheus│     │ • Vectorise │     │ • Patterns  │
│ • InfluxDB  │     │ • Store     │     │ • Surface   │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `CHRONOS_DB_TYPE` | `sqlite` | Database type: `sqlite`, `postgres`, `memory` |
| `CHRONOS_DB_CONN` | `chronos.db` | Connection string or file path |
| `CHRONOS_SIM_THRESHOLD` | `0.85` | Minimum cosine similarity for pattern matching |
| `CHRONOS_MIN_SAMPLE` | `2` | Minimum similar cases to generate an insight |
| `CHRONOS_MAX_INSIGHTS` | `10` | Maximum insights per computation run |
| `CHRONOS_HTTP_PORT` | `7778` | HTTP API port |
| `CHRONOS_HTTP_HOST` | `127.0.0.1` | HTTP API host |

## API

### Health Check

```bash
GET /health
```

### List Insights

```bash
GET /v1/insights?scope_id=<uuid>
```

### Dismiss Insight

```bash
POST /v1/insights/<id>
Content-Type: application/json

{
  "dismissed_by": "<user-uuid>"
}
```

## Development

```bash
# Run tests
make test

# Run checks (fmt, vet, test)
make check

# Regenerate sqlc code
make sqlc

# Build
make build

# Run with memory store for testing
CHRONOS_DB_TYPE=memory ./bin/chronos serve
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Chronos is released under the [MIT License](LICENSE).

## Acknowledgments

Chronos was created as a companion to [Ascend](https://github.com/felixgeelhaar/ascend) and [Mnemos](https://github.com/felixgeelhaar/Mnemos), forming a complete evidence stack for coaching decisions:

- **Ascend** = The coaching platform
- **Mnemos** = "What does the literature say?"
- **Chronos** = "What do athletes like mine actually do?"
