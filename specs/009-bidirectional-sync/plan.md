# Implementation Plan: Bidirectional Sync

**Branch**: `009-bidirectional-sync` | **Date**: 2026-08-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/009-bidirectional-sync/spec.md`

## Summary

Interactive sync follows the existing tree direction: inbound `s` stays today’s parent→child cascade; outbound `s` offers the **same descendant edges** as child→parent MRs, deepest-first, with outbound confirm counts. Standalone `branchy sync` gets the same `d` toggle on the root picker (starts downward). Scripted `--from` / `-y` stay downward-only. One cascade path: tree owns walks, `sync.Direction` owns MR ends, `SyncFlowModel` maps UI mode onto that.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `internal/tree` edge walk; `internal/sync` + `internal/mr` / `gitlab.FindOpenMR`; Bubble Tea `SyncFlowModel`; 008 `diffDirection` + dual count loaders

**Storage**: N/A (ephemeral direction per sync run; standalone picker not persisted)

**Testing**: `go test ./internal/tree/... ./internal/sync/... ./internal/tui/... ./internal/cli/...` — post-order membership, `Ends`, embedded follow-tree, standalone picker `d`, downward/CLI isolation

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: No extra git on the confirm key path (both count maps already loaded, same as 008 tree); walk is in-memory on a small tree (≤50 edges)

**Constraints**: Same subtree as `CollectEdges` (FR-003); no CLI direction flag (FR-013); no `d` on confirm (FR-012); skip only matching `(source, target)` (FR-014); 008 session tree direction unchanged

**Scale/Scope**: 1 tree walker (`CollectEdgesUpward`); `sync.Direction` + `Ends` / `EdgesBelow` / `Result.Source|Target`; `SyncFlowModel` dual maps + picker `d`; README; no new package

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–008 conventions |
| Library-first / separation | ✅ Pass | Walk in `internal/tree`. MR ends in `internal/sync`. UI mapping in `internal/tui`. CLI does not import upward. |
| CLI interface | ✅ Pass | No new flags; scripted `--from` / `-y` unchanged |
| Test coverage | ✅ Pass | Tree order + set equality; sync ends; TUI embedded/standalone; CLI isolation |
| Simplicity / YAGNI | ✅ Pass | One flow, one Direction enum, no second command or persisted picker mode |
| Consistency | ✅ Pass | Reuses 008 loaders/badges, 002 confirm/browser rules, existing `FindOpenMR` |

**Post-design re-check**: All gates pass. Deep modules: `CollectEdgesUpward` (hides post-order) and `sync.Ends` (hides source/target). `SyncFlowModel.direction` is the run owner so confirm screens stay toggle-free. 008 “sync isolation” is deliberately superseded, not left as a conflicting rule.

## Project Structure

### Documentation (this feature)

```text
specs/009-bidirectional-sync/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── bidirectional-sync.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/tree/
├── tree.go              # CollectEdgesUpward + shared private walker
└── tree_test.go         # Set equality + post-order vs sibling order

internal/sync/
├── sync.go              # Direction, Ends, EdgesBelow; Options.Direction; Result.Source/Target; RunEdge(dir)
└── sync_test.go         # Ends table; Result fields; existing URL-order tests still hold

internal/tui/
├── inbound.go           # Generalize confirm formatter (receiving branch)
├── syncflow.go          # direction, dual maps, picker d, Ends in copy, EdgesBelow
├── syncflow_test.go     # Standalone toggle; confirm copy both ways; no d on confirm
├── app.go               # newSyncFlowModel(..., m.diffDirection) on s
└── app_test.go          # Replace 008 isolation: outbound s → upward confirm

internal/cli/
└── root.go              # Unchanged (grep: no direction flag)

README.md                # Bidirectional interactive sync; scripted stays downward
```

**Structure Decision**: Single Go module. No new package. Tree gains the upward walk beside `CollectEdges`. Sync gains Direction as the MR-end owner. TUI extends `SyncFlowModel` instead of adding a second flow.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Implementation Notes

### Design sentence

**A sync run is the same parent/child edge list with a direction: the tree walks it, sync names the MR ends, the TUI only maps inbound/outbound onto that.**

### Phase ordering (for tasks.md)

1. **Foundation**: `CollectEdgesUpward` + set/order tests
2. **Sync core**: `Direction`, `Ends`, `EdgesBelow`, `Run`/`RunEdge`/`Result` source-target; Downward default
3. **TUI flow**: dual maps, confirm/picker copy, standalone `d`, embedded direction from `Model`
4. **App wiring**: `s` passes `diffDirection`; replace 008 isolation tests
5. **Docs / verify**: README; `go test`; quickstart scenarios

### Key patterns

**Post-order walk**:

```go
func collectEdgesUpward(parent string, edges *[]Edge) {
    children := sorted(node.Children)
    for _, child := range children {
        collectEdgesUpward(child, edges)
        *edges = append(*edges, Edge{Parent: parent, Child: child})
    }
}
```

**Ends**:

```go
// Downward: Parent → Child
// Upward:   Child  → Parent
```

**Embedded start**:

```go
newSyncFlowModel(p, selected, SyncFlowOptions{Embedded: true}, m.diffDirection)
```

**Standalone picker**:

```go
// d on stepSyncPickRoot only → flip direction; rebuild badges from the other map
```

### Files intentionally unchanged

- `internal/cli/root.go` — no direction flag (FR-013)
- `internal/mr/*`, `internal/gitlab/*` — already keyed by source/target
- Manual MR flow, link/unlink/init/projects
- Main-tree `d` binding and footer (008) — only the `s` handoff changes
- Fetch machinery — sync already refreshes counts; load both maps

### Verification

- `go test ./internal/tree/... ./internal/sync/... ./internal/tui/... ./internal/cli/...`
- Manual quickstart scenarios 1–7
- `branchy sync --help` lists no direction flag
