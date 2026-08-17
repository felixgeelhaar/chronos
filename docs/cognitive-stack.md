# Chronos in the Cognitive Stack

Chronos is the **time / pattern perception** layer next to **Mnemos** (memory). Downstream **agent runtimes** consume signals and decide what to do. Chronos perceives; it does not interpret, decide, or act.

| Layer | Role | Question it answers |
|----------|-----------------------------------|------------------------------------|
| Mnemos   | Memory & Knowledge                | *What happened? What do we know?*  |
| **Chronos** | **Time & Pattern Perception**  | ***What is changing? What patterns are emerging? What is unusual?*** |
| Agent runtimes | Interpretation & action      | *What matters? What should we do?* |

Optional policy library: **[decisionkit](https://github.com/felixgeelhaar/decisionkit)** — deterministic risk + intervention scoring that used to live in Nous.

**Archived.** [Nous](https://github.com/felixgeelhaar/nous) (coordination / intelligence) was archived 2026-05-31 at `v0.3.1-final`. Reasoning moved into agent runtimes; risk + intervention extracted to decisionkit. [Praxis](https://github.com/felixgeelhaar/praxis) and [Olymp](https://github.com/felixgeelhaar/olymp) are also archived. See [Mnemos ADR 0005](https://github.com/felixgeelhaar/mnemos/blob/main/docs/adr/0005-archive-nous.md).

The flow is one direction at a time, with feedback:

```
observe  →  understand  →  detect  →  decide  →  act  →  learn
  ↑                                                        │
  └────────────────────  outcome  ─────────────────────────┘

[Input streams]
    ↓
Mnemos         (events → memories, knowledge graph)
    ↓
Chronos        (time series → signals, patterns, anomalies)
    ↓
Agent runtimes (memory + signals → goals, plans, actions)
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
- `ChangePoint` — a sustained mean shift between two regimes (best-split test).
- `OutlierCluster` — several series in a scope go anomalous around the same time (cohort-level).
- `CrossScopeCorrelation` — two series in *different* scopes move together.

Each signal carries:

- **Pattern** (PatternType enum)
- **Series** (the entity the pattern was detected in)
- **Window** (the time interval analysed; Anomaly is a snapshot, so its window is degenerate — `Start == End == subject.Timestamp`)
- **Strength** (intensity of the pattern, 0..1)
- **Confidence** (how sure the detector is, 0..1)
- **ConfidenceClass** (qualitative grade: tentative / established / strong)
- **Evidence** (per-detector supporting observations)
- **Metrics** (free-form numeric measurements: `avg_similarity`, `slope`, `z`, …)
- **Explanation** (numeric/structured only: feature evolution, peer count, threshold, detector version — never Title/Summary/Suggestion)

The full list of stable string keys consumers may rely on — `Pattern` values, `Evidence.Kind`s, and `Metrics` keys per detector — is in [`wire-contract.md`](wire-contract.md).

A signal **does not interpret itself.** "This is a recurrence with strength 0.92, confidence 0.85" is the engine's full statement. Whether that recurrence is good news, bad news, or noise is the consumer's call — typically an agent runtime, optionally with [decisionkit](https://github.com/felixgeelhaar/decisionkit) for risk scoring.

## What Chronos is *not*

Chronos deliberately does not:

- **Decide what to do.** No suggestion strings, no priority, no urgency. Those are decisions; agent runtimes own decisions.
- **Render prose.** No `Title`, no `Summary`, no `Suggestion`. Presentation is a downstream concern that varies by surface (UI, Slack, voice, email) and audience.
- **Track dismissal or feedback.** Once a signal is detected and persisted it is immutable. Whether a signal is acted upon, suppressed, or used to update a belief lives in the agent layer (decisions) and Mnemos (knowledge), respectively.
- **Store long-term knowledge.** Chronos retains observations only long enough to detect patterns; durable facts ("athlete X improved their technique") belong in Mnemos with provenance.
- **Execute actions.** Chronos cannot send a notification or call an API. Acting on a signal is the consumer's job.

These are not limitations to be lifted later — they are the boundary that makes the stack composable.

## How the layers exchange data

| From → To              | Payload                                                                 |
|------------------------|-------------------------------------------------------------------------|
| World → Mnemos         | Raw events                                                              |
| Mnemos → agents        | Memories, entities, historical context                                  |
| World → Chronos        | Time-series observations (`chronos.EntityState`)                        |
| Chronos → agents       | Signals (this document's subject)                                       |
| Agents → world         | Actions (tools, APIs, notifications)                                    |
| Agents → Mnemos        | Outcomes, decisions, and updated memories                               |
| Mnemos ↔ agents        | Bidirectional: agents query facts; their decisions become new memories  |

## Independence guarantee

Each layer must work standalone. Chronos used alone is a generic time-series pattern-detection engine — adapters in, signals out, no knowledge of the rest of the stack. Tests run on `:memory:` SQLite with no external dependencies.

The stack composes not because Chronos *knows* about Mnemos or any agent runtime, but because their contracts (Signal, Memory) are stable and decoupled.

An end-to-end runnable version of the example below — with real curl payloads, the actual JSON a consumer receives, and Go SDK code on the consumer side — lives in [`cognitive-stack-example.md`](cognitive-stack-example.md).

## Worked example

> A user says: *"I'll follow up with Alex tomorrow."*

1. **Mnemos** persists the raw event, extracts a Memory `{kind:"commitment", content:"follow up with Alex", deadline:"tomorrow"}`.
2. **Chronos** sees, over time, that the commitment's "days since opened" series is climbing without resolution. It emits a `Stall` signal on that series.
3. An **agent runtime** combines the Memory (a commitment) with the Stall signal (no progress) and decides this is a commitment risk worth surfacing. It may score the risk with [decisionkit](https://github.com/felixgeelhaar/decisionkit) and plan an intervention: draft a follow-up message.
4. The agent **acts** (drafts the message, surfaces it to the user).
5. The outcome — sent or dismissed — flows back into **Mnemos**, updating the commitment's status and feeding future patterns.

Chronos's contribution is exactly one structured signal. It does not say *"this commitment is at risk"* — that interpretation requires combining the signal with the memory, which is the consumer's job.

## Where to look in the codebase

- `chronos.EntityState`, `chronos.Source` — adapter ingest contract.
- `internal/domain.Signal`, `domain.Evidence`, `domain.PatternType` — the perception output.
- `internal/detect/` — detectors (one file per pattern) and the engine that fans out across them.
- `internal/ports/SignalRepository` — query surface (`List`, `Get`, `Count`) for downstream consumers.
- `client/` — the public Go SDK agents, dashboards, and runtimes use to read signals.

The `client/` package is the canonical wire shape consumers ingest; it is decoupled from internal types so the engine can evolve without breaking the stack.
