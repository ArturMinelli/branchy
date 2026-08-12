# Quickstart: Diff Direction Toggle

**Feature**: `008-diff-direction-toggle` | **Date**: 2026-08-12

Validation guide for main-tree inbound/outbound toggle. Rules: [contracts/diff-direction.md](./contracts/diff-direction.md). Entities: [data-model.md](./data-model.md). Inbound baseline: [006 inbound-count](../../006-parity-diff-display/contracts/inbound-count.md).

## Prerequisites

- Go 1.26+ installed
- A registered branchy project whose git repo has local (or fetched) branches matching the tree
- At least one parent→child pair where **inbound ≠ outbound** (diverged: parent-only files and child-only files)
- Preferably three children with different outbound sizes
- Terminal ≥ 80×24

## Setup

```bash
go build -o branchy ./cmd/branchy
go test ./internal/git/... ./internal/tui/... -count=1
```

Automated tests cover outbound three-dot semantics, symmetry with inbound, toggle badges, footer mode cue, and session direction. Scenarios below are the manual visual check.

---

## Scenario 1: Default inbound, then flip

```bash
./branchy
```

Open the project tree.

**Expected**:

- Non-root behind children show inbound muted numbers (same as today’s 006 behavior)
- Footer indicates inbound (`counts: inbound (parent→child)`) and lists `d: show outbound`
- Press `d` once: badges switch to outbound numbers; a child that was inbound-only quiet may gain a number (or vice versa)
- Footer now indicates outbound and `d: show inbound`
- Press `d` again: inbound badges and footer restore

Spot-check one pair on GitLab: outbound N matches files changed on an MR **child → parent**.

---

## Scenario 2: Zero hide is per direction

Find a child with inbound > 0 and outbound == 0 (or the reverse).

**Expected**:

- In the direction where the count is zero, the row shows **no** badge
- In the other direction, the non-zero badge appears
- Roots never show a badge in either mode

---

## Scenario 3: Session scope

1. Press `d` so the tree is outbound
2. Press `s`, cancel or finish sync, return to the tree
3. Press `esc` to projects, open the same or another project

**Expected**:

- Tree is still outbound after sync return and after project switch
- Quit (`q`) and relaunch: tree opens inbound again

---

## Scenario 4: Sync stays inbound

1. Flip the tree to outbound (note a child whose outbound ≠ inbound)
2. Press `s` and inspect the sync root picker
3. Confirm an edge

**Expected**:

- Picker badges match **inbound** (hide zero), not the outbound tree numbers
- Confirm line is still `{N} files would change on {child}` with inbound N (including known zero)
- No direction toggle / outbound footer on sync screens

---

## Scenario 5: Unknown still unknown

With a tree child missing locally (and not on the default remote after fetch), open the tree and press `d`.

**Expected**:

- That row shows `?` in both inbound and outbound modes
- Other rows still toggle between real numbers
- Navigation and `d` remain usable

---

## Done when

- [ ] Scenarios 1–5 pass
- [ ] `go test ./internal/git/... ./internal/tui/...` is green
- [ ] README documents `d` for count direction on the main tree
