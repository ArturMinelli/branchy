# Contract: `branchy unlink` CLI

**Feature**: `003-unlink-branch` | **Version**: 1.0 (draft)

## Command

```text
branchy unlink <parent> <child>
```

Removes `child` and its entire subtree from the current project's branch tree configuration. Does not delete Git branches or GitLab merge requests.

**Prerequisites**: Registered project (`branchy init`), valid `branch-tree.yaml` with the specified parent→child edge.

---

## Arguments

| Arg | Required | Description |
|-----|----------|-------------|
| `parent` | yes | Parent branch in the edge to remove |
| `child` | yes | Child branch (root of subtree to delete) |

**Arity**: Exactly 2 arguments (`cobra.ExactArgs(2)`).

---

## Behavior

1. Resolve project from current working directory (`project.ResolveFromCWD()`).
2. Validate `parent` and `child` are non-empty and differ.
3. Validate edge `parent → child` exists in the branch tree.
4. Call `Tree.UnlinkSubtree(child)` — removes `child` and all descendants.
5. Persist via `project.SaveTree()`.
6. Print success message to stdout.

**No interactive confirmation** in CLI (consistent with `branchy link`).

---

## Success Output

```text
Unlinked develop → feature-a (3 branches removed)
```

- Message names the edge (`parent → child`).
- Branch count includes `child` and all descendants removed.
- Exit code: `0`

---

## Error Output

All errors written to stderr (cobra default). Exit code: `1`.

| Condition | Message (representative) |
|-----------|--------------------------|
| Not in registered repo | Project resolution error from `project.ResolveFromCWD()` |
| `parent == child` | `parent and child must differ` |
| Empty parent or child | `parent and child are required` |
| Child not in tree | `branch "<child>" not in tree` |
| Edge does not exist | `edge not found: <parent> → <child>` |
| Save failure | Underlying filesystem error |

---

## Symmetry with `branchy link`

| Aspect | `link` | `unlink` |
|--------|--------|----------|
| Args | `<parent> <child>` | `<parent> <child>` |
| Effect | Add edge; ensure nodes exist | Remove child subtree |
| Confirmation | None | None |
| Persistence | `SaveTree()` | `SaveTree()` |
| Success line | `Linked %s → %s` | `Unlinked %s → %s (%d branches removed)` |

---

## TUI Contract (reference)

Not a separate CLI command; documented for cross-surface consistency.

| Key | Screen | Action |
|-----|--------|--------|
| `u` | Tree view | Open unlink confirmation for selected branch |
| `y` / `Enter` | Unlink confirm | Execute unlink + save |
| `n` / `Esc` | Unlink confirm | Cancel, return to tree |

**Confirm message**:

```text
Remove "feature-a" and 3 branches from tree? [y/N]
```

Help footer on tree view MUST include `u: unlink`.

---

## Examples

```bash
# Remove feature-a and its descendants from under develop
branchy unlink develop feature-a

# Fail: edge does not exist
branchy unlink main feature-a
# stderr: edge not found: main → feature-a

# Fail: same branch
branchy unlink main main
# stderr: parent and child must differ
```

---

## Registration

```go
rootCmd.AddCommand(unlinkCmd)
```

Placed alongside `linkCmd` in `internal/cli/root.go` `init()`.
