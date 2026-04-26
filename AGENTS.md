# AGENTS.md

## Project Context

Chronos is a **generic time-series pattern detection engine** written in Go. It is designed to be data-source agnostic through an adapter architecture.

**Key principle:** The core engine knows nothing about weightlifting, athletes, or coaching. All domain knowledge lives in adapters.

## Architecture

```
cmd/chronos/           CLI entrypoint (compute, serve)
internal/
  adapter/             Adapter interface + registry
  config/              Chronos configuration
  extract/             Generic feature extraction pipeline
  similarity/          Cosine similarity + distance metrics
  insights/            Pattern detection + insight generation
  store/sqlite/        SQLite persistence (sqlc-generated)
  api/                 HTTP REST API
pkg/
  vector/              Generic feature vector types
  insight/             Generic insight types
adapters/
  ascend/              Ascend coaching platform adapter
```

## Adapter Interface

Adapters implement `adapter.Source`:

```go
type Source interface {
    Name() string
    Fetch(ctx context.Context, cfg map[string]string) ([]vector.EntityState, error)
}
```

The core engine treats all data as `vector.EntityState` — timestamps, entity IDs, scope (coach/team/tenant), and arbitrary float64 features.

## Coding Conventions

- Go 1.22+
- No external dependencies beyond sqlite3, uuid
- All numeric features are `float64`
- Time is `time.Time`, stored as RFC3339
- IDs are `uuid.UUID` from `github.com/google/uuid`
- Configuration is env-var + flag based, no config files
- Tests are table-driven, use in-memory SQLite

## Building

```bash
make build      # Build cmd/chronos
make test       # Run tests
make sqlc       # Regenerate sqlc code
make check      # fmt + vet + test
```
