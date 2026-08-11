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

Imports `repo/branch-tree.yaml` if present, otherwise scaffolds a minimal tree.

### TUI (default)

```bash
branchy
```

Opens the branch tree viewer. Keys:

- `s` — sync MRs from selected branch
- `m` — create a manual MR between two branches
- `l` — link a new child branch (config only)
- `esc` — back to project picker
- `q` — quit

### CLI commands

```bash
branchy mr                          # interactive TUI branch picker
branchy mr --source A --target B -y   # create MR without TUI
branchy sync --from develop           # interactive MR creation
branchy sync --from develop -y        # create all without prompting
branchy link <parent> <child>         # add tree edge
branchy projects                      # list registered projects
```

## Manual MR (`branchy mr`)

Create a GitLab merge request between any two branches in the project's branch tree:

```bash
branchy mr
```

The TUI guides you through source branch, target branch, title (editable), confirmation, and an optional browser open prompt.

**Flags** (skip the TUI when both branches are known):

| Flag | Description |
|------|-------------|
| `--source` | Source branch name |
| `--target` | Target branch name |
| `--title` | MR title (auto-generated if omitted) |
| `--yes` / `-y` | Skip confirmation prompt |

From the main TUI, press `m` on a selected branch to start an MR with that branch as source.

## Sync behavior

For a chosen root branch, branchy walks the tree depth-first and offers to create GitLab MRs for each parent→child edge (`parent` → `child`). Open MRs are skipped.
