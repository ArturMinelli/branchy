# Research: Split CLI Commands

**Feature**: `012-split-cli-commands` | **Date**: 2026-08-21

## 1. File layout under `internal/cli`

**Decision**: Flat package — `root.go`, existing `mode.go` / `mode_test.go`, and one file per subcommand: `init.go`, `sync.go`, `mr.go`, `link.go`, `unlink.go`, `projects.go`.

**Rationale**: Grilling locked flat layout. All symbols stay in package `cli`; no import cycles. Matches Go convention for small cobra apps. Subfolder would add path noise without a second package boundary.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `internal/cli/commands/*.go` | Grilling chose flat; extra directory with no package split |
| One file per command + `prompt.go` | Grilling: helpers stay with their command until duplication warrants extraction |
| Move commands to `cmd/branchy` | Spec edge case forbids; `main.go` stays thin |

---

## 2. What stays in `root.go`

**Decision**: `rootCmd` definition (Use, Short, Long, default `RunE` → `tui.Run`), `Execute()`, and `init()` that `AddCommand`s the six subcommands. No command bodies, no flag registration for subcommands.

**Rationale**: FR-002 — program entry / tree assembly in one obvious place. Default no-subcommand TUI behavior is part of the root command, not a separate subcommand file.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Empty root.go; default TUI in `tui.go` | Adds a seventh file for one struct; root command belongs with tree assembly |
| Each command self-registers via `init()` only | Still need one `AddCommand` block or scattered registration — harder to see full tree |

---

## 3. Scripted helper placement

**Decision**: `runSyncCLI` stays in `sync.go` (unexported). `runMRFlags` stays in `mr.go`. Per-edge confirm closure and browser prompt stay inside `runSyncCLI` as today. No new `prompt.go`.

**Rationale**: Grilling locked same-file private helpers. Sync and MR prompts differ enough (per-edge loop vs single confirm) that a shared helper would be thin abstraction over `bufio.Reader` reads.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Shared `prompt.go` with `ConfirmYN` | Grilling rejected; only two call sites, slightly different messages |
| Move printing into sync/mr packages | Violates surface vs domain split from 011 |

---

## 4. Command variable visibility

**Decision**: Keep package-level `var syncCmd`, `linkCmd`, etc. (exported within package only — lowercase names, same as today). Tests that reference `syncCmd` (`mode_test.go`) continue to work without exporting commands.

**Rationale**: `TestSyncCmdFlagForcesCLI` already uses `syncCmd`. Moving to constructor functions would churn tests without user benefit.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `NewSyncCmd()` per test | Spec allows constructing root per test but does not require it; current test works |
| Export commands for tests | Unnecessary API surface |

---

## 5. Flag registration and `init()` order

**Decision**: Each command file has its own `func init()` registering flags on its command var. Go merges `init()` order within the package; flag names remain unique per command (no collision).

**Rationale**: Colocating flags with command definition satisfies FR-001 (“use line, flags, run path in dedicated definition”). Cobra already scopes flags per command.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Central flags.go | Splits command definition across files — opposite of spec intent |
| Register flags in root.go init | Keeps root.go as a second “god” init block |

---

## 6. Handler thinness audit (post-split target)

**Decision**: After split, each `RunE` path MUST contain only: resolve project (if needed), `UseTUI` branch → TUI runner OR scripted parse → single operation call (`project.Init`, `project.Link`, `project.Unlink`, `sync.Run`, `mr.Create`, `project.ListAll`) → printf. Grep gate: no `Tree.Link`, `SaveTree`, `UnlinkSubtree`, `HasEdge`, `gitlab.Client` in `internal/cli`.

**Exception (grilling 2026-08-21)**: `sync.go` may call read-only `p.Tree.Names()` and `p.Tree.CollectEdges()` for the interactive `--from` picker and pre-sync plan print — copied verbatim for outcome freeze. No mutation, persist, or GitLab access.

**Rationale**: FR-003 restates 011 contract for commands. Split is worthless if logic is duplicated across files. Read-only tree reads in `sync.go` are display-only preludes to `sync.Run`, not duplicate domain logic.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Forbid all `p.Tree` in CLI | Would require moving picker/plan print into `sync` package — wider scope than 012 |
| Allow sync.go to call `EdgesBelow` directly | Already inside `sync.Run`; CLI should not re-collect edges for the sync operation itself |

---

## 7. What this spec does not change

**Decision**: `cmd/branchy/main.go` unchanged. TUI flows unchanged. No new flags or subcommands. Main-tree default still via `rootCmd.RunE`. Specs 013–015 files untouched.

**Rationale**: FR-007; independently mergeable refactor.
