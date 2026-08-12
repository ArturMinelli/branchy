# Implementation Plan: Diff Direction Toggle

**Branch**: `008-diff-direction-toggle` | **Date**: 2026-08-12 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/008-diff-direction-toggle/spec.md`

## Summary

Add a main-tree-only keystroke (`d`) that flips child file-change badges between inbound (parent → child, today’s default) and outbound (child → parent — the same files-changed number a child→parent MR would show). Both counts load eagerly from local three-dot diffs; direction is session-scoped on the root TUI model and named in the help footer. Sync picker and edge confirm stay inbound-only.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `os/exec` git seam (`InboundFiles` / `ResolveRef`); bubbletea key bindings; `internal/tui` inbound helpers from 006; remote refresh from 007

**Storage**: N/A (ephemeral in-memory direction + count maps; nothing persisted)

**Testing**: `go test ./internal/git/... ./internal/tui/...` — temp-repo outbound semantics; injected dual-map tests for toggle, footer, session direction, sync isolation

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: First paint still loads counts without a noticeable stall (both directions for ≤50 edges); toggle swaps cached maps with no git on the key path

**Constraints**: Tree-only toggle (FR-007); session-only direction (FR-006); same hide-zero / `?` / no-root rules (FR-004); scripted CLI unchanged (FR-010); remote refresh must refresh **both** maps in the current direction

**Scale/Scope**: 1 git helper (`OutboundFiles`) + dual loaders; Model direction + tree dual maps + footer; README key line; no sync/CLI/config changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–007 conventions |
| Library-first / separation | ✅ Pass | Comparison stays in `internal/git`. Direction and display policy stay in `internal/tui`. Sync does not import outbound. |
| CLI interface | ✅ Pass | No new flags; scripted paths untouched |
| Test coverage | ✅ Pass | Temp-repo outbound + symmetry; TUI toggle/footer/session/sync-isolation tests |
| Simplicity / YAGNI | ✅ Pass | Named `OutboundFiles` + mirrored loader; no persistence; no badge glyphs; one unused key |
| Consistency | ✅ Pass | Same three-dot / badge / remote-refresh patterns as 006–007; footer via existing `RenderHelp` |

**Post-design re-check**: All gates pass. Deep module is `git.OutboundFiles` (hides reverse three-dot order). Session direction on `Model` is the natural home so project switches do not reset mode. Sync isolation is a deliberate non-wire, not a missing feature.

## Project Structure

### Documentation (this feature)

```text
specs/008-diff-direction-toggle/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── diff-direction.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/git/
├── inbound.go           # Extend or sibling: OutboundFiles (parent...child three-dot)
└── inbound_test.go      # Extend: outbound cases + Inbound/Outbound symmetry

internal/tui/
├── inbound.go           # Rename count type; add loadOutboundCounts; tree help footer helper
├── inbound_test.go      # Badge reuse + footer strings for both modes
├── treeview.go          # Hold inbound+outbound maps; direction; render active map
├── treeview_test.go     # Toggle / zero-hide per direction / roots / unknown
├── app.go               # Model.diffDirection; d key; load both maps; footer; remote refresh both
├── app_test.go          # Session direction survives leave/return & project switch; remote keeps direction
└── syncflow.go          # Unchanged count path (inbound only) — regression covered from app/sync tests

README.md                # Document d + inbound/outbound on main tree
```

**Structure Decision**: Single Go module. Outbound comparison lives beside `InboundFiles` in `internal/git` (same seam, same ref resolution). TUI owns direction on `Model` and dual caches on `BranchTreeView`. No new package.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Implementation Notes

### Phase ordering (for tasks.md)

1. **Foundation**: `git.OutboundFiles` + temp-repo / symmetry tests
2. **Loaders**: `fileChangeCount` rename (or shared type), `loadOutboundCounts`, shared badge formatter
3. **Tree view**: dual maps + direction; render active; unit tests
4. **App wiring**: default inbound; `d` toggles; footer helper; load both on `selectProject` and remote success; session field not reset on project change
5. **Docs / verify**: README keys; `go test`; quickstart scenarios

### Key patterns

**Outbound comparison**:

```go
// OutboundFiles(dir, parent, child): git diff --name-only parent...child
// Same ResolveRef rules as InboundFiles. Error → unknown, never 0.
```

**Eager dual cache**:

```go
in := loadInboundCounts(path, tree)
out := loadOutboundCounts(path, tree)
treeView.setFileCounts(in, out) // direction unchanged
```

**Session direction**:

```go
// Model.diffDirection defaults to inbound
// updateTree: d → flip; treeView.setDirection(m.diffDirection)
// selectProject: reload maps; do NOT reset m.diffDirection
```

**Footer**:

```text
inbound:  … d: show outbound … | counts: inbound (parent→child)
outbound: … d: show inbound …  | counts: outbound (child→parent)
```

### Files intentionally unchanged

- `internal/cli/*`, `internal/sync/sync.go` — scripted CLI (FR-010)
- Sync picker / confirm display rules — stay inbound-only (FR-007)
- `internal/tree/*` — `ParentOf` sufficient
- MR / link / unlink / init / projects flows — no direction toggle
- Fetch / remote-update machinery — only the apply path gains a second load

### Verification

- `go test ./internal/git/... ./internal/tui/...`
- Manual quickstart scenarios 1–5
- With tree outbound, sync picker/confirm still match inbound GitLab parent→child counts
