---

description: "Task list for Unify Domain Types feature implementation"
---

# Tasks: Unify Domain Types

**Input**: Design documents from `/specs/015-unify-domain-types/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/unify-domain-types.md, quickstart.md

**Tests**: With implementation — no TDD-first ordering (grilling 2026-08-24). Retarget existing `internal/tui/*_test.go` direction cases. OpenableURLs / SkipReason cases stay in `internal/sync/sync_test.go`. Create validation cases go in existing `internal/mr/mr_test.go`.

**Organization**: Tasks grouped by user story. **MVP is US1 only** (grilling). US1 is TUI-only (no `mr` / `sync.Result` edits). Action + SkipReason types land in US2. US3 is Create godoc + validation tests, not a control-flow rewrite. US1 can overlap US2/US3 on different files.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3)
- All tasks include exact file paths

## Path Conventions

Go module at repository root:

- `internal/tui/inbound.go` — delete `diffDirection`; `treeHelpFooter(sync.Direction)`
- `internal/tui/inbound_test.go` — `sync.Downward` / `sync.Upward`
- `internal/tui/treeview.go` — `direction sync.Direction`; import `sync`
- `internal/tui/treeview_test.go` — `setDirection(sync.Upward)` etc.
- `internal/tui/app.go` — rename `diffDirection` → `direction sync.Direction`; import `sync`
- `internal/tui/app_test.go` — compare `model.direction` to `sync.Downward` / `sync.Upward`
- `internal/tui/syncflow.go` — delete `syncDirection`
- `internal/mr/mr.go` — `type Action`, `type SkipReason`; typed `CreateResult`; Create godoc
- `internal/mr/mr_test.go` — Create validation returns `(nil, err)`
- `internal/sync/sync.go` — `Result.Action mr.Action`; `SkipReason`; `OpenableURLs` by reason
- `internal/sync/sync_test.go` — fixtures set `SkipReason`; user-declined-with-URL excluded
- `internal/cli/sync.go` — `switch` on `mr.Action*`
- `internal/cli/mr.go` — failed result → `fmt.Errorf` (verify, no rewrite)
- `internal/git` inbound/outbound helpers stay closed
- Specs 001–014 stay closed except this feature’s docs

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm scope and file targets before edits

- [x] T001 Verify implementation targets in `specs/015-unify-domain-types/plan.md` — keep `sync.Direction` (`Downward`/`Upward`); delete TUI `diffDirection` and `syncDirection`; type `mr.Action` + `mr.SkipReason` in US2; Create control flow unchanged (US3 godoc + tests); no new packages; `internal/git` helpers closed

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Green baseline so story diffs are attributable

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Run `go test ./internal/mr/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1` and confirm green as the pre-unify baseline

**Checkpoint**: Affected packages pass; `diffDirection` and string `Result.Action` still exist (stories will remove them)

---

## Phase 3: User Story 1 - One merge direction everywhere (Priority: P1) 🎯 MVP

**Goal**: One `sync.Direction` for main tree, treeview, help footers, arrows, and standalone picker. Delete `diffDirection` and `syncDirection()`. Help copy, keys, and arrows stay. Scripted sync stays downward-only (no flag).

**Independent Test**: Outbound on the main tree → sync walks child→parent using that same value. Inbound → parent→child. Standalone Control+Up / Control+Down still set `sync.Upward` / `sync.Downward`. `rg 'type diffDirection|syncDirection|diffInbound|diffOutbound' internal/` empty. (`quickstart.md` Scenario 1; SC-001/004)

### Implementation for User Story 1

- [x] T003 [US1] In `internal/tui/inbound.go`, delete `diffDirection`, `diffInbound`, and `diffOutbound`. Change `treeHelpFooter` to take `sync.Direction`. Keep copy: `counts: inbound (parent→child)` for `sync.Downward`, `counts: outbound (child→parent)` for `sync.Upward`. Import `branchy/internal/sync`. Update `internal/tui/inbound_test.go` to call `treeHelpFooter(sync.Downward)` / `sync.Upward` and still assert the same two footer strings
- [x] T004 [US1] In `internal/tui/treeview.go`, change `direction`, `setDirection`, `activeCounts`, `directionArrow`, and `annotateCount` to `sync.Direction` (`Upward` → outbound counts and `↑`; else inbound and `↓`). Import `sync`. Update `internal/tui/treeview_test.go` `setDirection` calls the same way; keep glyph/badge assertions unchanged
- [x] T005 [US1] In `internal/tui/app.go`, rename field `diffDirection` → `direction sync.Direction`, import `sync`, map Control+Up → `sync.Upward` and Control+Down → `sync.Downward`, pass `Direction: m.direction` into `SyncFlowOptions` (no remap). Delete `syncDirection` from `internal/tui/syncflow.go`. Update `internal/tui/app_test.go` to use `model.direction` and `sync.Downward` / `sync.Upward` (`TestSyncFollowsTreeOutbound` / `Inbound` still pass)

- [x] T006 [US1] Run `go test ./internal/tui/... ./internal/sync/... -count=1` and `rg 'type diffDirection|func syncDirection|diffInbound|diffOutbound' internal/` (expect no matches)

**Checkpoint**: TUI and sync share one direction type. **MVP is shippable here** without US2/US3

---

## Phase 4: User Story 2 - One outcome vocabulary for create and sync (Priority: P1)

**Goal**: `mr.Action` typed on `CreateResult` and `sync.Result`. `mr.SkipReason` distinguishes already-open vs user-declined. `OpenableURLs` uses `SkipReason`, not `Message`. CLI sync switches on `mr.Action*` constants. User-visible created/skipped/failed words stay.

**Independent Test**: Fixtures for created, skipped-already-open, skipped-user (with URL), and failed. Scripted and interactive summaries classify via `mr.Action`. User-declined is not openable even with a URL. (`quickstart.md` Scenario 2; SC-002/005)

### Implementation for User Story 2

- [x] T007 [US2] In `internal/mr/mr.go`, add `type Action string` and `type SkipReason string` with constants `ActionCreated`/`Skipped`/`Failed` (same `"created"` / `"skipped"` / `"failed"` values) and `SkipNone` / `SkipAlreadyOpen` (`"already_open"`) / `SkipUserDeclined` (`"user_declined"`). Change `CreateResult.Action` to `Action` and add `SkipReason SkipReason`. Set `SkipAlreadyOpen` on every already-open return; leave `SkipNone` on created/failed. Do not change Create control flow
- [x] T008 [US2] In `internal/sync/sync.go`, change `Result.Action` to `mr.Action`, add `SkipReason mr.SkipReason`. `SkippedByUser` sets `ActionSkipped` + `SkipUserDeclined` + Message `"skipped by user"`. `processEdge` uses `mr.ActionFailed` (not `"failed"`), copies `SkipReason` from `mr.Create`. `OpenableURLs` includes created-with-URL and skipped-with-URL unless `SkipReason == mr.SkipUserDeclined`; MUST NOT compare `Message`. Update `internal/sync/sync_test.go`: set `SkipReason` on fixtures; add a skipped+URL+`SkipUserDeclined` case that is excluded
- [x] T009 [P] [US2] In `internal/cli/sync.go`, `switch r.Action` on `mr.ActionCreated` / `ActionSkipped` / `ActionFailed` (import `branchy/internal/mr`); keep print format and counts
- [x] T010 [P] [US2] Add `SkipReason` on skipped fixtures in `internal/tui/syncflow_test.go` and `internal/tui/embedded_sync_test.go` (user-declined → `SkipUserDeclined`; already-open → `SkipAlreadyOpen` where a URL is present). Leave `internal/tui/mrflow.go` / `syncflow.go` switches on `mr.Action*` (now typed)
- [x] T011 [US2] Run `go test ./internal/mr/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1` and `rg 'r\.Message == "skipped by user"' internal/` (expect no matches)

**Checkpoint**: One action type; skip subtypes drive browser-open filtering

---

## Phase 5: User Story 3 - One error policy for merge-request create (Priority: P2)

**Goal**: Lock the existing Create matrix in godoc and tests. Validation/auth → `(nil, err)`; already-open → skipped+nil; unrecovered GitLab failure → failed result+nil. Scripted `mr` still maps failed → non-zero. No Create rewrite.

**Independent Test**: Unknown/same branch → `Create` returns error and nil result (no GitLab). CLI `mr.go` still `ActionFailed` → `fmt.Errorf`. (`quickstart.md` Scenario 3; SC-003)

### Implementation for User Story 3

- [x] T012 [US3] Document the return matrix on `Create` in `internal/mr/mr.go` (validation/auth → error; already-open → skipped+`SkipAlreadyOpen`+nil error; unrecovered failure → failed result+nil error; created → created+nil). Do not change the function body
- [x] T013 [P] [US3] In `internal/mr/mr_test.go`, add `TestCreateValidationReturnsError`: `Create` with missing source, missing target, and same source/target on a `project.Project{Tree: testTree()}` returns `err != nil` and `result == nil` (validation runs before GitLab). Keep existing `TestValidateBranches`
- [x] T014 [P] [US3] Confirm `internal/cli/mr.go` still maps `mr.ActionFailed` → `return fmt.Errorf(...)`, skipped/created → exit 0, and `Create` error → `return err`. No code change unless a case is missing

**Checkpoint**: Policy is documented and validation is tested; CLI mapping unchanged

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Contract audit, full regression, quickstart. After US1 at minimum; after US2 for action/skip greps; full after US3

- [x] T015 Run the grep audit in `specs/015-unify-domain-types/contracts/unify-domain-types.md` (`type diffDirection`; `syncDirection`; `diffInbound|diffOutbound`; `case "created"|case "skipped"|case "failed"` in `internal/cli` and `internal/tui`; `r.Message == "skipped by user"`)
- [x] T016 Run `go test ./... -count=1` and fix regressions (no new flags; help/arrows/prompts unchanged)
- [x] T017 Run `specs/015-unify-domain-types/quickstart.md` Scenarios 1–5
- [x] T018 [P] Confirm no new packages, `internal/git/inbound.go` signatures unchanged, and `specs/001-mr-command/` through `specs/014-domain-ui-split/` were not edited

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational — T003 then T004 then T005 (same package; will not compile until T005); T006 last. **MVP after T006**
- **User Story 2 (Phase 4)**: Depends on Foundational only (not US1). T007 first; T008 after T007; T009 ∥ T010 after T008; T011 last
- **User Story 3 (Phase 5)**: T013 ∥ T014 can overlap US1. T012 edits `mr.go` — after T007 (or fold godoc into T007 if one implementer)
- **Polish (Phase 6)**: After MVP (US1) for direction greps; after US2 for action/skip greps; full quickstart after US1–US3

### User Story Dependencies

| Story | Depends on | Can start after |
|-------|------------|-----------------|
| US1 (one Direction) 🎯 MVP | Foundational | T002 |
| US2 (Action + SkipReason) | Foundational | T002 (parallel with US1; different files) |
| US3 (Create policy tests) | T007 for godoc on typed Create | T002 for T013/T014; T007 for T012 |
| Polish | Desired stories | T006 for MVP; T011 for full 015 |

### Within Each User Story

- US1: inbound type deletion → treeview → app/syncflow; tests travel with those files; T006 validates
- US2: types in `mr.go` first; then `sync.Result` / `OpenableURLs`; then CLI ∥ TUI fixtures
- US3: godoc after types exist; validation tests do not need GitLab
- Tests travel with the implementation task (no TDD-first tasks)

### Parallel Opportunities

- **Phase 3+4**: US1 (`internal/tui/*`) ∥ US2 (`internal/mr/mr.go` then `internal/sync`)
- **Phase 4**: T009 ∥ T010 after T008
- **Phase 5**: T013 ∥ T014; T012 after T007
- **Phase 6**: T018 ∥ T015; T016–T017 after greps

---

## Parallel Example: Foundational + US1 ∥ US2

```bash
# After T002:
# Developer A (US1):
Task T003: inbound.go + inbound_test.go
Task T004: treeview.go + treeview_test.go
Task T005: app.go + app_test.go + delete syncDirection in syncflow.go
Task T006: go test ./internal/tui/... ./internal/sync/...

# Developer B (US2), overlap T003–T006:
Task T007: internal/mr/mr.go types
Task T008: internal/sync/sync.go + sync_test.go
Task T009: internal/cli/sync.go
Task T010: tui syncflow_test.go + embedded_sync_test.go
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: `go test ./internal/tui/... ./internal/sync/...`; `rg 'type diffDirection|syncDirection' internal/`; `quickstart.md` Scenario 1
5. Demo: same arrows/help; TUI has no private direction type

Do not block MVP on Action/SkipReason (US2) or Create godoc/tests (US3).

### Incremental Delivery

1. Setup + Foundational → baseline green
2. US1 → one `sync.Direction` (**MVP**)
3. US2 → typed `mr.Action` + `SkipReason`
4. US3 → Create policy godoc + validation tests
5. Polish → full `go test ./...` + quickstart 1–5

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational (T001–T002)
2. Developer A: US1 (T003–T006)
3. Developer B: US2 (T007–T011) in parallel
4. Either: US3 (T012 after T007; T013–T014 anytime after T002)
5. Polish on the merged tree

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work in the same file
- [Story] label maps task to US1–US3
- Outcome freeze: inbound/outbound copy, arrows, keys, created/skipped/failed words, exit codes
- Grilling (2026-08-24, plan): `sync.Direction` kept; TUI drops `diffDirection`; typed Action + SkipReason; Create failures stay results
- Grilling (2026-08-24, tasks): tests with impl; MVP = US1 only; types in US2; US3 = godoc + tests, no Create rewrite
- Avoid: new packages; GitLab interface; renaming Downward/Upward; extra Action values; TDD-first ordering; editing specs 001–014
