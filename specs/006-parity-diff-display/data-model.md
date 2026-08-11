# Data Model: Parent Diff Display

**Feature**: `006-parity-diff-display` | **Date**: 2026-08-11

## Entities

### InboundCount (ephemeral, per child branch)

How many files the child would receive from its immediate parent — GitLab MR files-changed for parent→child.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Child` | string | Branch that would receive the changes |
| `Parent` | string | Immediate tree parent (sync source) |
| `Files` | int | File-change count when `Status` is `ok` |
| `Status` | enum | `ok`, `unknown`, `hidden` |

**Status rules**:

| Status | When | Tree / picker | Edge confirm |
|--------|------|---------------|--------------|
| `ok` and `Files > 0` | Comparable, inbound files exist | show `Files` | show `Files` |
| `ok` and `Files == 0` | Comparable, already matching | **hidden** (no badge) | show `0` |
| `unknown` | Missing ref, git error, unresolvable names | show `?` | `File count unavailable` |
| *(omitted)* | Branch is a tree root (no parent) | no badge | n/a (roots are not confirm targets as children) |

**Validation**:

- `Files` MUST be ≥ 0 when `Status` is `ok`
- `Files` MUST NOT be interpreted when `Status` is `unknown` (never coerce errors to 0)
- Roots MUST NOT appear in the count map as inbound targets

**Identity**: One inbound count per child. The tree is single-parent, so `(Child)` is a unique key.

---

### InboundCountMap (session cache)

In-memory map built once per tree display / sync session.

| Attribute | Type | Description |
|-----------|------|-------------|
| `ByChild` | map[string]InboundCount | Keyed by child branch name |
| `RepoDir` | string | Project path used for `git -C` |
| `ComputedAt` | session | Not persisted; discarded when the flow/tree is rebuilt |

**Lifecycle**:

```text
selectProject / newSyncFlowModel
        → walk doc via ParentOf
        → InboundFiles(dir, parent, child) per child
        → ByChild cache
        → View() reads cache only
```

No refresh after sync: creating MRs does not change local refs.

---

### Branch row display (main tree)

Extends existing `branchRow` with an optional badge.

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Existing |
| `prefix` / `connector` | string | Existing tree chrome |
| `badge` | string | `""`, `"N"`, or `"?"` after formatting |

**Derivation**: `badge = formatInboundBadge(ByChild[name])`. Roots and hidden zeros produce `""`.

---

### Picker item display (sync root picker)

Extends existing `branchItem` with an optional title suffix.

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Branch name; `FilterValue()` stays the bare name |
| `badge` | string | Same formatter as the tree |
| `Title()` | string | `name` or `name  badge` |

Link / unlink / MR pickers leave `badge` empty.

---

### Edge confirm context

Existing `ConfirmOptions.Context` gains one inbound line.

| Surface | Context lines (after this feature) |
|---------|------------------------------------|
| Sync edge | `Sync from {root}`; `{N} files would change on {child}` **or** `File count unavailable`; progress `Edge N of M` remains on the panel |

Question text is unchanged: `Create MR {parent} → {child}?`

---

## Relationships

```text
tree.Document.ParentOf(child)
        └── parent + child
                └── git.InboundFiles(dir, parent, child)
                        └── InboundCount
                                ├── BranchTreeView.badge
                                ├── branchItem.badge (sync picker only)
                                └── ConfirmOptions.Context (sync edge)
```

---

## Unchanged entities

- `tree.Document` / `tree.Edge` — parent lookup already exists
- `InvocationMode` / `UseTUI` — TUI-only feature
- Sync `Result` / summary rows — counts not shown
- Confirm / Loading chrome — confirm gains a context line only
