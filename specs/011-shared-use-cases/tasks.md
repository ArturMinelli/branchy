---

description: "Task list for Shared Use Cases feature implementation"
---

# Tasks: Shared Use Cases

**Input**: Design documents from `/specs/011-shared-use-cases/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/shared-operations.md, quickstart.md

**Tests**: Included per plan.md quality gates (`internal/project/project_test.go`, `internal/sync/sync_test.go`, existing TUI/CLI tests). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/project/project.go`, `internal/project/project_test.go` (new)
- `internal/sync/sync.go`, `internal/sync/sync_test.go`
- `internal/cli/root.go`
- `internal/tui/linkflow.go`, `internal/tui/unlinkflow.go`, `internal/tui/app.go`, `internal/tui/syncflow.go`, `internal/tui/initflow.go`
- `internal/tree/*` primitives stay; `internal/mr/mr.go` Create stays the create owner
- No new package; `internal/cli` file split, TUI embed, browser move, type unify stay closed (012–015)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before implementation

- [x] T001 Verify implementation targets in `specs/011-shared-use-cases/plan.md` — `project.Link` / `project.Unlink` in `internal/project/project.go`; `sync.Begin` / `sync.SkippedByUser` in `internal/sync/sync.go`; surfaces in `internal/cli/root.go` and `internal/tui/{linkflow,unlinkflow,app,syncflow}.go`; no `app`/`usecase` package; `tree.Link` / `UnlinkSubtree` stay exported; specs 012–015 files stay closed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared persist-test fixture for Link, Unlink, and Init tests

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Create `internal/project/project_test.go` with a helper that `t.Setenv("HOME", t.TempDir())` and returns a `*Project` with a stable `ID`, in-memory `*tree.Document`, and `Path` so `SaveTree` writes `~/.config/branchy/projects/<id>/branch-tree.yaml` under that HOME (see `internal/config/index_test.go` HOME pattern)

**Checkpoint**: Persist tests can reload via `tree.Load` + `config.TreePath` without touching the developer’s real config

---

## Phase 3: User Story 1 - One link operation for every surface (Priority: P1) 🎯 MVP

**Goal**: Scripted `branchy link`, standalone link confirm, and main-tree link all call `project.Link` (validate + `tree.Link` + `SaveTree`). Duplicate-edge / empty / self-link errors have one owner.

**Independent Test**: Change duplicate-edge rejection only in `project.Link` (or the primitive it calls). Scripted + standalone + main-tree accept/reject the same pairs and persist the same file (`quickstart.md` Scenario 1; SC-005)

### Implementation for User Story 1

- [x] T003 [US1] Add `(*Project) Link(parent, child string) error` in `internal/project/project.go`: call `p.Tree.Link` then `p.SaveTree`; return the first error; do not roll back in-memory mutation on save failure (research §6)
- [x] T004 [P] [US1] In `internal/project/project_test.go`, cover `Link`: success reloads the same edge from disk; empty name / self-link / duplicate edge return the existing `tree.Link` messages and do not add a second edge
- [x] T005 [P] [US1] In `internal/cli/root.go` scripted `linkCmd`, replace `p.Tree.Link` + `p.SaveTree` with `p.Link`; keep `Linked %s → %s` and the same arg/TTY rules
- [x] T006 [P] [US1] In `internal/tui/linkflow.go` confirm-Yes, call `m.project.Link` instead of `Tree.Link` + `SaveTree`; still put `err.Error()` on `stepLinkError`
- [x] T007 [P] [US1] In `internal/tui/app.go` `updateLink` Enter, call `m.current.Link` instead of `Tree.Link` + `SaveTree`; keep cancel/back with no persist

**Checkpoint**: Three link surfaces, one persist owner — `go test ./internal/project/... ./internal/tui/ ./internal/cli/` still green for link flows

---

## Phase 4: User Story 2 - One unlink operation, including validation (Priority: P1)

**Goal**: One `project.Unlink(parent, child)` that requires the edge when parent is set, allows empty parent for a root, persists, and returns removed count. Interactive paths pass `ParentOf(child)`. Scripted path drops its private validation block.

**Independent Test**: Unlink the same parent→child from CLI and both interactive paths — same members gone, same missing-edge / unknown-branch rejection (`quickstart.md` Scenario 2; SC-004)

### Implementation for User Story 2

- [x] T008 [US2] Add `UnlinkResult` and `(*Project) Unlink(parent, child string) (*UnlinkResult, error)` in `internal/project/project.go` per `specs/011-shared-use-cases/contracts/shared-operations.md` (edge check, root path, `Removed` = pre-delete `len(SubtreeNames(child))`, then `UnlinkSubtree` + `SaveTree`)
- [x] T009 [P] [US2] In `internal/project/project_test.go`, cover `Unlink`: missing child; parent set but not an edge (`edge not found: P → C`); success count + reload; empty parent on a root succeeds; empty parent on a non-root errors; tree unchanged on disk when validation fails
- [x] T010 [P] [US2] In `internal/cli/root.go` scripted `unlinkCmd`, delete the private empty/self/not-in-tree/`HasEdge`/count block; call `p.Unlink(parent, child)` and print `Unlinked %s → %s (%d branches removed)` from `UnlinkResult`
- [x] T011 [P] [US2] In `internal/tui/unlinkflow.go` confirm-Yes, `parent, _ := m.project.Tree.ParentOf(m.target)` then `m.project.Unlink(parent, m.target)`; use `Removed` for the success line; keep cancel with no persist
- [x] T012 [P] [US2] In `internal/tui/app.go` `updateUnlink` confirm-Yes, same `ParentOf` + `Unlink` as T011 (no `UnlinkSubtree` + `SaveTree`); keep `TestUpdateUnlinkConfirmRemovesSubtree` in `internal/tui/app_test.go` passing (in-memory mutate still happens before save failure)

**Checkpoint**: Interactive unlink of a non-edge now fails like CLI; TUI root unlink still works

---

## Phase 5: User Story 3 - Sync and merge-request create own GitLab auth and edge work (Priority: P1)

**Goal**: Interactive sync starts via `sync.Begin` (auth + collect when there are edges) instead of `gitlab.Client` in the TUI. Decline uses `sync.SkippedByUser`. `Run` / `RunEdge` / `mr.Create` stay the processing/create owners. TUI does not import `internal/gitlab`.

**Independent Test**: Logged-out glab: scripted and interactive sync/MR fail with `glab auth` meaning and create 0 MRs. Decline one TUI edge → skipped by user, cascade continues (`quickstart.md` Scenarios 4–5)

### Implementation for User Story 3

- [x] T013 [US3] In `internal/sync/sync.go`, add `Begin(p *project.Project, from string, dir Direction) ([]tree.Edge, error)`: same from-branch errors as `Run`; `EdgesBelow`; **AuthOK only if `len(edges) > 0`**; wrap auth errors as `glab auth: %w (run: glab auth login)`. Add `SkippedByUser(edge tree.Edge, dir Direction) Result` matching `processEdge` when Confirm returns false (`Action` skipped, `Message` `skipped by user`, ends from `Ends`)
- [x] T014 [P] [US3] In `internal/sync/sync_test.go`, test `Begin`: unknown from-branch errors; zero edges → empty slice, nil error (no GitLab). Test `SkippedByUser` Parent/Child/Source/Target/Action/Message (downward and upward)
- [x] T015 [US3] In `internal/tui/syncflow.go` `prepareSyncFrom`, call `sync.Begin` instead of `EdgesBelow` + `gitlab.Client.AuthOK`; map Begin error to `stepSyncError`; empty edges → `stepSyncEmpty`. On confirm-No, append `sync.SkippedByUser(...)` instead of a hand-built `sync.Result`. Remove the `gitlab` import
- [x] T016 [P] [US3] Update `internal/tui/syncflow_test.go` so decline still records skipped-by-user and continues; empty-edge fixture still reaches `stepSyncEmpty` without needing GitLab; keep existing confirm/summary/browser tests green

**Checkpoint**: `grep gitlab internal/tui` is empty; `mr.Create` remains the only create path (no signature change in `internal/mr/mr.go`)

---

## Phase 6: User Story 4 - Init stays one registration path (Priority: P2)

**Goal**: Lock `project.Init` as the only register operation. Add persist tests so already-registered / `--force` cannot silently fork. Interactive wizard stays confirm-then-`Init`.

**Independent Test**: Scripted `init` and interactive confirm produce the same id/path/tree on a fresh repo; without force both refuse already-registered; `--force` reuses id (`quickstart.md` Scenario 3)

### Implementation for User Story 4

- [x] T017 [US4] In `internal/project/project_test.go`, add `Init` tests with a temp git repo (`git init`) + `t.Setenv("HOME", ...)` + `t.Chdir`: first Init registers; second without Force returns `already registered as %q (use --force to re-import)`; Force reuses the same id
- [x] T018 [P] [US4] Confirm `internal/tui/initflow.go` `runInitCmd` still calls `project.Init(InitOptions{Force: false})` and cancel/No never calls it (`internal/tui/initflow_test.go`); confirm `internal/cli/root.go` `initCmd` still calls `project.Init` with the `--force` flag — do not add a second register helper

**Checkpoint**: Init has tests at the operation; TUI is still only a confirm wrapper

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Call-site audit, full test run, quickstart

- [x] T019 Grep audit: no `Tree.Link` + `SaveTree` and no `UnlinkSubtree` + `SaveTree` in `internal/cli/` or `internal/tui/`; no `internal/gitlab` import under `internal/tui/`; `sync.OpenURLs` still the browser helper (do not move it — spec 014)
- [x] T020 Run `go test ./internal/project/... ./internal/sync/... ./internal/mr/... ./internal/tui/... ./internal/cli/... ./internal/tree/... -count=1` and fix failures
- [x] T021 [P] Run quickstart validation per `specs/011-shared-use-cases/quickstart.md` Scenarios 1–5

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — MVP
- **User Story 2 (Phase 4)**: Depends on US1 because `internal/project/project.go`, `internal/cli/root.go`, and `internal/tui/app.go` are shared
- **User Story 3 (Phase 5)**: Depends on Foundational only — **can run in parallel with US1/US2** (`internal/sync` + `syncflow.go`)
- **User Story 4 (Phase 6)**: Depends on US2 so `project_test.go` Init tests land after Link/Unlink tests in the same file
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (link) | Foundational | T002 |
| US2 (unlink) | US1 (shared files) | T007 |
| US3 (sync/MR auth) | Foundational | T002 (parallel with US1) |
| US4 (init lock) | US2 (`project_test.go`) | T012 |

### Within Each User Story

- Operation function before surface wiring
- Persist tests before considering the story done
- TUI/CLI must call the owner, not re-validate then save
- `Begin` empty-edge path must not call GitLab
- Do not roll back in-memory tree on `SaveTree` failure

### Parallel Opportunities

- **Phase 3**: After T003, T004 ∥ T005 ∥ T006 ∥ T007 (four files)
- **Phase 4**: After T008, T009 ∥ T010 ∥ T011 ∥ T012
- **Phase 5**: T014 after T013; T016 after T015; US3 whole phase ∥ US1 if staffed
- **Phase 6**: T018 ∥ T017 only if T017 is not still editing `project_test.go` — prefer T018 after T017
- **Phase 7**: T021 ∥ T019; T020 after T019 fixes

---

## Parallel Example: User Story 1

```bash
# After T003 (project.Link exists):
Task T004: project_test.go Link persist + validation
Task T005: cli/root.go scripted link → p.Link
Task T006: linkflow.go confirm → project.Link
Task T007: app.go main-tree link → current.Link
```

---

## Parallel Example: User Story 3

```bash
# After T002, in parallel with US1:
Task T013: sync.go Begin + SkippedByUser
Task T014: sync_test.go (after T013)
# Then:
Task T015: syncflow.go wire Begin + skip helper, drop gitlab
Task T016: syncflow_test.go (after T015)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 — `project.Link` + three surfaces
4. **STOP and VALIDATE**: `go test ./internal/project/... ./internal/tui/ ./internal/cli/`; quickstart Scenario 1
5. Demo: duplicate-edge fails the same from CLI and TUI

### Incremental Delivery

1. Setup + Foundational → persist fixture ready
2. US1 → one link owner (MVP)
3. US2 → one unlink owner + interactive edge check
4. US3 → TUI drops GitLab client; shared skip + Begin
5. US4 → Init tests lock the existing owner
6. Polish → grep + full quickstart

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 then US2 then US4 (project/cli/app)
   - Developer B: US3 (sync + syncflow)
3. Integrate on `project.Link`/`Unlink` and `sync.Begin`/`SkippedByUser` only

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work
- [Story] label maps task to US1–US4 for traceability
- Outcome freeze except interactive unlink gaining the scripted edge check
- `tree.Link` / `UnlinkSubtree` remain for primitives and `project` internals — not for CLI/TUI
- Interactive root unlink: empty parent; do not make roots fail
- `sync.Run` still auths even when the subtree is empty (CLI unchanged); TUI empty screen still skips auth
- Commit after each task or logical group
- Stop at any checkpoint to validate the story independently
- Avoid: new packages, Cobra splits, embedding main-tree wizards, moving `OpenURLs`, renaming created/skipped/failed
