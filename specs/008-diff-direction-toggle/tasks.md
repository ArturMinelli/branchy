---

description: "Task list for Diff Direction Toggle feature implementation"
---

# Tasks: Diff Direction Toggle

**Input**: Design documents from `/specs/008-diff-direction-toggle/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/diff-direction.md, quickstart.md

**Tests**: Included per plan.md quality gates (`internal/git/inbound_test.go`, `internal/tui/inbound_test.go`, `treeview_test.go`, `app_test.go`, sync isolation checks). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/git/inbound.go`, `internal/git/inbound_test.go`
- `internal/tui/inbound.go`, `internal/tui/inbound_test.go`, `internal/tui/treeview.go`, `internal/tui/treeview_test.go`, `internal/tui/app.go`, `internal/tui/app_test.go`
- `internal/tui/syncflow.go` stays inbound-only (FR-007)
- `internal/cli/*` and `internal/sync/sync.go` stay unchanged (FR-010)
- `README.md` keys line updated in polish

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before implementation

- [x] T001 Verify implementation targets in `specs/008-diff-direction-toggle/plan.md` — extend `internal/git/inbound.go` with `OutboundFiles`; dual maps + direction in `internal/tui/{inbound,treeview,app}.go`; no outbound wiring in `internal/tui/syncflow.go`, `internal/cli/`, or `internal/sync/sync.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Outbound comparison API and shared count loaders that every story uses

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Implement `OutboundFiles(dir, parent, child string) (int, error)` in `internal/git/inbound.go` as three-dot `git diff --name-only <parent>...<child>` with the same `ResolveRef` / error rules as `InboundFiles` per `specs/008-diff-direction-toggle/contracts/diff-direction.md`
- [x] T003 [P] Add temp-repo tests in `internal/git/inbound_test.go`: child-only files → N outbound; identical trees → 0; parent-only after diverge → 0 outbound; missing ref → error; `OutboundFiles(dir,p,c) == InboundFiles(dir,c,p)`
- [x] T004 Rename `inboundCount` → `fileChangeCount` (or equivalent shared type) and add `loadOutboundCounts(dir, doc)` mirroring `loadInboundCounts` via `git.OutboundFiles` in `internal/tui/inbound.go`; keep badge formatter shared (rename to `formatFileChangeBadge` if needed) and update call sites in `internal/tui/treeview.go`, `internal/tui/syncflow.go`, and related tests so they still compile
- [x] T005 [P] Add loader/formatter tests in `internal/tui/inbound_test.go`: outbound map omits roots; unknown on git error; badge still hides zero / shows `N` / shows `?`

**Checkpoint**: Outbound git + dual loaders ready — tree toggle and footer can wire independently

---

## Phase 3: User Story 1 - Flip the main tree to outbound counts (Priority: P1) 🎯 MVP

**Goal**: Main tree starts inbound; `d` flips all child badges to outbound (and back) from cached maps; zeros hidden per active direction; roots stay badge-free; unknown stays `?`

**Independent Test**: Open the main TUI on a project with known different inbound vs outbound sizes; verify default inbound badges, press `d`, verify outbound badges, press `d` again (`specs/008-diff-direction-toggle/quickstart.md` Scenario 1–2)

### Implementation for User Story 1

- [x] T006 [US1] Extend `BranchTreeView` in `internal/tui/treeview.go` to hold inbound + outbound maps and a direction field; `setFileCounts` (or paired setters) replaces `setInbound`; `renderRow` reads the active map through the shared badge formatter
- [x] T007 [US1] Add `diffDirection` (default inbound) and `d` key binding on `Model` / `keyMap` in `internal/tui/app.go`; in `updateTree`, flip direction, mirror onto `treeView`, and redraw from cache without calling git
- [x] T008 [US1] In `selectProject` and `applyRemoteUpdate` in `internal/tui/app.go`, load both `loadInboundCounts` and `loadOutboundCounts`, set both maps on the tree, and keep the current direction (do not reset direction here — US3 owns persistence rules; default remains inbound for a new Model)
- [x] T009 [P] [US1] Add tests in `internal/tui/treeview_test.go` and `internal/tui/app_test.go`: default inbound badges; after `d`, outbound badges; second `d` restores inbound; zero hide per direction; roots never badged; `?` in both modes; remote success refreshes both maps without flipping direction

**Checkpoint**: Tree toggle MVP works — flip badges without footer polish or sync checks yet

---

## Phase 4: User Story 2 - Tell which direction is active (Priority: P1)

**Goal**: Main-tree help footer names the current direction and lists `d` with the action toward the other mode

**Independent Test**: Read footer before toggle (inbound + `d: show outbound`); press `d`; footer shows outbound + `d: show inbound` (`quickstart.md` Scenario 1 footer expectations)

### Implementation for User Story 2

- [x] T010 [US2] Add `treeHelpFooter(direction)` (or equivalent) in `internal/tui/inbound.go` returning key hints including `d: show outbound|inbound` plus mode cue `counts: inbound (parent→child)` / `counts: outbound (child→parent)` per `contracts/diff-direction.md`
- [x] T011 [US2] Use `treeHelpFooter` in `Model.View` for `screenTree` in `internal/tui/app.go` instead of the hard-coded help string
- [x] T012 [P] [US2] Add footer string tests in `internal/tui/inbound_test.go` and/or `internal/tui/app_test.go`: inbound and outbound View/footer contain the correct mode cue and toggle hint

**Checkpoint**: Direction is discoverable from the footer alone (FR-005, SC-006)

---

## Phase 5: User Story 3 - Direction lasts for this session only (Priority: P2)

**Goal**: Direction survives leaving the tree (sync/MR/projects) and switching projects in the same process; a new session always starts inbound

**Independent Test**: Toggle outbound → sync cancel → still outbound; switch project → still outbound; quit/relaunch → inbound (`quickstart.md` Scenario 3)

### Implementation for User Story 3

- [x] T013 [US3] Confirm `selectProject` in `internal/tui/app.go` rebuilds the tree and reloads both maps but never resets `Model.diffDirection`; new `Model` / `newModel` still defaults to inbound
- [x] T014 [P] [US3] Add tests in `internal/tui/app_test.go`: after flipping to outbound, simulating return from sync (`screenTree` restore) and `selectProject` to another project keeps outbound; a freshly constructed `Model` starts inbound

**Checkpoint**: Session-scoped mode matches FR-006 / SC-004

---

## Phase 6: User Story 4 - Sync still shows inbound (Priority: P2)

**Goal**: With the tree in outbound mode, sync picker and edge confirm still show inbound counts only; no direction toggle on sync screens

**Independent Test**: Flip tree outbound, open sync; picker/confirm match inbound numbers, not outbound (`quickstart.md` Scenario 4)

### Implementation for User Story 4

- [x] T015 [US4] Confirm `SyncFlowModel` in `internal/tui/syncflow.go` still loads and displays only `loadInboundCounts` (picker badges + confirm line); do not thread `diffDirection` or outbound maps into sync
- [x] T016 [P] [US4] Add regression tests in `internal/tui/app_test.go` and/or `internal/tui/syncflow_test.go`: with Model direction outbound, starting sync still presents inbound badges/confirm text; sync View has no `d: show` / outbound mode cue
- [x] T017 [P] [US4] Grep audit: `internal/cli/` and `internal/sync/sync.go` have no outbound/direction helpers; syncflow has no `OutboundFiles` / `loadOutboundCounts` / `diffDirection`

**Checkpoint**: Sync isolation holds (FR-007, SC-005)

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Docs and full validation across all stories

- [x] T018 [P] Update main TUI keys documentation in `README.md` to mention `d` and inbound/outbound count direction on the tree
- [x] T019 Run `go test ./internal/git/... ./internal/tui/... -count=1` and fix any failures
- [x] T020 [P] Run quickstart validation per `specs/008-diff-direction-toggle/quickstart.md` Scenarios 1–5

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — no dependency on US2–US4
- **User Story 2 (Phase 4)**: Depends on Foundational + US1 direction field existing on Model/tree so footer can read it
- **User Story 3 (Phase 5)**: Depends on US1 toggle wiring (`diffDirection` on Model)
- **User Story 4 (Phase 6)**: Depends on US1 (outbound tree mode exists) so isolation can be asserted
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (toggle badges) | Foundational | Phase 2 complete |
| US2 (footer) | Foundational + T007 direction | T007 |
| US3 (session) | US1 Model direction | T008–T009 |
| US4 (sync isolation) | US1 outbound mode | T009 |

### Within Each User Story

- Git `OutboundFiles` before TUI loaders
- Dual maps on tree before `d` key wiring
- Toggle before footer / session / sync-isolation assertions
- Do not recompute git on the `d` key path

### Parallel Opportunities

- **Phase 2**: T003 ∥ T004 after T002; T005 after T004
- **Phase 3**: T009 after T006–T008
- **Phase 4**: T012 after T010–T011
- **Phase 5**: T014 after T013
- **Phase 6**: T016 ∥ T017 after T015
- **Phase 7**: T018 ∥ T020; T019 sequential with fixes

---

## Parallel Example: User Story 1

```bash
# After T006–T008 (tree + app wiring):
Task T009: treeview_test.go + app_test.go toggle / zero-hide / remote dual refresh
```

---

## Parallel Example: User Story 4

```bash
Task T016: app_test.go / syncflow_test.go inbound-only under outbound tree
Task T017: grep cli/sync/syncflow for outbound leakage
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 — tree dual maps + `d` toggle
4. **STOP and VALIDATE**: Quickstart Scenario 1–2; `go test ./internal/git/... ./internal/tui/...`
5. Demo flipping inbound ↔ outbound badges

### Incremental Delivery

1. Setup + Foundational → `OutboundFiles` + loaders ready
2. US1 → toggle works (MVP)
3. US2 → footer makes mode obvious
4. US3 → session persistence verified
5. US4 → sync isolation verified
6. Polish → README + full quickstart

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (tree toggle)
   - After T007: Developer B: US2 (footer)
   - After US1: Developer C: US3 + US4 (session + sync isolation)
3. Stories integrate on shared `Model.diffDirection` / dual maps

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work
- [Story] label maps task to US1–US4 for traceability
- Exact toggle key is `d` (research decision); footer copy from contract
- Sync must never gain outbound badges or the direction key
- Commit after each task or logical group
- Stop at any checkpoint to validate the story independently
