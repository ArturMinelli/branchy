# Research: Unified Command TUI

**Feature**: `004-command-tui` | **Date**: 2026-08-11

## R1: Standalone sync entry point

**Decision**: Add `RunSync(p *project.Project, opts SyncFlowOptions)` in `internal/tui/syncflow.go` mirroring `RunMR`. Extend `SyncFlowModel` with an optional `stepSyncPickRoot` when `fromBranch` is empty; standalone CLI calls `RunSync` with empty from branch.

**Rationale**: `SyncFlowModel` already implements per-edge confirm, summary, and browser batch with `Embedded: false` → `tea.Quit` on finish. Only missing pieces are root-branch picker and a public `Run` wrapper. Reusing the model guarantees SC-003 (equivalent semantics to embedded `s` sync).

**Alternatives considered**:
- *Launch main `tui.Run` with sync pre-selected* — rejected per grilling decision (dedicated mini-flow, exit to shell).
- *Duplicate sync logic in new file* — rejected; violates DRY and risks semantic drift.
- *Refactor sync CLI to embed tree TUI* — rejected; user would land in tree app, violating FR-012.

---

## R2: Interactive vs scripted mode detection

**Decision**: Add `internal/cli/mode.go` with:

```go
func UseTUI(cmd *cobra.Command) bool
func IsTTY() bool
```

- `IsTTY()` — `term.IsTerminal(int(os.Stdout.Fd()))` via `github.com/charmbracelet/x/term` (already in module graph).
- `UseTUI(cmd)` — returns `IsTTY() && !anyFlagChanged(cmd)`.
- `anyFlagChanged` — `cmd.Flags().Visit(func(f *pflag.Flag) { if f.Changed { ... } })`.

Link/unlink: TUI when `len(args) == 0 && UseTUI(cmd)`; scripted when `len(args) == 2`.

**Rationale**: Spec FR-005/FR-007 require any-flag → plain CLI. Cobra's `Flag.Changed` is the standard way to detect explicit flags (including `--from` alone). Non-TTY fallback prevents CI hangs (FR-006).

**Alternatives considered**:
- *Check only `-y` flag* — rejected per grilling (any-flag rule).
- *Environment variable `BRANCHY_NO_TUI`* — rejected; YAGNI; TTY detection suffices.
- *Always TUI when no args* — rejected; breaks `sync --from develop` scripting.

---

## R3: Shared TUI chrome extraction

**Decision**: Extract lipgloss styles and helpers from `app.go` into `internal/tui/chrome.go`:

- `titleStyle`, `helpStyle`, `okStyle`, `warnStyle`, `errStyle` (package-level, shared)
- `RenderTitle(command string) string` — e.g. `"branchy sync"`
- `RenderHelp(text string) string`
- `MinSizeOK(width, height int) bool` — `width >= 80 && height >= 24`
- `RenderTooSmall() string` — single-line friendly message

All flow models (`syncflow`, `mrflow`, new flows) use shared chrome.

**Rationale**: FR-010 requires consistent styling. Today styles live in `app.go` as package-level vars already shared within package — formalizing in `chrome.go` makes the contract explicit and eases new flow files without growing `app.go`.

**Alternatives considered**:
- *Separate `internal/tui/styles` package* — rejected; over-abstraction for one package.
- *Leave styles in app.go* — rejected; new flow files would duplicate or create import cycles.

---

## R4: Dedicated flow models for link, unlink, init, projects

**Decision**: One Bubble Tea model per command, each with `Run*` entry point:

| Command | File | Steps |
|---------|------|-------|
| `sync` | `syncflow.go` (extend) | pick root → edges → summary → browser |
| `projects` | `projectsflow.go` | list browse → detail (read-only) → exit |
| `link` | `linkflow.go` | pick parent → type child → confirm → save |
| `unlink` | `unlinkflow.go` | pick branch → confirm (subtree count) → save |
| `init` | `initflow.go` | confirm register → run Init → success/error |

**Rationale**: Matches proven `MRFlowModel` / `RunMR` pattern. Each flow is independently testable. Link/unlink standalone flows improve on main-app link screen (raw typing) by using branch list pickers per SC-007.

**Alternatives considered**:
- *Reuse embedded `screenLink` from app.go* — rejected; raw text entry, not picker-based; doesn't work standalone.
- *Single generic wizard framework* — rejected; YAGNI; 4 small models are simpler than abstraction.

---

## R5: Link flow UX (standalone)

**Decision**: Parent = searchable branch list (all tree names). Child = text input with validation (reuse char input pattern from `mrflow.go` title step). Confirm screen: `Link <parent> → <child>? [y/N]` before `Tree.Link` + `SaveTree`.

**Rationale**: `tree.Link` accepts new child names; list-only child picker would block adding new branches. Parent picker satisfies SC-007 for parent selection; child name typed once with confirm guard.

**Alternatives considered**:
- *Both parent and child as lists* — rejected; cannot add new branch names.
- *Copy main app link screen* — rejected; spec requires picker quality, not raw dual-field typing.

---

## R6: Unlink flow UX (standalone)

**Decision**: Searchable branch list → confirm screen matching embedded TUI message:

```text
Remove "<branch>" and N branches from tree? [y/N]
```

On confirm: `UnlinkSubtree(branch)` → `SaveTree()` → success result screen → quit.

**Rationale**: Reuses `SubtreeNames` count from 003-unlink-branch. Consistent confirm pattern with embedded `u` key and sync/MR confirms.

**Alternatives considered**:
- *Require parent+child like CLI* — rejected; TUI user picks subtree root directly (same as embedded `u` on selected branch).

---

## R7: Init flow UX

**Decision**: Steps:
1. Detect git root; show project path and proposed ID.
2. Confirm: `Register this repository with branchy? [y/N]`
3. Run `project.Init({Force: false})`.
4. Success: show imported/scaffolded message + branch count; Enter to exit.
5. Error states: already registered (suggest `init --force`), not a git repo.

**Rationale**: Init is the only command that runs outside a registered project. Wizard replaces one-line `fmt.Printf` for discoverability. `--force` flag forces plain CLI per any-flag rule.

**Alternatives considered**:
- *Init always plain CLI* — rejected per grilling (all interactive commands in scope).
- *Auto-init without confirm* — rejected; registration is a meaningful action.

---

## R8: Projects flow UX

**Decision**: Read-only searchable list of registered projects (id + path). Enter on item shows detail panel (id, path, branch count). No navigation into main tree app. `q`/Esc exits.

**Rationale**: Spec assumption: "selecting a project may show details but need not launch main tree." Browse-only keeps scope bounded.

**Alternatives considered**:
- *Open main tree on select* — rejected; violates FR-012 pattern and blurs command boundaries.
- *Keep plain text list* — rejected; core feature ask.

---

## R9: `branchy mr` any-flag behavior change

**Decision**: Update `mrCmd` so `UseTUI(cmd)` is false when any flag is set — including `--source` without `--target`. Partial-flag invocations print usage error or require `--target` in plain CLI mode.

**Rationale**: Spec FR-011 + edge case "any single flag → plain CLI." Current code allows `--source` with TUI for target pick; must align.

**Alternatives considered**:
- *Grandfather mr partial-flag TUI* — rejected; inconsistent with any-flag rule across commands.

---

## R10: Minimum terminal size

**Decision**: On `tea.WindowSizeMsg`, if `!MinSizeOK(w, h)`, flows render `RenderTooSmall()` full-screen with `q` to quit. No partial broken layouts.

**Rationale**: Spec edge case + FR-014. 80×24 is standard minimum referenced in acceptance scenarios.

**Alternatives considered**:
- *Reflow for any size* — rejected; disproportionate effort for rare tiny terminals.
- *Silent truncation* — rejected; poor UX.

---

## R11: CLI link/unlink arg arity change

**Decision**: Change `linkCmd` and `unlinkCmd` from `cobra.ExactArgs(2)` to `cobra.RangeArgs(0, 2)` with validation in `RunE`:
- 0 args + TUI → dedicated flow
- 2 args → existing one-shot logic
- 1 arg → error: `requires 0 or 2 arguments`

**Rationale**: Cobra must accept zero args for interactive mode while preserving scripted two-arg form.

**Alternatives considered**:
- *Separate `branchy link-tui` command* — rejected; user expects `branchy link` without args.

---

## R12: Non-TTY fallback for interactive commands

**Decision**: When `!IsTTY()` and no flags/args would force scripted mode:

| Command | Fallback |
|---------|----------|
| `sync` | Error: `sync requires --from in non-interactive mode` |
| `link` / `unlink` | Error: requires 2 positional args |
| `init` | Plain one-liner init (current behavior) |
| `projects` | Plain tab-separated list (current behavior) |
| `mr` | Error: requires `--source` and `--target` |

**Rationale**: FR-006/SC-006 — never hang. Init/projects have safe plain fallbacks; others need explicit args.

**Alternatives considered**:
- *Fail all non-TTY without flags* — acceptable but worse DX for `init`/`projects` in CI.
