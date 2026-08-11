---

description: "Task list for Interactive Sync Confirmations feature implementation"
---

# Tasks: Interactive Sync Confirmations

**Input**: Design documents from `/specs/002-interactive-sync-confirm/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/sync-cli.md, quickstart.md

**Tests**: Unit tests included for `internal/sync` URL batch helpers and TUI step transitions per plan.md quality gates. No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/sync/sync.go`
- `internal/sync/sync_test.go`
- `internal/tui/syncflow.go`
- `internal/tui/syncflow_test.go`
- `internal/tui/app.go`
- `internal/cli/root.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Scaffold new sync flow model and test file

- [x] T001 Create `internal/tui/syncflow.go` with `SyncFlowModel`, step enum (`stepSyncEdgeConfirm`, `stepSyncProcessing`, `stepSyncSummary`, `stepSyncBrowser`, `stepSyncEmpty`), and `SyncFlowOptions` type stubs
- [x] T002 [P] Create `internal/sync/sync_test.go` test package scaffold for URL batch tests

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared sync service changes used by TUI and CLI — remove auto-browser, add single-edge API and browser batch helpers

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Remove auto-browser loop and `OnStatus` browser messaging from `sync.Run` in `internal/sync/sync.go`
- [x] T004 Export `RunEdge(p *project.Project, edge tree.Edge) Result` wrapping `processEdge` without confirm callback in `internal/sync/sync.go`
- [x] T005 Implement `CreatedURLs(summary *Summary) []string` filtering `Action == "created"` in DFS `Results` order in `internal/sync/sync.go`
- [x] T006 Implement `OpenURLs(urls []string) string` with sequential `browser.Open` and 50ms inter-tab delay in `internal/sync/sync.go`
- [x] T007 [P] Add unit tests for `CreatedURLs` inclusion/exclusion/order and `OpenURLs` warning behavior in `internal/sync/sync_test.go`

**Checkpoint**: `internal/sync` ready — TUI and CLI callers own browser decisions

---

## Phase 3: User Story 1 - Per-edge sync confirmation in TUI (Priority: P1) 🎯 MVP

**Goal**: Press `s` in tree view → immediate per-edge `Create MR parent → child? [y/N]` prompts; decline skips edge and continues; summary at end; no bulk confirm

**Independent Test**: Open branchy TUI, select branch with ≥3 child edges, press `s`, decline one edge and accept others; verify summary shows correct created/skipped per edge. No bulk "confirm entire sync" screen.

### Implementation for User Story 1

- [x] T008 [US1] Implement `newSyncFlowModel` collecting `tree.CollectEdges(from)` and auth check in `internal/tui/syncflow.go`
- [x] T009 [US1] Implement `stepSyncEdgeConfirm` view with `Create MR <parent> → <child>? [y/N]` and `(edge X of N)` counter in `internal/tui/syncflow.go`
- [x] T010 [US1] Implement `y`/Enter handler dispatching async `sync.RunEdge` via `tea.Cmd` and `stepSyncProcessing` in `internal/tui/syncflow.go`
- [x] T011 [US1] Implement `n` decline handler appending skipped result (`skipped by user`) and advancing to next edge in `internal/tui/syncflow.go`
- [x] T012 [US1] Implement `Esc` cancel handler stopping remaining edges and transitioning to partial summary in `internal/tui/syncflow.go`
- [x] T013 [US1] Implement `stepSyncSummary` per-edge results display (created/skipped/failed + URLs) in `internal/tui/syncflow.go`
- [x] T014 [US1] Implement `stepSyncEmpty` message `No child branches below "<from>"` with return to tree in `internal/tui/syncflow.go`
- [x] T015 [US1] Replace bulk `updateSync` logic with embedded `SyncFlowModel` delegated Update/View when `screen == screenSync` in `internal/tui/app.go`
- [x] T016 [US1] Wire `s` key to initialize `SyncFlowModel` with `syncFrom` from selected branch (no bulk confirm state) in `internal/tui/app.go`
- [x] T017 [P] [US1] Add step transition tests (edge confirm → processing → next edge → summary, decline, cancel) in `internal/tui/syncflow_test.go`

**Checkpoint**: TUI per-edge sync works — core MVP deliverable

---

## Phase 4: User Story 2 - Optional browser open at end of sync (Priority: P1)

**Goal**: After sync completes, prompt `Open created MRs in browser? [y/N]`; open created-only URLs in DFS order; skip prompt when zero creates

**Independent Test**: Complete TUI sync creating two MRs; answer `y` at browser prompt; verify tabs open in DFS order. Decline all edges → no browser prompt.

### Implementation for User Story 2

- [x] T018 [US2] Implement `stepSyncBrowser` view with `Open created MRs in browser? [y/N]` in `internal/tui/syncflow.go`
- [x] T019 [US2] Transition from `stepSyncSummary` to `stepSyncBrowser` only when `len(sync.CreatedURLs(summary)) > 0` in `internal/tui/syncflow.go`
- [x] T020 [US2] On browser confirm call `sync.OpenURLs(sync.CreatedURLs(summary))` in `internal/tui/syncflow.go`
- [x] T021 [US2] Display non-fatal browser open warning from `OpenURLs` return value in `internal/tui/syncflow.go`
- [x] T022 [P] [US2] Add browser step transition tests (skip when no creates, y opens, n skips) in `internal/tui/syncflow_test.go`

**Checkpoint**: Full P1 TUI sync complete with per-edge confirm + optional ordered browser batch

---

## Phase 5: User Story 3 - Unified sync behavior in CLI (Priority: P2)

**Goal**: `branchy sync` stops auto-opening tabs; adds end stdin browser prompt; `-y` skips per-edge only

**Independent Test**: Run `branchy sync --from develop`, confirm edges, answer `y` at end browser prompt. Run with `-y` — no per-edge prompts but end browser prompt still appears if MRs created.

### Implementation for User Story 3

- [x] T023 [US3] Remove `OnStatus` browser status callback from `sync.Run` call in `internal/cli/root.go`
- [x] T024 [US3] Add stdin prompt `Open created MRs in browser? [y/N]` after sync loop when `len(sync.CreatedURLs(summary)) > 0` in `internal/cli/root.go`
- [x] T025 [US3] Call `sync.OpenURLs(sync.CreatedURLs(summary))` on yes and print warning if returned in `internal/cli/root.go`
- [x] T026 [US3] Verify `-y` flag only bypasses per-edge `Confirm` callback, not end browser prompt in `internal/cli/root.go`

**Checkpoint**: TUI and CLI share identical browser-batch semantics

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, regression checks, and full validation

- [x] T027 [P] Update `README.md` sync behavior section (per-edge TUI, end browser prompt, breaking change from auto-open)
- [x] T028 Run all scenarios in `specs/002-interactive-sync-confirm/quickstart.md` and fix any gaps found
- [x] T029 [P] Verify `branchy mr` / `m` shortcut flow unchanged in `internal/tui/mrflow.go` and `internal/tui/app.go`
- [x] T030 Run `go test ./...` and ensure all packages pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs `RunEdge`)
- **User Story 2 (Phase 4)**: Depends on Phase 3 + T005/T006 (needs `SyncFlowModel` summary step and URL helpers)
- **User Story 3 (Phase 5)**: Depends on Phase 2 only — can run in parallel with Phase 3/4 (different file: `root.go` vs `syncflow.go`)
- **Polish (Phase 6)**: Depends on Phases 3–5

### User Story Dependencies

- **US1 (P1)**: Starts after Foundational — no dependency on US2/US3
- **US2 (P1)**: Starts after US1 checkpoint (browser step extends `SyncFlowModel`)
- **US3 (P2)**: Starts after Foundational — independently testable via CLI; parallel with US1/US2 after Phase 2

### Within Each User Story

- Foundational sync API before any caller changes
- TUI per-edge loop before browser step
- CLI changes isolated to `root.go`

### Parallel Opportunities

- **Phase 1**: T001 and T002 in parallel
- **Phase 2**: T007 parallel with T003–T006 (after T005–T006 defined)
- **Phase 3**: T017 parallel once T008–T016 complete
- **Phase 4**: T022 parallel once T018–T021 complete
- **Cross-story**: US3 (T023–T026) parallel with US1/US2 after Phase 2 (different files)

---

## Parallel Example: User Story 1

```bash
# After T016 completes, run in parallel:
# Task T017: "Add step transition tests in internal/tui/syncflow_test.go"
# Manual validation: quickstart Scenario 1 (per-edge confirm, no bulk screen)
```

## Parallel Example: After Phase 2 Checkpoint

```bash
# Developer A — User Story 1 + 2 (T008–T022 in internal/tui/syncflow.go, app.go)

# Developer B — User Story 3 (T023–T026 in internal/cli/root.go)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T002)
2. Complete Phase 2: Foundational (T003–T007)
3. Complete Phase 3: User Story 1 (T008–T017)
4. **STOP and VALIDATE**: quickstart Scenario 1 (per-edge confirm, selective skip)
5. Demo if ready

### Incremental Delivery

1. Setup + Foundational → shared sync API without auto-browser
2. Add US1 → TUI per-edge sync → **MVP**
3. Add US2 → end browser prompt in TUI (completes P1)
4. Add US3 → CLI unified browser behavior
5. Polish → docs + quickstart validation

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Split after Phase 2:
   - Developer A: US1 + US2 (`syncflow.go`, `app.go`)
   - Developer B: US3 (`root.go`)
3. Rejoin for Polish phase

---

## Notes

- `[P]` tasks touch different files with no incomplete-task dependencies
- `[USn]` labels map to spec.md user stories for traceability
- `sync.Run` no longer opens browser — TUI and CLI callers use `CreatedURLs` + `OpenURLs`
- Breaking change: sync auto-browser removed intentionally per spec FR-004/FR-007
- `internal/mr` and manual MR flow (`m` / `branchy mr`) are out of scope — verify unchanged in T029
- Commit after each task or logical group; stop at any checkpoint to validate independently
