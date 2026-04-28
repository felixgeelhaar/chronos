# Chronos

<p align="center">
  <strong>Time & Pattern Perception in the Cognitive Stack</strong>
</p>

<p align="center">
  <a href="https://github.com/felixgeelhaar/chronos/actions"><img src="https://github.com/felixgeelhaar/chronos/workflows/CI/badge.svg" alt="CI Status"></a>
  <a href="https://goreportcard.com/report/github.com/felixgeelhaar/chronos"><img src="https://goreportcard.com/badge/github.com/felixgeelhaar/chronos" alt="Go Report Card"></a>
  <a href="https://github.com/felixgeelhaar/chronos/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

---

Chronos ingests time-series observations from any source and emits structured **signals** describing the patterns it sees — recurrences, trends, spikes, drops, stalls. It does not decide, act, or render prose. Signals are perception, not opinion.

Chronos sits between **Mnemos** (memory) and **Nous** (decisions) in the cognitive stack, alongside **Praxis** (execution). See [`docs/cognitive-stack.md`](docs/cognitive-stack.md) for how the four systems compose.

## Design principles

- **Signals, not opinions.** Each signal carries Pattern, Strength, Confidence, Window, and Evidence. There is no Title, no Summary, no Suggestion. Interpretation is Nous's job.
- **Domain-agnostic.** Athletes, servers, sensors, stocks — all flow through the `chronos.Source` adapter port.
- **Loosely coupled.** Chronos works standalone. The stack composes through stable contracts, not internal coupling.
- **Lightweight.** Single Go binary. Pure-Go SQLite (no CGO). Optional PostgreSQL for production.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Adapter    │     │  Engine     │     │  API + SDK  │
│  (Source)   │ ──▶ │ (detectors) │ ──▶ │  (signals)  │
│             │     │             │     │             │
│ • Ascend    │     │ • Detect    │     │ • REST      │
│ • Prometheus│     │ • Score     │     │ • client/   │
│ • InfluxDB  │     │ • Persist   │     │ • Ingest    │
└─────────────┘     └─────────────┘     └─────────────┘
```

Detailed layering and invariants: [`docs/architecture.md`](docs/architecture.md). Cognitive-stack context: [`docs/cognitive-stack.md`](docs/cognitive-stack.md).

## Quick start

### Install

Chronos ships as a single static binary (no CGO, no runtime dependencies). Pick whichever channel suits your environment.

**Homebrew (macOS, Linux)**

```bash
brew tap felixgeelhaar/tap
brew install chronos
```

**Docker (any OCI runtime)**

```bash
docker run --rm -p 7778:7778 ghcr.io/felixgeelhaar/chronos:latest
# Multi-arch image: linux/amd64 + linux/arm64. Distroless, ~2 MB.
```

**Debian / Ubuntu (.deb)**

```bash
# Replace <version> and <arch> (amd64|arm64) with the desired release.
curl -fsSL -o chronos.deb \
  https://github.com/felixgeelhaar/chronos/releases/download/v<version>/chronos_<version>_linux_<arch>.deb
sudo dpkg -i chronos.deb
```

**RHEL / Fedora (.rpm)** and **Alpine (.apk)** are produced for the same OS/arch matrix; substitute the file extension.

**Prebuilt binary archive (any OS)**

```bash
# Linux/macOS/Windows × amd64/arm64 (windows-arm64 is intentionally skipped).
curl -fsSL -o chronos.tar.gz \
  https://github.com/felixgeelhaar/chronos/releases/download/v<version>/chronos_<version>_<os>_<arch>.tar.gz
# Verify against the published checksums.
curl -fsSL -O \
  https://github.com/felixgeelhaar/chronos/releases/download/v<version>/checksums.txt
shasum -a 256 -c checksums.txt --ignore-missing
tar -xzf chronos.tar.gz && sudo install -m 0755 chronos /usr/local/bin/chronos
```

**Go install (HEAD)**

```bash
go install github.com/felixgeelhaar/chronos/cmd/chronos@latest   # requires Go 1.23+
```

**Source build**

```bash
git clone https://github.com/felixgeelhaar/chronos.git
cd chronos
make build   # binary lands in ./bin/chronos with version/commit/buildDate ldflags
```

**Supported targets**

| OS      | amd64 | arm64 | Distribution channels                              |
|---------|:-----:|:-----:|----------------------------------------------------|
| linux   |  ✓    |  ✓    | Homebrew, Docker, .deb, .rpm, .apk, tar.gz, source |
| darwin  |  ✓    |  ✓    | Homebrew, tar.gz, source                           |
| windows |  ✓    |  —    | zip archive, source                                |

### Run

```bash
# Configure
export CHRONOS_DB_TYPE=sqlite
export CHRONOS_DB_CONN=chronos.db

# Compute signals from a pull-based adapter
./bin/chronos compute --adapter=ascend --scope-id=<uuid>

# Start the HTTP API (also accepts streaming /v1/ingest)
./bin/chronos serve --port=7778

# Query signals
curl 'http://localhost:7778/v1/signals?scope_id=<uuid>&pattern=recurrence&min_confidence=0.7'
```

Full configuration reference: [`docs/configuration.md`](docs/configuration.md).

## Writing an adapter

```go
package myadapter

import (
    "context"

    "github.com/felixgeelhaar/chronos"
)

type Source struct{}

func (s *Source) Name() string { return "my-source" }

func (s *Source) Fetch(ctx context.Context, cfg map[string]string) ([]chronos.EntityState, error) {
    // Map your domain into chronos.EntityState. Last feature is the outcome metric.
    return states, nil
}

func init() { chronos.Register(&Source{}) }
```

Adapters self-register. Add a blank import in your binary so `init()` fires:

```go
import _ "example.com/myadapter"
```

Full guide: [`docs/adapters.md`](docs/adapters.md).

## Reading signals

```go
import "github.com/felixgeelhaar/chronos/client"

c, _ := client.New("http://chronos.local:7778",
    client.WithToken(os.Getenv("CHRONOS_TOKEN")),
    client.WithTimeout(10*time.Second),
)

// Recent recurrence signals for a scope
signals, err := c.Signals().
    Scope(scopeID).
    Pattern(client.PatternTypeRecurrence).
    MinConfidence(0.7).
    Limit(20).
    List(ctx)
```

For streaming sources you can ingest single observations:

```go
_, err := c.Ingest(ctx, client.IngestRequest{
    EntityID:  entityID,
    ScopeID:   scopeID,
    Timestamp: time.Now(),
    Features:  []float64{f1, f2, f3, outcome},
    Adapter:   "my-source",
})
```

## API

```
GET  /health                              Liveness/readiness
POST /v1/ingest                           Stream a single observation
GET  /v1/signals                          List signals (filter by scope/pattern/series/since/until/min_confidence/limit)
GET  /v1/signals/<id>                     Fetch a single signal with evidence
```

Full filter reference is in the API docs.

## Pattern detectors

| Pattern         | What it detects                                                            | Evidence kind          |
|-----------------|----------------------------------------------------------------------------|------------------------|
| `recurrence`    | Subject is in a state other entities have been in before (cosine peers)    | `similar_state`        |
| `trend`         | Sustained directional movement of the outcome metric (linear regression)   | `regression_summary`   |
| `spike`         | Sharp positive deviation from the rolling baseline (z-score)               | `baseline_deviation`   |
| `drop`          | Sharp negative deviation from the rolling baseline (z-score)               | `baseline_deviation`   |
| `stall`         | Outcome variance falls below threshold over a window (normalised stddev)   | `variance_window`      |
| `anomaly`       | Subject is unlike its peers' current states (cross-entity dual of `recurrence`) | `peer_distance`   |
| `seasonality`   | Periodic structure in the outcome series (autocorrelation peak)            | `autocorrelation_peak` |
| `correlation`   | Two series in the same scope move together (pairwise Pearson)              | `pair_correlation`     |

## Bundled adapters

| Name     | Status        | Purpose                                              |
|----------|---------------|------------------------------------------------------|
| `ascend` | First-party   | Ascend coaching platform (PostgreSQL source)         |
| `memory` | Built-in      | In-memory backend used by tests                      |

## Development

```bash
make test          # go test -race -count=1 ./...
make check         # fmt + vet + test
make sqlc          # Regenerate SQLite query code
make build         # Builds with version/commit/buildDate ldflags
```

### Pre-commit hooks

Pre-commit catches style and lint failures locally, before CI. One-time setup per clone:

```bash
pip install pre-commit                   # or: brew install pre-commit
make precommit-install                   # installs pre-commit + commit-msg hooks
make precommit                           # run all hooks against the working tree
```

The hook set (gofmt, go vet, go mod tidy, golangci-lint, file hygiene, Conventional Commits) is a strict subset of CI; passing locally guarantees CI will not reject on style or lint.

Working conventions for human and agent contributors: [`AGENTS.md`](AGENTS.md). Contribution guidelines: [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

MIT — see [`LICENSE`](LICENSE).

## Companion projects

- **[Mnemos](https://github.com/felixgeelhaar/Mnemos)** — Memory & Knowledge ("what happened, what do we know")
- **Chronos** — Time & Pattern Perception ("what is changing, what's emerging")
- **Praxis** — Execution / Capabilities ("what can be done")
- **Nous** — Coordination / Intelligence ("what should happen, by whom, when")
