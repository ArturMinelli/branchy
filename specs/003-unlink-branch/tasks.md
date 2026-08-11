---

description: "Task list for Unlink Branch from Tree feature implementation"
---

# Tasks: Unlink Branch from Tree

**Input**: Design documents from `/specs/003-unlink-branch/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/unlink-cli.md, quickstart.md

**Tests**: Unit tests included for `internal/tree` unlink primitives and TUI confirm/cancel per plan.md quality gates. No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/tree/tree.go`
- `internal/tree/tree_test.go`
- `internal/tui/app.go`
- `internal/tui/app_test.go`
- `internal/cli/root.go`
- `README.md`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm extension points; no new packages or dependencies

- [x] T001 Verify implementation targets in `specs/003-unlink-branch/plan.md` — extend `internal/tree`, `internal/tui/app.go`, and `internal/cli/root.go` only

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared tree mutation primitives used by TUI and CLI

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Implement `SubtreeNames(root string) []string` via DFS in `internal/tree/tree.go`
- [x] T003 Implement `ParentOf(child string) (string, bool)` in `internal/tree/tree.go`
- [x] T004 Implement `HasEdge(parent, child string) bool` in `internal/tree/tree.go`
- [x] T005 Implement `UnlinkSubtree(root string) error` (delete members, remove from parent's children) in `internal/tree/tree.go`
- [x] T006 [P] Add unit tests for `SubtreeNames`, `ParentOf`, and `HasEdge` in `internal/tree/tree_test.go`
- [x] T007 [P] Add unit tests for `UnlinkSubtree` (nested subtree, leaf, root unlink, parent children cleanup, not-in-tree error) in `internal/tree/tree_test.go`

**Checkpoint**: `internal/tree` unlink API ready — user story implementation can begin

---

## Phase 3: User Story 1 - Unlink subtree from TUI (Priority: P1) 🎯 MVP

**Goal**: Select a branch, press `u`, confirm subtree removal, persist updated tree

**Independent Test**: Open branchy TUI on a project with a branch that has descendants; select it, press `u`, confirm; verify branch and descendants removed, siblings and ancestors remain.

### Implementation for User Story 1

- [x] T008 [US1] Add `screenUnlinkConfirm` to screen enum, `unlinkTarget`/`unlinkCount` model fields, and `Unlink` keybinding (`u`) in `internal/tui/app.go`
- [x] T009 [US1] Handle `u` in `updateTree`: require selected branch, set `unlinkTarget`, compute `unlinkCount` via `SubtreeNames`, transition to `screenUnlinkConfirm` in `internal/tui/app.go`
- [x] T010 [US1] Implement `View` for `screenUnlinkConfirm` with message `Remove "<branch>" and N branch(es) from tree? [y/N]` in `internal/tui/app.go`
- [x] T011 [US1] Implement `updateUnlink`: `y`/Enter calls `UnlinkSubtree` + `SaveTree` + `selectProject`; `n`/Esc returns to tree unchanged in `internal/tui/app.go`
- [x] T012 [US1] Wire `screenUnlinkConfirm` into `Update` switch and handle save-failure `errMsg` without inconsistent state in `internal/tui/app.go`
- [x] T013 [P] [US1] Add confirm/cancel key handling tests for unlink screen in `internal/tui/app_test.go`

**Checkpoint**: TUI unlink with confirmation fully functional — MVP deliverable

---

## Phase 4: User Story 2 - Unlink subtree via CLI (Priority: P1)

**Goal**: `branchy unlink <parent> <child>` validates edge, removes child subtree, persists tree

**Independent Test**: Run `branchy unlink develop feature-a` where `feature-a` has children; verify subtree removed, parent edge gone, exit 0. Invalid edge exits 1 with error.

### Implementation for User Story 2

- [x] T014 [US2] Add `unlinkCmd` with `Use: "unlink <parent> <child>"` and `cobra.ExactArgs(2)` in `internal/cli/root.go`
- [x] T015 [US2] Implement validation: non-empty names, parent≠child, child in tree, `HasEdge(parent, child)` in `internal/cli/root.go`
- [x] T016 [US2] Call `UnlinkSubtree(child)`, `SaveTree()`, and print `Unlinked %s → %s (%d branches removed)` per `specs/003-unlink-branch/contracts/unlink-cli.md` in `internal/cli/root.go`
- [x] T017 [US2] Register `unlinkCmd` in `init()` via `rootCmd.AddCommand(unlinkCmd)` in `internal/cli/root.go`

**Checkpoint**: CLI unlink works independently of TUI

---

## Phase 5: User Story 3 - Discoverability in TUI help (Priority: P2)

**Goal**: Tree view footer documents `u: unlink` alongside existing shortcuts

**Independent Test**: Open branchy TUI; verify help line includes `u: unlink`.

### Implementation for User Story 3

- [x] T018 [US3] Update tree view help line to include `u: unlink` alongside `l: link` in `internal/tui/app.go`

**Checkpoint**: Unlink discoverable from TUI without external docs

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and regression checks

- [x] T019 [P] Update `README.md` with `u` TUI shortcut and `branchy unlink <parent> <child>` CLI documentation
- [x] T020 Run all scenarios in `specs/003-unlink-branch/quickstart.md` and fix any gaps found
- [x] T021 Run `go test ./...` and ensure all packages pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Phase 2 — MVP
- **User Story 2 (Phase 4)**: Depends on Phase 2 — parallel with US1 after checkpoint
- **User Story 3 (Phase 5)**: Depends on US1 T008 (keybinding exists); help text can ship after US1
- **Polish (Phase 6)**: Depends on Phases 3–5

### User Story Dependencies

- **US1 (P1)**: Starts after Foundational — no dependency on US2/US3
- **US2 (P1)**: Starts after Foundational — independently testable via CLI; parallel with US1
- **US3 (P2)**: Depends on US1 keybinding being present — one-line help update

### Within Each User Story

- Foundational: helpers before `UnlinkSubtree`; tests after implementation
- US1: model fields → tree key handler → view → update handler → tests
- US2: command stub → validation → mutation/save → registration

### Parallel Opportunities

- **Phase 2**: T006 and T007 in parallel (after T002–T005)
- **Phase 3 + 4**: US1 (`internal/tui/app.go`) and US2 (`internal/cli/root.go`) in parallel after Phase 2 checkpoint
- **Phase 3**: T013 parallel once T011–T012 complete
- **Phase 6**: T019 parallel with T020

---

## Parallel Example: After Foundational Checkpoint

```bash
# Developer A — User Story 1 (TUI):
# Tasks T008–T013 in internal/tui/app.go

# Developer B — User Story 2 (CLI):
# Tasks T014–T017 in internal/cli/root.go
```

## Parallel Example: Foundational Tests

```bash
# After T005 completes, run in parallel:
# Task T006: "Add unit tests for SubtreeNames, ParentOf, HasEdge in internal/tree/tree_test.go"
# Task T007: "Add unit tests for UnlinkSubtree in internal/tree/tree_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002–T007)
3. Complete Phase 3: User Story 1 (T008–T013)
4. **STOP and VALIDATE**: Run quickstart Scenarios 4–5
5. Demo if ready

### Incremental Delivery

1. Setup + Foundational → shared tree unlink API ready
2. Add US1 → TUI unlink with confirm → **MVP**
3. Add US2 → CLI `branchy unlink` for scripting
4. Add US3 → help line discoverability
5. Polish → README + full quickstart validation

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Split after Phase 2:
   - Developer A: US1 (`internal/tui/app.go`)
   - Developer B: US2 (`internal/cli/root.go`)
3. US3 + Polish together

---

## Notes

- `[P]` tasks touch different files with no incomplete-task dependencies
- `[USn]` labels map to spec.md user stories for traceability
- Unlink is config-only — does not delete Git branches or GitLab MRs
- Empty tree after unlink is valid; no minimum branch count
- Commit after each task or logical group; stop at any checkpoint to validate independently
