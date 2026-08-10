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
- `l` — link a new child branch (config only)
- `esc` — back to project picker
- `q` — quit

### CLI commands

```bash
branchy sync --from develop       # interactive MR creation
branchy sync --from develop -y    # create all without prompting
branchy link <parent> <child>     # add tree edge
branchy projects                # list registered projects
```

## Sync behavior

For a chosen root branch, branchy walks the tree depth-first and offers to create GitLab MRs for each parent→child edge (`parent` → `child`). Open MRs are skipped.
