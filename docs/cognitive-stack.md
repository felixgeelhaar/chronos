# Chronos in the Cognitive Stack

Chronos is one of four loosely-coupled systems that, together, form a cognitive stack:

| Layer    | Role                              | Question it answers                |
|----------|-----------------------------------|------------------------------------|
| Mnemos   | Memory & Knowledge                | *What happened? What do we know?*  |
| **Chronos** | **Time & Pattern Perception**  | ***What is changing? What patterns are emerging? What is unusual?*** |
| Praxis   | Execution / Capabilities          | *What can be done? What happened when we did it?* |
| Nous     | Coordination / Intelligence       | *What matters? What should we do?* |

The flow is one direction at a time, with feedback:

```
observe  →  understand  →  detect  →  decide  →  act  →  learn
  ↑                                                        │
  └────────────────────  outcome  ─────────────────────────┘

[Input streams]
    ↓
Mnemos    (events → memories, knowledge graph)
    ↓
Chronos   (time series → signals, patterns, anomalies)
    ↓
Nous      (memory + signals → goals, plans, decisions)
    ↓
Praxis    (decisions → actions, results)
    ↓
[Outcomes flow back to Mnemos]
```

## What Chronos *is*

Chronos perceives change. It accepts time-series observations from any source and emits structured **Signals** describing the patterns it sees:

- `Recurrence` — the subject is in a state others have been in before (cosine peers).
- `Trend` — the metric is sustainedly moving in a direction (linear regression slope + R²).
- `Spike` / `Drop` — sharp deviations from a recent baseline (z-score).
- `Stall` — the metric is no longer changing meaningfully (normalised stddev).
- `Anomaly` — the subject is unlike its peers' *current* states (cross-entity dual of Recurrence).
- `Seasonality` — periodic structure in the series (autocorrelation peak).
- `Correlation` — two series move together (pairwise Pearson).

Each signal carries:

- **Pattern** (PatternType enum)
- **Series** (the entity the pattern was detected in)
- **Window** (the time interval analysed; Anomaly is a snapshot, so its window is degenerate — `Start == End == subject.Timestamp`)
- **Strength** (intensity of the pattern, 0..1)
- **Confidence** (how sure the detector is, 0..1)
- **Evidence** (per-detector supporting observations)
- **Metrics** (free-form numeric measurements: `avg_similarity`, `slope`, `z`, …)

The full list of stable string keys consumers may rely on — `Pattern` values, `Evidence.Kind`s, and `Metrics` keys per detector — is in [`wire-contract.md`](wire-contract.md).

A signal **does not interpret itself.** "This is a recurrence with strength 0.92, confidence 0.85" is the engine's full statement. Whether that recurrence is good news, bad news, or noise is Nous's call.

## What Chronos is *not*

Chronos deliberately does not:

- **Decide what to do.** No suggestion strings, no priority, no urgency. Those are decisions; Nous owns decisions.
- **Render prose.** No `Title`, no `Summary`, no `Suggestion`. Presentation is a downstream concern that varies by surface (UI, Slack, voice, email) and audience.
- **Track dismissal or feedback.** Once a signal is detected and persisted it is immutable. Whether a signal is acted upon, suppressed, or used to update a belief lives in Nous (decisions) and Mnemos (knowledge), respectively.
- **Store long-term knowledge.** Chronos retains observations only long enough to detect patterns; durable facts ("athlete X improved their technique") belong in Mnemos with provenance.
- **Execute actions.** Praxis owns capabilities and execution. Chronos cannot send a notification or call an API.

These are not limitations to be lifted later — they are the boundary that makes the stack composable.

## How the layers exchange data

| From → To       | Payload                                                                 |
|-----------------|-------------------------------------------------------------------------|
| World → Mnemos  | Raw events                                                              |
| Mnemos → Nous   | Memories, entities, historical context                                  |
| World → Chronos | Time-series observations (`chronos.EntityState`)                        |
| Chronos → Nous  | Signals (this document's subject)                                       |
| Nous → Praxis   | Action requests with capability + payload                               |
| Praxis → Mnemos | Execution results, outcomes, failures                                   |
| Mnemos ↔ Nous   | Bidirectional: Nous queries facts, Nous's decisions become new memories |

## Independence guarantee

Each layer must work standalone. Chronos used alone is a generic time-series pattern-detection engine — adapters in, signals out, no knowledge of the rest of the stack. Tests run on `:memory:` SQLite with no external dependencies.

The stack composes not because Chronos *knows* about Nous and Mnemos, but because their contracts (Signal, Memory, Action) are stable and decoupled.

An end-to-end runnable version of the example below — with real curl payloads, the actual JSON Nous receives, and Go SDK code on the consumer side — lives in [`cognitive-stack-example.md`](cognitive-stack-example.md).

## Worked example

> A user says: *"I'll follow up with Alex tomorrow."*

1. **Mnemos** persists the raw event, extracts a Memory `{kind:"commitment", content:"follow up with Alex", deadline:"tomorrow"}`.
2. **Chronos** sees, over time, that the commitment's "days since opened" series is climbing without resolution. It emits a `Stall` signal on that series.
3. **Nous** combines the Memory (a commitment) with the Stall signal (no progress) and decides this is a commitment risk worth surfacing. It plans an intervention: draft a follow-up message.
4. **Praxis** executes the action: drafts the message, surfaces it to the user.
5. The outcome — sent or dismissed — flows back into **Mnemos**, updating the commitment's status and feeding future patterns.

Chronos's contribution is exactly one structured signal. It does not say *"this commitment is at risk"* — that interpretation requires combining the signal with the memory, which is Nous's job.

## Where to look in the codebase

- `chronos.EntityState`, `chronos.Source` — adapter ingest contract.
- `internal/domain.Signal`, `domain.Evidence`, `domain.PatternType` — the perception output.
- `internal/detect/` — detectors (one file per pattern) and the engine that fans out across them.
- `internal/ports/SignalRepository` — query surface (`List`, `Get`, `Count`) for Nous integrators.
- `client/` — the public Go SDK consumers (Nous, dashboards, runtimes) use to read signals.

The `client/` package is the canonical wire shape Nous integrators consume; it is decoupled from internal types so the engine can evolve without breaking the stack.
