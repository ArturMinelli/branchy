# Research: Bidirectional Sync

**Feature**: `009-bidirectional-sync` | **Date**: 2026-08-13

## 1. One cascade, not a second sync

**Decision**: Keep a single sync path. `tree.Edge` stays a parent/child pair. Direction only chooses (1) walk order of that same set, (2) MR source/target, (3) which file-change map the TUI shows. Do not add a parallel “upward sync” package, command, or flow model.

**Rationale**: Spec FR-003 / FR-006 — same descendant edges, same confirm/skip/summary/browser rules. A second flow would fork every edge case already covered in `SyncFlowModel` and `sync.processEdge`.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Separate `sync up` command / key | Grilling locked `s` + tree `d`; two entry points drift |
| New `UpwardSyncFlowModel` | Duplicates picker, confirm, summary, browser |
| Reverse source/target only in the TUI | Scripted `RunEdge` would stay downward; skip/existing-MR logic would split |

---

## 2. Walk order lives in `internal/tree`

**Decision**:

- Keep `CollectEdges(root)` as today’s DFS **pre-order** (parent edge, then descendants). Downward sync and scripted CLI keep calling it.
- Add `CollectEdgesUpward(root)`: same edges, **post-order** (all descendants of a child, then that child’s edge). Sibling order stays the existing sorted-children order.
- Share one private walker so the two public functions cannot drift on membership.

Post-order is required. Reversing the pre-order slice is **wrong**: a shallow later sibling would appear before a deeper earlier subtree (spec FR-005: deepest first, then existing sibling order).

Example (`main → develop → {feat-a → leaf, feat-b}`):

```text
CollectEdges("develop"):          develop→feat-a, feat-a→leaf, develop→feat-b
CollectEdgesUpward("develop"):    feat-a→leaf, develop→feat-a, develop→feat-b
```

**Rationale**: Traversal is the tree document’s job (same as today). Sync/TUI must not invent a third walk.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `slices.Reverse(CollectEdges(...))` | Breaks deepest-first vs sibling order |
| TUI-only reorder | `sync.Run` and tests would not share the contract |
| Ancestor walk from the selected branch | Grilling rejected; ceiling is the start branch |

---

## 3. Direction type lives in `internal/sync`

**Decision**: Add `sync.Direction` (`Downward` default, `Upward`) on `sync.Options`. Helpers:

```text
Ends(edge, dir) → (source, target)     // Downward: parent→child; Upward: child→parent
EdgesBelow(doc, root, dir) → []Edge    // CollectEdges vs CollectEdgesUpward
```

`processEdge` / `RunEdge` use `Ends` for `mr.Create` title and GitLab find-or-create. `Run` uses `EdgesBelow` then `processEdge`. Zero-value direction is Downward, so CLI `Run` / `CollectEdges` stay unchanged if they never set the field.

TUI maps `diffInbound → Downward`, `diffOutbound → Upward`. Do not move `diffDirection` into `sync` (display mode vs MR direction; TUI already owns the session toggle).

**Rationale**: Creating the MR is sync’s job. `FindOpenMR(source, target)` already matches one direction, so FR-014 (opposite open MR does not skip) falls out of passing the real ends — no extra GitLab query.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Direction only on `SyncFlowModel` | `RunEdge` would stay parent→child; TUI would bypass the owner |
| Reuse `diffDirection` inside `sync` | `sync` must not import `tui`; names differ (counts vs cascade) |
| CLI `--direction` flag | FR-013 / grilling: scripted path stays downward-only |

---

## 4. Result display uses actual MR ends

**Decision**: Keep `Result.Parent` / `Result.Child` as the tree edge. Set `Result.Source` / `Result.Target` from `Ends` at process time. Confirm question, loading line, and summary render `Source → Target`. Tests that identify an edge can still use Parent/Child.

**Rationale**: Today’s `Parent → Child` line is correct only for downward. Upward would lie if the summary kept that arrow.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Swap display in the TUI only | CLI summary (if ever upward) and tests would disagree |
| Drop Parent/Child | Loses stable edge identity independent of direction |

---

## 5. SyncFlowModel becomes direction-aware (supersedes 008 isolation)

**Decision**: `SyncFlowModel` holds `direction`, plus **both** inbound and outbound maps (same loaders as the tree). Active map follows direction.

| Start | Direction source | `d` on picker | `d` on confirm |
|-------|------------------|---------------|----------------|
| Embedded `s` | `Model.diffDirection` at press time | N/A (no picker) | ignored |
| Standalone `branchy sync` | `diffInbound` / Downward | flips pending direction + badges | ignored |

`prepareSyncFrom` calls `sync.EdgesBelow`. Confirm copy:

- Downward: `Create MR {parent} → {child}?` + `{N} files would change on {child}` (inbound of child)
- Upward: `Create MR {child} → {parent}?` + `{N} files would change on {parent}` (outbound of child)

Generalize `formatInboundConfirm` to take the **receiving** branch. Picker badges use `formatFileChangeBadge` on the active map (already shared). Remote refresh reloads **both** maps and keeps direction (`applyInbound` → apply both).

Picker help (standalone only) names current sync direction and `d`, e.g. `sync: downward (parent→child)` / `d: show upward`. Confirm/summary help stays as today (no direction key).

Context line: `Sync down from {ceiling}` vs `Sync up to {ceiling}` so the ceiling’s role is obvious.

**Rationale**: 008 FR-007 (sync always inbound) is **replaced** by 009 FR-008/FR-009 — counts must match the MR. Embedded has no picker, so the tree toggle is the only switch (grilling). Standalone has no tree, so the same `d` on the picker is the grilled equivalent.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Keep sync inbound-only | Contradicts this spec; 008 US4 is obsolete |
| Toggle on confirm screens | FR-012; direction is chosen before the first edge |
| Load only the active map | Remote refresh and a standalone toggle would hitch or show stale badges |

---

## 6. Scripted CLI isolation

**Decision**: No new flags. `internal/cli` keeps `CollectEdges` + `sync.Run` without `Direction`. Grep audit: `internal/cli` has no `Upward` / direction flag. `sync.Run` remains safe if called with zero Options (Downward).

**Rationale**: FR-013 / SC-006. Interactive TUI is the only upward surface.

---

## 7. Testing strategy

**Decision**:

- **`internal/tree`**: same fixture as `TestCollectEdges`; assert `CollectEdgesUpward` membership equals `CollectEdges` (as a set) and order is post-order (deeper edge before its parent edge; siblings keep sorted order).
- **`internal/sync`**: table-test `Ends`; `processEdge` / `RunEdge` with a fake or inspected `CreateRequest` if the seam allows, otherwise unit-test `Ends` + title formatting and keep existing URL-order tests (order follows `Results` insertion, which follows offer order).
- **`internal/tui`**: replace 008 `TestSyncStaysInboundWhileTreeOutbound` with “embedded sync follows tree direction” (outbound `s` → child→parent confirm + outbound count, no `d` on confirm). Standalone: picker starts downward; `d` flips badges and pending direction; after root select, confirms are upward and help has no `d`. Downward regression: inbound `s` still parent→child + inbound count.
- **CLI**: existing sync CLI tests unchanged; optional grep test that `root.go` has no direction flag.

**Rationale**: Same split as 006–008 — tree owns order, sync owns ends, TUI owns copy and keys.

---

## 8. README

**Decision**: Sync section documents both directions: tree `d` then `s`; standalone picker `d`; scripted `--from` remains downward. Do not invent a CLI flag in the docs.

**Rationale**: README is the external key/behavior reference (same as 003/008).
