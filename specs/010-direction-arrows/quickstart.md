# Quickstart: Merge Direction Arrows

**Feature**: `010-direction-arrows` | **Date**: 2026-08-20

Validation guide. Rules: [contracts/direction-arrows.md](./contracts/direction-arrows.md). Entities: [data-model.md](./data-model.md). Direction semantics: [008](../../008-diff-direction-toggle/contracts/diff-direction.md) + [009](../../009-bidirectional-sync/contracts/bidirectional-sync.md).

## Prerequisites

- Go 1.26+ installed
- A registered project whose tree has at least one **root with children** and one **childless root** (or add a spare root in config)
- At least one parent/child pair with inbound ≠ outbound file-change counts
- A terminal that actually sends Control+Up / Control+Down (xterm-style CSI). If the chords do nothing, the glyphs should still show inbound `↓`.
- Terminal ≥ 80×24

## Setup

```bash
go build -o branchy ./cmd/branchy
go test ./internal/tui/... -count=1
```

Automated tests cover glyphs + padding, set-not-toggle chords, `d` no-op, footer/picker copy, and confirm ignore. Scenarios below are the manual visual check.

---

## Scenario 1: Arrows on the main tree

```bash
./branchy
```

Open the project tree (default inbound).

**Expected**:

- Participating rows (every child, plus roots that have children) show `↓` immediately before the name
- Childless roots show **no** arrow; their names line up with `↓ main`-style names (blank pad)
- File-change counts still appear after the name as today
- Selected row still uses `▸ `, which is not the same mark as `↓`
- Footer: `counts: inbound (parent→child)` and `ctrl+↑: outbound` / `ctrl+↓: inbound`
- Footer does **not** mention `d` as a direction key

---

## Scenario 2: Control+Up / Control+Down set direction

1. From inbound, press Control+Up
2. Press Control+Up again
3. Press Control+Down
4. Press plain Up / Down (and `j` / `k` if you use them)
5. Press `d`

**Expected**:

- First Control+Up: all participating glyphs become `↑`; badges switch to outbound; footer mode is outbound
- Second Control+Up: still outbound (not a toggle back)
- Control+Down: glyphs `↓`, inbound badges and footer
- Plain arrows / `j`/`k` only move the cursor
- `d` does nothing to direction, glyphs, or footer mode

---

## Scenario 3: Session scope still holds

1. Control+Up so the tree is outbound (`↑`)
2. Press `s`, cancel or finish, return to the tree
3. Press `esc` to projects, open the same or another project

**Expected**:

- Tree still outbound with `↑` after sync return and after project switch
- Quit (`q`) and relaunch: `↓` inbound again

---

## Scenario 4: Sync still follows the tree

1. Control+Up, select a ceiling with descendants, press `s`
2. Inspect the first confirm (decline or cancel after checking)
3. Return, Control+Down, press `s` again on the same ceiling

**Expected**:

- Outbound tree → first confirm is child → parent (upward), outbound count, no direction chords in confirm help
- Inbound tree → first confirm is parent → child (downward)
- Confirm screens have no `↓`/`↑` tree prefixes and no `ctrl+↑` / `d: show`

---

## Scenario 5: Standalone picker chords, no arrows

```bash
./branchy sync
```

**Expected**:

- Picker starts downward; help has `sync: downward (parent→child)` and `ctrl+↑: outbound` / `ctrl+↓: inbound`
- List rows are names + badges only — no `↓` / `↑` prefix
- Control+Up: badges become outbound; help becomes `sync: upward …`; Control+Up again stays upward
- `d` does not change picker direction
- Choose a root: confirms are child → parent; help has no direction chords
- Next `./branchy sync` starts downward again

---

## Done when

- [ ] Scenarios 1–5 pass
- [ ] `go test ./internal/tui/...` is green
- [ ] README documents Control+Up / Control+Down and tree `↓`/`↑`, not `d`
