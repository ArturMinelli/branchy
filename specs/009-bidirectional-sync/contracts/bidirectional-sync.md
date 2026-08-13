# Contract: Bidirectional Sync

**Feature**: `009-bidirectional-sync` | **Version**: 1.0 (draft)

## Overview

Extends [002 sync CLI](../../002-interactive-sync-confirm/contracts/sync-cli.md) and replaces the sync-isolation clause of [008 diff-direction](../../008-diff-direction-toggle/contracts/diff-direction.md). Interactive sync can run **downward** (today) or **upward**. Scripted/flagged `branchy sync` stays downward-only.

---

## Direction

| SyncDirection | Tree mode (embedded) | MR | Walk | Counts |
|---------------|----------------------|----|------|--------|
| Downward | inbound (default) | parent → child | `CollectEdges` pre-order | inbound |
| Upward | outbound | child → parent | `CollectEdgesUpward` post-order | outbound |

```text
Ends(edge, Downward) = (edge.Parent, edge.Child)
Ends(edge, Upward)   = (edge.Child, edge.Parent)
```

---

## Tree walk

### `CollectEdges(root)` — unchanged

DFS pre-order: append `parent→child`, then recurse. Children sorted as today.

### `CollectEdgesUpward(root)` — new

DFS post-order: recurse each sorted child, then append `parent→child`.

**Membership**: for a given `root`, the two functions MUST return the same multiset of edges. Neither includes edges whose parent is an ancestor of `root`.

**Order (upward)**: if `C` is a descendant of `P`, every edge in `C`’s subtree appears before `P→C`.

---

## Sync package

```text
Options.Direction          // default Downward
EdgesBelow(doc, root, dir) → []Edge
Ends(edge, dir)            → (source, target)
Run(p, opts)               // walk + process; CLI omits Direction
RunEdge(p, edge, dir)      // single edge, no confirm
```

`mr.Create` / `FindOpenMR` receive `Ends`. Title: `Sync: {source} → {target}`.

`Result` includes `Source` and `Target` in addition to `Parent` and `Child`.

### Skip

Skip only when an **open** MR exists for `(Source, Target)`. The reverse pair is a different MR.

### CLI (scripted)

```text
branchy sync --from <branch> [-y]
```

No direction flag. Always `CollectEdges` + Downward. Behavior otherwise unchanged from 002.

---

## TUI

### Embedded (`s` on main tree)

1. Direction = current tree `diffDirection` (inbound→Downward, outbound→Upward).
2. Ceiling = selected branch. No root picker.
3. Confirm screens do not bind `d` for direction.

### Standalone (`branchy sync`, no flags)

1. Root picker starts Downward with inbound badges.
2. `d` on the picker flips direction and badges immediately.
3. After a root is chosen, direction is frozen; confirm has no `d` toggle.
4. Next process starts Downward again.

### Confirm copy

| Direction | Question | Count line |
|-----------|----------|------------|
| Downward | `Create MR {parent} → {child}?` | `{N} files would change on {child}` (inbound, show known zero) |
| Upward | `Create MR {child} → {parent}?` | `{N} files would change on {parent}` (outbound, show known zero) |

Unknown comparison: `File count unavailable` (same as today).

### Picker badges

Same hide-zero / `?` rules as the tree, applied to the **active** map. Roots may appear in the list; they still have no badge.

### Picker help (standalone)

MUST include current direction (`sync: downward (parent→child)` or `sync: upward (child→parent)`) and `d: show upward` / `d: show downward`.

Confirm / summary / browser help MUST NOT advertise a direction toggle.

### Context line

- Downward: `Sync down from {ceiling}`
- Upward: `Sync up to {ceiling}`

### Empty ceiling

`No child branches below "{ceiling}".` — same for both directions. No browser prompt.

---

## Browser

Inclusion rules unchanged from 002 (`CreatedURLs` / `OpenableURLs`). Tab order is offer order: pre-order when downward, post-order when upward.

---

## Non-goals

- No scripted `--direction` / `--up`
- No ancestor walk above the ceiling
- No direction toggle on confirm
- No change to manual `branchy mr`
- No persistence of standalone picker direction

---

## 008 amendments

| 008 clause | 009 |
|------------|-----|
| Sync picker / confirm always inbound | Follow SyncDirection |
| Sync screens have no `d` | Standalone **picker** binds `d`; confirm still does not |
| Tree `d` does not affect sync | Tree `d` selects embedded sync direction |
