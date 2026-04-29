## Summary

<!-- One or two sentences on what this PR changes. Describe the change,
not the work — readers know it took effort. -->

## Why

<!-- Optional but encouraged: the motivation. Link issues, design docs,
or vision references where relevant. -->

## Scope

- [ ] Detector logic (`internal/detect/`)
- [ ] Persistence (`internal/store/`, schema, sqlc)
- [ ] Public surface (`chronos.go`, `client/`)
- [ ] HTTP API (`internal/api/`)
- [ ] CLI (`cmd/chronos/`)
- [ ] Docs / README / cognitive-stack
- [ ] CI / release / Docker

## Notes for the reviewer

<!-- Anything non-obvious: trade-offs taken, follow-ups deferred,
new config knobs introduced, migration concerns. -->

## Checklist

- [ ] `make check` passes locally (`fmt + vet + test -race`)
- [ ] No prose in domain or wire types — Title/Summary/Suggestion belong to Nous
- [ ] If a new detector: registered in `detect.DefaultDetectors`, has tests, has config knobs documented in `docs/configuration.md`
- [ ] If a schema change: migration file updated, sqlc regenerated where applicable
- [ ] If a new public type: stable JSON shape, decoupled from internal/domain
