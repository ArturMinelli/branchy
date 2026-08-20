---

description: "Task list for Merge Direction Arrows feature implementation"
---

# Tasks: Merge Direction Arrows

**Input**: Design documents from `/specs/010-direction-arrows/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/direction-arrows.md, quickstart.md

**Tests**: Included per plan.md quality gates (`internal/tui/treeview_test.go`, `internal/tui/app_test.go`, `internal/tui/inbound_test.go`, `internal/tui/syncflow_test.go`). No TDD-first ordering required.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/tui/treeview.go`, `internal/tui/treeview_test.go`
- `internal/tui/inbound.go`, `internal/tui/inbound_test.go`
- `internal/tui/app.go`, `internal/tui/app_test.go`
- `internal/tui/syncflow.go`, `internal/tui/syncflow_test.go`
- `README.md` keys + sync section updated in polish
- `internal/git/*`, `internal/sync/*`, `internal/tree/*`, `internal/cli/*` stay closed

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before implementation

- [x] T001 Verify implementation targets in `specs/010-direction-arrows/plan.md` — participating flag + two-cell `↓ `/`↑ `/pad column in `internal/tui/treeview.go`; `treeHelpFooter` in `internal/tui/inbound.go`; `DirectionUp`/`DirectionDown` (`ctrl+up`/`ctrl+down`) replacing `d` in `internal/tui/app.go`; same set-chords on standalone picker only in `internal/tui/syncflow.go`; no git/sync/cli/tree-walk changes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared row metadata and glyph-column helper that US1 rendering and later key tests reuse

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Add `participating` on `branchRow` in `internal/tui/treeview.go` during `collectRows` (`true` if the row has a parent or at least one child; childless roots `false`) and add `directionColumn(direction, participating)` returning `"↓ "`, `"↑ "`, or two spaces per `specs/010-direction-arrows/contracts/direction-arrows.md`
- [x] T003 [P] Extend `TestFlattenTree` (or add a sibling test) in `internal/tui/treeview_test.go` so `main`/`release`/`develop` are participating and `orphan` is not, on the existing four-row fixture

**Checkpoint**: Flatten knows who gets an arrow — `renderRow` can paint the column

---

## Phase 3: User Story 1 - Read merge direction from the tree itself (Priority: P1) 🎯 MVP

**Goal**: Every participating main-tree row shows a muted `↓` (inbound) or `↑` (outbound) before the name; childless roots are padded so names line up; counts and `▸` stay as today

**Independent Test**: Tree with a root that has children and a childless root; default inbound shows `↓` on participating rows only; `setDirection(outbound)` flips those to `↑`; zeros/`?` still show the glyph (`specs/010-direction-arrows/quickstart.md` Scenario 1)

### Implementation for User Story 1

- [x] T004 [US1] In `renderRow` in `internal/tui/treeview.go`, insert `directionColumn` immediately after the connector and before the branch name; paint it with `connectorStyle`; keep file-change badges after the name; keep `▸ ` as the selection marker only
- [x] T005 [P] [US1] Add tests in `internal/tui/treeview_test.go`: inbound `↓` on participating rows; outbound `↑`; childless root has two-space pad and no `↓`/`↑`; zero and unknown counts still show the glyph; selected row still contains `▸` distinct from the direction glyph

**Checkpoint**: Direction is readable from the tree without the footer — `d` still toggles until US2

---

## Phase 4: User Story 2 - Set direction with Control+Up and Control+Down (Priority: P1)

**Goal**: Control+Up sets outbound; Control+Down sets inbound; already-active chord is a no-op; plain Up/Down/`j`/`k` only move the cursor; `d` no longer changes direction

**Independent Test**: From inbound, Control+Up → outbound glyphs and badges; Control+Up again stays outbound; Control+Down restores inbound; `d` and plain Up do not flip (`quickstart.md` Scenario 2)

### Implementation for User Story 2

- [x] T006 [US2] In `internal/tui/app.go`, replace `keys.Direction` (`d`) with `DirectionUp` (`ctrl+up`) and `DirectionDown` (`ctrl+down`); in `updateTree` assign `diffOutbound` / `diffInbound` (set, not flip) and `treeView.setDirection`; match these bindings **before** `keys.Up` / `keys.Down`
- [x] T007 [P] [US2] Rewrite direction key tests in `internal/tui/app_test.go`: `tea.KeyMsg{Type: tea.KeyCtrlUp}` / `KeyCtrlDown` set direction and badges (and tree arrows if visible in `View`); second `KeyCtrlUp` stays outbound; `KeyRunes{'d'}` leaves direction unchanged; plain `KeyUp` moves the cursor only — do not yet require the new footer strings (US3)

**Checkpoint**: Tree direction is set by Control+arrows, not `d` (FR-005–FR-008, SC-003)

---

## Phase 5: User Story 3 - Help names the chords, not `d` (Priority: P1)

**Goal**: Main-tree help always lists `ctrl+↑: outbound` and `ctrl+↓: inbound` plus the current `counts: inbound|outbound …` cue; `d: show …` is gone

**Independent Test**: Read the footer inbound; Control+Up; footer names outbound and still lists both chords, never `d` as a direction key (`quickstart.md` Scenario 1 footer + Scenario 2)

### Implementation for User Story 3

- [x] T008 [US3] Update `treeHelpFooter` in `internal/tui/inbound.go` to list both set-chords in both modes (`ctrl+↑: outbound` / `ctrl+↓: inbound`) and keep `counts: inbound (parent→child)` / `counts: outbound (child→parent)` per `specs/010-direction-arrows/contracts/direction-arrows.md`; drop `d: show outbound` / `d: show inbound`
- [x] T009 [US3] Update footer tests in `internal/tui/inbound_test.go` and the main-tree `View` assertions in `internal/tui/app_test.go` so both modes contain the new chords and mode cue and contain no `d: show`

**Checkpoint**: First-time users can discover the chords from the footer (FR-009, SC-005)

---

## Phase 6: User Story 4 - Standalone sync picker uses the same chords (Priority: P2)

**Goal**: Standalone picker starts downward; Control+Up / Control+Down set pending direction and badges (not a toggle); no tree arrows on list rows; `d` is unbound; confirms ignore direction chords

**Independent Test**: `branchy sync` picker starts downward; Control+Up → upward badges/help; Control+Up again stays upward; `d` does nothing; after a root, confirms have no direction keys (`quickstart.md` Scenario 5)

### Implementation for User Story 4

- [x] T010 [US4] In `internal/tui/syncflow.go`, replace picker `d` / `flipPickerDirection` with `ctrl+up` → `sync.Upward` and `ctrl+down` → `sync.Downward` (set, rebuild badges from the active map); handle these in `updatePicker` before `branchList.Update`; update `syncPickerHelp` to the same `ctrl+↑: outbound` / `ctrl+↓: inbound` set-chords; do not bind them on confirm/summary/browser steps; picker rows stay name + badge only
- [x] T011 [P] [US4] Update tests in `internal/tui/syncflow_test.go`: standalone picker starts downward; `KeyCtrlUp` sets upward badges/help; second `KeyCtrlUp` stays upward; `KeyRunes{'d'}` does not flip; picker `View` has no tree-style `↓ ` / `↑ ` prefix; after root select, confirm `View` has no `ctrl+↑` / `d: show`; a fresh `newSyncFlowModel` is downward again

**Checkpoint**: Standalone interactive sync uses the same chords without tree arrows (FR-013–FR-014, SC-006)

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Docs, leftover-`d` audit, and full validation

- [x] T012 [P] Update TUI keys and the Sync behavior section in `README.md`: Control+Up / Control+Down set outbound/inbound; main tree shows `↓`/`↑` on participating branches; standalone picker uses the same chords; remove `d` as a direction key
- [x] T013 Grep audit: no `d: show` in `internal/tui/`; `keys.Direction` / `syncKeys.Direction` are not bound to `"d"`; `internal/cli/` still has no direction flag
- [x] T014 Run `go test ./internal/tui/... -count=1` and fix any failures
- [x] T015 [P] Run quickstart validation per `specs/010-direction-arrows/quickstart.md` Scenarios 1–5

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — no dependency on US2–US4
- **User Story 2 (Phase 4)**: Depends on US1 so Control+Up also flips visible arrows, not only badges
- **User Story 3 (Phase 5)**: Depends on US2 so footer copy matches the live chords; T009 also needs T007’s `app_test.go` rewrite
- **User Story 4 (Phase 6)**: Depends on US2 set-chord behavior (same mapping) and US3 help wording
- **Polish (Phase 7)**: Depends on all user stories

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (tree arrows) | Foundational | Phase 2 complete |
| US2 (Control+arrows) | US1 glyphs | T005 |
| US3 (footer copy) | US2 keys | T007 |
| US4 (standalone picker) | US2 set-chords + US3 copy | T009 |

### Within Each User Story

- Participating flag before `renderRow` glyph
- Glyph column before key tests that assert `View` arrows
- Set-chords before footer strings that name them
- Picker keys handled before `branchList.Update`
- Do not bind Control+Up / Control+Down / `d` on confirm screens

### Parallel Opportunities

- **Phase 2**: T003 after T002
- **Phase 3**: T005 after T004
- **Phase 4**: T007 after T006
- **Phase 5**: T009 after T008 and T007 (`app_test.go` shared with T007)
- **Phase 6**: T011 after T010
- **Phase 7**: T012 ∥ T015; T013 then T014 sequential with fixes

---

## Parallel Example: User Story 1

```bash
# After T004 (renderRow glyph column):
Task T005: treeview_test.go inbound/outbound glyphs, pad, zero/unknown, ▸ distinct
```

---

## Parallel Example: User Story 2

```bash
# After T006 (app.go set-chords):
Task T007: app_test.go KeyCtrlUp / KeyCtrlDown / d no-op / plain KeyUp
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 — arrows on the main tree
4. **STOP and VALIDATE**: Quickstart Scenario 1; `go test ./internal/tui/...`
5. Demo inbound `↓` / outbound `↑` (existing `d` still flips until US2)

### Incremental Delivery

1. Setup + Foundational → participating + glyph helper ready
2. US1 → tree arrows (MVP)
3. US2 → Control+Up / Control+Down set direction; `d` gone
4. US3 → footer names both chords
5. US4 → standalone picker same chords, no arrows
6. Polish → README + grep + full quickstart

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (tree glyphs)
   - After T005: Developer B: US2 then US3 (keys + footer)
   - After T009: Developer C: US4 (picker)
3. Stories integrate on `diffDirection` / `directionColumn` / the same `ctrl+up`/`ctrl+down` mapping

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work
- [Story] label maps task to US1–US4 for traceability
- Control+Up/Down are **set**, not toggle — a second Control+Up must stay outbound
- No fallback `d` or Shift+arrows
- Glyph column is always two cells (`↓ `/`↑ `/pad); style matches connectors
- Scripted CLI must never gain a direction flag
- Commit after each task or logical group
- Stop at any checkpoint to validate the story independently
