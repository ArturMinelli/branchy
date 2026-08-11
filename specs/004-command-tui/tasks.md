---

description: "Task list for Unified Command TUI feature implementation"
---

# Tasks: Unified Command TUI

**Input**: Design documents from `/specs/004-command-tui/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/command-tui.md, quickstart.md

**Tests**: Unit tests included per plan.md quality gates (`mode_test.go`, per-flow step transition tests). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/cli/mode.go`, `internal/cli/root.go`
- `internal/tui/chrome.go`, `internal/tui/*flow.go`
- `README.md`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and extension points; no new external dependencies

- [x] T001 Verify implementation targets in `specs/004-command-tui/plan.md` — extend `internal/cli`, `internal/tui/*flow.go`, and `internal/tui/chrome.go` only; no changes to `internal/sync` orchestration

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared TUI chrome and interactive/scripted mode detection used by all commands

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Create shared lipgloss styles and helpers (`RenderTitle`, `RenderHelp`, `MinSizeOK`, `RenderTooSmall`) in `internal/tui/chrome.go`
- [x] T003 Migrate style variables from `internal/tui/app.go` to `internal/tui/chrome.go` and update imports in `internal/tui/app.go`
- [x] T004 Implement `IsTTY()`, `UseTUI(cmd *cobra.Command)`, and `anyFlagChanged()` in `internal/cli/mode.go`
- [x] T005 [P] Add unit tests for TTY detection and any-flag-changed logic in `internal/cli/mode_test.go`
- [x] T006 [P] Update `internal/tui/mrflow.go` to use shared chrome helpers from `internal/tui/chrome.go`

**Checkpoint**: Mode detection and shared styling ready — user story implementation can begin

---

## Phase 3: User Story 1 - Standalone sync TUI from CLI (Priority: P1) 🎯 MVP

**Goal**: `branchy sync` without flags launches a dedicated full-screen sync TUI with root-branch picker, per-edge confirm, summary, and browser prompt

**Independent Test**: Run `branchy sync` in a registered project; complete flow via keyboard; verify MR outcomes match embedded tree sync (`s` key) for same choices; process exits to shell (not main tree app)

### Implementation for User Story 1

- [x] T007 [US1] Add `stepSyncPickRoot` with searchable branch list picker to `internal/tui/syncflow.go`
- [x] T008 [US1] Implement `RunSync(p *project.Project, opts SyncFlowOptions) error` entry point in `internal/tui/syncflow.go`
- [x] T009 [US1] Wire `syncCmd` to call `UseTUI(cmd)` → `tui.RunSync` when interactive in `internal/cli/root.go`
- [x] T010 [P] [US1] Add root-picker and standalone-quit step transition tests in `internal/tui/syncflow_test.go`

**Checkpoint**: Standalone sync TUI fully functional — MVP deliverable

---

## Phase 4: User Story 2 - TUI for projects, link, unlink, and init (Priority: P1)

**Goal**: Interactive `branchy link`, `branchy unlink`, `branchy init`, and `branchy projects` launch dedicated TUI flows with list pickers and confirmations

**Independent Test**: Run each command without flags in a TTY; complete via keyboard; verify tree mutations (link/unlink/init) or project browse (projects); exit to shell

### Implementation for User Story 2

- [x] T011 [P] [US2] Implement `RunLink` and `LinkFlowModel` (parent picker → child name → confirm → save) in `internal/tui/linkflow.go`
- [x] T012 [P] [US2] Implement `RunUnlink` and `UnlinkFlowModel` (branch picker → subtree confirm → save) in `internal/tui/unlinkflow.go`
- [x] T013 [P] [US2] Implement `RunInit` and `InitFlowModel` (confirm register → run Init → success/error) in `internal/tui/initflow.go`
- [x] T014 [P] [US2] Implement `RunProjects` and `ProjectsFlowModel` (searchable list → detail view) in `internal/tui/projectsflow.go`
- [x] T015 [US2] Change `linkCmd` and `unlinkCmd` to `cobra.RangeArgs(0, 2)` with 0/1/2 arg validation in `internal/cli/root.go`
- [x] T016 [US2] Wire `linkCmd` to `UseTUI(cmd) && len(args)==0` → `tui.RunLink` in `internal/cli/root.go`
- [x] T017 [US2] Wire `unlinkCmd` to `UseTUI(cmd) && len(args)==0` → `tui.RunUnlink` in `internal/cli/root.go`
- [x] T018 [US2] Wire `initCmd` to `UseTUI(cmd)` → `tui.RunInit` in `internal/cli/root.go`
- [x] T019 [US2] Wire `projectsCmd` to `UseTUI(cmd)` → `tui.RunProjects` in `internal/cli/root.go`
- [x] T020 [P] [US2] Add step transition tests in `internal/tui/linkflow_test.go`
- [x] T021 [P] [US2] Add step transition tests in `internal/tui/unlinkflow_test.go`
- [x] T022 [P] [US2] Add step transition tests in `internal/tui/initflow_test.go`
- [x] T023 [P] [US2] Add step transition tests in `internal/tui/projectsflow_test.go`

**Checkpoint**: All four command TUIs work independently in interactive mode

---

## Phase 5: User Story 3 - Scriptable non-TUI paths preserved (Priority: P1)

**Goal**: Any flag, full positional args, or non-TTY invocation uses plain CLI with no full-screen TUI and no hangs

**Independent Test**: Run `branchy sync --from develop -y`, `branchy link parent child`, `branchy init --force`, `branchy mr --source A --target B -y`, and piped `branchy sync`; verify no TUI and correct exit behavior per `specs/004-command-tui/contracts/command-tui.md`

### Implementation for User Story 3

- [x] T024 [US3] Ensure `syncCmd` plain CLI path runs when any flag set or non-TTY (no `RunSync`) in `internal/cli/root.go`
- [x] T025 [US3] Add non-TTY fallback errors for sync, link, and unlink when args insufficient in `internal/cli/root.go`
- [x] T026 [US3] Align `mrCmd` to any-flag rule: no TUI when any flag set; require `--source` and `--target` for plain creation in `internal/cli/root.go`
- [x] T027 [US3] Preserve plain CLI fallbacks for `initCmd` (non-TTY) and `projectsCmd` (non-TTY tab-separated list) in `internal/cli/root.go`
- [x] T028 [P] [US3] Add non-TTY and flag-forces-CLI integration tests in `internal/cli/mode_test.go`

**Checkpoint**: All scripted and CI paths verified; no interactive hang scenarios

---

## Phase 6: User Story 4 - Visual consistency and simplicity (Priority: P2)

**Goal**: All dedicated command TUIs share consistent title, help, confirm, success/warning/error styling and minimum terminal size guard

**Independent Test**: Run `branchy sync`, `branchy mr`, `branchy link` in sequence; verify matching headers, help footers, and `y`/`n`/`esc` confirm behavior; resize below 80×24 shows readable too-small message

### Implementation for User Story 4

- [x] T029 [US4] Add `MinSizeOK` guard on `tea.WindowSizeMsg` in `internal/tui/syncflow.go`
- [x] T030 [P] [US4] Add `MinSizeOK` guard in `internal/tui/linkflow.go`
- [x] T031 [P] [US4] Add `MinSizeOK` guard in `internal/tui/unlinkflow.go`
- [x] T032 [P] [US4] Add `MinSizeOK` guard in `internal/tui/initflow.go`
- [x] T033 [P] [US4] Add `MinSizeOK` guard in `internal/tui/projectsflow.go`
- [x] T034 [US4] Audit confirm screens and help footers across all `*flow.go` files for consistent `y`/`n`/`esc` bindings and chrome usage

**Checkpoint**: Visual language unified across all command flows

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and regression checks

- [x] T035 [P] Update `README.md` with interactive TUI behavior for sync, link, unlink, init, projects, and any-flag scripted mode rules
- [x] T036 Run all scenarios in `specs/004-command-tui/quickstart.md` and fix any gaps found
- [x] T037 Run `go test ./...` and ensure all packages pass
- [x] T038 Verify embedded main tree TUI (`branchy` with no subcommand; `s`, `m`, `l`, `u` keys) unchanged in `internal/tui/app.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Phase 2 — **MVP**
- **User Story 2 (Phase 4)**: Depends on Phase 2 — parallel with US1 after checkpoint
- **User Story 3 (Phase 5)**: Depends on US1/US2 CLI wiring (T009, T015–T019); completes scripted path coverage
- **User Story 4 (Phase 6)**: Depends on US1/US2 flow files existing — can parallel with late US3
- **Polish (Phase 7)**: Depends on Phases 3–6

### User Story Dependencies

- **US1 (P1)**: Starts after Foundational — no dependency on US2/US3/US4
- **US2 (P1)**: Starts after Foundational — independently testable; parallel with US1
- **US3 (P1)**: Requires CLI wiring from US1/US2; validates scripted contracts
- **US4 (P2)**: Requires flow files from US1/US2; polish layer only

### Within Each User Story

- Foundational: `chrome.go` before style migration; `mode.go` before any CLI wiring
- US1: root picker → `RunSync` → CLI wire → tests
- US2: flow models (parallel) → `RangeArgs` → CLI wiring per command → tests
- US3: ensure plain paths after TUI branches exist; mr alignment last
- US4: min-size guards (parallel) → confirm audit

### Parallel Opportunities

- **Phase 2**: T005 and T006 in parallel (after T002–T004)
- **Phase 3 + 4**: US1 (`syncflow.go`) and US2 flow files (T011–T014) in parallel after Phase 2
- **Phase 4**: T011–T014 in parallel; T020–T023 in parallel after respective flows
- **Phase 6**: T030–T033 in parallel
- **Phase 7**: T035 parallel with T036

---

## Parallel Example: After Foundational Checkpoint

```bash
# Developer A — User Story 1 (sync TUI):
# Tasks T007–T010 in internal/tui/syncflow.go and internal/cli/root.go

# Developer B — User Story 2 (other flows):
# Tasks T011–T014 in parallel (linkflow, unlinkflow, initflow, projectsflow)
```

## Parallel Example: Foundational Tests

```bash
# After T004 completes, run in parallel:
# Task T005: "Add unit tests in internal/cli/mode_test.go"
# Task T006: "Update internal/tui/mrflow.go to use chrome.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002–T006)
3. Complete Phase 3: User Story 1 (T007–T010)
4. **STOP and VALIDATE**: Run quickstart Scenarios 1–2
5. Demo if ready

### Incremental Delivery

1. Setup + Foundational → shared chrome and mode detection ready
2. Add US1 → standalone `branchy sync` TUI → **MVP**
3. Add US2 → link, unlink, init, projects TUIs
4. Add US3 → scripted path hardening and mr any-flag alignment
5. Add US4 → min-size guards and visual audit
6. Polish → README + full quickstart validation

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Split after Phase 2:
   - Developer A: US1 sync TUI
   - Developer B: US2 flow models (link, unlink, init, projects)
3. US3 developer wires scripted fallbacks after US1/US2 CLI branches land
4. US4 + Polish together

---

## Notes

- `[P]` tasks touch different files with no incomplete-task dependencies
- `[USn]` labels map to spec.md user stories for traceability
- Embedded main-app shortcuts (`s`, `m`, `l`, `u`) in `internal/tui/app.go` must remain unchanged
- `branchy mr --source` alone becomes plain CLI error (breaking alignment per contract)
- Dedicated flows MUST exit to shell — never drop user into main tree app
- Commit after each task or logical group; stop at any checkpoint to validate independently
