# Contract: Merge Direction Arrows

**Feature**: `010-direction-arrows` | **Version**: 1.0 (draft)

## Overview

Supersedes the **key binding** and **mode-cue-only** clauses of [008 diff-direction](../../008-diff-direction-toggle/contracts/diff-direction.md) and the **picker `d` toggle** clauses of [009 bidirectional-sync](../../009-bidirectional-sync/contracts/bidirectional-sync.md). File-change comparison, session ownership of direction, and sync walk/ends are unchanged.

---

## Direction keys

### States

| State | Tree counts | Tree glyph | Embedded sync | Default |
|-------|-------------|------------|---------------|---------|
| `inbound` | inbound map | `↓` | Downward (parent → child) | yes (session start) |
| `outbound` | outbound map | `↑` | Upward (child → parent) | no |

### Bindings

| Chord | Bubble Tea | Main tree | Standalone picker | Other screens |
|-------|------------|-----------|-------------------|---------------|
| Control+Up | `ctrl+up` / `KeyCtrlUp` | Set outbound | Set upward | Ignore (do not change direction) |
| Control+Down | `ctrl+down` / `KeyCtrlDown` | Set inbound | Set downward | Ignore |
| `d` | *(unbound)* | No direction change | No direction change | No direction change |
| Up / Down / `k` / `j` | `up` / `down` | Move cursor only | Move list cursor only | Unchanged |

Set means assign the target state. If it is already active, remain in that state (not a flip).

On the main tree and picker, Control+Up / Control+Down MUST be handled before navigation so the cursor does not also move.

After a sync root is chosen (standalone) or sync is started from the main tree, confirm / summary / browser steps MUST NOT change direction in response to these chords or `d`.

### Scope

| In scope | Out of scope |
|----------|--------------|
| Main tree screen | Project list, link, unlink, manual MR, init |
| Standalone sync **picker** (keys only) | Scripted / flagged `branchy sync` |
| Session (root `Model`) | Disk / project config / CLI flags |

---

## Tree glyph contract

### Participating rows

A row participates if it has a tree parent **or** at least one child. Childless roots do not participate.

### Prefix column

Every main-tree row has a two-cell column immediately after the connector and before the branch name:

| Row | Inbound | Outbound |
|-----|---------|----------|
| Participating | `↓ ` | `↑ ` |
| Not participating | `  ` | `  ` |

- Style: same muted connector color as `├──` / `│`.
- Selection marker `▸ ` remains a separate leading column; it MUST NOT be reused as the direction glyph.
- File-change badges stay after the name; zero / unknown / missing counts MUST NOT hide the glyph column.
- Standalone picker, confirms, and other lists MUST NOT use this prefix.

### Footer (main tree)

MUST include, in both modes:

1. Both set-chords: `ctrl+↑: outbound` and `ctrl+↓: inbound`
2. Current mode: `counts: inbound (parent→child)` or `counts: outbound (child→parent)`
3. Navigation still named as `↑/↓` (plain arrows)

MUST NOT include `d` as a direction control.

Exact layout may stay two muted lines; both pieces MUST fit a normal 80×24 tree view.

### Picker help (standalone)

MUST include current sync direction (`sync: downward (parent→child)` or `sync: upward (child→parent)`) and the same `ctrl+↑: outbound` / `ctrl+↓: inbound` set-chords.

MUST NOT include tree glyphs on list rows or `d: show upward` / `d: show downward`.

Confirm / summary / browser help MUST NOT advertise direction keys.

---

## Unchanged contracts (008 / 009)

- `InboundFiles` / `OutboundFiles` semantics and dual eager cache
- Session direction survives project switch and return from flows; new process starts inbound
- Embedded `s` uses tree direction at the moment sync starts
- Confirm copy, counts, walks, skip-on-matching-open-MR
- Scripted sync: no direction flag, downward only

---

## 008 / 009 amendments

| Prior clause | 010 |
|--------------|-----|
| 008: `d` flips inbound ↔ outbound | Control+Up sets outbound; Control+Down sets inbound; `d` unbound |
| 008: footer names `d: show …` (the other mode) | Footer names both set-chords every time |
| 008: no directional glyph on the badge / row | Two-cell `↓`/`↑`/pad column before the name on the main tree |
| 009: standalone picker `d` flips direction | Same set-chords as the tree; still no confirm toggle |
| 009: picker help `d: show upward` / `d: show downward` | `ctrl+↑: outbound` / `ctrl+↓: inbound` |

---

## Error / degradation

- Unknown or zero file-change counts still show the glyph column on participating rows
- If the terminal does not deliver Control+Up / Control+Down, direction stays as last set; glyphs still reflect that state
- Failed remote refresh MUST NOT reset direction (008)
