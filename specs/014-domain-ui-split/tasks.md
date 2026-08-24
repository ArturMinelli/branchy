---

description: "Task list for Domain / UI Split feature implementation"
---

# Tasks: Domain / UI Split

**Input**: Design documents from `/specs/014-domain-ui-split/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/domain-ui-split.md, quickstart.md

**Tests**: With implementation — no TDD-first ordering (grilling 2026-08-24). Walk tests go in existing `internal/tree/tree_test.go`. OpenURLs tests move to new `internal/browser/open_test.go`. Glyph/badge outcomes stay in existing `internal/tui/treeview_test.go`.

**Organization**: Tasks grouped by user story. **MVP is US1 only** (grilling). US2 (browser helper) and US3 (GitLab grep) are independent and can overlap US1 on different files. Spec 015 stays closed.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/tree/tree.go` — `WalkDisplay` + `DisplayNode`
- `internal/tree/tree_test.go` — walk membership / order / `IsLast` / `LastAtDepth`
- `internal/tree/render.go` — delete
- `internal/tui/treeview.go` — `flattenTree` maps the walk to glyphs
- `internal/tui/treeview_test.go` — unchanged connector/badge outcomes
- `internal/tui/syncflow.go` — `browser.OpenURLs` after confirm
- `internal/browser/open.go` — `Open` + `OpenURLs` + test hook
- `internal/browser/open_test.go` — new; sequential open + warning
- `internal/sync/sync.go` — keep `OpenableURLs`; remove `OpenURLs`
- `internal/sync/sync_test.go` — drop `OpenURLs` tests
- `internal/cli/sync.go` — prompt unchanged; open via `browser.OpenURLs`
- Spec 015 files stay closed

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before edits

- [x] T001 Verify implementation targets in `specs/014-domain-ui-split/plan.md` — delete `internal/tree/render.go`; add `WalkDisplay` on `Document` in `internal/tree/tree.go`; `flattenTree` in `internal/tui/treeview.go` consumes it; move `OpenURLs` to `internal/browser/open.go`; `OpenableURLs` stays in `internal/sync/sync.go`; US3 is grep-only; `specs/015-unify-domain-types/` stays closed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Green baseline so story diffs are attributable

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Run `go test ./internal/tree/... ./internal/tui/... ./internal/sync/... ./internal/browser/... ./internal/cli/... -count=1` and confirm green as the pre-split baseline

**Checkpoint**: Affected packages pass; `CollectEdges` remains the sync-edge walk and is not merged with display

---

## Phase 3: User Story 1 - Branch tree meaning stays in the tree; drawing stays in the UI (Priority: P1) 🎯 MVP

**Goal**: Delete unused styled `RenderASCII`. One unstyled `WalkDisplay` on `Document`. Interactive tree paints glyphs from that walk. Outcome freeze on membership, child order, connectors, badges.

**Independent Test**: Change a lipgloss color in `internal/tui/treeview.go` without editing `internal/tree/tree.go` or persist/link/unlink. `WalkDisplay` on a fixture matches interactive membership and sorted child order. `rg 'RenderASCII|lipgloss' internal/tree` is empty. (`quickstart.md` Scenario 1; SC-001/004)

### Implementation for User Story 1

- [x] T003 [P] [US1] Delete `internal/tree/render.go` (`RenderASCII` and lipgloss). Confirm nothing in the module still imports `github.com/charmbracelet/lipgloss`
- [x] T004 [P] [US1] Add `DisplayNode` (`Name`, `Depth`, `IsRoot`, `IsLast`, `LastAtDepth`) and `Document.WalkDisplay()` in `internal/tree/tree.go` — DFS pre-order from `Roots()`, `sort.Strings` on each node’s children (same as today’s `flattenTree` / `CollectEdges`); empty document → empty slice; no box-drawing characters. Add tests in `internal/tree/tree_test.go` for membership, root/child sort, `IsLast`, and `LastAtDepth` length == `Depth`
- [x] T005 [US1] Rewrite `flattenTree` in `internal/tui/treeview.go` to map `doc.WalkDisplay()` → `[]branchRow` (glyphs only here: root → no connector; last → `└── `; else `├── `; spine from `LastAtDepth` → `"    "` / `"│   "`). Delete local `collectRows` parent/child recursion. Do not change `View`, badges, or selection
- [x] T006 [US1] Run `go test ./internal/tree/... ./internal/tui/... -count=1` and keep existing `internal/tui/treeview_test.go` connector/badge cases green (same `└── ` / `├── ` and counts)

**Checkpoint**: Document has no styling; one display walk; interactive tree looks the same. **MVP is shippable here** without US2/US3

---

## Phase 4: User Story 2 - Sync returns URLs; surfaces open the browser (Priority: P1)

**Goal**: Sequential tab-opening lives in `browser.OpenURLs`. Sync still exposes `OpenableURLs` and never opens a browser. CLI and TUI still ask, then call the helper. Pause 200ms, warning wording, and decline/accept unchanged.

**Independent Test**: Decline open prompt → no tabs, summary still lists URLs. Accept → same order and pause. `rg 'func OpenURLs' internal/sync` empty; `rg 'browser.OpenURLs' internal/cli/sync.go internal/tui/syncflow.go` both match. (`quickstart.md` Scenarios 2–4; SC-003)

### Implementation for User Story 2

- [x] T007 [US2] Move sequential open from `internal/sync/sync.go` (`OpenURLs`, `browserOpen`, 200ms pause, `could not open %s: %v` joined by `"; "`) to `internal/browser/open.go` as `OpenURLs`; keep `Open` for single-URL MR. Add test hook `var openURL = Open`. Create `internal/browser/open_test.go` with the sequential-order and warning cases currently in `internal/sync/sync_test.go` (`TestOpenURLsEmpty`, `TestOpenURLsSequentialAndWarning`)
- [x] T008 [P] [US2] In `internal/cli/sync.go`, after the existing [y/N] prompt, call `browser.OpenURLs(urls)` instead of `sync.OpenURLs`; do not change prompt copy or when the prompt appears
- [x] T009 [P] [US2] In `internal/tui/syncflow.go`, on accept call `browser.OpenURLs` instead of `sync.OpenURLs`; keep confirm copy, loading text, and non-fatal warning display. Leave `internal/tui/mrflow.go` on single `browser.Open`
- [x] T010 [US2] Remove `OpenURLs`, `browserOpen`, and the `internal/browser` import from `internal/sync/sync.go`; delete `TestOpenURLsEmpty` and `TestOpenURLsSequentialAndWarning` from `internal/sync/sync_test.go`; keep `OpenableURLs` and its tests
- [x] T011 [US2] Run `go test ./internal/browser/... ./internal/sync/... ./internal/cli/... ./internal/tui/... -count=1` until green including existing decline/accept syncflow cases

**Checkpoint**: Sync does not import `browser`; both surfaces open via `browser.OpenURLs`; MR still uses `Open`

---

## Phase 5: User Story 3 - Interactive screens never start a GitLab session (Priority: P1)

**Goal**: Verification only (grilling). No expected code change. Confirm zero GitLab construction in TUI; file-change badges still use local git.

**Independent Test**: `rg 'gitlab' internal/tui` has no matches. Logged-out interactive sync still shows the operation’s auth error (spec 011). (`quickstart.md` Scenario 5; SC-002)

### Implementation for User Story 3

- [x] T012 [US3] Run `rg 'gitlab' internal/tui` per `specs/014-domain-ui-split/contracts/domain-ui-split.md` — expect no matches; if a hit exists, remove session construction from that flow file so auth/create stay in `internal/sync` / `internal/mr` only (do not add a GitLab interface)
- [x] T013 [P] [US3] Confirm file-change badges still load via local git in `internal/tui/inbound.go` / `internal/tui/treeview.go` (no GitLab). No behavior change

**Checkpoint**: SC-002 holds; US3 adds no product change when the grep is already clean

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Contract audit, full regression, quickstart. After US1 at minimum; after US2 for browser greps

- [x] T014 Run the grep audit in `specs/014-domain-ui-split/contracts/domain-ui-split.md` (`RenderASCII`/`lipgloss` in `internal/tree`; `func OpenURLs` in `internal/sync`; `sync.OpenURLs` anywhere; `browser.OpenURLs` at both surfaces; `gitlab` in `internal/tui`; `WalkDisplay` used from `internal/tui/treeview.go`)
- [x] T015 Run `go test ./... -count=1` and fix regressions (015 types and CLI flags must stay unchanged)
- [x] T016 Run `specs/014-domain-ui-split/quickstart.md` Scenarios 1–6 (walk, decline/accept open, sync has no OpenURLs, GitLab grep, `render.go` gone)
- [x] T017 [P] Confirm no new packages, `internal/tui/mrflow.go` still uses `browser.Open`, and `specs/015-unify-domain-types/` was not edited

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — T003 ∥ T004; T005 after T004; T006 last. **MVP after T006**
- **User Story 2 (Phase 4)**: Depends on Foundational only (not US1). T007 first; T008 ∥ T009 after T007; T010 after callers retargeted; T011 last
- **User Story 3 (Phase 5)**: Depends on Foundational only; meaningful after US1/US2 if any TUI files were touched
- **Polish (Phase 6)**: After MVP (US1) for tree greps; after US2 for OpenURLs greps; full quickstart after US1–US3

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (display walk) 🎯 MVP | Foundational | T002 |
| US2 (browser OpenURLs) | Foundational | T002 (parallel with US1) |
| US3 (GitLab grep) | Foundational | T002 (usually no code) |
| Polish | Desired stories | T006 for MVP; T011 for full 014 |

### Within Each User Story

- US1: delete unused renderer ∥ add walk; then map glyphs; then tests
- US2: helper + new tests first; retarget CLI ∥ TUI; then delete from sync
- Tests travel with the implementation task (no TDD-first tasks)

### Parallel Opportunities

- **Phase 3**: T003 ∥ T004 (`render.go` vs `tree.go`)
- **Phase 3+4**: US1 (`tree.go` / `treeview.go`) ∥ US2 (`browser` / `sync` / `cli/sync.go` / `syncflow.go`)
- **Phase 4**: T008 ∥ T009 after T007
- **Phase 5**: T012 ∥ T013
- **Phase 6**: T017 ∥ T014; T015–T016 after greps

---

## Parallel Example: Foundational + US1 ∥ US2

```bash
# After T002:
Task T003: delete internal/tree/render.go
Task T004: WalkDisplay in internal/tree/tree.go + tree_test.go

# Then US1 sequential:
Task T005: flattenTree in internal/tui/treeview.go
Task T006: go test ./internal/tree/... ./internal/tui/...

# Overlap with US1 (different files):
Task T007: internal/browser/open.go + open_test.go
Task T008: internal/cli/sync.go
Task T009: internal/tui/syncflow.go
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: `go test ./internal/tree/... ./internal/tui/...`; `rg 'RenderASCII|lipgloss' internal/tree`; `quickstart.md` Scenario 1
5. Demo: tree still looks the same; `internal/tree` has no lipgloss

Do not block MVP on moving `OpenURLs` (US2) or the GitLab grep (US3).

### Incremental Delivery

1. Setup + Foundational → baseline green
2. US1 → display walk (**MVP**)
3. US2 → `browser.OpenURLs`
4. US3 → grep audit
5. Polish → full `go test ./...` + quickstart 1–6

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational (T001–T002)
2. Developer A: US1 (T003–T006)
3. Developer B: US2 (T007–T011) in parallel
4. Either: US3 grep (T012–T013)
5. Polish on the merged tree

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work in the same file
- [Story] label maps task to US1–US3
- Outcome freeze: membership, glyphs, prompts, 200ms pause, warning text, [y/N]
- Grilling (2026-08-24, plan): delete `RenderASCII`; `WalkDisplay` on `Document`; `OpenURLs` in `browser`; US3 verify-only
- Grilling (2026-08-24, tasks): tests with impl; new `internal/browser/open_test.go`; MVP = US1 only
- Avoid: ports/adapters; merging `WalkDisplay` with `CollectEdges`; changing MR to `OpenURLs`; TDD-first ordering; editing spec 015
