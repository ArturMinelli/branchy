# Research: Merge Direction Arrows

**Feature**: `010-direction-arrows` | **Date**: 2026-08-20

## 1. Where the glyph lives

**Decision**: `BranchTreeView` owns the direction arrow. At flatten time each `branchRow` records whether it **participates** (has a parent, or is a root with children). `renderRow` prefixes the name with a two-cell column: `↓ ` / `↑ ` / two spaces. Style the column with `connectorStyle` (muted, same as `├──` / `│`). The selection marker stays `▸ ` and is painted separately as today.

**Rationale**: Spec FR-001–004 put the cue on the main-tree rows, not in the footer. Flatten already walks parent/child structure, so participating is known without a second pass. Putting the glyph in `renderRow` keeps `View()` free of `Model`. Muted chrome was grilled so arrows do not compete with `▸`.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Replace `├──` / `└──` with arrow connectors | Grilling locked a mark *next to the name*, after the connector |
| Glyph only on the selected row | Spec requires every participating row (SC-001 / SC-002) |
| Accent or two-tone colors | Grilling: muted like connectors |
| Draw arrows in `Model.View` | Leaks tree-row policy out of `BranchTreeView` |

---

## 2. Padding vs ragged names

**Decision**: Always emit a two-cell prefix before the name. Participating rows get `↓ ` or `↑ ` (glyph + space). Childless roots get two spaces. Child rows all participate (they have a parent), so only childless roots need the blank pad.

**Rationale**: Grilling required names to stay in one column. Unicode `↓`/`↑` are one terminal cell in the fonts branchy already uses (`▸`, box-drawing). A trailing space after the glyph keeps the name from colliding with the arrow.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| No pad | `orphan` would sit left of `↓ main` |
| Pad with a fullwidth space or three cells | Over-indents; two cells match glyph+space |
| Show a dummy arrow on childless roots | Spec: no arrow when the row cannot be a source or destination |

---

## 3. Set chords, not a toggle

**Decision**: Replace `keys.Direction` (`d`) with two bindings:

| Chord | Binding string | Tree | Standalone picker |
|-------|----------------|------|-------------------|
| Control+Up | `ctrl+up` | `diffOutbound` | `sync.Upward` |
| Control+Down | `ctrl+down` | `diffInbound` | `sync.Downward` |

Assign the target value even when it is already active (no-op, not a flip). Match these bindings **before** plain `Up`/`Down` (`up`/`k`, `down`/`j`). Remove `d` from both key maps.

Bubble Tea v1.3.10 already maps CSI `\x1b[1;5A` → `KeyCtrlUp` and `\x1b[1;5B` → `KeyCtrlDown`. Tests construct `tea.KeyMsg{Type: tea.KeyCtrlUp}` (and `KeyCtrlDown`). `key.Matches` compares against the `"ctrl+up"` / `"ctrl+down"` strings.

**Rationale**: Spec FR-005–008 and grilling: keys should match tree geometry. `d` as toggle was the unintuitive part. Charm bindings distinguish `up` from `ctrl+up`, so navigation does not steal the chord.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Keep `d` as a hidden toggle | Grilling: no fallback; spec FR-008 |
| Shift+Up / Shift+Down aliases | Grilling: Control+arrows only |
| Single chord that still toggles | Spec: Control+Up must stay outbound on a second press |
| Bind `ctrl+k` / `ctrl+j` | User asked for arrows, not vim-modified aliases |

---

## 4. Terminals that never send Control+arrows

**Decision**: No fallback chord. If the terminal does not emit Control+Up/Down, direction stays at its last value (inbound on a new session). Tree arrows still show that value. Document the chords in README; do not promise they work in every tmux/SSH setup.

**Rationale**: Grilling accepted this risk rather than resurrect `d` or add Shift+arrows. Bubble Tea’s sequence table already covers xterm-style CSI and a few urxvt variants; that is the support bar.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Unlisted `d` toggle | Contradicts “remove `d` everywhere” |
| Shift+Up / Shift+Down | Extra chords to teach; grilling declined |
| Detect missing sequences and print a warning | Cannot reliably detect “user tried and the terminal ate it” |

---

## 5. Help copy lists both set actions

**Decision**: Main-tree footer and standalone picker help always name **both** chords and the **current** mode. They do not adopt 008’s “key shows the other mode” pattern (`d: show outbound`).

```text
# tree (inbound)
↑/↓: navigate  ctrl+↑: outbound  ctrl+↓: inbound  s: sync  …
counts: inbound (parent→child)

# picker (downward)
↑/↓: navigate  enter: select  ctrl+↑: outbound  ctrl+↓: inbound  …
sync: downward (parent→child)
```

Confirm, summary, and browser help MUST NOT list these chords (direction already frozen). `d` MUST NOT appear as a direction control.

**Rationale**: Spec FR-009 lists Control+Up as the way to set outbound **and** Control+Down as the way to set inbound — both, in both modes. Set-not-toggle copy would be lying if it said “show the other.”

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `ctrl+↑: show outbound` only when inbound | Implies a toggle; second press is a no-op |
| Keep `d: show …` plus the new chords | Spec removes `d` from help |

---

## 6. Picker has keys, not arrows

**Decision**: Standalone picker binds the same set-chords and rebuilds badges from the other map (same as today’s `flipPickerDirection`, but **set** instead of flip). List rows stay a flat name + count badge. Handle Control+Up/Down in `updatePicker` before `branchList.Update` so the list cannot treat them as navigation. After `prepareSyncFrom`, those bindings are not consulted (confirm path already ignores `Direction`).

**Rationale**: Spec US4 / FR-013: same mental model, no tree arrows on a flat list. 009 already froze direction at root-select; only the key identity changes.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `↓` / `↑` on picker rows | Grilling: no arrows on the list |
| Leave picker on `d` | Same unintuitive toggle the feature is removing |
| Forward ctrl+up to the list then also set direction | List might move the cursor; chords must be exclusive |

---

## 7. What does not change

**Decision**: Dual count maps, session-scoped `Model.diffDirection`, embedded sync following the tree, confirm source→target copy, skip rules, and scripted CLI stay exactly as 008/009. This feature does not add loaders, walkers, or flags.

**Rationale**: Spec assumptions: arrows and keys only. YAGNI forbids a parallel direction type.

---

## 8. Testing strategy

**Decision**:

- **`treeview_test.go`**: injected maps; inbound `↓` on participating rows; outbound `↑`; childless root padded with two spaces and no glyph; zero/unknown counts still show the arrow; `▸` still present on the selected row and is not the direction glyph.
- **`app_test.go`**: replace `d` toggle tests with `KeyCtrlUp` / `KeyCtrlDown`; second `KeyCtrlUp` stays outbound; `KeyRunes{'d'}` leaves direction unchanged; plain `KeyUp` moves cursor only; footer has `ctrl+↑` / `ctrl+↓` and never `d: show`; session survival tests keep using the field, not `d`.
- **`inbound_test.go`**: footer helper strings.
- **`syncflow_test.go`**: standalone picker starts downward; `KeyCtrlUp` sets upward badges/help; second `KeyCtrlUp` stays upward; `d` does not flip; picker View has no `↓ ` / `↑ ` tree-style prefix; confirm View has no direction chords.

**Rationale**: Same split as 008 — injected maps for presentation; key messages for the new chords. No git tests (comparison unchanged).

---

## 9. README

**Decision**: Replace the main TUI `d` bullet with Control+Up / Control+Down (set outbound / inbound) and mention that the tree shows `↓` / `↑` on participating branches. Sync prose and the `branchy sync` comment drop “`d` toggles direction.”

**Rationale**: Footer is in-app discoverability (FR-009); README is the external key reference, same as 003/008/009.
