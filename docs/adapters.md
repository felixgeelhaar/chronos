# Writing a Chronos adapter

Adapters are the only place in a Chronos deployment that knows about a domain. They turn external data into a slice of `chronos.EntityState`s and let the engine do everything else.

## The contract

```go
package chronos

type Source interface {
    Name() string
    Fetch(ctx context.Context, cfg map[string]string) ([]EntityState, error)
}
```

`Name()` returns a stable identifier (`"ascend"`, `"prometheus"`). It is used to look the adapter up by name and is persisted alongside each entity state for retention queries.

`Fetch` reads from the external system and returns observations. `cfg` is a free-form string→string map that comes through the CLI / API verbatim — typically used to parameterise the query (a coach ID, a tenant ID, a time window).

## EntityState shape

```go
type EntityState struct {
    ID        uuid.UUID         // unique observation ID — generate with uuid.New()
    EntityID  uuid.UUID         // the thing being observed (athlete, server, …)
    ScopeID   uuid.UUID         // the grouping (coach, team, tenant, …)
    Timestamp time.Time         // when the observation was made
    Features  []float64         // numeric feature vector
    Labels    []string          // optional human-readable feature names; len(Labels)==len(Features) when set
    Meta      map[string]string // opaque adapter metadata; not used for similarity
}
```

`Validate()` enforces the invariants:

- `EntityID` and `ScopeID` are non-zero
- `Features` is non-empty
- `Labels`, when set, has the same length as `Features`

## Engine conventions adapters must follow

1. **Last feature is the outcome metric.** Detectors emit metrics derived from this convention: Recurrence's evidence carries `outcome_diff = peer.outcome - subject.outcome`; Trend reports `slope` on it; Spike/Drop compute z-scores on it. **Higher is conventionally better.** If your domain treats "lower is better" (e.g. error rate, latency), invert it before producing the feature.
2. **Features should be on comparable scales.** Cosine similarity is scale-insensitive in direction but feature magnitudes still influence which dimensions dominate. Normalise where it matters (the bundled Ascend adapter divides tonnage by bodyweight).
3. **`ScopeID` is your authority boundary.** Detectors operate within a scope only; the engine never compares across scopes. Pick the right grain — usually the tenant or owner, not a finer slice like "active session."
4. **`Meta` is for downstream consumers, not for similarity.** Anything you put in `Meta` is preserved through the persistence layer and visible at the API; it never enters cosine, regression, or z-score computations.

## Skeleton

```go
// adapters/myadapter/myadapter.go
package myadapter

import (
    "context"
    "fmt"

    "github.com/felixgeelhaar/chronos"
    "github.com/google/uuid"
)

type Source struct {
    // external client / db handle / etc
}

func NewSource(...) (*Source, error) { /* ... */ }

func (s *Source) Name() string { return "myadapter" }

func (s *Source) Fetch(ctx context.Context, cfg map[string]string) ([]chronos.EntityState, error) {
    tenant, ok := cfg["tenant_id"]
    if !ok {
        return nil, fmt.Errorf("myadapter: tenant_id required")
    }
    tenantID, err := uuid.Parse(tenant)
    if err != nil {
        return nil, fmt.Errorf("myadapter: invalid tenant_id: %w", err)
    }

    // Read from your external system, mapping each row into an EntityState.
    var states []chronos.EntityState
    for _, row := range rows {
        states = append(states, chronos.EntityState{
            ID:        uuid.New(),
            EntityID:  row.EntityID,
            ScopeID:   tenantID,
            Timestamp: row.ObservedAt,
            Features:  []float64{row.f1, row.f2, row.f3, row.outcome},
            Labels:    []string{"f1", "f2", "f3", "outcome"},
            Meta:      map[string]string{"source": "external-system"},
        })
    }
    return states, nil
}

// Optional: implement chronos.Closer if your source owns external resources.
func (s *Source) Close() error { /* close db / http client */ return nil }

func init() { chronos.Register(&Source{}) }
```

## Wiring the adapter into the CLI binary

Adapters self-register via `init()`, but `init()` only runs when the package is imported. The bundled CLI imports its adapters in `cmd/chronos/main.go` with blank imports:

```go
import (
    _ "github.com/felixgeelhaar/chronos/adapters/ascend"
    _ "github.com/felixgeelhaar/chronos/adapters/myadapter"
)
```

Out-of-tree adapters (in your own repo) follow the same pattern in your own `main.go`. You can build a custom binary that imports both Chronos's CLI subcommands and your adapter packages.

## Using cfg

`Fetch` accepts a `cfg map[string]string`. The CLI populates `cfg["scope_id"]` (and `cfg["coach_id"]` as the legacy alias) from the `--scope-id` / `--coach-id` flag. If your adapter needs more, pull it from environment variables inside `Fetch`:

```go
func (s *Source) Fetch(ctx context.Context, cfg map[string]string) ([]chronos.EntityState, error) {
    window := os.Getenv("MYADAPTER_LOOKBACK")
    // …
}
```

Document the variables your adapter reads. Avoid putting secrets in `cfg`; environment variables are the right place.

## Testing

A typical adapter test fakes the upstream system and asserts on the produced `EntityState` slice:

```go
func TestSource_Fetch(t *testing.T) {
    fake := newFakeUpstream(t, fixtures)
    src := NewSourceFromFake(fake)
    states, err := src.Fetch(context.Background(), map[string]string{"tenant_id": tenantID.String()})
    // assert shape, ordering, feature ordering, scope/entity assignment
}
```

The `adapters/ascend/` package is a working example backed by PostgreSQL.

## Pitfalls

- **Forgetting to register.** Without `chronos.Register(...)` in `init()` the CLI cannot find your adapter by name.
- **Different feature lengths per row.** The engine assumes a stable feature dimensionality per scope. If your data sometimes has six features and sometimes seven, normalise to a canonical set before producing states.
- **Outcome semantics drift.** If you change which feature is the outcome (or its direction), historical insights may misrepresent the new convention. Keep the contract stable for a given adapter.
- **Including domain prose in `Meta`.** The engine does not render that string anywhere. Keep `Meta` small and structured; let the API layer handle copy.
