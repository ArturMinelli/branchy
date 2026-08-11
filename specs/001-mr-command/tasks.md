---

description: "Task list for Manual MR Command feature implementation"
---

# Tasks: Manual MR Command

**Input**: Design documents from `/specs/001-mr-command/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/mr-cli.md, quickstart.md

**Tests**: Unit tests included for `internal/mr` validation logic and TUI step transitions per plan.md quality gates. No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `cmd/branchy/main.go`
- `internal/cli/root.go`
- `internal/mr/mr.go`
- `internal/tui/mrflow.go`
- `internal/tui/app.go`
- `internal/sync/sync.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Scaffold new packages and command registration point

- [x] T001 Create `internal/mr/` package with `CreateRequest`, `CreateResult` type stubs in `internal/mr/mr.go`
- [x] T002 [P] Create `internal/tui/mrflow.go` with `MRFlowModel`, step enum, and `MROptions` type stubs

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared MR creation service used by TUI, CLI flags, and sync refactor

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Implement branch validation helpers (`ValidateBranches`, same-branch check, tree membership) in `internal/mr/mr.go`
- [x] T004 Implement `DefaultTitle(source, target string)` and `DefaultDescription()` helpers in `internal/mr/mr.go`
- [x] T005 Implement `Create(p *project.Project, req CreateRequest) (*CreateResult, error)` with glab find-existing → create flow in `internal/mr/mr.go`
- [x] T006 Refactor `processEdge` in `internal/sync/sync.go` to delegate MR creation to `internal/mr.Create()` while preserving sync auto-browser behavior
- [x] T007 [P] Add unit tests for validation, default title/description, and result action mapping in `internal/mr/mr_test.go`

**Checkpoint**: `internal/mr` package ready — user story implementation can begin

---

## Phase 3: User Story 1 - Create MR via TUI branch picker (Priority: P1) 🎯 MVP

**Goal**: `branchy mr` launches a multi-step TUI (source → target → title → confirm → result → browser prompt) using branch-tree branches only

**Independent Test**: Run `branchy mr` in a registered project with ≥2 tree branches; complete the flow and verify MR created on GitLab with chosen source/target/title. Browser opens only on explicit `y` at browser prompt.

### Implementation for User Story 1

- [x] T008 [US1] Implement source branch picker step using `bubbles/list` over `tree.Document.Names()` in `internal/tui/mrflow.go`
- [x] T009 [US1] Implement target branch picker step excluding selected source with source/target labels in `internal/tui/mrflow.go`
- [x] T010 [US1] Implement title edit step with default `MR: <source> → <target>` and backspace/type input in `internal/tui/mrflow.go`
- [x] T011 [US1] Implement confirm step calling `internal/mr.Create()` on `y`/Enter in `internal/tui/mrflow.go`
- [x] T012 [US1] Implement result display and browser prompt (`Open in browser? [y/N]`) calling `internal/browser/open.go` on confirm in `internal/tui/mrflow.go`
- [x] T013 [US1] Handle edge cases in TUI: fewer than 2 branches message, creation failure screen, browser-open failure warning in `internal/tui/mrflow.go`
- [x] T014 [US1] Implement `RunMR(p *project.Project, opts MROptions) error` standalone Bubble Tea entry in `internal/tui/mrflow.go`
- [x] T015 [US1] Register `mrCmd` subcommand in `internal/cli/root.go` dispatching to `tui.RunMR()` when branch flags are incomplete
- [x] T016 [P] [US1] Add step transition tests (source→target→title→confirm→browser) in `internal/tui/mrflow_test.go`

**Checkpoint**: `branchy mr` TUI flow fully functional — MVP deliverable

---

## Phase 4: User Story 2 - Create MR via CLI flags (Priority: P2)

**Goal**: Non-interactive `branchy mr --source A --target B [--title T] [--yes]` bypasses TUI and prints contract-formatted stdout

**Independent Test**: Run `branchy mr --source feature-x --target develop --yes` — no TUI, MR created, URL printed, no browser opened.

### Implementation for User Story 2

- [x] T017 [US2] Add `--source`, `--target`, `--title`, `--yes`/`-y` flags to `mrCmd` in `internal/cli/root.go`
- [x] T018 [US2] Implement flag-mode routing: both flags set → non-TUI path; `--target` without `--source` → error in `internal/cli/root.go`
- [x] T019 [US2] Implement stdin confirmation prompt `Create MR <source> → <target>? [y/N]` when `--yes` omitted in flag mode in `internal/cli/root.go`
- [x] T020 [US2] Implement flag-mode stdout formatting (`Created:`/`Skipped:`/`Failed:` + URL) per `specs/001-mr-command/contracts/mr-cli.md` in `internal/cli/root.go`
- [x] T021 [US2] Wire `--title` flag through to `internal/mr.CreateRequest` in `internal/cli/root.go`

**Checkpoint**: Flag mode works independently of TUI; power users can script MR creation

---

## Phase 5: User Story 3 - Launch MR flow from main TUI (Priority: P2)

**Goal**: Press `m` from tree view to enter MR flow with source pre-filled from selected branch; Esc returns to tree

**Independent Test**: Open `branchy`, select a branch, press `m`, complete MR flow. Esc during flow returns to tree without creating MR.

### Implementation for User Story 3

- [x] T022 [US3] Add `screenMR` to screen enum and `MR` keybinding (`m`) in `internal/tui/app.go`
- [x] T023 [US3] Embed `MRFlowModel` in main `Model` with delegated Update/View when `screen == screenMR` in `internal/tui/app.go`
- [x] T024 [US3] Pre-fill source from `treeView.selectedName()` and skip to target step when branch selected in `internal/tui/app.go`
- [x] T025 [US3] Implement cancel/Esc handler returning to `screenTree` without side effects in `internal/tui/app.go`
- [x] T026 [US3] Update tree view help line to include `m: mr` in `internal/tui/app.go`

**Checkpoint**: Main TUI integration complete; all three user stories independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and consistency across all entry points

- [x] T027 [P] Update `README.md` with `branchy mr` command, flags, and `m` TUI shortcut documentation
- [x] T028 Run all scenarios in `specs/001-mr-command/quickstart.md` and fix any gaps found
- [x] T029 [P] Verify `branchy sync` behavior unchanged (auto-browser, batch confirm) after `internal/sync/sync.go` refactor
- [x] T030 Run `go test ./...` and ensure all packages pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Phase 2 — MVP
- **User Story 2 (Phase 4)**: Depends on Phase 2 + T015 (`mrCmd` exists from US1)
- **User Story 3 (Phase 5)**: Depends on Phase 3 (`MRFlowModel` and `RunMR` must exist)
- **Polish (Phase 6)**: Depends on Phases 3–5

### User Story Dependencies

- **US1 (P1)**: Starts after Foundational — no dependency on US2/US3
- **US2 (P2)**: Starts after US1 T015 (needs `mrCmd` registration) — independently testable via flags
- **US3 (P2)**: Starts after US1 (needs `MRFlowModel`) — independently testable via main TUI

### Within Each User Story

- TUI steps built bottom-up: pickers → title → confirm → result → browser
- CLI flags extend existing `mrCmd` after TUI path works
- Main TUI embeds completed `MRFlowModel`

### Parallel Opportunities

- **Phase 1**: T001 and T002 in parallel
- **Phase 2**: T007 parallel with T006 (after T003–T005)
- **Phase 3**: T016 parallel once T008–T014 complete
- **Phase 6**: T027 and T029 in parallel
- **Cross-story**: US2 and US3 can proceed in parallel after US1 checkpoint (different files: `root.go` vs `app.go`)

---

## Parallel Example: User Story 1

```bash
# After T014 completes, run in parallel:
# Task T016: "Add step transition tests in internal/tui/mrflow_test.go"
# Manual validation: "Run branchy mr and complete full TUI flow"
```

## Parallel Example: After US1 Checkpoint

```bash
# Developer A — User Story 2:
# Tasks T017–T021 in internal/cli/root.go

# Developer B — User Story 3:
# Tasks T022–T026 in internal/tui/app.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T002)
2. Complete Phase 2: Foundational (T003–T007)
3. Complete Phase 3: User Story 1 (T008–T016)
4. **STOP and VALIDATE**: Run quickstart Scenario 1
5. Demo/deploy if ready

### Incremental Delivery

1. Setup + Foundational → shared MR service ready
2. Add US1 → `branchy mr` TUI works → **MVP**
3. Add US2 → flag mode for scripting
4. Add US3 → `m` shortcut in main TUI
5. Polish → docs + full quickstart validation

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. One developer completes US1 (required first)
3. Then split:
   - Developer A: US2 (`internal/cli/root.go`)
   - Developer B: US3 (`internal/tui/app.go`)
4. Rejoin for Polish phase

---

## Notes

- `[P]` tasks touch different files with no incomplete-task dependencies
- `[USn]` labels map to spec.md user stories for traceability
- `internal/mr.Create()` never opens browser — TUI and sync callers own that behavior
- Flag mode browser open intentionally out of scope for v1 per spec assumptions
- Commit after each task or logical group; stop at any checkpoint to validate independently
