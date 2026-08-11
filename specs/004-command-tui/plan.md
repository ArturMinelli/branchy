# Implementation Plan: Unified Command TUI

**Branch**: `004-command-tui` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-command-tui/spec.md`

## Summary

Bring all branchy subcommands to a consistent dedicated TUI experience when run interactively (TTY, no flags). Primary gap: `branchy sync` still uses plain numbered stdin prompts while embedded tree sync (`s`) uses `SyncFlowModel`. Add `RunSync` with root-branch picker, plus new `RunLink`, `RunUnlink`, `RunInit`, `RunProjects` flows mirroring `RunMR`. Extract shared TUI chrome. Wire CLI commands through `UseTUI(cmd)` helper enforcing the any-flag → plain CLI rule. Preserve all scripted paths for CI and automation.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: cobra (CLI), bubbletea + bubbles + lipgloss (TUI), charmbracelet/x/term (TTY detection), existing `internal/sync`, `internal/mr`, `internal/project`, `internal/tree`

**Storage**: File-based (`~/.config/branchy/`); no schema changes

**Testing**: `go test` — unit tests for `UseTUI`/`IsTTY` flag detection; flow model step transitions (mirror `mrflow_test.go` / `syncflow_test.go`); non-TTY fallback errors; manual quickstart scenarios

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: TUI screen transitions <100ms perceived; list filtering responsive on trees <100 branches; MR creation bounded by glab latency

**Constraints**: Any flag disables TUI; dedicated flows exit to shell (not main tree); embedded main-app shortcuts unchanged; minimum terminal 80×24; non-TTY must never hang

**Scale/Scope**: ~10–12 new/modified source files; no new external dependencies; `mr` partial-flag behavior alignment (minor breaking change for `--source` alone)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates derived from existing codebase conventions |
| Library-first / separation of concerns | ✅ Pass | Mode detection in `internal/cli`; flow models in `internal/tui/*flow.go`; domain logic stays in `internal/sync`, `internal/tree`, `internal/project` |
| CLI interface | ✅ Pass | Scripted paths unchanged; TUI is additive interactive layer |
| Test coverage for new logic | ✅ Pass | `mode_test.go` + per-flow step transition tests |
| Simplicity / YAGNI | ✅ Pass | Reuses `SyncFlowModel`, `MRFlowModel` patterns; one model per command, no wizard framework |
| Consistency with existing flows | ✅ Pass | `RunMR` is reference; shared chrome extracted once |

**Post-design re-check**: All gates pass. `UseTUI` centralizes mode selection; each `Run*` is a thin tea.Program wrapper; link/unlink/init/projects flows are independent and testable.

## Project Structure

### Documentation (this feature)

```text
specs/004-command-tui/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── command-tui.md
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
│   ├── root.go              # Wire UseTUI branches per command; link/unlink RangeArgs(0,2)
│   ├── mode.go              # New: IsTTY, UseTUI, anyFlagChanged
│   └── mode_test.go         # New: flag/TTY detection tests
├── tui/
│   ├── chrome.go            # New: shared styles, MinSizeOK, RenderTooSmall
│   ├── syncflow.go          # Add stepSyncPickRoot, RunSync
│   ├── syncflow_test.go     # Extend: root pick + standalone quit
│   ├── linkflow.go          # New: RunLink
│   ├── linkflow_test.go     # New
│   ├── unlinkflow.go        # New: RunUnlink
│   ├── unlinkflow_test.go   # New
│   ├── initflow.go          # New: RunInit
│   ├── initflow_test.go     # New
│   ├── projectsflow.go      # New: RunProjects
│   ├── projectsflow_test.go # New
│   ├── mrflow.go            # Use chrome.go styles; no behavior change
│   ├── app.go               # Import chrome; embedded flows unchanged
│   └── ...
├── sync/                    # Unchanged
├── project/                 # Unchanged
└── tree/                    # Unchanged
```

**Structure Decision**: Single Go module. New `mode.go` owns interactive vs scripted routing. Each subcommand gets a dedicated `*flow.go` with `Run*` entry point following `RunMR`. `SyncFlowModel` extended rather than duplicated.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Implementation Notes

### Phase ordering (for tasks.md)

1. **Foundation**: `chrome.go`, `mode.go` + tests
2. **Sync**: `RunSync` + root picker + CLI wiring (highest value)
3. **Link / Unlink**: dedicated flows + `RangeArgs` change
4. **Init / Projects**: wizards + non-TTY fallbacks
5. **MR alignment**: any-flag rule in `mrCmd`
6. **Polish**: min-size guard across flows; README update

### Key integration points

```go
// internal/cli/root.go (sync example)
if UseTUI(cmd) {
    p, err := project.ResolveFromCWD()
    if err != nil { return err }
    return tui.RunSync(p, tui.SyncFlowOptions{})
}
// existing plain CLI path below
```

### Files intentionally unchanged

- `internal/tui/app.go` embedded sync/MR/link/unlink shortcuts
- `internal/sync/sync.go` orchestration
- GitLab MR creation logic
