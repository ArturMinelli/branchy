---

description: "Task list for Split CLI Commands feature implementation"
---

# Tasks: Split CLI Commands

**Input**: Design documents from `/specs/012-split-cli-commands/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-commands.md, quickstart.md

**Tests**: Verification only — keep `internal/cli/mode_test.go`; two-step grep audit per contracts; no per-command test files (grilling 2026-08-21). No TDD-first ordering.

**Organization**: Tasks grouped by user story for independent validation. File extractions T003–T008 can run in parallel (new files only); T009 trims `root.go` once all copies exist.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `cmd/branchy/main.go` — unchanged
- `internal/cli/root.go` — assembly only after split
- `internal/cli/mode.go`, `internal/cli/mode_test.go` — unchanged
- `internal/cli/{init,sync,mr,link,unlink,projects}.go` — new command files
- Specs 013–015 stay closed

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope, prerequisites, and baseline before moving code

- [x] T001 Verify implementation targets in `specs/012-split-cli-commands/plan.md` — flat `internal/cli/` layout; spec 011 owners present (`project.Link`, `project.Unlink`, `project.Init`, `sync.Run`, `mr.Create`); `cmd/branchy/main.go` stays thin; no `commands/` subfolder; no `prompt.go`; specs 013–015 files stay closed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Baseline green tests before refactor

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Run `go test ./internal/cli/... -count=1` and confirm green; note `internal/cli/root.go` line count (~340) as pre-split baseline per `specs/012-split-cli-commands/quickstart.md`

**Checkpoint**: CLI tests pass; monolithic `root.go` is the only command source file besides `mode.go`

---

## Phase 3: User Story 1 - Edit one command without opening the others (Priority: P1) 🎯 MVP

**Goal**: Each subcommand lives in its own file; `root.go` becomes tree assembly + default TUI entry only.

**Independent Test**: `ls internal/cli/{init,sync,mr,link,unlink,projects}.go`; `wc -l internal/cli/root.go` ≈ 40; `./branchy --help` and each subcommand `--help` unchanged (`quickstart.md` Scenario 1; SC-001, SC-004)

### Implementation for User Story 1

- [x] T003 [P] [US1] Create `internal/cli/link.go` with `linkCmd` copied from `internal/cli/root.go` (Use, Args, RunE — TUI via `UseTUI` + `tui.RunLink`, scripted via `p.Link`); do not edit `root.go` yet
- [x] T004 [P] [US1] Create `internal/cli/unlink.go` with `unlinkCmd` copied from `internal/cli/root.go` (scripted via `p.Unlink` + `UnlinkResult` print); do not edit `root.go` yet
- [x] T005 [P] [US1] Create `internal/cli/projects.go` with `projectsCmd` copied from `internal/cli/root.go`; do not edit `root.go` yet
- [x] T006 [P] [US1] Create `internal/cli/init.go` with `initCmd`, its `func init()` registering `--force`, and scripted/TUI paths calling `project.Init`; do not edit `root.go` yet
- [x] T007 [P] [US1] Create `internal/cli/mr.go` with `mrCmd`, private `runMRFlags`, and `func init()` for `--source`, `--target`, `--title`, `-y`; do not edit `root.go` yet
- [x] T008 [P] [US1] Create `internal/cli/sync.go` with `syncCmd`, private `runSyncCLI`, and `func init()` for `--from`, `-y`; do not edit `root.go` yet
- [x] T009 [US1] Trim `internal/cli/root.go` to assembly only: keep `rootCmd` (Short, Long, default `RunE` → `tui.Run`), `Execute()`, and `init()` with `rootCmd.AddCommand(initCmd, syncCmd, mrCmd, linkCmd, unlinkCmd, projectsCmd)`; remove all subcommand bodies, helpers, and flag inits now living in the new files; fix imports
- [x] T010 [US1] Run `go build ./cmd/branchy` and `go test ./internal/cli/... -count=1`; fix duplicate symbol or init-order issues until green

**Checkpoint**: Eight CLI source files exist; contributor can open `sync.go` without scrolling past `mr` or `link`; default `./branchy` still launches main TUI

---

## Phase 4: User Story 2 - Command handlers only parse, call, and print (Priority: P1)

**Goal**: Every command `RunE` and private helper follows parse → spec-011 operation → print; no domain logic reintroduced by the move.

**Independent Test**: narrow grep `rg 'Tree\.Link|SaveTree|UnlinkSubtree|HasEdge|internal/gitlab' internal/cli/` returns no matches; read-only grep `rg 'p\.Tree\.(Names|CollectEdges)' internal/cli/` matches only `sync.go`; scripted paths match pre-split behavior (`quickstart.md` Scenarios 3, 5; SC-002, SC-003)

### Implementation for User Story 2

- [x] T011 [P] [US2] Audit `internal/cli/link.go` and `internal/cli/unlink.go` — scripted paths call `project.Link` / `project.Unlink` only; no `Tree.Link`, `UnlinkSubtree`, or `SaveTree`
- [x] T012 [P] [US2] Audit `internal/cli/init.go` — calls `project.Init` only; no tree/git logic beyond that operation
- [x] T013 [P] [US2] Audit `internal/cli/sync.go` — `runSyncCLI` calls `sync.Run`, `sync.OpenableURLs`, `sync.OpenURLs`; read-only `p.Tree.Names()` / `p.Tree.CollectEdges()` allowed for picker/plan print only; direct `IsTTY()` for interactive `--from` picker allowed; no GitLab client, tree mutation, or persist
- [x] T014 [P] [US2] Audit `internal/cli/mr.go` — `runMRFlags` calls `mr.Create` only; no GitLab imports
- [x] T015 [P] [US2] Audit `internal/cli/projects.go` — calls `project.ListAll` only; no index mutation
- [x] T016 [US2] Run two-step contract grep from `specs/012-split-cli-commands/contracts/cli-commands.md`: (1) `rg 'Tree\.Link|SaveTree|UnlinkSubtree|HasEdge|internal/gitlab' internal/cli/` must be empty; (2) `rg 'p\.Tree\.(Names|CollectEdges)' internal/cli/` must match only `internal/cli/sync.go`

**Checkpoint**: Handler thinness holds on every command file; scripted link/unlink/sync/mr/init/projects outputs unchanged

---

## Phase 5: User Story 3 - Shared TTY and prompt rules stay in one place (Priority: P2)

**Goal**: `mode.go` remains the sole owner of `UseTUI` and flag-changed detection; command files do not copy `anyFlagChanged` or duplicate mode-selection logic.

**Independent Test**: `rg 'anyFlagChanged' internal/cli/` matches only `mode.go` and `mode_test.go`; `TestSyncCmdFlagForcesCLI` passes with `syncCmd` in `sync.go` (`quickstart.md` Scenario 4); direct `IsTTY()` in `sync.go` is limited to picker UX inside `runSyncCLI`

### Implementation for User Story 3

- [x] T017 [US3] Confirm `internal/cli/mode.go` is unchanged except if an import cleanup was required; grep command files for duplicated `anyFlagChanged` / reimplemented `UseTUI` — must find none outside `mode.go`; confirm `sync.go` uses direct `IsTTY()` only for interactive `--from` picker (not a second mode rule)
- [x] T018 [US3] Run `go test ./internal/cli/... -count=1` — `TestAnyFlagChanged`, `TestUseTUIRequiresTTY`, `TestSyncCmdFlagForcesCLI`, and related mode tests must pass with split layout

**Checkpoint**: Flag-forces-scripted and TTY rules behave as before; no six copies of mode helpers

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Full regression, quickstart validation, entry-point check

- [x] T019 Confirm `cmd/branchy/main.go` is unchanged (still `cli.Execute()` + stderr exit 1)
- [x] T020 Run `go test ./... -count=1` and fix any failures outside `internal/cli` caused by the split
- [x] T021 Run `specs/012-split-cli-commands/quickstart.md` Scenarios 1–5 (help, default TUI, scripted paths, mode rules, two-step grep audit)
- [x] T022 [P] Verify `specs/012-split-cli-commands/contracts/cli-commands.md` layout table matches on-disk files: `ls internal/cli/{root,mode,init,sync,mr,link,unlink,projects}.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — MVP; T009 depends on T003–T008
- **User Story 2 (Phase 4)**: Depends on US1 (files must exist to audit)
- **User Story 3 (Phase 5)**: Depends on US1 (`syncCmd` location for mode test); can run parallel with US2 audits
- **Polish (Phase 6)**: Depends on US1–US3

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (file split) | Foundational | T002 |
| US2 (thin handlers) | US1 | T010 |
| US3 (shared mode) | US1 | T010 |
| Polish | US1–US3 | T018 |

### Within Each User Story

- T003–T008: copy to new files first; do not trim `root.go` until T009
- T009 must compile: one `var` per command across the package
- Each command file owns its `func init()` flag registration
- Outcome freeze: no flag/help/output changes during the move

### Parallel Opportunities

- **Phase 3**: T003 ∥ T004 ∥ T005 ∥ T006 ∥ T007 ∥ T008 (six new files, no `root.go` edits)
- **Phase 4**: T011 ∥ T012 ∥ T013 ∥ T014 ∥ T015 (five independent file audits)
- **Phase 5**: T017 ∥ T018 can overlap with Phase 4 audits once T010 is green
- **Phase 6**: T022 ∥ T019; T020–T021 after audits

---

## Parallel Example: User Story 1

```bash
# After T002, in parallel (six agents / six branches merged before T009):
Task T003: internal/cli/link.go
Task T004: internal/cli/unlink.go
Task T005: internal/cli/projects.go
Task T006: internal/cli/init.go
Task T007: internal/cli/mr.go
Task T008: internal/cli/sync.go
# Then sequentially:
Task T009: trim internal/cli/root.go
Task T010: go build + go test ./internal/cli/...
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 — six new files + trim `root.go`
4. **STOP and VALIDATE**: `go test ./internal/cli/...`; `quickstart.md` Scenarios 1–2
5. Demo: open `sync.go` without scrolling past other commands

### Incremental Delivery

1. Setup + Foundational → baseline green
2. US1 → file isolation (MVP)
3. US2 → handler grep audit
4. US3 → mode tests + no duplicated helpers
5. Polish → full quickstart + `go test ./...`

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: T003 link + T004 unlink
   - Developer B: T005 projects + T006 init
   - Developer C: T007 mr + T008 sync
3. One developer runs T009–T010 merge trim
4. US2/US3 audits can split by file (T011–T015)

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work in the same file
- [Story] label maps task to US1–US3 for traceability
- Outcome freeze: move code verbatim where possible; no flag/help/copy redesign
- `mode.go` / `mode_test.go` should not need logic changes — only fix imports if tests break
- Do not add `prompt.go`, `commands/` subfolder, or per-command test files
- Grilling (2026-08-21): tasks.md is an unchecked executable checklist; CLI tests = `mode_test.go` + grep only; direct `IsTTY()` in `sync.go` allowed for picker UX
- Commit after T009 (split complete) or after each story checkpoint
- Avoid: editing `root.go` in parallel during T003–T008; moving commands to `cmd/branchy`; TUI or domain changes
