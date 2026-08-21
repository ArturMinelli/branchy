# Implementation Plan: Split CLI Commands

**Branch**: `012-split-cli-commands` | **Date**: 2026-08-21 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/012-split-cli-commands/spec.md`

## Summary

Split the monolithic `internal/cli/root.go` into one file per subcommand while keeping handlers thin: parse flags/args, choose interactive vs scripted via shared `mode.go`, call the spec-011 operations, print results. `root.go` assembles the tree and hosts the default TUI entry. Scripted helpers (`runSyncCLI`, `runMRFlags`) stay private in their command files. Outcome freeze — no flag, help, or output changes.

## Technical Context

**Language/Version**: Go 1.26.4

**Primary Dependencies**: cobra/pflag (existing); spec-011 owners in `internal/project`, `internal/sync`, `internal/mr`; TUI launchers in `internal/tui`. No new modules.

**Storage**: N/A — commands are wiring only; persistence stays in project/sync/mr operations.

**Testing**: `go test ./internal/cli/... -count=1` — keep `mode_test.go`; add per-command tests only where behavior is non-trivial (sync flag-forces-CLI already references `syncCmd`). Full regression: `go test ./... -count=1` and help/output spot checks from quickstart.

**Target Platform**: Linux/macOS terminal (existing branchy platforms)

**Project Type**: CLI tool with interactive TUI default

**Performance Goals**: Same as today — file split is compile-time organization only

**Constraints**: Outcome freeze (FR-006); handlers MUST NOT mutate tree, persist, or construct GitLab clients (FR-003); shared TTY/prompt rules stay in `mode.go` (FR-005); no move to `cmd/branchy` (edge case); product usable without specs 013–015 (FR-007); depends on spec 011 merged

**Scale/Scope**: Six subcommands + default root; ~340 lines today in `root.go` → ~8 files; `cmd/branchy/main.go` unchanged

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Constitution ratified | ⚠️ N/A | `.specify/memory/constitution.md` is still a template — interim gates from 004–011 conventions |
| Library-first / separation | ✅ Pass | Commands are thin surfaces; domain stays in project/sync/mr |
| CLI interface | ✅ Pass | Same subcommand names, flags, shorthands, help purpose, stdout/stderr, exit codes |
| Test coverage | ✅ Pass | Existing `mode_test.go` + sync flag test; grep audit that handlers call owners only |
| Simplicity / YAGNI | ✅ Pass | Flat package, no `commands/` subfolder; no shared `prompt.go` until a third duplicate appears |
| Consistency | ✅ Pass | Every scripted path follows parse → operation → print; interactive path delegates to existing TUI runners |

**Post-design re-check**: Gates still pass. Flat layout matches grilling. `root.go` ~40 lines (Execute, rootCmd, AddCommand). Each command file owns its `init()` flag registration.

**Grilling follow-up (2026-08-21)**: Read-only `p.Tree.Names()` / `p.Tree.CollectEdges()` in `sync.go` only — allowed for outcome freeze. Grep audit stays narrow; no per-command test files beyond `mode_test.go`.

## Project Structure

### Documentation (this feature)

```text
specs/012-split-cli-commands/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── cli-commands.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 (/speckit-tasks — not created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/branchy/
└── main.go              # unchanged: cli.Execute()

internal/cli/
├── root.go              # rootCmd, Execute(), init() AddCommand only
├── mode.go              # IsTTY, UseTUI, anyFlagChanged (unchanged)
├── mode_test.go         # unchanged
├── init.go              # initCmd + --force flag init()
├── sync.go              # syncCmd, runSyncCLI (private), flags init()
├── mr.go                # mrCmd, runMRFlags (private), flags init()
├── link.go              # linkCmd
├── unlink.go            # unlinkCmd
└── projects.go          # projectsCmd
```

**Structure Decision**: Single `internal/cli` package, flat files (grilling 2026-08-21). No `commands/` subdirectory. Private scripted helpers live in the same file as their command. TUI embed (013), browser move (014), type unify (015) stay out of scope.

## Complexity Tracking

> No constitution violations to justify.
