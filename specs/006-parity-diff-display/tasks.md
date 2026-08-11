---

description: "Task list for Parent Diff Display feature implementation"
---

# Tasks: Parent Diff Display

**Input**: Design documents from `/specs/006-parity-diff-display/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/inbound-count.md, quickstart.md

**Tests**: Included per plan.md quality gates (`internal/git/inbound_test.go`, `internal/tui/inbound_test.go`, extended `treeview_test.go` / `syncflow_test.go`). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/git/inbound.go`, `internal/git/inbound_test.go`
- `internal/tui/inbound.go`, `internal/tui/inbound_test.go`, `internal/tui/treeview.go`, `internal/tui/app.go`, `internal/tui/mrflow.go`, `internal/tui/syncflow.go`
- `internal/cli/*` and `internal/sync/sync.go` stay unchanged (FR-010)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before implementation

- [x] T001 Verify implementation targets in `specs/006-parity-diff-display/plan.md` — new `internal/git/inbound.go` and `internal/tui/inbound.go`; modify `internal/tui/treeview.go`, `internal/tui/app.go`, `internal/tui/mrflow.go`, and `internal/tui/syncflow.go` only; no behavior changes to `internal/sync/sync.go` or `internal/cli/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Local GitLab-comparable comparison API and shared TUI formatters used by every story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Implement `InboundFiles(dir, parent, child string) (int, error)` as three-dot `git diff --name-only <child>...<parent>` line count in `internal/git/inbound.go` per `specs/006-parity-diff-display/contracts/inbound-count.md`
- [x] T003 [P] Add temp-repo tests in `internal/git/inbound_test.go`: N files only on parent → N; identical trees → 0; missing ref → error (not 0); files only on child after diverge → 0 inbound
- [x] T004 Implement `inboundCount`, `loadInboundCounts`, `formatInboundBadge`, and `formatInboundConfirm` in `internal/tui/inbound.go` — walk `tree.Document` via `ParentOf`, map git errors to unknown (never 0), omit roots
- [x] T005 Add formatter tests in `internal/tui/inbound_test.go`: badge hides zero, shows `N`, shows `?` for unknown; confirm line includes known zero and `File count unavailable`

**Checkpoint**: Comparison + display helpers ready — tree, picker, and confirm can be wired independently

---

## Phase 3: User Story 1 - See pending sync size on the main tree (Priority: P1) 🎯 MVP

**Goal**: Each non-root tree row shows inbound file-change count vs its parent; hide zeros; roots have no inbound badge

**Independent Test**: Open the main TUI on a project whose children have known different inbound sizes; behind children show N, in-sync children show no count, roots show no inbound count (`specs/006-parity-diff-display/quickstart.md` Scenario 1)

### Implementation for User Story 1

- [x] T006 [US1] Attach an inbound count map to `BranchTreeView` and render a muted `N` / hidden-zero / no-root badge in `internal/tui/treeview.go` using `formatInboundBadge`
- [x] T007 [US1] Compute `loadInboundCounts(p.Path, p.Tree)` once in `selectProject` and pass it into the tree view in `internal/tui/app.go` (do not recompute in `View()`)
- [x] T008 [P] [US1] Add tree badge tests in `internal/tui/treeview_test.go`: positive N visible, zero hidden, root has no badge

**Checkpoint**: Main tree alone is a usable MVP — largest pending sync target is visible without starting sync

---

## Phase 4: User Story 2 - See pending sync size while choosing a sync root (Priority: P1)

**Goal**: Sync root picker shows the same inbound badges as the main tree (hide zero, no badge on roots)

**Independent Test**: Open embedded `s` or `branchy sync`; listed non-root behind branches show the same N as the tree; roots and in-sync children have no badge (`quickstart.md` Scenario 2)

### Implementation for User Story 2

- [x] T009 [US2] Add an optional badge field to `branchItem` so `Title()` is `name` or `name  badge` and `FilterValue()` stays the bare name in `internal/tui/mrflow.go`; keep `newBranchList` badges empty so link/unlink/MR pickers stay name-only
- [x] T010 [US2] Load `loadInboundCounts` when starting sync and apply badges on the root picker list in `internal/tui/syncflow.go`
- [x] T011 [P] [US2] Add picker badge assertions in `internal/tui/syncflow_test.go`: non-zero child shows N, zero/root have no badge, filter value is the bare name

**Checkpoint**: Sync root picker matches main-tree inbound numbers (FR-002, FR-005)

---

## Phase 5: User Story 3 - See pending sync size on each edge confirm (Priority: P1)

**Goal**: Each parent→child confirm includes the inbound file-change count, including a known zero

**Independent Test**: Walk sync to an edge confirm; context includes `{N} files would change on {child}` matching the tree/picker number; in-sync edges still show `0 files would change` (`quickstart.md` Scenario 3)

### Implementation for User Story 3

- [x] T012 [US3] Add `formatInboundConfirm` as a confirm context line in `resetEdgeConfirm` in `internal/tui/syncflow.go` (keep question `Create MR {parent} → {child}?`; always show known zero)
- [x] T013 [P] [US3] Add edge-confirm context assertions in `internal/tui/syncflow_test.go`: positive N, known zero text, same integer as the picker badge for that child

**Checkpoint**: Every comparable edge confirm shows inbound size before yes/no (FR-003)

---

## Phase 6: User Story 4 - Stay usable when comparison is unavailable (Priority: P2)

**Goal**: Missing/stale refs show a clear unknown placeholder (not a fake zero); other edges still show numbers; navigation and sync keep working

**Independent Test**: Remove a local branch still listed in the tree; main TUI and sync show `?` / `File count unavailable` for that child only; other counts remain; quit and skip still work (`quickstart.md` Scenario 4)

### Implementation for User Story 4

- [x] T014 [US4] Add unknown-badge assertions (`?`, not `0`) for a failed child while siblings still show N in `internal/tui/treeview_test.go`
- [x] T015 [P] [US4] Add unknown picker `?` and confirm `File count unavailable` assertions, plus a successful sibling still showing N, in `internal/tui/syncflow_test.go`
- [x] T016 [P] [US4] Confirm scripted paths stay count-free: no inbound helpers imported from `internal/cli/` or `internal/sync/sync.go` (comment at top of `internal/tui/inbound.go` that counts are TUI-only)

**Checkpoint**: Unknown state is recognizable and non-blocking (FR-008, SC-004, SC-005)

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [x] T017 Run `go test ./internal/git/... ./internal/tui/... -count=1` and fix any failures
- [x] T018 [P] Run quickstart validation per `specs/006-parity-diff-display/quickstart.md` Scenarios 1–5
- [x] T019 [P] Grep audit: `internal/cli/` and `internal/sync/sync.go` have no inbound-count display; badges stay compact enough for width 80 in `internal/tui/treeview.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — no dependency on US2/US3/US4
- **User Story 2 (Phase 4)**: Depends on Foundational; can start in parallel with US1 (different files except shared helpers already done)
- **User Story 3 (Phase 5)**: Depends on Foundational; same file as US2 (`syncflow.go`) — implement after T010
- **User Story 4 (Phase 6)**: Depends on US1–US3 display wiring so unknown paths can be asserted on real surfaces
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (main tree) | Foundational | Phase 2 complete |
| US2 (sync picker) | Foundational | Phase 2 complete (parallel with US1) |
| US3 (edge confirm) | Foundational + US2 picker load in `syncflow.go` | T010 |
| US4 (unknown) | US1–US3 surfaces | T008, T011, T013 |

### Within Each User Story

- Foundational helpers before surface wiring
- Surface implementation before that surface’s tests
- Do not recompute git diffs in `View()`

### Parallel Opportunities

- **Phase 2**: T003 ∥ T004 after T002 (test file vs TUI helper file)
- **Phase 3**: T007 ∥ T008 after T006 (`app.go` vs `treeview_test.go`)
- **Phase 4**: T011 after T010; US2 can overlap US1 (different files)
- **Phase 5**: T013 after T012
- **Phase 6**: T015 ∥ T016 after wiring
- **Phase 7**: T018 ∥ T019

---

## Parallel Example: User Story 1

```bash
# After T006 (treeview render):
Task T007: app.go selectProject cache
Task T008: treeview_test.go badge cases
```

---

## Parallel Example: User Stories 1 + 2

```bash
# After Phase 2, different files:
Task T006: treeview.go
Task T009: mrflow.go branchItem badge field
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 — tree badges
4. **STOP and VALIDATE**: Quickstart Scenario 1; `go test ./internal/git/... ./internal/tui/...`
5. Demo glanceable inbound size on the main tree before sync wiring

### Incremental Delivery

1. Setup + Foundational → `InboundFiles` + formatters ready
2. US1 → main tree badges (MVP)
3. US2 → sync picker badges
4. US3 → edge confirm context (including known zero)
5. US4 → unknown placeholder, no fake zeros
6. Polish → full quickstart + test suite

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 tree + app (`treeview.go`, `app.go`)
   - Developer B: US2/US3 sync (`mrflow.go`, `syncflow.go`)
3. Merge, then US4 tests on both surfaces

---

## Notes

- Metric is **files changed**, not commits (`git diff --name-only child...parent`)
- Hide zero on tree/picker; **show** known zero on edge confirm
- Unknown is `?` / `File count unavailable` — never coerce git errors to 0
- `newBranchList` default stays name-only so MR/link/unlink pickers do not gain badges
- Sync results summary is out of scope
- Commit after each task or logical group; stop at any checkpoint to validate the story independently
