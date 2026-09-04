# Quickstart: Unify Domain Types

**Feature**: `015-unify-domain-types` | **Date**: 2026-08-24

Validation guide. Contract: [contracts/unify-domain-types.md](./contracts/unify-domain-types.md). Entities: [data-model.md](./data-model.md).

## Prerequisites

- Specs 011–014 merged
- Go 1.26+
- Registered test project (temp HOME fixture from earlier quickstarts)

## Setup

```bash
go test ./internal/mr/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1
go build -o branchy ./cmd/branchy
```

---

## Scenario 1: One direction type

```bash
go test ./internal/tui/... ./internal/sync/... -count=1 -run 'Direction|Inbound|Outbound|SyncFollowsTree|Picker'
```

**Expected**:

- Main-tree Control+Up / Control+Down still show outbound/inbound help and arrows (SC-004)
- Starting sync from the tree uses the same `sync.Direction` value the tree holds — no `syncDirection` (SC-001)
- Standalone picker Control+Up / Control+Down still set `sync.Upward` / `sync.Downward`
- Scripted `branchy sync` still walks parent → child only

---

## Scenario 2: Shared outcome classification

Create, skip-already-open, skip-declined, and fail one edge each (unit fixtures are enough):

```bash
go test ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1 -run 'OpenableURLs|CreatedURLs|renderResults|Action'
```

**Expected**:

- Printers treat `mr.ActionCreated` / `Skipped` / `Failed` the same on scripted and interactive summaries (SC-002)
- User-declined is skipped in the summary and **not** in `OpenableURLs`, even if a URL is present on the fixture (SC-005)
- Already-open skip with URL **is** openable

---

## Scenario 3: Create error policy

```bash
go test ./internal/mr/... -count=1 -run 'ValidateBranches|Create'
```

Unknown / same / missing branch: `Create` returns an error, `result == nil`, no skipped/failed payload (US3 AS1).

Scripted `mr` (manual or existing CLI test): already-open → skipped, exit 0; unrecovered failure → non-zero with the failed message; validation error → non-zero before GitLab (SC-003).

---

## Scenario 4: No parallel types left

```bash
rg 'type diffDirection' internal/tui
rg 'syncDirection' internal/
rg 'diffInbound|diffOutbound' internal/
rg 'r\.Message == "skipped by user"' internal/
rg 'case "created"|case "skipped"|case "failed"' internal/cli internal/tui
```

**Expected**: no matches (contract grep audit).

---

## Scenario 5: Outcome freeze

```bash
go test ./... -count=1
```

**Expected**: existing direction, sync, MR, badge, and prompt tests still pass (SC-004). No intended copy or key changes.

---

## Done when

- [ ] Scenarios 1–5 pass
- [ ] `go test ./... -count=1` green
- [ ] Contract grep audit clean
- [ ] No intended visual, prompt, or exit-code changes
