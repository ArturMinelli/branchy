# Implementation Plan: Embed TUI Flows

**Branch**: `013-embed-tui-flows` | **Date**: 2026-08-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/013-embed-tui-flows/spec.md`

## Summary

Replace the main-tree inline link editor and unlink-only confirm with the dedicated `LinkFlowModel` / `UnlinkFlowModel` wizards, hosted the same way as embedded sync/MR. Operations stay spec 011 (`project.Link` / `project.Unlink`). Standalone `branchy link` / `unlink` still exit to the shell. Outcome freeze on persist, error meaning, and keys; wizard screens on the main tree are an allowed visible change.

Grilling (2026-08-24): from the tree, `l` prefills parent and starts at child name; `u` skips the picker and opens the dedicated confirm; success/error steps are shown before returning to the refreshed tree; esc from the prefilled child step opens the parent picker.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: Bubble Tea (existing `internal/tui`); spec-011 owners `project.Link` / `project.Unlink`; embed pattern already used by `SyncFlowModel` / `MRFlowModel` (`Embedded` + host forwards Update/View). No new modules.

**Storage**: Existing `branch-tree.yaml` via project operations. No schema change.

**Testing**: `go test ./internal/tui/... -count=1` — rewrite main-tree link/unlink tests to drive the dedicated flows; add prefill/skip-step and embedded cancel/finish cases; keep standalone `linkflow_test.go` / `unlinkflow_test.go`. Full regression: `go test ./... -count=1`.

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI default

**Performance Goals**: Same as today — embedding is session routing, not extra I/O

**Constraints**: Outcome freeze on persist, error meaning, `l`/`u`/`esc`/`q` existence (FR-001–007); do not split `app.go` or sync/MR flow files (FR-008); product usable without specs 014–015 (FR-009); depends on specs 011–012 merged; no second link/unlink implementation in the main session (FR-003)

**Scale/Scope**: Two flows (`linkflow.go`, `unlinkflow.go`) gain `Embedded` + prefill options; `app.go` drops inline `updateLink` / `updateUnlink` and hosts the models like sync/MR. No new files.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–012 conventions |
| Library-first / separation | ✅ Pass | Domain stays in `project.Link` / `Unlink`; this spec only deletes duplicate interactive paths |
| CLI interface | ✅ Pass | Standalone `branchy link` / `unlink` still dedicated sessions that exit to the shell; no new flags |
| Test coverage | ✅ Pass | Dedicated-flow tests stay; main-tree tests assert embed + prefill + cancel/success, not the old two-field editor |
| Simplicity / YAGNI | ✅ Pass | Copy the existing sync/MR host pattern; no shared embed dispatcher; no new files |
| Consistency | ✅ Pass | `Embedded` + `tea.Quit` only when not embedded; `q`/`esc` match MR `cancelOrQuit` / `finish` |

**Post-design re-check**: Gates still pass. Natural home is `LinkFlowOptions` / `UnlinkFlowOptions` next to `MROptions`. Prefill is the same skip-step idea as `MROptions.PrefilledSource`. Main session stays one file.

**Grilling follow-up (2026-08-24)**: Prefill skips the first picker; esc from child still opens that picker. Success/error steps stay visible; host refreshes the tree only after a successful persist and a done key.

## Project Structure

### Documentation (this feature)

```text
specs/013-embed-tui-flows/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── embedded-flows.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/tui/
├── app.go               # Host link/unlink flows like sync/MR; delete inline editor/confirm fields
├── app_test.go          # Drive dedicated flows from `l`/`u`; cancel and success→tree
├── linkflow.go          # LinkFlowOptions{Embedded, PrefillParent}; finish/cancel without tea.Quit when embedded
├── linkflow_test.go     # Keep standalone; add prefill starts at child; esc from child opens picker
├── unlinkflow.go        # UnlinkFlowOptions{Embedded, PrefillTarget}; skip to confirm when prefilled
├── unlinkflow_test.go   # Keep standalone; add prefill starts at confirm
├── syncflow.go          # unchanged
├── mrflow.go            # unchanged (pattern to copy)
└── embedded_sync_test.go # unchanged; optional sibling tests may live in app_test.go

internal/cli/
└── link.go / unlink.go  # unchanged: still tui.RunLink / RunUnlink (standalone)

internal/project/
└── project.go           # unchanged: Link / Unlink remain the persist owners
```

**Structure Decision**: Single Go module. No new TUI files. Deepen the existing flow models with the same `Embedded` + prefill options MR already has. `app.go` stays the router (FR-008).

## Complexity Tracking

> No constitution violations to justify.
