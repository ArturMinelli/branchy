# Data Model: Unlink Branch from Tree

**Feature**: `003-unlink-branch` | **Date**: 2026-08-11

## Entities

### BranchTreeDocument (existing)

From `internal/tree.Document`. No YAML schema changes.

| Field | Type | Description |
|-------|------|-------------|
| `Branches` | map[string]BranchNode | All known branches keyed by name |

### BranchNode (existing)

| Field | Type | Description |
|-------|------|-------------|
| `Children` | []string | Sorted child branch names |

---

### Subtree (derived, in-memory)

The unit of removal for unlink operations.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Root` | string | Branch selected for removal (TUI) or `child` arg (CLI) |
| `Members` | []string | Root plus all descendants via DFS |
| `Count` | int | `len(Members)`; shown in TUI confirmation |

**Computation**: `SubtreeNames(root)` — DFS from `root` following `Children` edges; returns sorted or traversal-order names (order irrelevant for deletion; count matters for UI).

---

### Parent→Child Edge (existing, validated)

| Field | Type | Description |
|-------|------|-------------|
| `Parent` | string | Parent branch name |
| `Child` | string | Direct child (subtree root to remove) |

**CLI validation**: `HasEdge(parent, child)` MUST be true before `UnlinkSubtree(child)`.

**TUI cleanup**: After subtree deletion, if `ParentOf(root)` returns a parent, remove `root` from that parent's `Children` (redundant if root already deleted from map — parent cleanup removes string from slice only).

---

### UnlinkRequest (ephemeral, TUI session)

| Field | Type | Description |
|-------|------|-------------|
| `Branch` | string | Selected branch name |
| `SubtreeCount` | int | Precomputed count for confirm message |
| `Confirmed` | bool | User accepted prompt |

---

## Validation Rules

| Rule | Applies to | Error behavior |
|------|------------|----------------|
| `root` non-empty | TUI + CLI | No-op / error |
| `root` exists in `Branches` | TUI + CLI | Error: branch not in tree |
| `parent != child` | CLI | Error (same as link) |
| `parent` and `child` non-empty | CLI | Error |
| Edge `parent → child` exists | CLI | Error: edge not found |
| Selected branch non-empty | TUI | `u` ignored when nothing selected |

---

## State Transitions

### TUI

```text
[Tree view, branch selected]
   │
   ├─ press u ──► screenUnlinkConfirm
   │                  │
   │                  ├─ y / Enter ──► UnlinkSubtree(branch) → SaveTree → refresh → Tree view
   │                  ├─ n / Esc ───► Tree view (unchanged)
   │                  └─ save error ─► errMsg + Tree view
   │
   └─ press u (no selection) ──► no transition
```

### CLI

```text
branchy unlink <parent> <child>
   │
   ├─ validation fail ──► stderr message, exit 1
   └─ success ─────────► UnlinkSubtree(child) → SaveTree → stdout success, exit 0
```

---

## Mutation Algorithm (`UnlinkSubtree`)

```text
1. If root not in Branches → return error
2. members ← SubtreeNames(root)
3. For each name in members → delete Branches[name]
4. If parent, ok ← ParentOf(root); ok → remove root from Branches[parent].Children
5. Return nil
```

**Note**: Step 4 runs after step 3; parent node still exists in map (unless parent is inside subtree — impossible since parent is ancestor, not descendant). For root with no parent, step 4 is skipped.

---

## Relationships

```text
Project 1──1 Document (branch tree)
UnlinkRequest ──targets──► Subtree (derived from Document)
CLI args (parent, child) ──validate──► Edge ──triggers──► UnlinkSubtree(child)
```

---

## Unchanged External Systems

- Git branches: not deleted
- GitLab merge requests: not closed
- `~/.config/branchy/index.yaml`: unchanged
