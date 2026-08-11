# Implementation Plan: Interactive Sync Confirmations

**Branch**: `002-interactive-sync-confirm` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-interactive-sync-confirm/spec.md`

## Summary

Change sync (TUI `s` shortcut and CLI `branchy sync`) so MRs are never created in a single bulk shot without per-edge consent. The TUI gains a multi-step sync flow (per-edge confirm → create → summary → optional browser batch). The CLI keeps its existing per-edge stdin prompts but stops auto-opening tabs; both surfaces prompt at the end to open **created-only** MR URLs in **DFS edge order**. Remove auto-browser from `internal/sync.Run`; callers own the end browser decision.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: cobra (CLI), bubbletea + bubbles + lipgloss (TUI), glab (GitLab MR via shell-out), existing `internal/mr`, `internal/sync`, `internal/browser`

**Storage**: File-based (`~/.config/branchy/` + per-project `branch-tree.yaml`); no schema changes

**Testing**: `go test` — unit tests for `CreatedURLs` / browser batch helpers; `SyncFlowModel` step transitions (mirror `mrflow_test.go`); CLI prompt behavior via stdin tests or integration smoke; manual quickstart scenarios

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: Per-edge TUI transitions feel instant (<100ms perceived); MR creation bounded by glab/GitLab latency; sequential browser opens complete within 2s for ≤10 tabs

**Constraints**: Must not change manual MR flow (`branchy mr` / `m`); `-y` skips per-edge prompts only, not browser consent; browser batch excludes skipped/declined/failed edges; tab order = DFS among created MRs

**Scale/Scope**: ~4–6 modified/new source files; no new external dependencies; breaking behavior change for sync auto-browser (intentional per spec)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates derived from existing codebase conventions |
| Library-first / separation of concerns | ✅ Pass | Sync orchestration in `internal/sync`; TUI wizard in `internal/tui/syncflow.go`; CLI wiring in `internal/cli` |
| CLI interface | ✅ Pass | `branchy sync` retains flags; new end browser prompt on stdout/stdin |
| Test coverage for new logic | ✅ Pass | Unit tests for URL batch helper + TUI model transitions |
| Simplicity / YAGNI | ✅ Pass | Reuses `MRFlowModel` patterns; no new deps; incremental edge API instead of rewriting tree walk |
| Consistency with MR flow | ✅ Pass | Browser prompt semantics align with manual MR (`[y/N]`, non-fatal open failure) |

**Post-design re-check**: All gates pass. `SyncFlowModel` mirrors proven `MRFlowModel` embedding pattern; shared `sync.CreatedURLs` + `sync.OpenURLs` keeps CLI/TUI DRY without over-abstracting.

## Project Structure

### Documentation (this feature)

```text
specs/002-interactive-sync-confirm/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── sync-cli.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not yet created)
```

### Source Code (repository root)

```text
cmd/branchy/
└── main.go

internal/
├── cli/
│   └── root.go              # syncCmd: remove reliance on auto-browser; add end prompt
├── sync/
│   ├── sync.go              # Remove auto-browser; add CreatedURLs, OpenURLs, RunEdge
│   └── sync_test.go         # New: URL batch + ordering tests
├── tui/
│   ├── app.go               # Delegate screenSync → SyncFlowModel (like screenMR)
│   ├── syncflow.go          # New: per-edge confirm + summary + browser prompt
│   ├── syncflow_test.go     # New: step transition tests
│   ├── mrflow.go            # Unchanged (reference pattern)
│   └── ...
├── browser/
│   └── open.go              # Unchanged; called sequentially from sync.OpenURLs
├── mr/
│   └── mr.go                # Unchanged
└── tree/
    └── tree.go              # CollectEdges (DFS order) — unchanged
```

**Structure Decision**: Single Go module. New `SyncFlowModel` in `internal/tui` drives interactive TUI sync; `internal/sync` exposes single-edge creation and shared browser-batch helpers used by both TUI and CLI.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
