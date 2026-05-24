# Integration harness (Refs #31, mnemos #46)

Cross-service smoke tests for the chronos ↔ mnemos cognitive stack.
The suite is build-tag-gated so a plain `go test ./...` never tries
to hit a live service.

## Run it

```sh
docker compose -f test/integration/docker-compose.yml up -d
go test -tags=integration ./test/integration/...
docker compose -f test/integration/docker-compose.yml down -v
```

## Repointing endpoints

The smoke tests honour two env vars when set, so CI can repoint at a
hosted instance without code changes:

- `CHRONOS_INTEGRATION_URL` (default `http://127.0.0.1:7778`)
- `MNEMOS_INTEGRATION_URL`  (default `http://127.0.0.1:7777`)

## Image overrides

Compose pulls the released images from GHCR by default. Override
when iterating on a local build:

```sh
CHRONOS_IMAGE=ghcr.io/felixgeelhaar/chronos:dev \
MNEMOS_IMAGE=ghcr.io/felixgeelhaar/mnemos:dev \
  docker compose -f test/integration/docker-compose.yml up -d
```

## What the tests cover

- `TestHealthBothServices` — `/health` round-trips on both services
  before any cross-talk fires. Failing here means the stack isn't up.
- `TestCrossTalk_IngestThenList` — chronos `POST /v1/ingest` →
  `GET /v1/signals?scope_id=…`. Asserts the HTTP contracts work;
  detection is async and may legitimately return zero rows.
- `TestCrossTalk_MnemosClaimRoundTrip` — mnemos `POST /v1/events` →
  `GET /v1/events?run_id=…`. Pins the persistence + tenant-filter
  contract end-to-end.

A failure here on a PR that touches either repo is the
version-skew warning signal the joint-harness issue called for.
