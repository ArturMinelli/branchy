# Implementation Plan: Domain / UI Split

**Branch**: `014-domain-ui-split` | **Date**: 2026-08-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/014-domain-ui-split/spec.md`

## Summary

Move presentation and environment side effects off domain types. Delete unused `Document.RenderASCII` (lipgloss on the tree document). Add an unstyled `Document.WalkDisplay` that both the interactive tree and any future listing consume. Move sequential tab-opening (`OpenURLs` + pause + warnings) from `internal/sync` into `internal/browser`; CLI and TUI keep asking, then call the helper. `OpenableURLs` stays on sync. Interactive GitLab construction is already gone (spec 011) — this spec only verifies it.

Outcome freeze: same membership, child order, glyphs, badges, prompts, tab order, pause, and warning wording.

Grilling (2026-08-24): plan 014 next; delete `RenderASCII` and extract the walk onto `Document`; sequential open lives in `browser`; US3 is a grep audit with no expected code change.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `internal/tree`, `internal/tui`, `internal/sync`, `internal/browser`, `internal/cli`; lipgloss stays in TUI only. No new modules.

**Storage**: Existing `branch-tree.yaml`. No schema change.

**Testing**: `go test ./internal/tree/... ./internal/browser/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1` — walk membership/order; OpenURLs sequential+warning in browser; treeview still paints the same glyphs from the walk; OpenableURLs stays in sync. Full regression: `go test ./... -count=1`.

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI default

**Performance Goals**: Same as today — one DFS for display; sequential opens still pause 200ms between tabs

**Constraints**: Outcome freeze (FR-007); no ports/adapters (spec assumption); no GitLab in TUI (FR-005, already true); product usable without spec 015 (FR-008); depends on specs 011–013 merged

**Scale/Scope**: Delete `internal/tree/render.go`; deepen `Document` with a display walk; move `OpenURLs` (~20 lines) to `internal/browser`; retarget two call sites (CLI sync, TUI syncflow). US3 is verification only.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–013 conventions |
| Library-first / separation | ✅ Pass | Drawing leaves `tree`; browser side effects leave `sync`; no new orchestration package |
| CLI interface | ✅ Pass | Same prompt, same [y/N], same warning print; no new flags |
| Test coverage | ✅ Pass | Walk tests on `Document`; OpenURLs tests move with the helper; existing treeview glyph/badge tests stay |
| Simplicity / YAGNI | ✅ Pass | Delete unused renderer; one walk method; deepen `browser` instead of a new opener type |
| Consistency | ✅ Pass | CLI and TUI both call `browser.OpenURLs`; MR still uses single `browser.Open` |

**Post-design re-check**: Gates still pass. Natural homes: `WalkDisplay` next to `Roots` on `Document`; `OpenURLs` next to `Open` in `browser`. TUI maps visits to glyphs. Sync keeps URL selection (`OpenableURLs`) and drops the open loop.

**Grilling follow-up (2026-08-24)**: Unused styled `RenderASCII` is deleted, not moved. US3 is audit-only.

## Project Structure

### Documentation (this feature)

```text
specs/014-domain-ui-split/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── domain-ui-split.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/tree/
├── tree.go              # Add WalkDisplay + DisplayNode; keep Roots / CollectEdges
├── tree_test.go         # Membership, child sort, isLast / LastAtDepth
└── render.go            # DELETE (RenderASCII + lipgloss)

internal/tui/
├── treeview.go          # flattenTree consumes WalkDisplay; glyphs stay here
├── treeview_test.go     # Unchanged outcomes (connectors, badges)
└── syncflow.go          # Call browser.OpenURLs instead of sync.OpenURLs

internal/browser/
├── open.go              # Open (unchanged) + OpenURLs (moved) + test hook
└── open_test.go         # Sequential order + warning (moved from sync)

internal/sync/
├── sync.go              # Keep OpenableURLs; delete OpenURLs and browserOpen
└── sync_test.go         # Keep OpenableURLs tests; drop OpenURLs tests

internal/cli/
└── sync.go              # Prompt unchanged; open via browser.OpenURLs
```

**Structure Decision**: Single Go module. No new packages. Delete the unused renderer file; deepen `Document` and `browser`. `CollectEdges` stays the sync-edge walk — do not merge it with display.

## Complexity Tracking

> No constitution violations to justify.
