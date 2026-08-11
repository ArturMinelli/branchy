# branchy

Local CLI for visualizing branch hierarchies and creating GitLab merge requests along parent→child edges.

## Prerequisites

- [Go](https://go.dev/) 1.22+
- [glab](https://gitlab.com/gitlab-org/cli) (authenticated: `glab auth login`)

## Install

From this directory:

```bash
go install ./cmd/branchy
```

## Config

branchy stores everything under `~/.config/branchy/`:

```
~/.config/branchy/
├── index.yaml
└── projects/
    └── <project-id>/
        └── branch-tree.yaml
```

No config files live inside application repos.

## Usage

### Register a project

Inside a git repository:

```bash
branchy init
```

Imports `repo/branch-tree.yaml` if present, otherwise scaffolds a minimal tree. In a TTY, `branchy init` opens a registration wizard; `branchy init --force` stays plain CLI.

### TUI (default)

```bash
branchy
```

Opens the branch tree viewer. Keys:

- `s` — sync MRs from selected branch
- `m` — create a manual MR between two branches
- `l` — link a new child branch (config only)
- `u` — unlink selected branch and its subtree (config only)
- `esc` — back to project picker
- `q` — quit

### CLI commands

**Interactive mode** (TTY, no flags): dedicated full-screen TUI per command, then exit to shell.

**Scripted mode** (any flag, full positional args, or non-TTY): plain stdout/stdin, no TUI.

```bash
branchy sync                         # interactive sync TUI (root picker → per-edge confirm)
branchy sync --from develop          # plain CLI: per-edge prompts + end browser prompt
branchy sync --from develop -y        # plain CLI: skip per-edge prompts; browser prompts at end

branchy mr                           # interactive MR TUI
branchy mr --source A --target B -y  # plain CLI: create MR without TUI

branchy link                         # interactive link TUI (parent picker → child name)
branchy link <parent> <child>         # plain CLI: add tree edge

branchy unlink                       # interactive unlink TUI (branch picker → confirm)
branchy unlink <parent> <child>       # plain CLI: remove child subtree

branchy init                         # interactive registration wizard
branchy init --force                 # plain CLI: re-import branch tree

branchy projects                     # interactive project browser
branchy projects | cat               # plain CLI: tab-separated list
```

## Manual MR (`branchy mr`)

Create a GitLab merge request between any two branches in the project's branch tree:

```bash
branchy mr
```

The TUI guides you through source branch, target branch, title (editable), confirmation, and an optional browser open prompt.

**Flags** (any flag disables TUI; both branches required for plain CLI):

| Flag | Description |
|------|-------------|
| `--source` | Source branch name |
| `--target` | Target branch name |
| `--title` | MR title (auto-generated if omitted) |
| `--yes` / `-y` | Skip confirmation prompt |

From the main TUI, press `m` on a selected branch to start an MR with that branch as source.

## Sync behavior

For a chosen root branch, branchy walks the tree depth-first and prompts to create a GitLab MR for each parent→child edge (`parent` → `child`). In the TUI (`s` key or `branchy sync`), each edge is confirmed individually — there is no bulk "create all" step. Edges that already have an open MR are skipped (with the existing URL shown), not failed.

After sync completes, if any confirmed edges have an MR URL (newly created or already open), branchy asks whether to open them in the browser. Tabs open in tree order. User-declined edges are excluded. The plain CLI behaves the same way: per-edge prompts (skipped with `-y`), then an optional end browser prompt.

```bash
branchy sync --from develop           # per-edge prompts + end browser prompt
branchy sync --from develop -y        # skip per-edge prompts; browser still prompts at end
```
