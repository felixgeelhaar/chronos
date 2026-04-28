# Worked Example: Mnemos → Chronos → Nous

The [cognitive-stack overview](cognitive-stack.md) describes the layers in the abstract: *Mnemos remembers, Chronos perceives change, Nous decides, Praxis acts*. This document walks one scenario end-to-end with real wire payloads so adopters can see exactly how the four systems compose.

The scenario is the same one summarised in [`cognitive-stack.md`](cognitive-stack.md): a user says "I'll follow up with Alex tomorrow", and we want Nous to surface a nudge if the commitment goes stale.

```
[user]      "I'll follow up with Alex tomorrow"
   │
   ▼
[Mnemos]    persists the event; extracts a Memory: kind=commitment,
            content="follow up with Alex", deadline="tomorrow"
   │
   ▼
[Chronos]   observes the series "days since the commitment was opened",
            sees it climb without resolution, emits a Stall signal
   │
   ▼
[Nous]      reads (memory + signal), decides this is a commitment risk,
            drafts a follow-up message
   │
   ▼
[Praxis]    sends or surfaces the message; outcome flows back to Mnemos
```

The remainder of this doc focuses on the **Chronos slice** — what it sees, what it emits, and how Nous reads it. The Mnemos and Praxis sides are sketched briefly so you can see where Chronos sits in the data flow without leaving the stack abstract.

## 0. Setup

Run a Chronos server with the in-process detection scheduler enabled (so the live `/v1/signals/stream` SSE endpoint has something to push):

```bash
export CHRONOS_DB_DSN="sqlite:///tmp/walkthrough.db"
export CHRONOS_DETECTION_INTERVAL=10s
chronos serve --port 7778 &
```

Throughout the example we use these IDs (replace with your own when running):

```bash
SCOPE_ID="11111111-1111-1111-1111-111111111111"   # tenant / coach / org
COMMIT_ID="22222222-2222-2222-2222-222222222222"  # the commitment entity
```

`SCOPE_ID` is the cognitive-stack equivalent of a tenant — it isolates the perception window. `COMMIT_ID` is the entity Chronos is watching.

## 1. Mnemos persists the commitment (sketch)

When the user utters the commitment, Mnemos records the raw event and extracts a structured Memory. The shape is Mnemos's concern; from Chronos's point of view the only thing that matters is that *something downstream* will start ingesting an observation series for this entity. A representative Memory:

```json
{
  "id": "mem-c1",
  "kind": "commitment",
  "content": "follow up with Alex",
  "deadline": "2026-04-29T17:00:00Z",
  "subject_entity_id": "22222222-2222-2222-2222-222222222222",
  "scope_id": "11111111-1111-1111-1111-111111111111",
  "created_at": "2026-04-28T09:00:00Z",
  "status": "open"
}
```

Mnemos is the source of truth for *what was committed and when*. Chronos doesn't need that text — it needs a numeric series describing how that commitment is moving over time.

## 2. Chronos observes the series

A small bridge service (or a Mnemos plugin) ingests one observation per day per open commitment into Chronos. The convention from [`docs/adapters.md`](adapters.md) applies: **the last feature is the outcome metric, higher is better**. For a "days since opened" series we want lower values to be better, so the bridge inverts the metric — `outcome = max_days - days_since_opened` so a fresh commitment scores high and a stalled one scores low.

Day 0 — the commitment was just opened:

```bash
curl -s -X POST http://localhost:7778/v1/ingest \
  -H 'Content-Type: application/json' \
  -d '{
    "entity_id": "22222222-2222-2222-2222-222222222222",
    "scope_id":  "11111111-1111-1111-1111-111111111111",
    "timestamp": "2026-04-28T09:00:00Z",
    "features":  [0, 14],
    "labels":    ["days_since_opened", "outcome"],
    "adapter":   "commitments-bridge"
  }'
# => 202 Accepted, {"id":"...","status":"accepted"}
```

Day 1 — still nothing happened:

```bash
curl -s -X POST http://localhost:7778/v1/ingest \
  -H 'Content-Type: application/json' \
  -d '{
    "entity_id": "22222222-2222-2222-2222-222222222222",
    "scope_id":  "11111111-1111-1111-1111-111111111111",
    "timestamp": "2026-04-29T09:00:00Z",
    "features":  [1, 13],
    "labels":    ["days_since_opened", "outcome"],
    "adapter":   "commitments-bridge"
  }'
```

…and so on for several days. After about a week, the outcome series for this entity is `[14, 13, 12, 11, 10, 9, 8]` — gradually trending downward but mostly *flat in normalised terms* once the commitment passes the deadline.

## 3. Chronos detects the stall

The detection scheduler ticks every `CHRONOS_DETECTION_INTERVAL`. It groups observations by scope, runs every detector ([`internal/detect/`](../internal/detect)), and persists each emitted signal. With seven flat observations the **Stall** detector trips: normalised stddev is below `CHRONOS_STALL_MAX_STDDEV` (default 0.05) over at least `CHRONOS_STALL_MIN_POINTS` (default 4).

Chronos's emitted signal — exactly what Nous will see when it queries `/v1/signals`:

```json
{
  "id": "33333333-3333-3333-3333-333333333333",
  "scope_id": "11111111-1111-1111-1111-111111111111",
  "series": "22222222-2222-2222-2222-222222222222",
  "pattern": "stall",
  "detected_at": "2026-05-04T09:00:01Z",
  "window": {
    "start": "2026-04-28T09:00:00Z",
    "end":   "2026-05-04T09:00:00Z"
  },
  "strength": 0.91,
  "confidence": 0.83,
  "metrics": {
    "normalised_stddev": 0.0072,
    "mean": 11.0,
    "n": 7
  },
  "evidence": [{
    "series": "22222222-2222-2222-2222-222222222222",
    "time":   "2026-05-04T09:00:00Z",
    "kind":   "variance_window",
    "score":  0.0072,
    "metrics": {
      "normalised_stddev": 0.0072,
      "mean": 11.0,
      "n": 7
    }
  }]
}
```

Things to notice — these are the parts of the contract Nous can rely on, all listed in [`docs/wire-contract.md`](wire-contract.md):

- **`pattern: "stall"`** — drives interpretation routing on the Nous side.
- **`strength: 0.91`** — how flat the series is (1.0 = perfectly flat).
- **`confidence: 0.83`** — strength scaled by sample-size factor; rises as more observations land.
- **`evidence[0].kind: "variance_window"`** — stable string Nous can switch on.
- **`metrics.normalised_stddev`**, **`metrics.mean`**, **`metrics.n`** — the underlying numbers Nous needs if it wants to reason about *how* stalled the commitment is.

There is no `title`, no `summary`, no `suggestion`. Chronos perceives; Nous interprets. That's the design rule.

## 4. Nous reads the signal

Nous integrates as a Chronos *consumer*. It uses the public Go SDK ([`client/`](../client)) — there is no internal coupling.

### 4a. Pull (poll for the latest)

```go
import (
    "context"
    "os"
    "time"

    "github.com/felixgeelhaar/chronos/client"
    "github.com/google/uuid"
)

func ReadStallSignals(ctx context.Context, scopeID uuid.UUID, since time.Time) error {
    c, err := client.New("http://chronos.local:7778",
        client.WithToken(os.Getenv("CHRONOS_API_TOKEN")),
        client.WithTimeout(10*time.Second),
    )
    if err != nil {
        return err
    }
    signals, err := c.Signals().
        Scope(scopeID).
        Pattern(client.PatternTypeStall).
        Since(since).
        MinConfidence(0.7).
        List(ctx)
    if err != nil {
        return err
    }
    for _, sig := range signals {
        nousInterpret(sig) // see 4c
    }
    return nil
}
```

### 4b. Push (subscribe to the live feed)

For low-latency interventions Nous can subscribe to the SSE endpoint instead of polling. Same wire shape, channel-based delivery:

```go
ctx, cancel := context.WithCancel(ctx)
defer cancel()

events, err := c.Signals().
    Scope(scopeID).
    Pattern(client.PatternTypeStall).
    Stream(ctx)
if err != nil {
    return err
}
for sig := range events {
    nousInterpret(sig)
}
// channel closes on ctx cancel, server EOF, or fatal protocol error
```

A common production pattern is to use **both**: stream for live awareness, plus a periodic `Since`-keyed `List` call as the gap-recovery path. Chronos guarantees at-most-once delivery for streams; persistence is the source of truth.

### 4c. Nous interprets and decides

This is the part Chronos deliberately **does not do**. Sketch:

```go
func nousInterpret(sig client.Signal) {
    // 1. Look up the Memory that owns this entity.
    memory, _ := mnemos.GetByEntity(sig.Series)
    if memory.Kind != "commitment" {
        return // Stall signals on non-commitment entities mean
               // something else; route accordingly.
    }

    // 2. Combine memory + signal into a decision.
    if sig.Confidence < 0.8 || memory.Status != "open" {
        return // not actionable yet
    }
    daysOpen := int(sig.Metrics["mean"]) // from the Stall metric

    // 3. Hand a structured action to Praxis.
    praxis.RequestAction(praxis.ActionRequest{
        Capability: "draft-followup-message",
        Payload: praxis.DraftPayload{
            Subject:    memory.Content,           // "follow up with Alex"
            DaysOpen:   daysOpen,                  // 11
            Confidence: sig.Confidence,            // 0.83
            SignalID:   sig.ID,                    // for audit traceability
        },
    })
}
```

Nous's decision is always a *combination*: a memory (the commitment) plus a perception (the Stall) plus an interpretation rule. None of those rules live in Chronos.

## 5. The outcome flows back to Mnemos

Once Praxis has done something — sent a draft for the user to review, posted to Slack, etc. — the outcome (sent, dismissed, edited) lands back in Mnemos. Chronos may then see the `days_since_opened` reset to zero on the next ingest tick, the next Stall signal goes silent, and the loop is closed.

```
World ──► Mnemos ──► Nous ◄── Chronos ◄── observations
                       │
                       ▼
                     Praxis ──► outcome ──► Mnemos
```

## What's deliberately *not* in this example

- **No prose in any Chronos payload.** No "consider following up with Alex" string. The signal is a numeric perception; the prose belongs to Nous-or-later.
- **No state about the *user*** in Chronos. Mnemos owns the commitment, the deadline, who it's with. Chronos sees a series of numbers tied to an entity ID and emits a perception.
- **No decision logic in Chronos.** The Stall signal would fire identically for a stalled deployment, a stalled training plan, or a stalled commitment — Nous is what makes those interpretations diverge.
- **No retries, dismissal, or feedback in Chronos.** Once a signal is detected and persisted it is immutable; any *acted on / suppressed / valid / invalid* status lives in Nous (decisions) or Mnemos (memory updates).

These are the same boundaries [`docs/cognitive-stack.md`](cognitive-stack.md) draws abstractly, made concrete with one scenario.

## Further reading

- [`docs/cognitive-stack.md`](cognitive-stack.md) — the layer roles and contracts in the abstract.
- [`docs/wire-contract.md`](wire-contract.md) — every stable string Nous can rely on (`Pattern`, `Evidence.Kind`, `Metrics` keys per detector).
- [`docs/adapters.md`](adapters.md) — write your own bridge if the commitments-bridge sketch above isn't enough.
- [`docs/configuration.md`](configuration.md) — DSN syntax, namespace contract, push-notification setup.
