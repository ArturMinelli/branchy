# Implementation Plan: Unify Domain Types

**Branch**: `015-unify-domain-types` | **Date**: 2026-08-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/015-unify-domain-types/spec.md`

## Summary

Unify the leftover dual vocabularies after specs 008–014. One merge-direction type: keep `sync.Direction` (`Downward` / `Upward`); delete TUI `diffDirection` and the `syncDirection()` conversion. One outcome type: `mr.Action` on both `mr.CreateResult` and `sync.Result`, plus `mr.SkipReason` so `OpenableURLs` no longer sniffs `Message == "skipped by user"`. Create’s error policy stays as today and is locked in tests: validation/auth abort with an error; already-open is skipped+nil; unrecovered GitLab failure is a failed result+nil; scripted `mr` still maps failed → non-zero exit.

Outcome freeze: inbound/outbound help copy, arrows, keys, created/skipped/failed wording, cascade order, and exit codes stay.

Grilling (2026-08-24): plan 015 next; direction lives on `sync` (TUI consumes it); typed action + skip reason; do not turn GitLab failures into `Create` errors.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `internal/sync`, `internal/mr`, `internal/tui`, `internal/cli`; no new modules or packages.

**Storage**: Existing `branch-tree.yaml`. No schema change. Direction remains session-scoped (not persisted).

**Testing**: `go test ./internal/mr/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1` — typed action switches; `OpenableURLs` by `SkipReason`; TUI direction uses `sync.Direction` with frozen help/arrow copy; Create validation still errors; CLI `mr` still non-zero on failed. Full regression: `go test ./... -count=1`.

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI default

**Performance Goals**: Same as today — no extra walks or GitLab calls

**Constraints**: Outcome freeze (FR-006); no new `app` package; no GitLab interface (spec assumption); product usable after merge as last of five (FR-007); depends on specs 011–014 merged; scripted sync remains downward-only

**Scale/Scope**: Delete `diffDirection` + `syncDirection`; type `mr.Action` and add `SkipReason`; retarget TUI/CLI switches and tests. Create control flow already matches the policy — US3 is contract tests + CLI mapping, not a rewrite.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–014 conventions |
| Library-first / separation | ✅ Pass | Deepen `sync.Direction` and `mr.Action`; no new orchestration package |
| CLI interface | ✅ Pass | No new flags; same copy, keys, arrows, exit codes |
| Test coverage | ✅ Pass | Direction tests retarget to `sync.Direction`; OpenableURLs tests use `SkipReason`; Create validation vs result matrix |
| Simplicity / YAGNI | ✅ Pass | One type each; no shared printer layer; no GitLab port |
| Consistency | ✅ Pass | CLI and TUI switch on `mr.Action`; both surfaces use `sync.Direction` |

**Post-design re-check**: Gates still pass. Natural homes: `Direction` stays next to `Ends` / `EdgesBelow` on `sync`; `Action` and `SkipReason` stay next to `Create` on `mr`. TUI stores `sync.Direction` (treeview imports `sync`). Sync copies `mr.Action` / `SkipReason` onto `Result`.

**Grilling follow-up (2026-08-24)**: Do not rename `Downward`/`Upward` to inbound/outbound. Do not move Direction onto `tree`. Do not encode skip subtypes as extra `Action` values.

## Project Structure

### Documentation (this feature)

```text
specs/015-unify-domain-types/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── unify-domain-types.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/sync/
├── sync.go              # Keep Direction/Downward/Upward; Result.Action is mr.Action; add SkipReason; OpenableURLs uses SkipReason
└── sync_test.go         # OpenableURLs fixtures set SkipReason; drop Message-as-discriminator

internal/mr/
├── mr.go                # type Action string; type SkipReason string; CreateResult fields typed; set SkipAlreadyOpen on skip
└── mr_test.go           # Validation → error (no result); already-open/fail matrix documented with existing Create paths

internal/tui/
├── inbound.go           # Delete diffDirection; treeHelpFooter(sync.Direction); inbound/outbound copy unchanged
├── inbound_test.go      # Footer tests use sync.Downward / sync.Upward
├── treeview.go          # direction sync.Direction; import sync
├── treeview_test.go     # setDirection(sync.Upward) etc.; glyph/badge outcomes unchanged
├── app.go               # Rename field diffDirection → direction sync.Direction; pass it to sync with no remap
├── app_test.go          # Compare against sync.Downward / sync.Upward
├── syncflow.go          # Delete syncDirection; Control+Up/Down already set sync.Direction
└── mrflow.go            # Switch stays on mr.Action (now typed)

internal/cli/
├── sync.go              # switch r.Action on mr.ActionCreated/Skipped/Failed (not "created" literals)
└── mr.go                # Unchanged mapping: failed result → fmt.Errorf (non-zero); Create error → return err
```

**Structure Decision**: Single Go module. No new packages. Direction stays in `sync`; outcomes stay in `mr`. `internal/git` inbound/outbound file-count helpers are unchanged (different concept). Do not merge `WalkDisplay` or browser helpers.

## Complexity Tracking

> No constitution violations to justify.
