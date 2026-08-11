# Implementation Plan: TUI Native Confirm & Loading

**Branch**: `005-sync-tui-polish` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-sync-tui-polish/spec.md`

## Summary

Replace inline `[y/N]` text prompts across all branchy TUI yes/no steps with a shared `ConfirmModel` (bordered panel, arrow-key Yes/No focus, y/n shortcuts). Add a shared `LoadingModel` using `bubbles/spinner` for network-bound work during sync MR creation, MR wizard creation, browser-open batches, and init registration. Refactor MR and init flows from synchronous `Update` calls to async `tea.Cmd` patterns (already used by sync). No changes to scripted CLI paths.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: bubbletea, bubbles (spinner, key, list), lipgloss; existing `internal/sync`, `internal/mr`, `internal/project`, `internal/browser`

**Storage**: N/A (presentation-only; no schema changes)

**Testing**: `go test ./internal/tui/...` — unit tests for `ConfirmModel`/`LoadingModel`; extend flow step tests in `syncflow_test.go`, `mrflow_test.go`, `initflow_test.go`, `linkflow_test.go`, `unlinkflow_test.go`, `app_test.go`

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: Spinner visible within 200ms of confirm; confirm focus toggle <16ms perceived; MR/init latency still bounded by glab/filesystem

**Constraints**: Block all input during loading; default confirm focus No; scripted CLI (`UseTUI` false) unchanged; minimum terminal 80×24; embedded and standalone flows must share identical chrome

**Scale/Scope**: 2 new source files + tests; 7 modified flow files (`syncflow`, `mrflow`, `linkflow`, `unlinkflow`, `initflow`, `app`, `chrome`); no new external dependencies

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004 codebase conventions |
| Library-first / separation | ✅ Pass | Shared `ConfirmModel`/`LoadingModel` in `internal/tui`; domain logic stays in `sync`/`mr`/`project` |
| CLI interface | ✅ Pass | Scripted paths untouched; TUI-only presentation layer |
| Test coverage | ✅ Pass | Component tests + flow step tests per research §8 |
| Simplicity / YAGNI | ✅ Pass | Two small shared models; no wizard framework; link/unlink skip loading (fast I/O) |
| Consistency | ✅ Pass | Extends `chrome.go` tokens; all flows adopt same components |

**Post-design re-check**: All gates pass. Async refactor limited to flows that currently block UI (MR, init, browser batch). Confirm/loading contracts documented in `contracts/tui-chrome.md`.

## Project Structure

### Documentation (this feature)

```text
specs/005-sync-tui-polish/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── tui-chrome.md    # Phase 1 output
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not yet created)
```

### Source Code (repository root)

```text
internal/tui/
├── chrome.go            # Extend: confirmBorder, buttonFocus, buttonBlur styles
├── confirm.go           # New: ConfirmModel, ConfirmOptions, ConfirmChoice
├── confirm_test.go      # New: focus, shortcuts, view lacks [y/N]
├── loading.go           # New: LoadingModel, spinner integration
├── loading_test.go      # New: tick updates, Active()
├── flowutil.go          # Add Left/Right bindings for confirm navigation
├── syncflow.go          # Confirm + loading on edge/browser; stepSyncBrowserLoading
├── syncflow_test.go     # Extend: confirm view, loading ignores keys
├── mrflow.go            # Confirm; async stepMRLoading, stepMRBrowserLoading
├── mrflow_test.go       # Extend: async confirm → loading → result
├── linkflow.go          # Confirm panel only
├── linkflow_test.go     # Extend: confirm view
├── unlinkflow.go        # Confirm panel only
├── unlinkflow_test.go   # Extend: confirm view
├── initflow.go          # Confirm; async stepInitLoading
├── initflow_test.go     # Extend: loading step
└── app.go               # screenUnlinkConfirm uses ConfirmModel
```

**Structure Decision**: Single Go module. Shared components live alongside `chrome.go`. Each flow embeds `ConfirmModel`/`LoadingModel` fields and delegates Update/View. Domain packages unchanged.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Implementation Notes

### Phase ordering (for tasks.md)

1. **Foundation**: `confirm.go`, `loading.go`, `chrome.go` style extensions, `flowutil.go` left/right keys + tests
2. **Sync**: Wire confirm + enhance `stepSyncProcessing` spinner; add `stepSyncBrowserLoading`
3. **MR**: Async refactor + confirm + loading (biggest behavior change)
4. **Init**: Async refactor + confirm + loading
5. **Link / Unlink / App**: Confirm-only adoption (lower risk)
6. **Polish**: Visual audit per quickstart Scenario 9; grep TUI views for `[y/N]`

### Key patterns

**Confirm delegation in flow Update**:

```go
case stepSyncEdgeConfirm:
    if m.loading.Active() { return m, nil }
    var choice ConfirmChoice
    m.confirm, choice = m.confirm.Update(msg)
    switch choice {
    case ConfirmYes:
        m.step = stepSyncProcessing
        m.loading = NewLoading(LoadingOptions{...})
        return m, tea.Batch(m.loading.Init(), runEdgeCmd(...))
    case ConfirmNo:
        // skip edge...
    }
```

**MR async (new)**:

```go
type mrCreateResultMsg struct { result *mr.CreateResult; err error }

func runMRCreateCmd(p *project.Project, req mr.CreateRequest) tea.Cmd {
    return func() tea.Msg {
        r, err := mr.Create(p, req)
        return mrCreateResultMsg{result: r, err: err}
    }
}
```

### Files intentionally unchanged

- `internal/cli/*` — mode detection unchanged
- `internal/sync/sync.go` — plain CLI stdin prompts
- `internal/mr/mr.go`, `internal/project/project.go` — domain logic
- `internal/tui/projectsflow.go` — no yes/no confirm in scope

### Verification

- `go test ./internal/tui/...`
- Manual quickstart scenarios 1–9
- `rg '\[y/N\]' internal/tui/` should return zero matches in View strings after implementation
