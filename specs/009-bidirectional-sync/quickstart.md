# Quickstart: Bidirectional Sync

**Feature**: `009-bidirectional-sync` | **Date**: 2026-08-13

Validation guide. Rules: [contracts/bidirectional-sync.md](./contracts/bidirectional-sync.md). Entities: [data-model.md](./data-model.md). Downward baseline: [002 sync](../../002-interactive-sync-confirm/contracts/sync-cli.md).

## Prerequisites

- Go 1.26+ installed
- A registered project whose tree has **at least two levels** below some ceiling (e.g. `develop → feat-a → leaf` and a sibling)
- At least one parent/child pair with **inbound ≠ outbound** file-change counts
- `glab` authenticated (for real MR creation in manual scenarios)
- Terminal ≥ 80×24

## Setup

```bash
go build -o branchy ./cmd/branchy
go test ./internal/tree/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1
```

Automated tests cover post-order membership, `Ends`, embedded follow-tree, standalone picker toggle, and downward/CLI isolation. Scenarios below are the manual end-to-end check.

---

## Scenario 1: Upward cascade from the tree

```bash
./branchy
```

1. Open the project, press `d` so the tree is outbound
2. Select the mid-level ceiling (not a leaf)
3. Press `s`
4. Accept every edge

**Expected**:

- First confirm is a **deeper** child → parent (not the ceiling → child)
- Each question is `Create MR {child} → {parent}?`
- Count line is files that would change on the **parent** (outbound number, including known zero)
- No `d: show` on the confirm help
- Context reads `Sync up to {ceiling}`
- No edge whose parent is above the ceiling
- Summary arrows are child → parent
- After return, the tree is still outbound (008 session rule)

---

## Scenario 2: Downward still matches today

1. Relaunch or press `d` so the tree is inbound
2. Select the same ceiling, press `s`
3. Inspect the first confirm (decline or cancel after checking)

**Expected**:

- First edge is ceiling → first child (top-down)
- Question is `Create MR {parent} → {child}?`
- Count line is files that would change on the **child** (inbound)
- Context reads `Sync down from {ceiling}`

---

## Scenario 3: Counts match the MR

Pick a pair where inbound N ≠ outbound M.

- Inbound tree + `s`: confirm shows N on the child
- Outbound tree + `s`: confirm shows M on the parent

**Expected**: the two numbers differ and each matches the GitLab files-changed count for that MR direction.

---

## Scenario 4: Standalone picker toggle

```bash
./branchy sync
```

**Expected**:

- Picker starts downward; help has `sync: downward (parent→child)` and `d: show upward`
- Badges match inbound (hide zero)
- Press `d`: help flips to upward; badges match outbound
- Choose the ceiling: confirms are child → parent, deepest first, outbound counts
- Confirm help has no direction toggle
- Quit and run `./branchy sync` again: picker is downward once more

---

## Scenario 5: Scripted CLI stays downward

```bash
./branchy sync --from develop -y
```

(Answer the end browser prompt `n`.)

**Expected**:

- Plan and prompts are parent → child, top-down
- No direction flag exists (`./branchy sync --help` has `--from` and `-y` only)
- MRs created (if any) are parent → child

---

## Scenario 6: Opposite open MR does not skip

With an open parent → child MR already on a pair, run **upward** on that ceiling and reach the same pair.

**Expected**: upward still offers `child → parent` (not skipped as “already exists”), unless a child → parent MR is also open.

---

## Scenario 7: Empty ceiling

Select a leaf, outbound, press `s`.

**Expected**: `No child branches below "{leaf}".` — no confirms, no browser question. Same message inbound.

---

## Done when

- [ ] Scenarios 1–7 pass
- [ ] `go test ./internal/tree/... ./internal/sync/... ./internal/tui/... ./internal/cli/...` is green
- [ ] README documents bidirectional interactive sync and downward-only scripted sync
