# Contract: CLI Command Layout

**Feature**: `012-split-cli-commands` | **Version**: 1.0 (draft)

## Overview

Each user-facing subcommand lives in its own file under `internal/cli`. Handlers parse input, call spec-011 operations, and print. Shared TTY and flag-changed detection live in `mode.go`. User-visible behavior is frozen.

---

## Package layout

| File | MUST contain | MUST NOT contain |
|------|--------------|------------------|
| `root.go` | `rootCmd`, `Execute`, `AddCommand` init | Subcommand bodies, flag defs, sync/MR printing |
| `mode.go` | `IsTTY`, `UseTUI`, `anyFlagChanged` | Command-specific logic |
| `init.go` | `initCmd`, `--force`, scripted + TUI paths | Tree/git logic beyond `project.Init` |
| `sync.go` | `syncCmd`, flags, `runSyncCLI` | `sync.Run` reimplementation, GitLab client |
| `mr.go` | `mrCmd`, flags, `runMRFlags` | Direct GitLab calls |
| `link.go` | `linkCmd`, arg rules, `project.Link` | `Tree.Link`, `SaveTree` |
| `unlink.go` | `unlinkCmd`, arg rules, `project.Unlink` | `UnlinkSubtree`, edge validation block |
| `projects.go` | `projectsCmd`, list printing | Index mutation |

---

## Handler contract (all commands)

1. Resolve `*project.Project` when the command needs a registered project (same errors as today).
2. If `UseTUI(cmd)` (and positional rules allow), delegate to the existing `tui.Run*` entry — no duplicate wizard.
3. Otherwise run the scripted path: validate flags/args → **one** operation call → print using existing format strings.
4. Return errors to cobra unchanged (stderr + exit 1 via `main.go`).

**Forbidden in every command file**:

- `p.Tree.Link` / `UnlinkSubtree` + `SaveTree`
- `gitlab.Client` or `internal/gitlab` import
- Inline `HasEdge` / subtree walks for unlink validation
- Copy-pasted `IsTTY` / flag-changed logic

**Allowed exception (`sync.go` only)**: read-only `p.Tree.Names()` and `p.Tree.CollectEdges()` for the interactive `--from` picker and pre-sync plan print (outcome freeze; no mutation).

---

## Subcommand invariants (unchanged behavior)

### `branchy` (no args)

- TTY: `tui.Run(preselected)` with optional CWD project
- Same Short/Long as today

### `init`

- TUI when `UseTUI`
- Scripted: `project.Init(InitOptions{Force: force})` → `Registered %q at %s`
- Flag: `--force`

### `sync`

- TUI when `UseTUI` → `tui.RunSync`
- Scripted: `runSyncCLI` — interactive `--from` picker when TTY and from empty; `sync.Run` with Confirm; summary counts; browser prompt at end
- Flags: `--from`, `-y` / `--yes`

### `mr`

- TUI when `UseTUI` → `tui.RunMR`
- Scripted: `runMRFlags` — y/N unless `-y`; `mr.Create`; exit non-zero on `ActionFailed`
- Flags: `--source`, `--target`, `--title`, `-y`

### `link`

- TUI when no args and `UseTUI` → `tui.RunLink`
- Scripted: two args → `p.Link` → `Linked %s → %s`
- Args: 0 or 2 (1 arg remains error)

### `unlink`

- TUI when no args and `UseTUI` → `tui.RunUnlink`
- Scripted: two args → `p.Unlink` → `Unlinked %s → %s (%d branches removed)`
- Args: 0 or 2

### `projects`

- TUI when `UseTUI` → `tui.RunProjects`
- Scripted: tab-separated id/path or `No projects registered.`

---

## Shared mode rules (`mode.go`)

| Rule | Contract |
|------|----------|
| `UseTUI(cmd)` | `IsTTY() && !anyFlagChanged(cmd)` |
| Flag changed | Any flag with `Changed == true` forces scripted path |
| Tests | `mode_test.go` remains the home for mode unit tests |

Commands MUST NOT reimplement these rules.

---

## Program entry

`cmd/branchy/main.go`:

```go
if err := cli.Execute(); err != nil { ... os.Exit(1) }
```

No command definitions in `main`.

---

## Compatibility

- Subcommand names, flags, shorthands, help text purpose, stdout lines, exit codes: unchanged
- Product usable without specs 013–015
- Depends on spec 011 operations being the sole domain entry points

---

## Verification grep (post-implementation)

```bash
# Must return no matches outside tests (narrow audit; sync.go read-only Tree.* is allowed):
rg 'Tree\.Link|SaveTree|UnlinkSubtree|HasEdge|internal/gitlab' internal/cli/

# Optional: confirm sync-only read-only tree access (expected matches in sync.go only):
rg 'p\.Tree\.(Names|CollectEdges)' internal/cli/

# Must show one file per command:
ls internal/cli/{root,mode,init,sync,mr,link,unlink,projects}.go
```
