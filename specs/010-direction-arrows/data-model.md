# Data Model: Merge Direction Arrows

**Feature**: `010-direction-arrows` | **Date**: 2026-08-20

## Entities

### CountDirection (session) — unchanged owner, new transitions

Same entity as 008: stored on root TUI `Model`, mirrored onto `BranchTreeView.direction` and mapped to `SyncFlowModel.direction` when sync starts.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Value` | enum | `inbound` \| `outbound` |
| `Owner` | root TUI `Model` | Survives screen changes and project switches |
| `Default` | `inbound` | Every new process |
| `Persistence` | none | Not written to disk or project config |

**State transitions**:

```text
[session start] → inbound
* --(ctrl+up on main tree)→ outbound
* --(ctrl+down on main tree)→ inbound
outbound --(ctrl+up)→ outbound          # no-op
inbound  --(ctrl+down)→ inbound         # no-op
* --(d)→ unchanged
* --(plain up/down/j/k)→ unchanged
* --(quit)→ (discarded)
* --(selectProject / leave sync / MR)→ unchanged
```

Standalone picker has its own run-scoped copy (`SyncFlowModel.direction`), default downward, with the same set-chord transitions until a root is chosen; then it freezes.

---

### Participating branch (derived, per tree row)

| Attribute | Type | Description |
|-----------|------|-------------|
| `Name` | string | Branch row |
| `HasParent` | bool | Not a tree root |
| `HasChildren` | bool | At least one child in the document |
| `Participating` | bool | `HasParent OR HasChildren` |

**Validation**:

- Childless roots: `Participating = false` — no direction glyph
- Every non-root: `Participating = true` (has a parent)
- Roots with children: `Participating = true`

Computed at flatten time; not persisted.

---

### Direction arrow (ephemeral, main tree only)

| Attribute | Type | Description |
|-----------|------|-------------|
| `Glyph` | string | `↓` when inbound, `↑` when outbound, empty when not participating |
| `Column` | 2 cells | `↓ ` / `↑ ` / two spaces — always present so names align |
| `Placement` | — | After the tree connector, immediately before the branch name |
| `Style` | muted | Same as tree connectors (`connectorStyle`) |
| `Surfaces` | main tree | Not on picker, confirm, summary, project list, or other flows |

**Display rules**:

| Condition | Prefix before name |
|-----------|--------------------|
| Inbound + participating | `↓ ` |
| Outbound + participating | `↑ ` |
| Not participating | `  ` (pad) |

Independent of file-change badge visibility (zero hide / `?` / missing). Selection marker `▸ ` stays a separate leading column.

---

### Main-tree help footer

| Attribute | Type | Description |
|-----------|------|-------------|
| `KeyHints` | string | Includes `ctrl+↑: outbound` and `ctrl+↓: inbound` in **both** modes; never `d: show …` |
| `ModeCue` | string | `counts: inbound (parent→child)` or `counts: outbound (child→parent)` |

Derived solely from `CountDirection`.

---

### Standalone picker help

| Attribute | Type | Description |
|-----------|------|-------------|
| `KeyHints` | string | Same `ctrl+↑: outbound` / `ctrl+↓: inbound` set-chords; no tree arrows |
| `ModeCue` | string | `sync: downward (parent→child)` or `sync: upward (child→parent)` |

Confirm / summary / browser help do not include these fields.

---

## Relationships

```text
Model.diffDirection (session)
        │
        ├── BranchTreeView.direction
        │         ├── renderRow → direction column (↓/↑/pad) + formatFileChangeBadge(active map)
        │         └── participating flag from flatten(Document)
        │
        └── embedded sync start → SyncFlowModel.direction
                  (inbound→Downward, outbound→Upward)

SyncFlowModel.direction (standalone run)
        │
        ├── picker: ctrl+up / ctrl+down set value + badges
        └── after root chosen: frozen (confirms ignore chords)
```

---

## Unchanged entities

- `FileChangeCount` maps and hide-zero / `?` / no-root-count rules (008)
- Sync walks, `Ends`, skip-on-matching-open-MR (009)
- `tree.Document` shape
- Project config / invocation mode — no direction field
- Scripted CLI — no counts, no direction keys
