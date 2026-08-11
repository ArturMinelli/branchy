---

description: "Task list for Lazy Auto-Fetch feature implementation"
---

# Tasks: Lazy Auto-Fetch

**Input**: Design documents from `/specs/007-lazy-auto-fetch/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/lazy-fetch.md, quickstart.md

**Tests**: Included per plan.md quality gates (`internal/git/fetch_test.go`, `internal/tui/remote_test.go`, extended `app_test.go` / `syncflow_test.go`). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/git/fetch.go`, `internal/git/fetch_test.go`
- `internal/tui/remote.go`, `internal/tui/remote_test.go`, `internal/tui/app.go`, `internal/tui/syncflow.go`
- `internal/cli/*` and `internal/sync/sync.go` stay unchanged (FR-008)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before implementation

- [x] T001 Verify implementation targets in `specs/007-lazy-auto-fetch/plan.md` — new `internal/git/fetch.go` and `internal/tui/remote.go`; modify `internal/tui/app.go` and `internal/tui/syncflow.go` only; no fetch calls in `internal/cli/` or `internal/sync/sync.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Default-remote fetch with single-flight dedupe, plus the shared TUI cmd/msg used by every story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Implement `DefaultRemote(dir) (name string, ok bool)` and `FetchDefaultRemote(dir) error` in `internal/git/fetch.go` — prefer `origin`, else first remote, skip with nil error when none; run `git fetch <remote>`; share one in-flight fetch per absolute path per `specs/007-lazy-auto-fetch/contracts/lazy-fetch.md`
- [x] T003 [P] Add temp-remote tests in `internal/git/fetch_test.go`: successful fetch updates tracking refs; no remotes → nil error; unreachable remote → error; two concurrent calls share one fetch
- [x] T004 Implement `remoteUpdateMsg` and `remoteUpdateCmd(p *project.Project)` in `internal/tui/remote.go` — cmd calls `git.FetchDefaultRemote` and returns `{projectID, path, err}`
- [x] T005 Add cmd/msg tests in `internal/tui/remote_test.go` that invoke the cmd against a temp repo with no remotes and assert a `remoteUpdateMsg` with nil error

**Checkpoint**: Fetch + TUI cmd ready — tree and sync can schedule updates independently

---

## Phase 3: User Story 1 - Open the tree on local data instantly (Priority: P1) 🎯 MVP

**Goal**: Main tree shows local inbound counts immediately, then refreshes in place after a successful background default-remote update; keys stay usable

**Independent Test**: Open the main TUI on a project with a remote-only tree child; tree appears immediately (possibly `?`); after the update that child shows a number without restarting (`specs/007-lazy-auto-fetch/quickstart.md` Scenario 1)

### Implementation for User Story 1

- [x] T006 [US1] Paint local inbound counts in `selectProject` then return `remoteUpdateCmd` (do not fetch synchronously) in `internal/tui/app.go`
- [x] T007 [US1] Return `remoteUpdateCmd` from `Model.Init` when a project is already selected in `internal/tui/app.go` so preselected/single-project startup still fetches after first paint
- [x] T008 [US1] Handle `remoteUpdateMsg` in `Model.Update` in `internal/tui/app.go`: ignore other project ids; on nil error recompute `loadInboundCounts` and `treeView.setInbound`; on error leave counts unchanged; do not block keys
- [x] T009 [P] [US1] Add tests in `internal/tui/app_test.go`: tree has local counts before any msg; success msg for the current project refreshes inbound; msg for a different project id is ignored

**Checkpoint**: Main tree is a usable MVP — instant local paint + in-place refresh

---

## Phase 4: User Story 2 - Sync sees the same refreshed counts (Priority: P1)

**Goal**: Sync picker and edge confirm show local counts immediately, join the in-flight project fetch, and refresh badges/confirm in place

**Independent Test**: Open sync before or after the tree fetch; picker/confirm appear immediately; when the update completes, badges and the confirm line refresh; no second overlapping fetch (`quickstart.md` Scenario 2)

### Implementation for User Story 2

- [x] T010 [US2] Add `applyInbound` on `SyncFlowModel` in `internal/tui/syncflow.go` to replace the count map, rebuild picker badges when on `stepSyncPickRoot`, and rewrite confirm context when on `stepSyncEdgeConfirm`
- [x] T011 [US2] Start `remoteUpdateCmd` from `SyncFlowModel.Init` in `internal/tui/syncflow.go` (standalone sync) and handle `remoteUpdateMsg` in `SyncFlowModel.Update` via `applyInbound` on success
- [x] T012 [US2] Intercept `remoteUpdateMsg` in `internal/tui/app.go` **before** delegating to `syncFlow` when `screen == screenSync`, apply tree refresh and `syncFlow.applyInbound` so embedded sync shares the tree’s in-flight fetch
- [x] T013 [P] [US2] Add tests in `internal/tui/syncflow_test.go`: success msg refreshes picker badge and confirm “files would change” line; keys still work while waiting for the msg

**Checkpoint**: Tree and sync share one fetch and the same refreshed numbers (FR-003–FR-005)

---

## Phase 5: User Story 3 - Stay usable when the remote update fails (Priority: P2)

**Goal**: Failed or missing remotes keep the local snapshot; no fetch error UI; scripted CLI does not fetch

**Independent Test**: Unreachable origin; tree and sync stay on local counts with no error banner; `?` remains `?` (`quickstart.md` Scenario 3–4)

### Implementation for User Story 3

- [x] T014 [US3] Confirm failure paths in `internal/tui/app.go` and `internal/tui/syncflow.go` never write fetch-error strings into `errMsg` or View
- [x] T015 [P] [US3] Add tests in `internal/tui/app_test.go` and `internal/tui/syncflow_test.go`: `remoteUpdateMsg` with a non-nil error leaves inbound counts unchanged and View contains no fetch-failure text
- [x] T016 [P] [US3] Confirm scripted paths stay fetch-free: no `FetchDefaultRemote` / `remoteUpdateCmd` imports in `internal/cli/` or `internal/sync/sync.go`

**Checkpoint**: Offline/auth failure is silent and non-blocking (FR-006–FR-008)

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [x] T017 Run `go test ./internal/git/... ./internal/tui/... -count=1` and fix any failures
- [x] T018 [P] Run quickstart validation per `specs/007-lazy-auto-fetch/quickstart.md` Scenarios 1–4
- [x] T019 [P] Grep audit: `internal/cli/` and `internal/sync/sync.go` have no auto-fetch; tree `Init`/`selectProject` still call `loadInboundCounts` before returning the fetch cmd

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — no dependency on US2/US3
- **User Story 2 (Phase 4)**: Depends on Foundational + US1 msg handling in `app.go` (T008) so embedded intercept can land cleanly
- **User Story 3 (Phase 5)**: Depends on US1–US2 surfaces so silence-on-failure can be asserted
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (tree) | Foundational | Phase 2 complete |
| US2 (sync) | Foundational + T008 | T008 |
| US3 (failure) | US1–US2 wiring | T009, T013 |

### Within Each User Story

- Fetch helper before TUI cmd
- TUI cmd before surface wiring
- Surface implementation before that surface’s tests
- Do not call `git fetch` inside `selectProject` / first paint

### Parallel Opportunities

- **Phase 2**: T003 ∥ T004 after T002
- **Phase 3**: T009 after T008
- **Phase 4**: T013 after T010–T012
- **Phase 5**: T015 ∥ T016
- **Phase 6**: T018 ∥ T019

---

## Parallel Example: User Story 1

```bash
# After T006–T008 (app wiring):
Task T009: app_test.go local paint + refresh + foreign-project ignore
```

---

## Parallel Example: User Story 3

```bash
Task T015: app_test.go + syncflow_test.go silent failure
Task T016: grep cli/sync for fetch
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 — tree local paint + refresh
4. **STOP and VALIDATE**: Quickstart Scenario 1; `go test ./internal/git/... ./internal/tui/...`
5. Demo `?` becoming a number after background fetch

### Incremental Delivery

1. Setup + Foundational → `FetchDefaultRemote` + cmd ready
2. US1 → main tree refresh (MVP)
3. US2 → sync join + in-place picker/confirm refresh
4. US3 → silent failure + CLI guard
5. Polish → full quickstart + test suite

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 `app.go`
   - Developer B: US2 `syncflow.go` (`applyInbound`) until T012 needs App intercept
3. Merge, then US3 tests

---

## Notes

- First paint is always local `loadInboundCounts` — never `git fetch` on the Update/Init stack before View
- App must handle `remoteUpdateMsg` while embedded sync is open or the result is dropped
- `FetchDefaultRemote` single-flight is what makes tree + sync share one network call
- No spinner, no status line, no fetch-error `errMsg`
- Commit after each task or logical group; stop at any checkpoint to validate the story independently
