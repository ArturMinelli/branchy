# Implementation Plan: Merge Direction Arrows

**Branch**: `010-direction-arrows` | **Date**: 2026-08-20 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/010-direction-arrows/spec.md`

## Summary

Make merge direction obvious on the main tree: a muted `↓` / `↑` before every participating branch name (padded on childless roots so names still line up), and replace the `d` toggle with **set** chords — Control+Up → outbound (child→parent), Control+Down → inbound (parent→child). Standalone sync picker uses the same chords and no arrows. Count meaning, session scope, and sync follow-tree behavior stay as in 008/009.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `internal/tui` tree + `diffDirection`; Bubble Tea `key.Binding` (`ctrl+up` / `ctrl+down` → `tea.KeyCtrlUp` / `tea.KeyCtrlDown`); lipgloss row styles from `treeview.go`

**Storage**: N/A (ephemeral session direction unchanged; nothing persisted)

**Testing**: `go test ./internal/tui/...` — tree glyphs + padding, set-not-toggle chords, `d` no-op, footer/picker help, confirm screens ignore direction keys

**Target Platform**: Linux/macOS terminal (existing branchy platforms). Control+Up/Down require a terminal that emits CSI `\x1b[1;5A` / `\x1b[1;5B` (Bubble Tea already maps these). No fallback chord.

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: Direction change is cache-only (same as today’s toggle) — no git on the key path; glyph is one extra cell per participating row

**Constraints**: Set-not-toggle (FR-005/006); `d` removed (FR-008); arrows only on the main tree (FR-013/016); confirm screens freeze direction (FR-014); scripted CLI unchanged (FR-015); muted connector color; pad non-arrow rows

**Scale/Scope**: `BranchTreeView` glyph + participating flag; replace one key binding with two set-chords on tree and standalone picker; footer/README copy; no new package, no git/sync/tree walk changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–009 conventions |
| Library-first / separation | ✅ Pass | Display policy stays in `internal/tui`. Git comparison and sync walks are untouched. |
| CLI interface | ✅ Pass | No new flags; scripted `--from` / `-y` unchanged |
| Test coverage | ✅ Pass | Tree render + key set/no-op + picker chords + confirm ignore; injected maps as in 008 |
| Simplicity / YAGNI | ✅ Pass | No fallback keys, no per-row colors, no connector replacement; pad + muted glyph only |
| Consistency | ✅ Pass | Reuses `diffDirection`, dual maps, `treeHelpFooter` / `syncPickerHelp`, existing `key.Matches` order |

**Post-design re-check**: All gates pass. Natural home is `BranchTreeView.renderRow` (glyph + pad) and the existing direction field on `Model` / `SyncFlowModel` (set chords instead of flip). `d` is deleted from those key maps, not left as a hidden alias.

## Project Structure

### Documentation (this feature)

```text
specs/010-direction-arrows/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── direction-arrows.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/tui/
├── treeview.go          # participating flag on rows; ↓/↑ or two-space pad before name; muted connector style
├── treeview_test.go     # glyphs, padding, childless roots, selected marker ≠ arrow, counts still work
├── inbound.go           # treeHelpFooter: ctrl+↑ / ctrl+↓ set-chords; drop `d: show …`
├── inbound_test.go      # footer strings both modes; no `d: show`
├── app.go               # DirectionUp / DirectionDown bindings; set outbound/inbound; match before Up/Down
├── app_test.go          # CtrlUp/CtrlDown set; second CtrlUp stays outbound; `d` and plain Up do not flip
├── syncflow.go          # picker: same set-chords; drop flip/`d`; help copy; confirms still unbound
└── syncflow_test.go     # picker CtrlUp/Down; `d` no-op; no tree arrows in picker View; confirm ignores chords

README.md                # Replace `d` with Control+Up / Control+Down; mention tree arrows
```

**Structure Decision**: Single Go module. This is a TUI chrome + keybinding change on the existing direction owner (`Model.diffDirection` / `SyncFlowModel.direction`). No new package. `internal/git`, `internal/sync`, `internal/tree`, and `internal/cli` stay closed.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Implementation Notes

### Phase ordering (for tasks.md)

1. **Tree glyphs**: participating rows + pad; render `↓ ` / `↑ ` before the name in connector style
2. **Tree keys + footer**: replace `d` flip with Control+Up / Control+Down set; update `treeHelpFooter`
3. **Standalone picker**: same set-chords and help; no arrows on list rows
4. **Docs / verify**: README; `go test ./internal/tui/...`; quickstart scenarios

### Key patterns

**Participating row** (flatten time, from the tree document):

```text
participating = has parent OR has at least one child
childless root → participating = false
```

**Glyph column** (always two cells so names line up):

```text
inbound  + participating     → "↓ "
outbound + participating     → "↑ "
not participating            → "  "
```

Place immediately after the connector, before the branch name. Style with `connectorStyle` (same as `├──`). Selection marker stays `▸ ` and is not reused as the direction glyph.

**Set chords** (match before plain Up/Down):

```go
// keys.DirectionUp:   "ctrl+up"   → diffOutbound / sync.Upward
// keys.DirectionDown: "ctrl+down" → diffInbound  / sync.Downward
// already-active chord: no-op (assign the same value)
// "d": not bound
```

Tests send `tea.KeyMsg{Type: tea.KeyCtrlUp}` / `tea.KeyCtrlDown`. Bindings use Bubble Tea’s `ctrl+up` / `ctrl+down` strings (already mapped from CSI `\x1b[1;5A` / `\x1b[1;5B`).

**Footer** (both chords always listed, not “the other mode”):

```text
↑/↓: navigate  ctrl+↑: outbound  ctrl+↓: inbound  s: sync  …
counts: inbound (parent→child)   |   counts: outbound (child→parent)
```

Picker help keeps `sync: downward|upward …` and the same `ctrl+↑` / `ctrl+↓` set hints. Confirm/summary/browser help stay free of direction keys.

### Files intentionally unchanged

- `internal/git/*`, `internal/sync/*`, `internal/tree/*` — counts, walks, MR ends
- `internal/cli/*` — no direction flag
- MR / link / unlink / init / projects flows — no direction chords or arrows
- Dual count loaders and session ownership of `diffDirection` — 008/009 behavior

### Verification

- `go test ./internal/tui/...`
- Manual quickstart scenarios 1–5
- `d` nowhere in main-tree or picker help; Control+Up twice does not return to inbound
