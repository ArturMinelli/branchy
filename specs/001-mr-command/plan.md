# Implementation Plan: Manual MR Command

**Branch**: `001-mr-command` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-mr-command/spec.md`

## Summary

Add a `branchy mr` subcommand and an `m` shortcut in the main TUI so users can manually create a GitLab merge request between any two branches in the project's branch tree. The flow uses a multi-step Bubble Tea TUI (source picker → target picker → title edit → confirm → optional browser open) and supports non-interactive `--source` / `--target` flags. Core MR creation logic is extracted into a shared `internal/mr` package reused by both the new command and the existing sync flow.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: cobra (CLI), bubbletea + bubbles + lipgloss (TUI), glab (GitLab MR via shell-out), gopkg.in/yaml.v3 (branch tree config)

**Storage**: File-based (`~/.config/branchy/` index + per-project `branch-tree.yaml`); no database

**Testing**: `go test` — unit tests for MR validation, title defaults, branch filtering; TUI step transitions via bubbletea model tests; integration tests with mocked glab where feasible

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: TUI step transitions feel instant (<100ms perceived); MR creation bounded by glab/GitLab network latency

**Constraints**: Must reuse existing gitlab/browser/project packages; no automatic browser open in MR flow (unlike sync); branch picker limited to `tree.Document.Names()`

**Scale/Scope**: Single-repo CLI feature; ~5 new/modified source files; one new internal package; no new external dependencies

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates derived from existing codebase conventions |
| Library-first / separation of concerns | ✅ Pass | MR business logic in `internal/mr`; TUI in `internal/tui`; CLI wiring in `internal/cli` |
| CLI interface | ✅ Pass | `branchy mr` subcommand with flags; human-readable stdout/stderr |
| Test coverage for new logic | ✅ Pass | Unit tests for validation + MR service; TUI model tests for flow steps |
| Simplicity / YAGNI | ✅ Pass | Reuses bubbles `list` for pickers; no new deps; flag-mode browser deferred |
| No breaking changes to sync | ✅ Pass | Sync behavior unchanged; shared MR helper preserves existing auto-browser in sync |

**Post-design re-check**: All gates still pass. Shared `internal/mr` package reduces duplication without expanding scope.

## Project Structure

### Documentation (this feature)

```text
specs/001-mr-command/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── mr-cli.md
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
│   └── root.go              # Add mrCmd + flags
├── mr/
│   ├── mr.go                  # CreateRequest, Create, validation
│   └── mr_test.go
├── tui/
│   ├── app.go                 # Add screenMR, `m` keybinding
│   ├── mrflow.go              # MRFlowModel (multi-step TUI)
│   ├── mrflow_test.go
│   ├── treeview.go
│   └── ...
├── sync/
│   └── sync.go                # Refactor to call internal/mr
├── gitlab/
│   └── glab.go                # Unchanged (used by internal/mr)
├── browser/
│   └── open.go                # Called from mr flow on user confirm
├── tree/
│   └── tree.go                # Names() used for branch lists
└── project/
    └── project.go
```

**Structure Decision**: Single Go module, extend existing `internal/` layout. New `internal/mr` package owns manual MR creation; new `internal/tui/mrflow.go` owns the reusable TUI wizard embeddable in both standalone `branchy mr` and main app.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
