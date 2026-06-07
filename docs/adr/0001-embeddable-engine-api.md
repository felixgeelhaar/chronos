# ADR 0001: Expose an embeddable Engine API

- **Status:** Accepted
- **Date:** 2026-05-31
- **Deciders:** Felix Geelhaar
- **Scope:** Chronos public Go package surface.

## Context

Chronos today is shaped as a CLI + HTTP service. The public Go package
(`chronos.go`) intentionally exposes only the adapter contract:
`EntityState`, `Source` interface, and the process-wide registry. All
engine logic — pattern detection, signal generation, persistence — lives
under `internal/` and is reachable only via the HTTP API or the
`cmd/chronos compute` CLI subcommand.

That shape worked when Chronos was always deployed as its own process.
It does not work for the new architecture: as part of the cognitive-stack
simplification (Mnemos ADR 0003-0006), Mnemos is becoming an embeddable
Go library that should bundle Chronos by default so consumers get
temporal memory out of the box. With the current public API, the only
way for Mnemos to talk to Chronos in-process is either to spin up an
in-process HTTP server (operationally ugly, two TCP listeners, JSON
serialization for every call) or to reach into `internal/` (forbidden
by the Go module system).

Neither is acceptable. We need a public, in-process Go API.

## Decision

We will promote Chronos's engine wiring into a small public `Engine`
type that lets Go consumers construct, drive, and query the engine
in-process without going through HTTP.

The public surface added in this ADR:

- `chronos.Engine` — the embeddable engine handle.
- `chronos.New(opts ...Option) (*Engine, error)` — constructor.
- `Engine.Process(ctx, EntityState) error` — ingest one observation.
- `Engine.ProcessBatch(ctx, []EntityState) error` — ingest a batch.
- `Engine.Detect(ctx, scopeIDs []uuid.UUID) ([]Signal, error)` — run
  detection synchronously.
- `Engine.Query(ctx, QueryOpts) ([]Signal, error)` — fetch stored
  signals filtered by scope / pattern / window / confidence.
- `Engine.Close() error` — release storage handle and detector
  resources.
- Option builders: `WithStorage(dsn)`, `WithDetectorConfig(cfg)`,
  `WithLogger(*slog.Logger)`.
- Public type aliases: `Signal`, `PatternType`, `TimeWindow`, `Evidence`
  re-exported from `internal/domain` in a new `chronos/signal.go`.

The new types and methods are thin wrappers around existing internal
packages (`internal/detect`, `internal/store`, `internal/pipeline`,
`internal/domain`). No detector or storage logic moves; only the
construction surface is publicised.

The existing `cmd/chronos serve` and `cmd/chronos compute` continue to
work unchanged but become consumers of the new public surface (refactor
in a follow-up commit).

## Consequences

**Positive:**

- Mnemos can bundle Chronos with `mnemos.New()` and offer
  `RememberEvent` / `Timeline` powered by Chronos without HTTP overhead.
- Other Go agent runtimes and operational systems can embed Chronos
  for incident timelines, anomaly detection, and audit trails without
  running a separate service.
- The public API stabilises the surface against the
  `cognitive-stack/library-first` direction.

**Negative / risks:**

- The public API becomes a versioned contract. Future detector or
  storage changes need to preserve `Engine` semantics. Mitigation: pin
  the option/type surface to value-typed inputs (no `*config.Config`
  leaks); document the `internal/*` packages as unstable.
- Process-wide adapter registry (`chronos.Register`) is unchanged. If a
  consumer imports both the library and the `cmd/chronos` binary in the
  same process (unusual but possible during embedded-test setups), the
  registry double-registration would panic today. We make `Register`
  idempotent (warn-on-duplicate, last-write-wins) as part of this ADR.
- `Engine` does not support multiple writers concurrently against the
  same scope (the underlying detect engine can produce duplicate
  signals). We document a single-writer expectation and rely on the
  store's transactional guarantees.

## Public API contract

The new package surface is covered by semantic versioning starting with
the release that ships this ADR. The contract:

1. `Engine.New`, `Engine.Process`, `Engine.ProcessBatch`,
   `Engine.Detect`, `Engine.Query`, `Engine.Close` and their parameter
   shapes are stable.
2. `Option` and the `With*` builders may grow but won't be removed
   without a major bump.
3. `Signal`, `PatternType`, `TimeWindow`, `Evidence` re-exports follow
   the internal source of truth; field additions are backward compatible,
   field removals are major bumps.
4. `internal/*` is **not** part of the public API; external callers must
   not import it.

## Implementation plan

Tracked in the parallel work plan:

1. Add `chronos/signal.go` re-exporting public domain types.
2. Add `chronos/engine.go` with the embeddable `Engine` type.
3. Extend `chronos/chronos.go` (or new `chronos/options.go`) with
   option builders.
4. Make `chronos.Register` idempotent.
5. Add an `engine_test.go` covering lifecycle (Process → Query) end-to-
   end against the memory store backend.
6. Refactor `cmd/chronos compute` to use the public surface as a
   dogfooding test.
7. Tag `v0.6.0` and release.

## Alternatives Considered

**1. Keep the public API minimal; require HTTP for all engine
interaction.** Rejected. Forces every Go consumer to run an in-process
HTTP server. Operationally ugly and adds latency / serialisation cost
for what should be a Go function call.

**2. Expose `internal/*` to selected consumers (Mnemos) via a Go
workspace replace directive.** Rejected. Brittle, doesn't survive
external consumers, and signals that the surface is unstable when in
practice we want consumers to depend on a versioned API.

**3. Split Chronos into a library module (`chronos/lib`) and a service
module.** Rejected. Adds module-management overhead. The single-module
shape with a clean public API is simpler.

## Related Work

- [Mnemos ADR 0003-0006 (cognitive-stack simplification)](https://github.com/klarlabs-studio/mnemos/blob/main/docs/adr/)
- Mnemos plan: `~/.claude/plans/agent-ready-i-would-like-snazzy-kay.md`
