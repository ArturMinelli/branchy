# Data Model: Diff Direction Toggle

**Feature**: `008-diff-direction-toggle` | **Date**: 2026-08-12

## Entities

### FileChangeCount (ephemeral, per child branch, per direction)

Shared value type for inbound and outbound badges (replaces the inbound-only mental model of 006’s `inboundCount`).

| Attribute | Type | Description |
|-----------|------|-------------|
| `Child` | string | Tree child (map key) |
| `Parent` | string | Immediate tree parent |
| `Files` | int | File-change count when `ok` |
| `ok` | bool | Comparable vs unknown |

**Direction-specific meaning**:

| Direction | `Files` means |
|-----------|---------------|
| Inbound | Files the child would receive from the parent (MR parent → child) |
| Outbound | Files the parent would receive from the child (MR child → parent) |

**Display rules** (both directions, main tree only for outbound):

| Condition | Badge |
|-----------|-------|
| Root (no parent) | *(none)* — omitted from maps |
| `ok` and `Files == 0` | hidden |
| `ok` and `Files = N > 0` | `N` |
| `!ok` | `?` |

**Validation**:

- `Files` ≥ 0 when `ok`
- Errors MUST NOT coerce to 0
- Inbound N and outbound M for the same child are independent (diverged edges)

---

### CountDirection (session)

| Attribute | Type | Description |
|-----------|------|-------------|
| `Value` | enum | `inbound` \| `outbound` |
| `Owner` | root TUI `Model` | Survives screen changes and project switches |
| `Default` | `inbound` | Every new process |
| `Persistence` | none | Not written to disk or project config |

**State transitions**:

```text
[session start] → inbound
inbound --(d on main tree)→ outbound
outbound --(d on main tree)→ inbound
* --(quit)→ (discarded)
* --(selectProject / leave sync / MR)→ unchanged
```

---

### BranchTreeView count cache

Extends the 006 cache to hold both directions.

| Attribute | Type | Description |
|-----------|------|-------------|
| `inbound` | map[string]FileChangeCount | Keyed by child |
| `outbound` | map[string]FileChangeCount | Keyed by child |
| `direction` | CountDirection | Which map `renderRow` reads |

**Lifecycle**:

```text
selectProject / applyRemoteUpdate(success)
        → loadInboundCounts + loadOutboundCounts
        → set both maps
        → direction unchanged (Model / view)

d key on main tree
        → flip Model.diffDirection
        → mirror onto treeView.direction
        → next View() reads the other map (no git)
```

---

### Main-tree help footer

| Attribute | Type | Description |
|-----------|------|-------------|
| `KeyHints` | string | Includes `d: show outbound` or `d: show inbound` |
| `ModeCue` | string | `counts: inbound (parent→child)` or `counts: outbound (child→parent)` |

Derived solely from `CountDirection`. Sync and other screens do not use this footer.

---

## Relationships

```text
Model.diffDirection (session)
        │
        ├── BranchTreeView.direction + inbound/outbound maps
        │         └── renderRow → formatFileChangeBadge(active map)
        │
        └── (does not feed)
                  SyncFlowModel.inbound  ← loadInboundCounts only
```

```text
tree.Document.ParentOf(child)
        ├── git.InboundFiles(dir, parent, child)  → inbound map
        └── git.OutboundFiles(dir, parent, child) → outbound map
```

---

## Unchanged entities

- Sync `InboundCountMap` / picker badges / confirm “files would change on {child}” — inbound only
- `tree.Document` — still single-parent
- Project config / invocation mode — no direction field
- Scripted CLI — no counts
