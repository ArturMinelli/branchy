---

description: "Task list for Manual Remote Count Reload feature implementation"
---

# Tasks: Manual Remote Count Reload

**Input**: Design documents from `/specs/016-tui-reload-counts/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tui-reload-counts.md, quickstart.md

**Tests**: Included per plan.md testing strategy (key wiring, help strings, non-blocking behavior). Fetch coalescing covered by existing `internal/git/fetch_test.go`.

**Organization**: Tasks grouped by user story. Reuses existing `remoteUpdateCmd` / `remoteUpdateMsg` / `applyRemoteUpdate` pipeline from 007 — no `internal/git` changes expected.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single Go module** at repository root
- TUI code: `internal/tui/`
- Git fetch (unchanged): `internal/git/fetch.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm baseline before adding keybindings

- [x] T001 Run `go test ./internal/tui/... -count=1` and confirm all tests pass before feature work

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared key definitions used by main tree and sync surfaces

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Add `Reload` field to `keyMap` with `key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reload"))` in `internal/tui/app.go`
- [x] T003 [P] Add `Reload` field to `syncKeyMap` with `key.NewBinding(key.WithKeys("r"))` in `internal/tui/syncflow.go`

**Checkpoint**: `keys.Reload` and `syncKeys.Reload` defined — user story phases can begin

---

## Phase 3: User Story 1 - Refresh stale counts on the main tree (Priority: P1) 🎯 MVP

**Goal**: Press `r` on the main tree to schedule a default-remote update and refresh inbound/outbound file-change badges in place without changing direction or blocking input.

**Independent Test**: Open main tree with stale counts, press `r`, verify footer lists `r: reload`, keyboard stays responsive, and after `remoteUpdateMsg` counts recompute (existing `applyRemoteUpdate` path).

### Implementation for User Story 1

- [x] T004 [US1] Handle `keys.Reload` in `updateTree`: when `m.current != nil`, return `m, remoteUpdateCmd(m.current)` in `internal/tui/app.go`
- [x] T005 [US1] Add `r: reload` to the first help line in `treeHelpFooter` in `internal/tui/inbound.go` (per contracts/tui-reload-counts.md)

### Tests for User Story 1

- [x] T006 [P] [US1] Add `TestUpdateTreeReloadSchedulesRemoteUpdateCmd` in `internal/tui/app_test.go` — `updateTree` with `r` returns non-nil cmd; invoking cmd yields `remoteUpdateMsg` with current project id
- [x] T007 [P] [US1] Extend `TestTreeHelpFooter` in `internal/tui/inbound_test.go` to assert both inbound and outbound footer strings contain `r: reload`

**Checkpoint**: Main tree `r` binding works; footer discoverable; existing `TestRemoteUpdateRefreshesCurrentProject` and `TestRemoteUpdateKeepsDirection` cover refresh semantics

---

## Phase 4: User Story 2 - Refresh counts during sync (Priority: P1)

**Goal**: Press `r` on sync root picker and edge confirm to schedule the same remote update; picker badges and confirm line refresh in place on success.

**Independent Test**: Open embedded sync with stale picker or confirm counts, press `r`, verify picker help lists `r: reload`, badges/confirm update after `remoteUpdateMsg` without leaving sync.

### Implementation for User Story 2

- [x] T008 [US2] Handle `syncKeys.Reload` in `updatePickRoot`: return `m, remoteUpdateCmd(m.project)` when `m.project != nil` in `internal/tui/syncflow.go`
- [x] T009 [US2] Handle `syncKeys.Reload` in `updateEdgeConfirm` (before yes/no handling): return `m, remoteUpdateCmd(m.project)` when `m.project != nil` in `internal/tui/syncflow.go`
- [x] T010 [US2] Append `r: reload` to `syncPickerHelp` first line in `internal/tui/syncflow.go`; leave edge confirm help unchanged (`esc` / `q` only)

### Tests for User Story 2

- [x] T011 [P] [US2] Add `TestSyncPickRootReloadSchedulesCmd` in `internal/tui/syncflow_test.go` — `updatePickRoot` with `r` returns cmd producing `remoteUpdateMsg`
- [x] T012 [P] [US2] Add `TestSyncEdgeConfirmReloadSchedulesCmd` in `internal/tui/syncflow_test.go` — `updateEdgeConfirm` with `r` returns cmd producing `remoteUpdateMsg` and does not confirm/skip edge
- [x] T013 [P] [US2] Add test in `internal/tui/syncflow_test.go` asserting `syncPickerHelp` contains `r: reload` for both directions

**Checkpoint**: Sync picker and confirm support `r`; existing `TestSyncFlowApplyInboundRefreshesPickerAndConfirm` covers refresh after msg

---

## Phase 5: User Story 3 - Stay usable when reload fails (Priority: P2)

**Goal**: Failed or no-remote reload keeps prior counts, shows no error UI, and does not block navigation or sync keys.

**Independent Test**: Inject `remoteUpdateMsg` with error (or press `r` with unreachable remote); counts unchanged, no fetch banner, `j`/`k`/`y`/`n` still work.

### Tests for User Story 3

- [x] T014 [US3] Add `TestUpdateTreeReloadDoesNotBlockNavigation` in `internal/tui/app_test.go` — after `r` schedules cmd, a subsequent `j`/`k` key still moves cursor and does not require fetch completion
- [x] T015 [P] [US3] Add `TestSyncEdgeConfirmReloadDoesNotBlockConfirm` in `internal/tui/syncflow_test.go` — after `r` schedules cmd on confirm step, `n` still skips edge (pattern from `TestSyncFlowKeysWorkDuringRemoteUpdate`)
- [x] T016 [US3] Verify existing silent-failure tests still pass unchanged: `TestRemoteUpdateFailureKeepsSnapshot` in `internal/tui/app_test.go` and `TestSyncFlowRemoteUpdateFailureKeepsSnapshot` in `internal/tui/syncflow_test.go`

**Checkpoint**: Manual reload uses same silent apply path as auto-fetch; no new error UI introduced

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Full verification and manual validation

- [x] T017 Run `go test ./internal/tui/... -count=1` and fix any failures from this feature
- [x] T018 [P] Execute manual quickstart scenarios 1–5 from `specs/016-tui-reload-counts/quickstart.md` (tree reload, sync picker, edge confirm, offline silence, auto-fetch coalesce)
- [x] T019 [P] Confirm `internal/git/fetch_test.go` `TestFetchDefaultRemoteConcurrentShare` still passes (FR-011 coalescing — no git changes expected)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — **MVP; no dependency on US2/US3**
- **User Story 2 (Phase 4)**: Depends on Foundational — independent of US1 (can parallel after Phase 2)
- **User Story 3 (Phase 5)**: Depends on US1 + US2 key wiring being in place (tests exercise `r` handlers)
- **Polish (Phase 6)**: Depends on desired user stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 2 — delivers MVP main-tree reload
- **User Story 2 (P1)**: Can start after Phase 2 — parallel with US1 (different files: `syncflow.go` vs `app.go`/`inbound.go`)
- **User Story 3 (P2)**: After US1 + US2 handlers exist — mostly test verification of existing apply path

### Within Each User Story

- Implementation before story-specific tests
- Key handler before help text (help is cosmetic; either order within same file is fine)

### Parallel Opportunities

- **Phase 2**: T002 (`app.go`) and T003 (`syncflow.go`) in parallel
- **US1 tests**: T006 and T007 in parallel
- **US2 tests**: T011, T012, T013 in parallel
- **US3 tests**: T014 and T015 in parallel
- **After Phase 2**: US1 and US2 implementation can proceed in parallel across developers
- **Polish**: T018 and T019 in parallel with T017 after code complete

---

## Parallel Example: User Story 1

```bash
# After T004–T005 land, run tests in parallel:
Task T006: "TestUpdateTreeReloadSchedulesRemoteUpdateCmd in internal/tui/app_test.go"
Task T007: "Extend TestTreeHelpFooter in internal/tui/inbound_test.go"
```

## Parallel Example: User Story 2

```bash
# Implementation sequence T008 → T009 → T010 (same file)
# Then parallel tests:
Task T011: "TestSyncPickRootReloadSchedulesCmd"
Task T012: "TestSyncEdgeConfirmReloadSchedulesCmd"
Task T013: "syncPickerHelp contains r: reload"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002–T003)
3. Complete Phase 3: User Story 1 (T004–T007)
4. **STOP and VALIDATE**: `go test ./internal/tui/...`; manual quickstart scenario 1
5. Demo main-tree reload

### Incremental Delivery

1. Setup + Foundational → key constants ready
2. User Story 1 → main tree `r` + footer (MVP)
3. User Story 2 → sync picker + confirm `r`
4. User Story 3 → non-blocking / silent-failure test coverage
5. Polish → full suite + quickstart

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (`app.go`, `inbound.go`, tests)
   - Developer B: US2 (`syncflow.go`, tests)
3. Developer C or A+B: US3 verification tests
4. Anyone: Polish phase

---

## Notes

- **No new files required** unless tests grow unwieldy — extend existing files per plan.md
- **Do not modify**: `internal/git/fetch.go`, `internal/tui/remote.go`, `internal/cli/*`, MR/link/unlink/projects flows
- **Do not add**: spinner, status line, error banner, CLI flag, or TUI in-flight tracker
- `remoteUpdateMsg` apply path is unchanged — manual `r` and auto-fetch share identical refresh logic
- Edge confirm intentionally has no `r` help line (minimal chrome per research.md §3)
- Commit after each task or logical group; stop at any checkpoint to validate story independently
