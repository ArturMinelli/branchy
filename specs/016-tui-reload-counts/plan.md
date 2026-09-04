# Implementation Plan: Manual Remote Count Reload

**Branch**: `016-tui-reload-counts` | **Date**: 2026-09-04 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/016-tui-reload-counts/spec.md`

## Summary

Add an `r` keybinding on the main tree and sync root picker that schedules the existing `remoteUpdateCmd` → `git.FetchDefaultRemote` → `remoteUpdateMsg` → count refresh path. No new fetch primitive, no loading UI, no CLI flag. Sync edge confirm also handles `r` (no help line there). Git single-flight coalesces overlapping manual and automatic fetches; the TUI always schedules `remoteUpdateCmd` on `r`.

Grilling (2026-09-04 specify): tree + sync surfaces; coalesce in-flight; silent feedback.
Grilling (2026-09-04 plan): `r: reload` on main tree footer + sync picker help; always schedule cmd on `r`.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `internal/git.FetchDefaultRemote` (007), `internal/tui/remote.go` cmd/msg, inbound/outbound count loaders (006/008), Bubble Tea `tea.Cmd`

**Storage**: N/A (updates remote-tracking refs only; no branchy schema)

**Testing**: `go test ./internal/tui/... -count=1` — key binding schedules cmd; injected `remoteUpdateMsg` still refreshes tree + sync; footer/picker help strings; `r` during in-flight does not block keys. Reuse `internal/git/fetch_test.go` for fetch coalescing (unchanged).

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: `r` returns within one frame (non-blocking `tea.Cmd`); no added network wait on startup

**Constraints**: Silent in progress and on failure (FR-006); one fetch per dir at a time via git layer (FR-011); refresh only matching project id (FR-007); MR/link/unlink/projects/CLI out of scope; preserve direction mode and cursor on refresh

**Scale/Scope**: ~4 touched files (`app.go`, `syncflow.go`, `inbound.go`, tests); no new packages; no `internal/git` changes expected

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–015 conventions |
| Library-first / separation | ✅ Pass | Fetch stays in `git`; TUI only binds a key to existing `remoteUpdateCmd` |
| CLI interface | ✅ Pass | No new flags; scripted paths unchanged |
| Test coverage | ✅ Pass | Key → cmd wiring + existing msg-driven refresh tests extended |
| Simplicity / YAGNI | ✅ Pass | Reuse 007 pipeline; no TUI in-flight tracker; no spinner |
| Consistency | ✅ Pass | Same `remoteUpdateMsg` apply path as auto-fetch; sync confirm help unchanged |

**Post-design re-check**: All gates pass. Feature is a thin keybinding layer over 007 infrastructure.

## Project Structure

### Documentation (this feature)

```text
specs/016-tui-reload-counts/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── tui-reload-counts.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/tui/
├── remote.go            # Unchanged: remoteUpdateCmd / remoteUpdateMsg
├── app.go               # Add keys.Reload; handle r in updateTree → remoteUpdateCmd
├── app_test.go          # r schedules cmd; footer lists r: reload
├── inbound.go           # treeHelpFooter: add r: reload
├── inbound_test.go      # Footer string includes r
├── syncflow.go          # Add Reload to syncKeys; handle r in updatePickRoot + updateEdgeConfirm
└── syncflow_test.go     # r on picker/confirm schedules cmd; picker help lists r

internal/git/
└── fetch.go             # Unchanged (single-flight already satisfies FR-011)
```

**Structure Decision**: Single Go module. No new files required unless tests grow large enough to split — prefer extending existing test files.

## Complexity Tracking

> No constitution violations requiring justification.

## Implementation Notes

### Phase ordering (for tasks.md)

1. **Keybinding**: `keys.Reload` / `syncKeys.Reload` with `r`
2. **Main tree**: `updateTree` returns `remoteUpdateCmd(m.current)`; update `treeHelpFooter`
3. **Sync**: `updatePickRoot` + `updateEdgeConfirm` return `remoteUpdateCmd(m.project)`; update `syncPickerHelp`
4. **Tests**: cmd scheduling, help strings, confirm/picker refresh via injected msg (existing pattern)
5. **Verify**: `go test ./internal/tui/...`; quickstart manual scenarios

### Key patterns

**Manual reload** (same as auto-fetch, user-triggered):

```go
if key.Matches(msg, keys.Reload) {
    return m, remoteUpdateCmd(m.current)
}
```

**Apply path** (unchanged from 007):

```go
case remoteUpdateMsg:
    return m.applyRemoteUpdate(ru), nil  // tree + embedded sync
```

**Coalescing** (git layer, no TUI state):

```go
// Each remoteUpdateCmd calls FetchDefaultRemote.
// Concurrent calls for same dir share one git fetch (007 research §4).
```

### Files intentionally unchanged

- `internal/git/fetch.go` — single-flight already correct
- `internal/cli/*` — no manual reload
- `internal/tui/{mr,link,unlink,projects}flow.go` — `r` not bound
- Sync edge confirm help line — stays minimal (`esc` / `q` only)

### Verification

- `go test ./internal/tui/... -count=1`
- Manual quickstart scenarios 1–4
