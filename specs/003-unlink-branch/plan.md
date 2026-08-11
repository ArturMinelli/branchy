# Implementation Plan: Unlink Branch from Tree

**Branch**: `003-unlink-branch` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-unlink-branch/spec.md`

## Summary

Add symmetric "unlink" capability to branchy: remove a branch and its entire subtree from the local `branch-tree.yaml` configuration. TUI users select a branch and press `u` to get a confirmation showing the branch name and subtree size; CLI users run `branchy unlink <parent> <child>` mirroring `link`. Core tree mutation lives in `internal/tree` (`UnlinkSubtree`, helpers); TUI adds a lightweight confirm screen; CLI adds `unlinkCmd` wired like `linkCmd`.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: cobra (CLI), bubbletea + lipgloss (TUI), gopkg.in/yaml.v3 (branch tree config)

**Storage**: File-based (`~/.config/branchy/` + per-project `branch-tree.yaml`); no schema changes

**Testing**: `go test` — unit tests for `UnlinkSubtree`, `SubtreeNames`, `ParentOf`, edge validation; TUI confirm/cancel transitions in `app_test.go` or dedicated test; manual quickstart scenarios

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: Subtree removal on typical trees (<100 branches) completes in <10ms; TUI confirm screen appears instantly

**Constraints**: Config-only mutation (no Git branch deletion); single-parent tree assumption; empty tree allowed after unlink; `u` key must not conflict with existing bindings

**Scale/Scope**: ~3–4 modified/new source files; no new external dependencies; mirrors existing `link` patterns

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates derived from existing codebase conventions |
| Library-first / separation of concerns | ✅ Pass | Tree mutation in `internal/tree`; TUI confirm in `internal/tui/app.go`; CLI wiring in `internal/cli` |
| CLI interface | ✅ Pass | `branchy unlink <parent> <child>` with stdout/stderr errors; exit code 1 on failure |
| Test coverage for new logic | ✅ Pass | Unit tests for subtree removal, parent edge cleanup, validation errors |
| Simplicity / YAGNI | ✅ Pass | Reuses link/save patterns; single confirm screen (no nested model); no new packages |
| Consistency with link | ✅ Pass | CLI args mirror `link`; TUI key sits alongside `l` in help |

**Post-design re-check**: All gates pass. `UnlinkSubtree` is the single mutation primitive shared by TUI (selected branch) and CLI (`child` after edge validation).

## Project Structure

### Documentation (this feature)

```text
specs/003-unlink-branch/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── unlink-cli.md
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
│   └── root.go              # Add unlinkCmd; register in init()
├── tree/
│   ├── tree.go              # UnlinkSubtree, SubtreeNames, ParentOf, HasEdge
│   └── tree_test.go         # New: unlink + helper tests
├── tui/
│   ├── app.go               # screenUnlinkConfirm, `u` key, updateUnlink, help text
│   └── app_test.go          # New (optional): confirm/cancel key handling
└── project/
    └── project.go           # SaveTree() unchanged
```

**Structure Decision**: Single Go module. Extend `internal/tree` with unlink primitives; TUI uses inline confirm screen (like early sync confirm, not a separate flow model); CLI mirrors `linkCmd` one-liner pattern.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
