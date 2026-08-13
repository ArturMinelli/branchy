---

description: "Task list for Bidirectional Sync feature implementation"
---

# Tasks: Bidirectional Sync

**Input**: Design documents from `/specs/009-bidirectional-sync/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/bidirectional-sync.md, quickstart.md

**Tests**: Included per plan.md quality gates (`internal/tree/tree_test.go`, `internal/sync/sync_test.go`, `internal/tui/syncflow_test.go`, `internal/tui/app_test.go`, CLI isolation). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/tree/tree.go`, `internal/tree/tree_test.go`
- `internal/sync/sync.go`, `internal/sync/sync_test.go`
- `internal/tui/inbound.go`, `internal/tui/syncflow.go`, `internal/tui/syncflow_test.go`, `internal/tui/app.go`, `internal/tui/app_test.go`
- `internal/cli/root.go` stays downward-only (FR-013)
- `README.md` sync section updated in polish

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before implementation

- [x] T001 Verify implementation targets in `specs/009-bidirectional-sync/plan.md` — `CollectEdgesUpward` in `internal/tree/tree.go`; `Direction` / `Ends` / `EdgesBelow` / `Result.Source|Target` in `internal/sync/sync.go`; direction + dual maps + picker `d` in `internal/tui/syncflow.go`; `s` passes `diffDirection` from `internal/tui/app.go`; no direction flag in `internal/cli/root.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared walk and MR-end helpers that every story uses

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Add `CollectEdgesUpward(root)` in `internal/tree/tree.go` as post-order (recurse sorted children, then append the edge); share sibling sort / membership with `CollectEdges` via one private walker so both return the same edge set for a given root per `specs/009-bidirectional-sync/contracts/bidirectional-sync.md`
- [x] T003 [P] Add tests in `internal/tree/tree_test.go` on a two-level-below-ceiling fixture (e.g. `develop → feat-a → leaf` plus sibling `feat-b`): `CollectEdgesUpward` membership equals `CollectEdges` as a set; `feat-a→leaf` appears before `develop→feat-a`; siblings keep existing sorted order (not a reversed pre-order slice)
- [x] T004 Add `Direction` (`Downward` default, `Upward`), `Ends(edge, dir)`, and `EdgesBelow(doc, root, dir)` in `internal/sync/sync.go`; put `Direction` on `Options`; set `Result.Source` / `Result.Target` from `Ends` in `processEdge`; `Run` uses `EdgesBelow`; change `RunEdge` to accept `Direction`; update `runEdgeCmd` in `internal/tui/syncflow.go` to pass `Downward` so the module still compiles
- [x] T005 [P] Add tests in `internal/sync/sync_test.go`: `Ends` table for both directions; `EdgesBelow` downward matches `CollectEdges` and upward matches `CollectEdgesUpward`; existing `CreatedURLs` / `OpenableURLs` order tests still pass when `Source`/`Target` are populated

**Checkpoint**: Walk + MR ends ready — TUI stories can map inbound/outbound onto them

---

## Phase 3: User Story 1 - Cascade child work up to the selected branch (Priority: P1) 🎯 MVP

**Goal**: Outbound tree + `s` on a ceiling walks that subtree deepest-first and creates child → parent MRs, stopping at the selected branch; per-edge confirm/skip/summary/browser rules stay the same

**Independent Test**: Two levels below the selected branch; flip outbound; press `s`; accept all; MRs are child → parent only, none above the ceiling, deeper edges first (`specs/009-bidirectional-sync/quickstart.md` Scenario 1)

### Implementation for User Story 1

- [x] T006 [US1] Extend `SyncFlowModel` in `internal/tui/syncflow.go` with `direction` (`sync.Direction`); `newSyncFlowModel` takes a direction (embedded uses the caller’s value); `prepareSyncFrom` calls `sync.EdgesBelow`; confirm question, loading line, and context use `Ends` (`Create MR {child} → {parent}?`, `Sync up to {ceiling}` vs `Sync down from {ceiling}`); `renderResults` prints `Source → Target`; `runEdgeCmd` passes the flow’s direction
- [x] T007 [US1] In `updateTree` in `internal/tui/app.go`, pass `m.diffDirection` mapped to `sync.Direction` (inbound→Downward, outbound→Upward) when constructing the embedded sync flow
- [x] T008 [P] [US1] Add tests in `internal/tui/syncflow_test.go` and `internal/tui/app_test.go`: outbound `s` offers deepest child→parent first; no edge above the ceiling; decline skips and continues; empty leaf ceiling shows `No child branches below` with no browser prompt; replace `TestSyncStaysInboundWhileTreeOutbound` so outbound `s` is upward (confirm has no `d: show`)

**Checkpoint**: Upward cascade MVP works from the main tree — counts and standalone picker can follow

---

## Phase 4: User Story 2 - Downward sync stays as it is today (Priority: P1)

**Goal**: Inbound tree `s` and flagged scripted `branchy sync --from` remain parent → child, top-down; opposite-direction open MRs do not skip

**Independent Test**: Inbound tree + `s` matches today’s first edge and arrow; `branchy sync --from … -y` stays downward (`quickstart.md` Scenarios 2 and 5)

### Implementation for User Story 2

- [x] T009 [US2] Confirm `internal/cli/root.go` still calls `CollectEdges` + `sync.Run` without setting `Direction` and prints parent → child; do not add a direction flag
- [x] T010 [P] [US2] Add regression tests in `internal/tui/app_test.go` and/or `internal/tui/syncflow_test.go`: inbound `s` still offers parent→child in pre-order with `Sync down from {ceiling}`
- [x] T011 [P] [US2] Grep audit: `internal/cli/root.go` has no `Upward` / `--direction` / `--up`; `branchy sync --help` documents only `--from` and `-y`

**Checkpoint**: Downward + scripted paths are unchanged (FR-001, FR-013, SC-003, SC-006)

---

## Phase 5: User Story 3 - Confirm numbers match the merge request (Priority: P1)

**Goal**: Upward confirms/picker show outbound counts (files the parent would receive); downward keeps inbound counts (files the child would receive)

**Independent Test**: Pair with inbound N ≠ outbound M; downward confirm shows N on the child; upward confirm shows M on the parent (`quickstart.md` Scenario 3)

### Implementation for User Story 3

- [x] T012 [US3] Load both `loadInboundCounts` and `loadOutboundCounts` in `newSyncFlowModel` / remote-update apply in `internal/tui/syncflow.go`; generalize `formatInboundConfirm` in `internal/tui/inbound.go` to take the receiving branch; downward confirm uses inbound[child] on the child; upward confirm uses outbound[child] on the parent (show known zero; unknown → `File count unavailable`); picker badges (when shown) use the active map via `formatFileChangeBadge`
- [x] T013 [P] [US3] Add tests in `internal/tui/syncflow_test.go`: injected maps with N ≠ M; downward confirm/picker show inbound; upward confirm/picker show outbound; known zero still appears on confirm

**Checkpoint**: Confirm numbers match the MR being created (FR-008, FR-009, SC-004)

---

## Phase 6: User Story 4 - Standalone sync can go either way (Priority: P2)

**Goal**: `branchy sync` (no flags) root picker starts downward and offers `d`; toggle flips pending direction and badges; after a root is chosen, confirms follow that direction and do not offer `d`

**Independent Test**: Launch standalone sync; starts downward; `d` flips badges; pick a root; cascade is upward (`quickstart.md` Scenario 4)

### Implementation for User Story 4

- [x] T014 [US4] On `stepSyncPickRoot` only in `internal/tui/syncflow.go`, bind `d` to flip `direction` and rebuild picker badges from the other map; default standalone direction is Downward; picker help names `sync: downward (parent→child)` / `sync: upward (child→parent)` and `d: show upward` / `d: show downward` per `contracts/bidirectional-sync.md`; confirm/summary/browser help must not list `d` as a direction toggle; direction freezes at `prepareSyncFrom`
- [x] T015 [P] [US4] Add tests in `internal/tui/syncflow_test.go`: standalone picker starts downward with inbound badges and downward help; `d` flips to upward badges/help; after root select, confirms are child→parent deepest-first and View has no picker `d: show` on the confirm step; a new `newSyncFlowModel(..., "", …)` is downward again

**Checkpoint**: Standalone interactive sync can go either way without the main tree (FR-010–FR-012, SC-005)

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Docs and full validation across all stories

- [x] T016 [P] Update the Sync behavior section and TUI keys in `README.md` to document inbound `s` = downward, outbound `s` = upward, standalone picker `d`, and scripted `--from` remaining downward-only
- [x] T017 Run `go test ./internal/tree/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1` and fix any failures
- [x] T018 [P] Run quickstart validation per `specs/009-bidirectional-sync/quickstart.md` Scenarios 1–7

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — no dependency on US2–US4
- **User Story 2 (Phase 4)**: Depends on Foundational + US1 `SyncFlowModel.direction` so inbound can be asserted against the same path
- **User Story 3 (Phase 5)**: Depends on US1 (confirm already uses `Ends` / direction)
- **User Story 4 (Phase 6)**: Depends on US1 direction field + US3 dual maps so picker badges can flip
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (upward cascade) | Foundational | Phase 2 complete |
| US2 (downward / CLI) | Foundational + T006 direction | T006 |
| US3 (matching counts) | US1 confirm path | T008 |
| US4 (standalone picker) | US1 direction + US3 maps | T013 |

### Within Each User Story

- Tree walk before `EdgesBelow`
- `Ends` / `RunEdge(dir)` before TUI confirm copy
- Embedded `s` handoff before standalone picker `d`
- Dual count maps before picker badge flip
- Do not bind `d` on confirm screens

### Parallel Opportunities

- **Phase 2**: T003 ∥ T004 after T002; T005 after T004
- **Phase 3**: T008 after T006–T007
- **Phase 4**: T010 ∥ T011 after T009
- **Phase 5**: T013 after T012
- **Phase 6**: T015 after T014
- **Phase 7**: T016 ∥ T018; T017 sequential with fixes

---

## Parallel Example: User Story 1

```bash
# After T006–T007 (syncflow + app handoff):
Task T008: syncflow_test.go + app_test.go upward offer order / ceiling / empty leaf
```

---

## Parallel Example: User Story 2

```bash
Task T010: app_test.go / syncflow_test.go inbound downward regression
Task T011: grep cli/root.go for direction flag leakage
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 — upward cascade from the tree
4. **STOP and VALIDATE**: Quickstart Scenario 1; `go test ./internal/tree/... ./internal/sync/... ./internal/tui/...`
5. Demo outbound `s` creating child → parent MRs deepest-first

### Incremental Delivery

1. Setup + Foundational → walk + `Ends` ready
2. US1 → upward cascade (MVP)
3. US2 → downward + CLI unchanged
4. US3 → confirm counts match the MR
5. US4 → standalone picker toggle
6. Polish → README + full quickstart

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (embedded upward)
   - After T006: Developer B: US2 (downward / CLI)
   - After US1: Developer C: US3 then US4 (counts + standalone)
3. Stories integrate on shared `sync.Direction` / `SyncFlowModel.direction`

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work
- [Story] label maps task to US1–US4 for traceability
- Reverse of `CollectEdges` is not a valid upward walk — use post-order
- 008 “sync stays inbound” tests must be replaced, not preserved
- Scripted CLI must never gain a direction flag
- Commit after each task or logical group
- Stop at any checkpoint to validate the story independently
