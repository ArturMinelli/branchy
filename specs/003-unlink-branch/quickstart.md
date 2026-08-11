# Quickstart: Unlink Branch from Tree

**Feature**: `003-unlink-branch` | **Date**: 2026-08-11

Validation guide for proving unlink works end-to-end after implementation.

## Prerequisites

- Go 1.22+ installed
- A git repo registered with branchy (`branchy init`)
- A branch tree with nested branches (see setup below)

## Setup

```bash
# From repo root
go build -o branchy ./cmd/branchy

# Inside a registered git repo
branchy init
```

Seed or edit `~/.config/branchy/projects/<id>/branch-tree.yaml`:

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

Or link via CLI:

```bash
./branchy link develop feature-a
./branchy link feature-a feature-b
```

## Scenario 1: CLI unlink subtree (P1)

```bash
./branchy unlink develop feature-a
```

**Expected**:
- stdout: `Unlinked develop → feature-a (3 branches removed)` (or similar count)
- `feature-a`, `feature-b` absent from tree; `develop` remains with empty children
- Exit code 0
- Re-open TUI: tree shows only `develop`

## Scenario 2: CLI validation — missing edge (P1)

```bash
./branchy unlink develop nonexistent
```

**Expected**:
- stderr error (branch not in tree or edge not found)
- Exit code 1
- Tree file unchanged

## Scenario 3: CLI validation — same branch (P1)

```bash
./branchy unlink develop develop
```

**Expected**:
- stderr: `parent and child must differ`
- Exit code 1

## Scenario 4: TUI unlink with confirmation (P1)

```bash
./branchy
```

**Steps**:
1. Open registered project
2. Navigate to `feature-a` (re-link first if removed in Scenario 1)
3. Press `u`
4. Confirm prompt shows branch name and branch count
5. Press `y`

**Expected**:
- `feature-a` and descendants disappear from tree
- Returns to tree view with updated hierarchy
- No YAML manual edit required

## Scenario 5: TUI cancel (P1)

**Steps**:
1. Select a branch, press `u`
2. Press `n` or `Esc`

**Expected**:
- Tree unchanged
- Back on tree view with no error

## Scenario 6: Unlink leaf branch (edge case)

With tree `develop → feature-a` (no children under `feature-a`):

```bash
./branchy unlink develop feature-a
```

**Expected**:
- Message reports `1 branches removed` (or `1 branch`)
- Only `feature-a` removed; `develop` remains

## Scenario 7: Unlink root subtree (edge case)

Tree with two roots `main` and `orphan`:

```bash
# Select main in TUI, press u, confirm
```

**Expected**:
- `main` subtree removed
- `orphan` (and any other roots) remain

## Scenario 8: Empty tree allowed (edge case)

Remove all branches via repeated unlink.

**Expected**:
- TUI shows `(empty tree)`
- No crash; link still works to rebuild tree

## Scenario 9: Help discoverability (P2)

```bash
./branchy
```

**Expected**:
- Tree view footer includes `u: unlink` alongside `l: link`

## Scenario 10: Unit tests

```bash
go test ./internal/tree/... -run Unlink
go test ./internal/tui/... -run Unlink
```

**Expected**: All tests pass.

## References

- CLI contract: [contracts/unlink-cli.md](./contracts/unlink-cli.md)
- Data model: [data-model.md](./data-model.md)
- Feature spec: [spec.md](./spec.md)
