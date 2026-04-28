# AGENTS.md

## Project intent

Chronos is the **Time / Pattern Perception** layer of the cognitive stack ([Mnemos / **Chronos** / Praxis / Nous](docs/cognitive-stack.md)). It ingests time-series observations from any source and emits **signals** — structured records describing patterns: `Recurrence`, `Trend`, `Spike`, `Drop`, `Stall`, …

The single hardest rule: **signals, not opinions.** Chronos perceives; it does not interpret. There is no Title, no Summary, no Suggestion, no dismissal workflow, no feedback. Those concerns belong to Nous (decisions) and Mnemos (knowledge), respectively.

The second hardest rule: **the core engine knows nothing about the domain.** Athletes, servers, sensors — all enter through `chronos.Source` as undifferentiated `EntityState` records.

## Architecture

```
chronos/                       Public adapter SDK: EntityState, Source, registry
client/                        Public HTTP-API SDK for consumers (Nous integrators, dashboards)
cmd/chronos/                   CLI: main.go + one file per subcommand + errors.go
internal/
  domain/                      Private domain model: Signal, Evidence, TimeWindow,
                               PatternType (Recurrence/Trend/Spike/Drop/Stall/...),
                               Validate, Normalise
  ports/                       Outbound interfaces (one file): Source,
                               EntityStateRepository, SignalRepository
  config/                      Env-var driven config + Validate
  similarity/                  Cosine, weighted cosine, Euclidean (pure math)
  detect/                      Detectors + Engine that fans observations out
  pipeline/                    Orchestration: fetch → save observations → detect → save signals
  api/                         HTTP REST layer (transport + DTO conversion)
  store/                       Persistence factory (Open dispatches on dbType)
    memory/                    In-memory backend (test default)
    sqlite/                    modernc.org/sqlite-backed; sqlcgen subpkg holds generated code
    postgres/                  PostgreSQL backend (hand-written queries)
sql/sqlite/                    sqlc query file
adapters/
  ascend/                      First-party adapter for the Ascend coaching platform
docs/                          Architecture, cognitive-stack, adapter authoring, configuration
```

## Public surface

- **`chronos`** — adapter authors import `chronos.EntityState` and implement `chronos.Source`. Adapters self-register via `init()` calling `chronos.Register(src)`.
- **`client`** — consumers (Nous, dashboards) import `client.New(...)` and use the fluent builder (`c.Signals().Scope(id).Pattern(client.PatternTypeRecurrence).MinConfidence(0.7).List(ctx)`).

Anything under `internal/` is private and may change without notice.

## Conventions

- **Go 1.23+**, minimal external deps: `google/uuid`, `lib/pq`, `modernc.org/sqlite` (pure Go — no CGO).
- Numeric features are always `[]float64`. Timestamps are `time.Time` (RFC3339Nano on disk). IDs are `uuid.UUID`.
- **Last feature is the outcome metric**, higher is better. Adapters that violate this convention will produce signals whose evidence metrics (`outcome_diff`, `slope`) are inverted.
- `internal/store/sqlite/sqlcgen/` is sqlc-generated; **do not hand-edit** under normal circumstances. To change SQL: edit `sql/sqlite/queries.sql` and/or `internal/store/sqlite/migrations/001_initial.sql`, then `make sqlc`.
- Tests: standard library only; in-memory SQLite for store integration tests via `Open(":memory:")`. Table-driven where it helps.
- Errors: wrap with `fmt.Errorf("...: %w", err)` in libraries; the CLI returns `*ChronosError` carrying an `ExitCode` and optional `Hint`.
- Logging: `log/slog`. Libraries do not log; only the CLI and HTTP server emit log lines.

## Building

```bash
make build      # ./bin/chronos with version/commit/buildDate baked in
make test       # go test -race -count=1 ./...
make check      # fmt + vet + test
make sqlc       # regenerate internal/store/sqlite/sqlcgen
```

## Adding a pattern detector

1. Add a file under `internal/detect/` (e.g. `trend.go`) implementing the `Detector` interface.
2. Add the corresponding `PatternType` constant in `internal/domain/types.go` (most are reserved already).
3. Wire the detector into `detect.DefaultDetectors`.
4. Add per-detector config knobs in `internal/config/config.go` (`CHRONOS_<DETECTOR>_*`).
5. Document the detector's evidence shape (the `Kind` value and `Metrics` keys).
6. Add unit tests covering the trigger and the no-trigger paths.

## Adding an adapter

1. Create a package under `adapters/<name>/`.
2. Implement `chronos.Source` (`Name`, `Fetch`).
3. Call `chronos.Register(&Source{})` in `init()`.
4. Add a blank import to `cmd/chronos/main.go` so `init()` runs in CLI builds.
5. Document the required `cfg` keys (e.g. `coach_id`).

See `adapters/ascend/ascend.go` for a complete example.
