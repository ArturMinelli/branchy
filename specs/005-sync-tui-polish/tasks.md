---

description: "Task list for TUI Native Confirm & Loading feature implementation"
---

# Tasks: TUI Native Confirm & Loading

**Input**: Design documents from `/specs/005-sync-tui-polish/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tui-chrome.md, quickstart.md

**Tests**: Unit and flow step tests included per plan.md quality gates (`confirm_test.go`, `loading_test.go`, extended `*_flow_test.go`). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/tui/confirm.go`, `internal/tui/loading.go`, `internal/tui/chrome.go`, `internal/tui/*flow.go`, `internal/tui/app.go`
- `internal/cli/mode_test.go` (regression only; no CLI behavior changes)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before implementation

- [x] T001 Verify implementation targets in `specs/005-sync-tui-polish/plan.md` — new `internal/tui/confirm.go` and `internal/tui/loading.go`; modify `internal/tui/*flow.go` and `internal/tui/app.go` only; no changes to `internal/sync/sync.go` or `internal/cli/root.go` behavior

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared `ConfirmModel`, `LoadingModel`, and chrome tokens used by all user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Extend `internal/tui/chrome.go` with confirm border, button focus/blur, and loading layout styles per `specs/005-sync-tui-polish/contracts/tui-chrome.md`
- [x] T003 Add `left`/`right`/`tab` key bindings for confirm navigation to `internal/tui/flowutil.go`
- [x] T004 Implement `ConfirmModel`, `ConfirmOptions`, `ConfirmChoice`, `NewConfirm`, `Update`, and `View` in `internal/tui/confirm.go`
- [x] T005 Implement `LoadingModel`, `LoadingOptions`, `NewLoading`, `Init`, `Update`, `View`, and `Active` using `bubbles/spinner` in `internal/tui/loading.go`
- [x] T006 [P] Add unit tests for confirm focus navigation, y/n/enter/esc shortcuts, and no `[y/N]` in view output in `internal/tui/confirm_test.go`
- [x] T007 [P] Add unit tests for spinner tick updates and `Active()` state in `internal/tui/loading_test.go`

**Checkpoint**: Shared confirm and loading components ready — user story wiring can begin

---

## Phase 3: User Story 1 - Native confirm dialogs across all TUI flows (Priority: P1) 🎯 MVP

**Goal**: Replace inline `[y/N]` text with shared `ConfirmModel` on every yes/no step in sync, MR, link, unlink, init, and embedded tree unlink

**Independent Test**: Run sync, MR, link, unlink, init, and tree unlink (`u`); verify framed Yes/No panel with arrow-key focus (default No); no literal `[y/N]` in question text; y/n/enter/esc behave per `specs/005-sync-tui-polish/contracts/tui-chrome.md`

### Implementation for User Story 1

- [x] T008 [US1] Replace per-edge and browser confirm text with embedded `ConfirmModel` in `internal/tui/syncflow.go`
- [x] T009 [US1] Replace create and browser confirm text with embedded `ConfirmModel` in `internal/tui/mrflow.go`
- [x] T010 [P] [US1] Replace link confirm text with embedded `ConfirmModel` in `internal/tui/linkflow.go`
- [x] T011 [P] [US1] Replace unlink confirm text with embedded `ConfirmModel` in `internal/tui/unlinkflow.go`
- [x] T012 [US1] Replace init registration confirm text with embedded `ConfirmModel` in `internal/tui/initflow.go`
- [x] T013 [US1] Replace `screenUnlinkConfirm` inline `[y/N]` with embedded `ConfirmModel` in `internal/tui/app.go`
- [x] T014 [P] [US1] Add confirm panel view assertions (no `[y/N]`, focus default No) in `internal/tui/syncflow_test.go`
- [x] T015 [P] [US1] Add confirm step transition tests in `internal/tui/mrflow_test.go`
- [x] T016 [P] [US1] Add confirm step tests in `internal/tui/linkflow_test.go`
- [x] T017 [P] [US1] Add confirm step tests in `internal/tui/unlinkflow_test.go`
- [x] T018 [P] [US1] Add confirm step tests in `internal/tui/initflow_test.go`
- [x] T019 [P] [US1] Add embedded unlink confirm tests in `internal/tui/app_test.go`

**Checkpoint**: All in-scope TUI flows use native confirm panels — MVP deliverable (loading may still be static on MR/init until US2)

---

## Phase 4: User Story 2 - Loading feedback during long-running TUI actions (Priority: P1)

**Goal**: Animated spinner with blocked input during sync MR creation, MR wizard creation, browser-open batches, and init registration

**Independent Test**: Confirm an edge in sync TUI and MR wizard; verify spinner with status message; keys ignored until completion; auto-advance to next step; browser-open and init show loading per `specs/005-sync-tui-polish/quickstart.md` Scenarios 1–2, 4, 6

### Implementation for User Story 2

- [x] T020 [US2] Wire `LoadingModel` into `stepSyncProcessing` with edge progress and input blocking in `internal/tui/syncflow.go`
- [x] T021 [US2] Add `stepSyncBrowserLoading` with async browser batch cmd and loading panel in `internal/tui/syncflow.go`
- [x] T022 [US2] Refactor MR create to async `stepMRLoading` via `tea.Cmd` and add `stepMRBrowserLoading` in `internal/tui/mrflow.go`
- [x] T023 [US2] Refactor init registration to async `stepInitLoading` via `tea.Cmd` in `internal/tui/initflow.go`
- [x] T024 [P] [US2] Add loading step tests (keys ignored, transitions on result msg) in `internal/tui/syncflow_test.go`
- [x] T025 [P] [US2] Add async MR loading step tests in `internal/tui/mrflow_test.go`
- [x] T026 [P] [US2] Add init loading step tests in `internal/tui/initflow_test.go`

**Checkpoint**: All network-bound TUI operations show animated loading with blocked input

---

## Phase 5: User Story 3 - Visual consistency via shared TUI chrome (Priority: P1)

**Goal**: Confirm and loading treatments are visually identical across all in-scope flows (border, button layout, spinner placement, help footers)

**Independent Test**: Step through confirm and loading screens in sync, MR, link, unlink, init, and tree unlink; verify matching chrome per quickstart Scenario 9 checklist

### Implementation for User Story 3

- [x] T027 [US3] Align `ConfirmModel.View` button order (No left, Yes right), focus highlight, and help footer text with contract in `internal/tui/confirm.go`
- [x] T028 [US3] Align `LoadingModel.View` spinner placement, message layout, and progress line format in `internal/tui/loading.go`
- [x] T029 [US3] Audit and normalize confirm/loading help footers across `internal/tui/syncflow.go`, `internal/tui/mrflow.go`, `internal/tui/linkflow.go`, `internal/tui/unlinkflow.go`, `internal/tui/initflow.go`, and `internal/tui/app.go`

**Checkpoint**: Visual consistency checklist passes for all six TUI surfaces (SC-007)

---

## Phase 6: User Story 4 - Scripted CLI paths unchanged (Priority: P2)

**Goal**: Flagged, positional-arg, and non-TTY invocations keep plain stdin prompts with no TUI confirm or spinner output

**Independent Test**: Run `branchy sync --from develop`, `branchy link parent child`, and piped sync; verify no full-screen TUI and no spinner per quickstart Scenario 8

### Implementation for User Story 4

- [x] T030 [US4] Verify `internal/sync/sync.go` stdin `[y/N]` prompts are untouched and document in code comment if needed at top of `internal/tui/confirm.go` that confirm/loading are TUI-only
- [x] T031 [P] [US4] Add regression test that `UseTUI` false paths do not import or invoke confirm/loading components in `internal/cli/mode_test.go`

**Checkpoint**: Scripted CLI semantics unchanged (FR-013)

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup across all stories

- [x] T032 Run full quickstart validation per `specs/005-sync-tui-polish/quickstart.md` Scenarios 1–9
- [x] T033 [P] Grep audit: confirm zero `[y/N]` literals in `internal/tui/` View strings (`rg '\[y/N\]' internal/tui/`)
- [x] T034 [P] Run `go test ./internal/tui/... ./internal/cli/... -count=1` and fix any failures

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — no dependency on US2/US3/US4
- **User Story 2 (Phase 4)**: Depends on Foundational; integrates best after US1 confirm wiring in sync/mr/init (T008, T009, T012)
- **User Story 3 (Phase 5)**: Depends on US1 + US2 complete (audits final integrated views)
- **User Story 4 (Phase 6)**: Can run in parallel with US1/US2 (verification only); finalize after US2
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (confirm) | Foundational | Phase 2 complete |
| US2 (loading) | Foundational, US1 confirm wiring in sync/mr/init | T008, T009, T012 (or full US1) |
| US3 (consistency) | US1 + US2 | Phase 4 complete |
| US4 (CLI guard) | None (verification) | Phase 2 (parallel with US1/US2) |

### Within Each User Story

- Foundational components before flow wiring
- Flow implementation before flow tests
- Sync/MR/init confirm (US1) before loading async refactor (US2) on same files

### Parallel Opportunities

- **Phase 2**: T006 ∥ T007 (separate test files)
- **Phase 3**: T010 ∥ T011; T014–T019 all parallel after T008–T013 (different test files)
- **Phase 4**: T024 ∥ T025 ∥ T026 after respective flow changes
- **Phase 6**: T031 parallel with US1/US2 work
- **Phase 7**: T033 ∥ T034

---

## Parallel Example: User Story 1

```bash
# After T008–T009 (sync + mr), wire confirm-only flows in parallel:
Task T010: linkflow.go
Task T011: unlinkflow.go

# Run all flow confirm tests together:
Task T014: syncflow_test.go
Task T015: mrflow_test.go
Task T016: linkflow_test.go
Task T017: unlinkflow_test.go
Task T018: initflow_test.go
Task T019: app_test.go
```

---

## Parallel Example: User Story 2

```bash
# After sync loading (T020–T021), MR and init async refactors touch different files:
Task T022: mrflow.go
Task T023: initflow.go

# Loading tests in parallel:
Task T024: syncflow_test.go
Task T025: mrflow_test.go
Task T026: initflow_test.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 — native confirm on all flows
4. **STOP and VALIDATE**: Quickstart confirm scenarios; grep for `[y/N]`
5. Demo improved confirm UX before loading work

### Incremental Delivery

1. Setup + Foundational → shared components ready
2. US1 → all flows have native confirm (MVP)
3. US2 → loading + async MR/init (perceived performance)
4. US3 → visual consistency audit
5. US4 → CLI regression guard
6. Polish → full quickstart + test suite

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 sync + mr + app (T008, T009, T013)
   - Developer B: US1 link + unlink + init (T010–T012)
   - Developer C: US4 verification (T030–T031) in parallel
3. Merge US1, then split US2 across sync (A) and mr/init (B)

---

## Notes

- Link and unlink get confirm only (no loading step) — fast file I/O per research.md §6
- `internal/tui/projectsflow.go` is out of scope (no yes/no confirm)
- MR `updateConfirm` synchronous call must move to `tea.Cmd` in US2 (T022) for spinner to animate
- Default confirm focus: **No** (safe default)
- Commit after each task or logical group; stop at any checkpoint to validate story independently
