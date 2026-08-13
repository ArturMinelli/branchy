# Data Model: Bidirectional Sync

**Feature**: `009-bidirectional-sync` | **Date**: 2026-08-13

## Entities

### SyncDirection

| Attribute | Type | Description |
|-----------|------|-------------|
| `Value` | enum | `Downward` (default) \| `Upward` |
| `Owner` | `sync.Options` / `SyncFlowModel` | One value per sync run |
| `Default` | `Downward` | Zero value; standalone picker start; scripted CLI |

**Mapping from tree count mode** (embedded `s` only):

| `Model.diffDirection` | SyncDirection |
|-----------------------|---------------|
| inbound | Downward |
| outbound | Upward |

**State transitions**:

```text
[standalone open] → Downward
Downward --(d on standalone picker)→ Upward
Upward   --(d on standalone picker)→ Downward
[embedded s] → tree direction at keypress (locked for the run)
* --(root chosen / first confirm)→ direction frozen
* --(process exit)→ discarded
```

Standalone picker direction is **not** written to `Model.diffDirection` and is not remembered after quit.

---

### Sync ceiling

| Attribute | Type | Description |
|-----------|------|-------------|
| `Branch` | string | Selected tree row (embedded) or picked root (standalone) |
| `Role (down)` | highest **source** | First edges are this branch → its children |
| `Role (up)` | highest **target** | Last edges land in this branch; no ancestor edges |

**Validation**: Must exist in `tree.Document.Branches`. Empty descendant set → empty run (no confirms, no browser prompt).

---

### Sync edge (unchanged identity, directed use)

`tree.Edge` is still `{Parent, Child}`.

| Direction | Offer order | MR source | MR target | Count map | Confirm receiving branch |
|-----------|-------------|-----------|-----------|-----------|--------------------------|
| Downward | `CollectEdges` (pre-order) | Parent | Child | inbound[Child] | Child |
| Upward | `CollectEdgesUpward` (post-order) | Child | Parent | outbound[Child] | Parent |

**Validation**:

- Upward edge set == downward edge set for the same ceiling (FR-003)
- Parent of every edge is the ceiling or a descendant of the ceiling
- Post-order: for every edge `P→C`, every edge in C’s subtree appears earlier

---

### Sync Result (extended)

| Attribute | Type | Description |
|-----------|------|-------------|
| `Parent` | string | Tree parent (edge identity) |
| `Child` | string | Tree child (edge identity) |
| `Source` | string | MR source (`Ends`) |
| `Target` | string | MR target (`Ends`) |
| `Action` | string | created / skipped / failed |
| `URL` | string | optional |
| `Message` | string | optional |

**Display**: `{Source} → {Target}: {Action}` — not Parent → Child.

**Skip rule**: `FindOpenMR(Source, Target)` only. An open MR with swapped ends does not skip.

**Browser batch**: `CreatedURLs` / offer-order insertion unchanged; upward runs are deepest-first because results append in offer order.

---

### SyncFlowModel (extended)

| Attribute | Type | Description |
|-----------|------|-------------|
| `direction` | SyncDirection | Mapped from tree or picker |
| `fromBranch` | string | Ceiling |
| `edges` | []Edge | From `EdgesBelow` |
| `inbound` | map[string]FileChangeCount | Keyed by child |
| `outbound` | map[string]FileChangeCount | Keyed by child |
| `step` | existing enum | Unchanged machine |

**Lifecycle**:

```text
newSyncFlowModel(embedded, from, treeDir)
        → direction = map(treeDir)
        → load both count maps
        → prepareSyncFrom(from) if from != ""

newSyncFlowModel(standalone, "", _)
        → direction = Downward
        → load both maps
        → picker badges = inbound
        → d flips direction + rebuilds badges from the other map
        → enter → prepareSyncFrom(name) with frozen direction

remoteUpdate success
        → reload both maps
        → direction unchanged
        → refresh picker or current confirm from the active map
```

---

### Picker help (standalone root step only)

| Attribute | Downward | Upward |
|-----------|----------|--------|
| Mode cue | `sync: downward (parent→child)` | `sync: upward (child→parent)` |
| Toggle hint | `d: show upward` | `d: show downward` |

Confirm, summary, browser, empty, and error steps MUST NOT list `d` as a direction toggle.

---

## Relationships

```text
Model.diffDirection  --(embedded s)→  SyncFlowModel.direction
Standalone picker d  --(before root)→  SyncFlowModel.direction

SyncFlowModel.direction
        ├── sync.EdgesBelow(tree, ceiling, dir)
        ├── confirm question + loading arrow via Ends
        ├── active count map (inbound vs outbound)
        └── runEdgeCmd → sync.RunEdge(..., dir)

tree.Document
        ├── CollectEdges(ceiling)         → downward
        └── CollectEdgesUpward(ceiling)   → upward
```

```text
008 CountDirection (tree badges)
        └── unchanged session rules
        └── now also selects embedded sync direction (009)
```

---

## Unchanged entities

- Scripted `branchy sync --from` / `-y` — no Direction field, no upward walk
- Manual MR flow (`branchy mr` / `m`)
- `FindOpenMR` / `mr.Create` contracts — callers pass the real ends
- Tree `d` key, footer, and dual count caches on the main tree
- Browser open timing and created-only (plus existing-open) inclusion rules from 002

## Superseded

- 008 FR-007 / US4 / contract “sync surfaces always inbound” — counts and copy now follow SyncDirection
