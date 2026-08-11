# Implementation Plan: Parent Diff Display

**Branch**: `006-parity-diff-display` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-parity-diff-display/spec.md`

## Summary

Show each child branch’s inbound file-change count versus its parent — the same files-changed number GitLab would show on the parent→child merge request — on the main tree, the sync root picker, and each sync edge confirm. Compute locally with a three-dot `git diff` (no fetch, no GitLab API). Hide zeros on browse surfaces; always show a known count (including zero) on confirm; show a clear unknown placeholder when comparison fails.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `os/exec` git wrapper; bubbletea / bubbles / lipgloss TUI; `internal/tree.ParentOf`; no new modules

**Storage**: N/A (ephemeral in-memory counts; no schema changes)

**Testing**: `go test ./internal/git/... ./internal/tui/...` — temp-repo tests for comparison semantics; injected-count tests for tree/picker/confirm rendering

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: Typical trees (≤50 edges) show counts on first paint without a noticeable stall; counts must not recompute on every `View()` frame

**Constraints**: Local refs only (FR-009); scripted CLI unchanged (FR-010); compact display for narrow terminals; same numeric value across all three surfaces (FR-005)

**Scale/Scope**: 1 new git comparison module + tests; tree view, sync picker, and sync confirm wiring; no changes to MR/link/unlink/init flows beyond shared `branchItem` remaining backward compatible

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004/005 codebase conventions |
| Library-first / separation | ✅ Pass | Comparison lives in `internal/git` (independently testable). TUI only formats and places counts. Tree document stays the source of parent edges. |
| CLI interface | ✅ Pass | Scripted paths untouched; TUI-only presentation |
| Test coverage | ✅ Pass | Temp-repo tests lock GitLab-comparable semantics; view tests lock hide-zero / unknown / confirm-zero rules |
| Simplicity / YAGNI | ✅ Pass | One function (`InboundFiles`) + one in-memory map. No GitLab API, no fetch, no async spinner for local git. |
| Consistency | ✅ Pass | Reuses `helpStyle` / `warnStyle`; confirm context lines; existing `ParentOf` |

**Post-design re-check**: All gates pass. Deep module is `git.InboundFiles` (hides three-dot / name-only details). Display rules live in one shared formatter so tree, picker, and confirm cannot drift.

## Project Structure

### Documentation (this feature)

```text
specs/006-parity-diff-display/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── inbound-count.md # Phase 1 output
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/git/
├── root.go              # Existing: Root, SlugFromPath
├── inbound.go           # New: InboundFiles, InboundCount
└── inbound_test.go      # New: temp-repo three-dot / zero / missing-ref

internal/tui/
├── inbound.go           # New: loadInboundCounts, formatInboundBadge, formatInboundConfirm
├── inbound_test.go      # New: hide-zero / unknown / confirm-zero formatting
├── treeview.go          # Extend: attach counts; render badge on non-root rows
├── treeview_test.go     # Extend: hide zero, show N, hide roots, unknown placeholder
├── app.go               # Compute counts in selectProject; pass into tree view
├── mrflow.go            # Extend branchItem + newBranchList to accept optional badge (default empty)
├── syncflow.go          # Load counts for picker titles and edge-confirm context
└── syncflow_test.go     # Extend: picker badge + confirm context include count
```

**Structure Decision**: Single Go module. Comparison is a deep module in the existing `internal/git` package (already the `os/exec` git seam). TUI owns display policy in one helper file so all three surfaces share formatting. No new package.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Implementation Notes

### Phase ordering (for tasks.md)

1. **Foundation**: `git.InboundFiles` + temp-repo tests (three-dot inbound, zero, missing ref, diverged child-only files ignored)
2. **Display helpers**: `loadInboundCounts`, badge/confirm formatters + unit tests
3. **Main tree**: attach map on `selectProject`; render badge; treeview tests
4. **Sync picker**: pass counts into `newBranchList` / `branchItem.Title`; other pickers stay name-only
5. **Sync confirm**: add inbound context line (including known zero); unknown placeholder
6. **Verify**: `go test ./internal/git/... ./internal/tui/...`; quickstart scenarios

### Key patterns

**Comparison (GitLab MR files-changed)**:

```go
// InboundFiles: git -C dir diff --name-only child...parent
// count non-empty lines. Missing/unresolvable refs → error (unknown, not 0).
```

**Eager cache, not per-frame**:

```go
counts := loadInboundCounts(p.Path, p.Tree) // once on selectProject / sync start
treeView.setInbound(counts)
```

**Display policy**:

| Surface | Zero | Unknown | Positive N |
|---------|------|---------|------------|
| Main tree / sync picker | hide | `?` | `N` |
| Edge confirm | show `0 files would change` | `File count unavailable` | `N files would change` |

### Files intentionally unchanged

- `internal/cli/*`, `internal/sync/sync.go` — scripted CLI (FR-010)
- `internal/mr/*`, `internal/gitlab/*` — no network compare
- `internal/tree/*` — `ParentOf` already sufficient
- `internal/tui/{link,unlink,init,mr,projects}flow.go` behavior — pickers remain name-only unless they reuse `newBranchList` with empty badges
- Sync results summary — out of scope

### Verification

- `go test ./internal/git/... ./internal/tui/...`
- Manual quickstart scenarios 1–5
- Same parent→child pair shows the same N on tree, picker, and confirm
