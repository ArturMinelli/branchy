# branchy

Local CLI for visualizing branch hierarchies and creating GitLab merge requests along parent→child edges.

## Getting started

1. Install [Go](https://go.dev/) 1.26+ and [glab](https://gitlab.com/gitlab-org/cli), then run `glab auth login`.
2. Clone and install branchy (see [Install](#install)).
3. In a GitLab git repo, run `branchy init` (see [Set up a git repo](#set-up-a-git-repo)).
4. Run `branchy` to open the branch tree TUI.

## Prerequisites

- [Go](https://go.dev/) 1.26+
- [glab](https://gitlab.com/gitlab-org/cli) — install it and authenticate once with `glab auth login`
- A git repository whose `origin` remote points at GitLab (branchy uses glab to create merge requests)

## Install

```bash
git clone https://github.com/ArturMinelli/branchy.git
cd branchy
go install ./cmd/branchy
```

Make sure `$HOME/go/bin` (or your `GOPATH/bin`) is on your `PATH`, then verify:

```bash
branchy --help
```

## Set up a git repo

branchy does not add config files to your application repository. Registration links your repo path to a local project entry and branch tree under `~/.config/branchy/`.

### 1. Register the repo

From the root of your git checkout:

```bash
branchy init
```

In a terminal (TTY), this opens a registration wizard. For scripted use:

```bash
branchy init --force   # re-import branch tree for an already registered repo
```

On first registration, branchy:

- Detects the git root and records it in `~/.config/branchy/index.yaml`
- Imports an existing team tree from `repo/branch-tree.yaml` if that file is present
- Otherwise scaffolds a minimal tree with a single `main` branch

### 2. (Optional) Share a branch tree with your team

Teams can commit a seed file at `repo/branch-tree.yaml` in the application repo. When a teammate runs `branchy init`, branchy imports that file into their local `~/.config/branchy/` store.

Example `repo/branch-tree.yaml`:

```yaml
branches:
  develop:
    children:
      - feature-a
  feature-a:
    children:
      - feature-b
  feature-b:
    children: []
```

After import, each developer's working copy of the tree lives locally. Edits made with `branchy link` / `branchy unlink` (or the TUI) are saved under `~/.config/branchy/`, not back into the repo. Re-run `branchy init --force` to pull in an updated `repo/branch-tree.yaml`.

### 3. Build the branch tree

Add parent→child edges for branches you want to sync or open MRs between:

```bash
branchy link develop feature-a
```

Or press `l` in the main TUI. Linking updates the local branch tree only — it does not create git branches.

### 4. Open the TUI

From anywhere inside the registered repo:

```bash
branchy
```

If you have multiple registered projects, branchy auto-selects the project matching the current directory.

List registered projects:

```bash
branchy projects
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

The index maps project IDs to local git checkout paths. The branch tree defines which branches exist and how they relate. Optional `repo/branch-tree.yaml` in a git repo is only read during `branchy init` (or `init --force`).

## Usage

### TUI (default)

```bash
branchy
```

Opens the branch tree viewer. Keys:

- `s` — sync MRs from selected branch (inbound = downward parent→child; outbound = upward child→parent)
- `m` — create a manual MR between two branches
- `l` — link a new child branch (config only)
- `u` — unlink selected branch and its subtree (config only)
- `ctrl+↑` — set outbound direction (child→parent); rows with a count show `↑` after it
- `ctrl+↓` — set inbound direction (parent→child); rows with a count show `↓` after it
- `r` — reload file-change counts from the default remote (`origin`)
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

Interactive sync follows the tree’s count direction. **Downward** (inbound, the default) walks descendants top-down and creates parent → child merge requests. **Upward** (outbound) walks the same descendant edges deepest-first and creates child → parent merge requests, stopping at the branch you started from.

In the TUI, press Control+Up on the main tree for outbound (child→parent) or Control+Down for inbound (parent→child), then `s` on a branch to sync in that direction. When a row shows a file-change count (or `?`), that badge is followed by `↑` or `↓`. Standalone `branchy sync` (no flags) starts downward and offers the same Control+arrow keys on the root picker. Each edge is confirmed individually — there is no bulk "create all" step. Edges that already have an open MR in that same direction are skipped (with the existing URL shown), not failed.

After sync completes, if any confirmed edges have an MR URL (newly created or already open), branchy asks whether to open them in the browser. Tabs open in offer order. User-declined edges are excluded.

Scripted sync (`--from` / `-y`) stays downward only — parent → child, top-down. There is no direction flag.

```bash
branchy sync                          # TUI: picker (ctrl+↑ / ctrl+↓ set direction) → per-edge confirm
branchy sync --from develop           # plain CLI: downward only; per-edge prompts + end browser prompt
branchy sync --from develop -y        # plain CLI: skip per-edge prompts; browser still prompts at end
```
