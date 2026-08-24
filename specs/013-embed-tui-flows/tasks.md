---

description: "Task list for Embed TUI Flows feature implementation"
---

# Tasks: Embed TUI Flows

**Input**: Design documents from `/specs/013-embed-tui-flows/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/embedded-flows.md, quickstart.md

**Tests**: Included in existing files only (`internal/tui/linkflow_test.go`, `internal/tui/unlinkflow_test.go`, `internal/tui/app_test.go`). Tests with implementation — no TDD-first ordering (grilling 2026-08-24). No new TUI test files.

**Organization**: Tasks grouped by user story. MVP is US1 + US2 together (grilling: do not ship with inline unlink still present). `app.go` is shared — US2 host wiring waits until US1 host wiring is done. Unlink flow-file work (T009) can overlap US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/tui/app.go` — router; delete inline link/unlink
- `internal/tui/linkflow.go`, `internal/tui/unlinkflow.go` — dedicated wizards + options
- `internal/tui/app_test.go`, `internal/tui/linkflow_test.go`, `internal/tui/unlinkflow_test.go`
- `internal/tui/syncflow.go`, `internal/tui/mrflow.go` — pattern to copy; do not split
- `internal/cli/link.go`, `internal/cli/unlink.go` — stay `tui.RunLink` / `RunUnlink`
- Specs 014–015 stay closed

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before editing flows

- [x] T001 Verify implementation targets in `specs/013-embed-tui-flows/plan.md` — deepen `internal/tui/linkflow.go` / `unlinkflow.go` with `Embedded` + prefill; host in `internal/tui/app.go` like sync/MR; no new TUI files; no split of `app.go` or `syncflow.go` / `mrflow.go`; persist stays `project.Link` / `project.Unlink`; specs 014–015 files stay closed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Options + embedded finish/cancel on both flow models so stories only wire the host

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Run `go test ./internal/tui/... -count=1` and confirm green as the pre-embed baseline
- [x] T003 [P] Add `LinkFlowOptions` (`Embedded`, `PrefillParent`) to `internal/tui/linkflow.go`; change `newLinkFlowModel` to take options; `RunLink` forces `Embedded: false`; add `cancelOrQuit` / `finish` copied from `internal/tui/mrflow.go` (embedded → no `tea.Quit`); keep current start-at-parent-picker when `PrefillParent` is empty; update existing constructors in `internal/tui/linkflow_test.go` to pass empty options
- [x] T004 [P] Add `UnlinkFlowOptions` (`Embedded`, `PrefillTarget`) to `internal/tui/unlinkflow.go`; same constructor / `RunUnlink` / `cancelOrQuit` / `finish` pattern as T003; keep current start-at-picker when `PrefillTarget` is empty; update `internal/tui/unlinkflow_test.go` constructors

**Checkpoint**: Standalone link/unlink tests still pass; empty options preserve today’s wizard start; `tea.Quit` still happens only when not embedded

---

## Phase 3: User Story 1 - Main-tree link is the same flow as standalone link (Priority: P1)

**Goal**: `l` on the tree runs `LinkFlowModel` (prefill parent, start at child name). Delete the two-field inline editor.

**Independent Test**: From a selected branch, `l` shows the child-name step (not “Link branch” two fields). Confirm Yes → success step → enter → tree has the child. Cancel → tree unchanged. Same pair via standalone `branchy link` persists the same. (`quickstart.md` Scenarios 1–2; SC-001/002/003 for link)

### Implementation for User Story 1

- [x] T005 [US1] In `internal/tui/linkflow.go`, when `PrefillParent` is non-empty and in `p.Tree.Branches`, start at `stepLinkChild` with that parent; esc from child still rebuilds the parent picker (even if entry skipped it); add tests in `internal/tui/linkflow_test.go` for prefill-starts-at-child and esc-from-child-opens-picker
- [x] T006 [US1] In `internal/tui/app.go`, add `linkFlow LinkFlowModel`; while `screenLink`, forward all `tea.Msg` to `linkFlow` and render `linkFlow.View()` (same early-return pattern as `screenSync`); `l` sets `LinkFlowOptions{Embedded: true, PrefillParent: selectedName}` (empty selection → no prefill); on `finished` after success step call `selectProject(m.current)`, otherwise return to `screenTree` and zero `linkFlow`; delete `updateLink`, `linkParent`, `linkChild`, `linkInput`
- [x] T007 [US1] Update `internal/tui/app_test.go`: `l` with selection opens the dedicated flow at child step; confirm decline / picker esc returns to tree with no persist; after Yes, assert success step then enter returns to a tree that includes the new child; `View()` must not contain the old two-field “Link branch” editor
- [x] T008 [US1] Run `go test ./internal/tui/... -count=1` and fix link-path failures until green (unlink tests may still expect the old host confirm until US2)

**Checkpoint**: Main-tree `l` is the standalone wizard; inline link editor gone; persist still `project.Link` inside `linkflow.go`

---

## Phase 4: User Story 2 - Main-tree unlink is the same flow as standalone unlink (Priority: P1) 🎯 MVP (with US1)

**Goal**: `u` on the tree runs `UnlinkFlowModel` starting at the dedicated confirm. Delete the host-only unlink confirm.

**Independent Test**: `u` on a selected in-tree branch opens the dedicated confirm (subtree size, yes/no). Yes → success step → enter → subtree gone. No/esc → tree unchanged. Empty selection does not launch. (`quickstart.md` Scenario 3; SC-001/002/003 for unlink)

### Implementation for User Story 2

- [x] T009 [P] [US2] In `internal/tui/unlinkflow.go`, when `PrefillTarget` is non-empty and in the tree, start at `stepUnlinkConfirm` with subtree size and the existing confirm copy; confirm No/esc still leaves the flow (does not reopen the picker); add tests in `internal/tui/unlinkflow_test.go` for prefill-starts-at-confirm
- [x] T010 [US2] In `internal/tui/app.go`, add `unlinkFlow UnlinkFlowModel` and rename `screenUnlinkConfirm` → `screenUnlink`; forward Update/View like link/sync; `u` with empty selection is a no-op; otherwise `UnlinkFlowOptions{Embedded: true, PrefillTarget: selected}`; success → `selectProject`, cancel/error → `screenTree`; delete `updateUnlink`, `unlinkTarget`, `unlinkCount`, `unlinkConfirm`; do not call `project.Unlink` from `app.go`
- [x] T011 [US2] Rewrite `TestUpdateTreeUnlinkOpensConfirm`, `TestUpdateUnlinkCancel`, and `TestUpdateUnlinkConfirmRemovesSubtree` in `internal/tui/app_test.go` to drive `unlinkFlow` — Yes now lands on the success step (not immediate `screenTree`); enter after success removes the subtree; cancel leaves the tree intact; confirm view still must not contain `[y/N]`
- [x] T012 [US2] Run `go test ./internal/tui/... -count=1` until green including rewritten unlink host tests

**Checkpoint**: Zero inline link/unlink implementations in `app.go` (FR-003). **MVP: stop here and validate US1+US2** before US3 polish.

---

## Phase 5: User Story 3 - Sync and MR stay embedded; main app stays a router (Priority: P2)

**Goal**: Do not split `app.go` or sync/MR files; standalone link/unlink still exit to the shell; picker, direction, and existing embeddings keep passing.

**Independent Test**: `go test ./internal/tui/...` includes `embedded_sync_test.go` and direction/picker tests; `internal/cli/link.go` / `unlink.go` still call `tui.RunLink` / `RunUnlink`; `app.go` remains one file (`quickstart.md` Scenarios 4–5; SC-004)

### Implementation for User Story 3

- [x] T013 [P] [US3] Confirm `internal/tui/syncflow.go`, `internal/tui/mrflow.go`, and `internal/tui/embedded_sync_test.go` have no split and no behavior changes required by this spec; `internal/tui/app.go` is still a single file (FR-008)
- [x] T014 [P] [US3] Confirm `internal/cli/link.go` and `internal/cli/unlink.go` still delegate TUI to `tui.RunLink` / `tui.RunUnlink` (FR-006); no main-tree launch from those commands
- [x] T015 [US3] Run `go test ./internal/tui/... -count=1` and `go test ./internal/cli/... -count=1` — embedded sync, MR, direction, and picker tests must pass

**Checkpoint**: Sync/MR embeddings and CLI process boundaries unchanged

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Contract grep, full regression, quickstart

- [x] T016 Run the grep audit in `specs/013-embed-tui-flows/contracts/embedded-flows.md` against `internal/tui/app.go` and `internal/tui/`: no `linkParent` / `linkChild` / `linkInput` / `unlinkTarget` / `unlinkConfirm`; no `updateLink` / `updateUnlink`; no `project.Link` / `project.Unlink` in `app.go`
- [x] T017 Run `go test ./... -count=1` and fix regressions outside TUI caused by constructor signature changes
- [x] T018 Run `specs/013-embed-tui-flows/quickstart.md` Scenarios 1–6 (embed wizards, cancel, standalone shell, sync/MR freeze, grep)
- [x] T019 [P] Confirm no new files under `internal/tui/` and that `specs/014-domain-ui-split/` and `specs/015-unify-domain-types/` were not edited

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — T005 → T006 → T007 → T008; T006 edits `app.go`
- **User Story 2 (Phase 4)**: T009 can start after T004 (parallel with US1); T010–T012 wait for T006 (`app.go`); **MVP complete after T012**
- **User Story 3 (Phase 5)**: Depends on US1+US2 so freeze checks are meaningful
- **Polish (Phase 6)**: Depends on US1–US3

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (embed link) | Foundational | T003 |
| US2 (embed unlink) | Foundational; host wiring after US1 | T004 (flow); T006 (app.go) |
| US3 (freeze sync/MR) | US1+US2 | T012 |
| Polish | US1–US3 | T015 |

### Within Each User Story

- Prefill + flow tests in the flow file, then host wiring in `app.go`, then rewrite `app_test.go`
- Success/error steps stay visible; host refreshes the tree only after a successful persist **and** a done key
- Embedded `q`/`esc` match MR: return to tree, do not `tea.Quit`

### Parallel Opportunities

- **Phase 2**: T003 ∥ T004 (different flow files)
- **Phase 3+4**: T009 ∥ T005–T008 (unlinkflow vs link/app)
- **Phase 5**: T013 ∥ T014 after T012
- **Phase 6**: T019 ∥ T016; T017–T018 after greps

---

## Parallel Example: Foundational + US1

```bash
# After T002:
Task T003: internal/tui/linkflow.go (+ linkflow_test.go constructors)
Task T004: internal/tui/unlinkflow.go (+ unlinkflow_test.go constructors)

# Then US1 sequential on link + app:
Task T005: internal/tui/linkflow.go prefill + tests
Task T006: internal/tui/app.go host link, delete updateLink
Task T007: internal/tui/app_test.go link host cases
Task T008: go test ./internal/tui/...

# Overlap with US1:
Task T009: internal/tui/unlinkflow.go prefill + tests
```

---

## Implementation Strategy

### MVP (User Stories 1 and 2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 (embedded link)
4. Complete Phase 4: User Story 2 (embedded unlink)
5. **STOP and VALIDATE**: `go test ./internal/tui/...`; grep audit for inline fields; `quickstart.md` Scenarios 1–3
6. Demo: `l` and `u` on the main tree use the standalone wizards

Do not treat US1 alone as shippable — inline unlink would remain (FR-003).

### Incremental Delivery

1. Setup + Foundational → options and finish helpers
2. US1 → embedded link
3. US2 → embedded unlink (**MVP**)
4. US3 → freeze sync/MR/CLI boundaries
5. Polish → full `go test ./...` + quickstart 1–6

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T003 ∥ T004)
2. Developer A: US1 (T005–T008) on `linkflow.go` / `app.go`
3. Developer B: T009 on `unlinkflow.go` in parallel
4. After T006: Developer B (or A) runs T010–T012 on `app.go`
5. US3 + polish on the merged tree

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work in the same file
- [Story] label maps task to US1–US3
- Outcome freeze on persist and error meaning; wizard chrome on the main tree is allowed
- `q` on embedded wizards returns to the tree (match sync/MR), which is a visible change from the old inline editor that quit the app
- Grilling (2026-08-24, plan): prefill skip-picker; success/error then tree; esc from child opens picker
- Grilling (2026-08-24, tasks): tests with impl, existing test files only, MVP = US1+US2
- Avoid: new TUI files; embed dispatcher; splitting `app.go`; `project.Link`/`Unlink` in `app.go`; TDD-first ordering
