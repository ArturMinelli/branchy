# Data Model: Domain / UI Split

**Feature**: `014-domain-ui-split` | **Date**: 2026-08-24

No on-disk schema change. This feature splits **what is stored** from **how it is drawn** and **which process opens URLs**.

## Entities

### Document (unchanged persist shape)

| Field | Role |
|-------|------|
| `Branches` | `map[string]BranchNode` — names and children on disk |
| `Roots()` | Sorted names that never appear as a child |
| `CollectEdges` / `CollectEdgesUpward` | Sync-edge walks (not display) |
| **`WalkDisplay()`** | New: unstyled DFS for drawing consumers |

**Invariant**: `WalkDisplay` MUST NOT import lipgloss, contain box-drawing characters, or know about selection/badges.

---

### DisplayNode (new, in-memory)

One visit in pre-order DFS (roots first, then each node’s children).

| Field | Meaning |
|-------|---------|
| `Name` | Branch name; membership equals `Document` keys reachable from `Roots` |
| `Depth` | 0 for roots; +1 per parent |
| `IsRoot` | `Depth == 0` |
| `IsLast` | Last among siblings at this parent (or last among roots) |
| `LastAtDepth` | For `i < Depth`, whether the ancestor at depth `i` was last among *its* siblings. Length equals `Depth`. Used to draw the spine without a second walk |

**Child order**: copy `node.Children`, `sort.Strings` — same as today’s `flattenTree` / `RenderASCII` / `CollectEdges`. Outcome freeze; YAML list order is not the display order.

**Empty tree**: `WalkDisplay` returns `nil` or empty slice. TUI still renders `(empty tree)` (existing `View`).

**Relationships**:

```text
Document.WalkDisplay()  -->  []DisplayNode
                                 │
                                 ▼
                    TUI flattenTree → []branchRow
                         (glyphs, then lipgloss + badges)
```

---

### branchRow (TUI, unchanged fields)

| Field | Source after this spec |
|-------|------------------------|
| `name` | `DisplayNode.Name` |
| `connector` | `""` / `"└── "` / `"├── "` from `IsRoot` / `IsLast` |
| `prefix` | spine spaces/`│   ` from `LastAtDepth` |

Selection, width, inbound/outbound counts, and direction arrows stay on `BranchTreeView`. Counts still come from local git, not GitLab.

---

### Openable URL (sync, unchanged policy)

`sync.OpenableURLs(summary)` — DFS result order:

| Include | Exclude |
|---------|---------|
| `created` with URL | user-declined skip (`"skipped by user"`) |
| `skipped` already-open with URL | `failed` without URL |

Sync MUST NOT call `browser.Open` / `OpenURLs`.

---

### Sequential open (browser)

| Piece | Owner after this spec |
|-------|------------------------|
| `browser.Open` | Single URL (MR flow unchanged) |
| `browser.OpenURLs` | Ordered list, 200ms between tabs, joined warnings |
| Ask step | CLI stdin prompt; TUI confirm — surfaces only |

**State after sync** (unchanged):

```text
summary
  → OpenableURLs
  → empty? skip ask
  → ask
       ├─ decline → no tabs, summary still lists URLs
       └─ accept  → OpenURLs (order + pause + non-fatal warn)
```

---

## Validation rules

- `WalkDisplay` visit names MUST equal the set of names `flattenTree` would have produced today (all branches; roots sorted; children sorted).
- Glyph tests in `treeview_test.go` MUST still see `└── ` / `├── ` on the TUI rows, not on `DisplayNode`.
- `OpenURLs` MUST NOT change warning text or pause duration.
- TUI MUST NOT import `internal/gitlab`.

---

## Unchanged entities

- `project.Project`, `project.Link` / `Unlink`
- `sync.Begin` / `Run` / `OpenableURLs`
- `mr.Create`
- On-disk YAML
- Direction types and action strings (spec 015)
