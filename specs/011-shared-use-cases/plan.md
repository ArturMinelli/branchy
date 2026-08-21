# Implementation Plan: Shared Use Cases

**Branch**: `011-shared-use-cases` | **Date**: 2026-08-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/011-shared-use-cases/spec.md`

## Summary

Give link, unlink, sync, merge-request create, and init a single owner each so scripted CLI, standalone TUI, and the main tree call the same functions. Deepen `internal/project` (link/unlink/init persist) and `internal/sync` (begin + skip helper); keep `mr.Create` as the create owner. No new package. User-visible outcomes stay frozen except interactive unlink, which gains the scripted “edge exists” check.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: existing `internal/project`, `internal/tree`, `internal/sync`, `internal/mr`; cobra CLI; Bubble Tea flows. No new modules.

**Storage**: Existing `~/.config/branchy/projects/<id>/branch-tree.yaml` via `Project.SaveTree`. No schema change.

**Testing**: `go test ./internal/project/... ./internal/sync/... ./internal/mr/... ./internal/cli/... ./internal/tui/...` — persist+validation on project ops; Begin/SkippedByUser; surfaces call the ops (no private Link+Save / gitlab.Client in TUI)

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI

**Performance Goals**: Same as today — link/unlink are one tree mutate + one YAML write; sync still one glab round-trip per confirmed edge

**Constraints**: Outcome freeze (FR-007); no `app`/`usecase` package (FR-006); TUI must not import `internal/gitlab` (FR-009); no Cobra file split / TUI embed / presentation split / type unify (specs 012–015); product remains fully usable after this merge (FR-010)

**Scale/Scope**: Five operations; three surfaces for link/unlink; two for sync/MR/init. New exported functions: `project.Link`, `project.Unlink`, `sync.Begin`, `sync.SkippedByUser`. Init and `mr.Create` already exist.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–010 conventions |
| Library-first / separation | ✅ Pass | Operations deepen `project` / `sync` / `mr`; surfaces collect input and print. No new orchestration layer |
| CLI interface | ✅ Pass | No new flags, subcommands, or copy redesign; scripted `-y` / positionals unchanged |
| Test coverage | ✅ Pass | Project persist tests + sync Begin/skip + TUI/CLI call-site tests; existing flow tests keep passing |
| Simplicity / YAGNI | ✅ Pass | Keep `sync.Run` and `RunEdge`; no channel-based Confirm in Bubble Tea; tree primitives stay exported |
| Consistency | ✅ Pass | Unlink uses the scripted edge check on every surface; auth error meaning stays `glab auth: … (run: glab auth login)` |

**Post-design re-check**: Gates still pass. Natural homes: `project.Link`/`Unlink` next to `SaveTree`; `sync.Begin` next to `Run`; skip helper next to `processEdge`. TUI sync drops `gitlab.Client`. Main-tree link/unlink stay inline screens (spec 013) but call the project ops.

## Project Structure

### Documentation (this feature)

```text
specs/011-shared-use-cases/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── shared-operations.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
internal/project/
├── project.go           # Link, Unlink (+ UnlinkResult); Init already here
└── project_test.go      # new: persist, duplicate edge, unlink edge/root, already-registered init

internal/tree/
├── tree.go              # Link / UnlinkSubtree remain in-memory primitives (unchanged rules)
└── tree_test.go         # keep covering primitives; surfaces must not be the only persist tests

internal/sync/
├── sync.go              # Begin (auth + collect when edges exist); SkippedByUser; Run / RunEdge stay
└── sync_test.go         # Begin empty-vs-auth, unknown from-branch; skip helper shape

internal/mr/
└── mr.go                # Create remains the sole GitLab create+auth for one pair (no signature change)

internal/cli/
└── root.go              # scripted link/unlink call project ops; drop private unlink validation block

internal/tui/
├── linkflow.go          # confirm → project.Link (not Tree.Link + SaveTree)
├── unlinkflow.go        # ParentOf + project.Unlink
├── app.go               # main-tree link/unlink same ops; no Tree.Link/UnlinkSubtree + SaveTree
├── syncflow.go          # prepareSyncFrom → sync.Begin; decline → SkippedByUser; drop gitlab import
├── initflow.go          # already project.Init — keep
└── mrflow.go            # already mr.Create — keep
```

**Structure Decision**: Single Go module. No new package. Command files stay in `internal/cli` (spec 012). Inline main-tree link/unlink screens stay in `app.go` (spec 013). Browser open stays in `sync.OpenURLs` (spec 014). Outcome string types stay as they are (spec 015).

## Complexity Tracking

> No constitution violations to justify.
