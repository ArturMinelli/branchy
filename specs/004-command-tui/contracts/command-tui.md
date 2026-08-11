# Contract: Unified Command TUI

**Feature**: `004-command-tui` | **Version**: 1.0 (draft)

## Overview

All branchy subcommands follow a dual-mode contract:

| Mode | Trigger | Behavior |
|------|---------|----------|
| **Interactive TUI** | TTY + no flags changed + args not fully specified (per command) | Dedicated full-screen Bubble Tea flow; exits to shell |
| **Scripted CLI** | Any flag set, full positional args, or non-TTY | Plain stdout/stderr; no alt-screen TUI |

**Flag detection**: `cmd.Flags().Visit` — any `Flag.Changed == true` forces scripted mode.

**TTY detection**: `term.IsTerminal(stdout)`.

---

## Commands

### `branchy sync`

```text
branchy sync [--from <branch>] [-y|--yes]
```

| Condition | Mode |
|-----------|------|
| No flags, TTY | TUI: root picker → per-edge confirm → summary → browser |
| Any flag or non-TTY | Plain CLI (existing stdin flow) |

**TUI flow** (`RunSync`):

| Step | Keys | Action |
|------|------|--------|
| Pick root | ↑/↓, enter, filter | Select sync root branch |
| Per-edge | y/n, esc | Create/skip MR per edge |
| Summary | enter | Continue to browser if created > 0 |
| Browser | y/n | Open created MRs in DFS order |

**Exit**: Returns to shell (not main tree app).

---

### `branchy mr`

```text
branchy mr [--source S] [--target T] [--title T] [-y|--yes]
```

| Condition | Mode |
|-----------|------|
| No flags, TTY | TUI (`RunMR`) — unchanged steps |
| Any flag | Plain CLI; requires `--source` AND `--target` for creation |

**Breaking alignment**: `--source` alone no longer opens TUI for target pick.

---

### `branchy link`

```text
branchy link [<parent> <child>]
```

| Condition | Mode |
|-----------|------|
| 0 args, TTY | TUI: parent picker → child name → confirm → save |
| 2 args | Plain: `Linked parent → child` |
| 1 arg | Error: `requires 0 or 2 arguments` |
| 0 args, non-TTY | Error: `link requires <parent> <child> in non-interactive mode` |

---

### `branchy unlink`

```text
branchy unlink [<parent> <child>]
```

| Condition | Mode |
|-----------|------|
| 0 args, TTY | TUI: branch picker → confirm (subtree count) → save |
| 2 args | Plain: `Unlinked parent → child (N branches removed)` |
| 1 arg | Error: `requires 0 or 2 arguments` |
| 0 args, non-TTY | Error: `unlink requires <parent> <child> in non-interactive mode` |

**TUI unlink** picks subtree root directly (no parent arg); CLI scripted mode still requires parent+child edge validation.

---

### `branchy init`

```text
branchy init [--force]
```

| Condition | Mode |
|-----------|------|
| No flags, TTY | TUI wizard: confirm → register → success |
| `--force` or non-TTY | Plain CLI (current one-liner) |

---

### `branchy projects`

```text
branchy projects
```

| Condition | Mode |
|-----------|------|
| TTY | TUI: searchable list → detail on enter |
| non-TTY | Plain tab-separated list (current) |

---

### `branchy` (no subcommand)

Unchanged — launches main tree application TUI.

---

## Shared TUI Chrome

All dedicated flows MUST use:

| Element | Pattern |
|---------|---------|
| Title | `branchy <command>` in bold accent |
| Help footer | Muted `helpStyle` with key hints |
| Confirm | `y` / `n` / `esc` consistent with sync/MR |
| Success | Green `okStyle` |
| Warning | Amber `warnStyle` |
| Error | Red `errStyle` |
| Min size | 80×24; below → "Terminal too small" message |

---

## Key Bindings (common)

| Key | Action |
|-----|--------|
| ↑/↓ or j/k | Navigate lists |
| enter | Select / confirm step |
| y | Yes on confirm screens |
| n | No / skip |
| esc | Back / cancel |
| q / ctrl+c | Quit to shell |
| / or type | Filter lists (bubbles list) |

---

## Exit Codes

| Code | When |
|------|------|
| 0 | Success or user cancel (no error) |
| 1 | Validation error, save failure, MR failure (command-specific) |

---

## Examples

```bash
# Interactive TUI sync
branchy sync

# Scripted sync (any flag disables TUI)
branchy sync --from develop
branchy sync --from develop -y

# Interactive link wizard
branchy link

# Scripted link
branchy link develop feature-x

# Init wizard
branchy init

# Scripted re-import
branchy init --force

# Projects browser
branchy projects

# CI-safe (no hang)
branchy sync --from main -y          # works
branchy sync                       # fails: requires --from in non-interactive mode
```

---

## Registration / Wiring

Each `*cobra.Command` `RunE` in `internal/cli/root.go`:

```text
1. if UseTUI(cmd) && <command-specific interactive condition>
2.     return tui.Run<Command>(...)
3. else
4.     return existing plain CLI logic (or non-TTY fallback error)
```

New exports in `internal/tui`:

- `RunSync(p, opts)`
- `RunLink(p)`
- `RunUnlink(p)`
- `RunInit()`
- `RunProjects()`
