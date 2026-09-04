# Quickstart: Manual Remote Count Reload

**Feature**: `016-tui-reload-counts` | **Date**: 2026-09-04

Validation guide for the `r` reload key. Contracts: [tui-reload-counts.md](./contracts/tui-reload-counts.md). Fetch/refresh base: [007 lazy-fetch](../../007-lazy-auto-fetch/contracts/lazy-fetch.md).

## Prerequisites

- Go 1.26+
- A registered branchy project with default remote (`origin`) reachable
- A tree where at least one child has stale or `?` counts relative to remote (e.g. teammate pushed; local refs behind)
- Terminal ≥ 80×24

## Setup

```bash
go build -o branchy ./cmd/branchy
go test ./internal/tui/... -count=1
```

---

## Scenario 1: Reload stale counts on the main tree

```bash
./branchy
```

Open a project with known stale counts (numbers wrong vs GitLab MR file counts).

1. Note current badges on child rows.
2. Press `r`.
3. Wait a few seconds (no spinner expected).

**Expected**:

- Tree footer lists `r: reload`
- Keyboard remains responsive immediately after `r`
- After remote update completes, badges update to match GitLab file-change counts
- Active direction (inbound/outbound) unchanged; toggle still works with fresh data
- No fetch error banner

---

## Scenario 2: Reload during sync picker

From the tree, press `s` to open sync (embedded).

1. Note inbound badges on the root picker.
2. Press `r`.
3. Wait for update.

**Expected**:

- Picker help lists `r: reload`
- Badges refresh in place without leaving sync
- Direction chords still work; enter still selects root

---

## Scenario 3: Reload on edge confirm

Walk sync to an edge confirm panel.

1. Note the “files would change” line.
2. Press `r`.
3. Wait for update.

**Expected**:

- Confirm line updates if remote refs changed
- `y` / `n` still work; `r` does not confirm or skip
- No extra help clutter on confirm footer

---

## Scenario 4: Offline reload is silent

Disconnect network or use unreachable `origin`.

1. Open tree with existing counts.
2. Press `r` several times.

**Expected**:

- Counts unchanged
- No error message
- UI remains usable (`j`/`k`, `s`, `q`)

---

## Scenario 5: Coalesce with auto-fetch

Open the tree and immediately press `r` (while startup background fetch may still run).

**Expected**:

- No hang; no parallel-fetch symptoms
- Counts refresh when the shared fetch completes (may refresh once or twice — harmless)

---

## Scenario 6: Scripted CLI unchanged

```bash
./branchy sync --from <root> -y
```

**Expected**:

- No `r` binding in plain CLI
- Behavior identical to pre-feature
