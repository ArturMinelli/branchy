# Implementation Plan: Lazy Auto-Fetch

**Branch**: `007-lazy-auto-fetch` | **Date**: 2026-08-11 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/007-lazy-auto-fetch/spec.md`

## Summary

Paint the main tree and sync TUI from local refs immediately, then run a background update of the default remote (`git fetch`). On success, recompute inbound counts and refresh the visible tree, picker, and edge confirm in place. On failure or missing remote, keep the local snapshot with no error UI. At most one in-flight fetch per project directory; scripted CLI does not fetch.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `os/exec` git seam; Bubble Tea `tea.Cmd` for background work (same pattern as sync MR `runEdgeCmd`); inbound count helpers from 006

**Storage**: N/A (updates remote-tracking refs in the user’s git repo only; no branchy schema)

**Testing**: `go test ./internal/git/... ./internal/tui/...` — temp repos with a file:// remote for fetch success/fail/no-remote; TUI tests inject a completed update msg and assert in-place refresh / no error text

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: First paint of tree + local inbound counts must not wait on the network (SC-001). Fetch runs after the model is constructed and the first frame can render.

**Constraints**: Do not block keys during fetch (FR-007); do not show fetch errors (FR-006); one in-flight fetch per project path (FR-004); refresh only the project that requested it (FR-009); scripted CLI unchanged (FR-008)

**Scale/Scope**: 1 new git fetch helper + tests; App `Init`/`selectProject` and SyncFlow `Init`/`Update` wiring; no MR/link/unlink/init/projects fetch; no CLI fetch

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–006 conventions |
| Library-first / separation | ✅ Pass | `git.FetchDefaultRemote` owns remotes + `git fetch` + in-flight dedupe. TUI only schedules a cmd and applies inbound refresh. |
| CLI interface | ✅ Pass | Scripted paths never call fetch |
| Test coverage | ✅ Pass | Temp remote repos for fetch; msg-driven TUI tests for refresh and silence-on-failure |
| Simplicity / YAGNI | ✅ Pass | No spinner, no status line, no per-branch fetch, no polling |
| Consistency | ✅ Pass | Async via `tea.Cmd` like existing sync/MR work; reuse `loadInboundCounts` |

**Post-design re-check**: All gates pass. Deep module is `FetchDefaultRemote` (hides remote name + single-flight). TUI refresh is a thin apply of existing inbound formatters.

## Project Structure

### Documentation (this feature)

```text
specs/007-lazy-auto-fetch/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── lazy-fetch.md    # Phase 1 output
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/git/
├── fetch.go             # New: DefaultRemote, FetchDefaultRemote (single-flight per dir)
├── fetch_test.go        # New: origin fetch, no remote, failure, concurrent share
├── ref.go               # Unchanged (used after fetch to resolve origin/branch)
└── inbound.go           # Unchanged

internal/tui/
├── remote.go            # New: remoteUpdateCmd, remoteUpdateMsg
├── remote_test.go       # New: cmd produces msg; failure is not an error string for View
├── app.go               # Init + selectProject start fetch; handle msg; refresh tree inbound
├── app_test.go          # Extend: local counts before msg; refresh on success; ignore other project
├── syncflow.go          # Init starts fetch if needed; handle msg; rebuild picker/confirm inbound
└── syncflow_test.go     # Extend: picker/confirm refresh on success; no error text on failure
```

**Structure Decision**: Single Go module. Fetch lives next to `ResolveRef` in `internal/git`. TUI scheduling is one small file so App and SyncFlow share the same cmd/msg types. No new package.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Implementation Notes

### Phase ordering (for tasks.md)

1. **Foundation**: `DefaultRemote` + `FetchDefaultRemote` with per-directory single-flight + temp-remote tests
2. **TUI cmd**: `remoteUpdateCmd` / `remoteUpdateMsg` (project id + path + err)
3. **Main tree**: `Init`/`selectProject` paint local then return cmd; on success refresh `treeView` inbound; drop msg if project changed
4. **Sync**: same cmd; rebuild picker badges / confirm context; App intercepts msg while embedded sync is open so one flight is shared
5. **Failure / no remote**: no View text; counts unchanged
6. **Verify**: `go test ./internal/git/... ./internal/tui/...`; quickstart

### Key patterns

**Single-flight fetch** (in `internal/git`, not the TUI):

```go
// Concurrent FetchDefaultRemote(dir) share one git fetch.
// No default remote → nil error (skip).
// git fetch failure → returned error (TUI ignores it).
```

**Non-blocking schedule**:

```go
// selectProject / SyncFlow.Init:
loadInboundCounts(...)           // local, synchronous, first paint
return remoteUpdateCmd(project)  // tea.Cmd → git.FetchDefaultRemote
```

**Apply only if still the same project**:

```go
case remoteUpdateMsg:
    if msg.projectID != current.ID { return } // FR-009
    if msg.err != nil { return }              // FR-006 silent
    counts := loadInboundCounts(...)
    treeView.setInbound(counts)
    // if screenSync: syncFlow.applyInbound(counts)
```

**App must intercept `remoteUpdateMsg` before delegating to `syncFlow`** when embedded sync is active, then forward or apply into the flow. Otherwise the result is dropped while `screen == screenSync`.

### Files intentionally unchanged

- `internal/cli/*`, `internal/sync/sync.go` — no auto-fetch in scripted mode
- `internal/tui/{mr,link,unlink,init,projects}flow.go` — do not trigger fetch
- Inbound display rules (006) — formatters reused as-is

### Verification

- `go test ./internal/git/... ./internal/tui/...`
- Manual quickstart scenarios 1–4
- Opening the TUI while offline still shows the tree immediately
