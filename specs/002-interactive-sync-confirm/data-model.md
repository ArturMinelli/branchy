# Data Model: Interactive Sync Confirmations

**Feature**: `002-interactive-sync-confirm` | **Date**: 2026-08-11

## Entities

### SyncSession (in-memory, flow state)

Represents one sync invocation from TUI or CLI. Lives for the duration of the command.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `FromBranch` | string | yes | Root branch selected for sync |
| `Edges` | []Edge | yes | DFS-ordered parent→child pairs from `CollectEdges` |
| `Results` | []Result | grows | Per-edge outcomes, appended in edge walk order |
| `CurrentIndex` | int | TUI only | Index of edge awaiting user confirm |
| `BrowserOpened` | bool | no | Whether user confirmed end browser batch |

**Validation rules**:
- `FromBranch` MUST exist in `project.Tree.Branches`
- `Edges` MAY be empty (no children below root)

---

### Edge (existing, `internal/tree`)

| Field | Type | Description |
|-------|------|-------------|
| `Parent` | string | Source branch for MR |
| `Child` | string | Target branch for MR |

**Ordering**: DFS below `FromBranch`; stable order from `Document.CollectEdges`.

---

### Result (existing, `internal/sync`)

| Field | Type | Description |
|-------|------|-------------|
| `Parent` | string | Edge parent |
| `Child` | string | Edge child |
| `Action` | enum | `created` \| `skipped` \| `failed` |
| `URL` | string | MR URL when created or existing MR found |
| `Message` | string | Skip reason, error detail, or `skipped by user` |

**Browser batch eligibility**: `Action == "created"` AND `URL != ""`.

---

### Summary (existing, `internal/sync`)

| Field | Type | Description |
|-------|------|-------------|
| `Results` | []Result | All edge outcomes in DFS order |

**Derived**:
- `CreatedURLs()` → filter `Results` where `Action == "created"`, preserve order
- `HasCreated()` → `len(CreatedURLs()) > 0`

---

### BrowserBatch (derived, ephemeral)

Ordered list of URLs to open at session end.

| Rule | Value |
|------|-------|
| Inclusion | `created` results only |
| Exclusion | `skipped` (user or existing MR), `failed`, declined-before-processing |
| Order | Same as `Results` iteration (DFS among created) |
| Open semantics | Sequential `browser.Open` with 50ms inter-tab delay when count > 1 |

---

## State Transitions (TUI SyncFlowModel)

```text
[Start: s pressed, edges collected]
   │
   ├─ len(edges)==0 ──► stepSyncEmpty ──► [Back to tree]
   │
   └─ len(edges)>0 ──► stepSyncEdgeConfirm (index=0)
         │
         ├─ y ──► stepSyncProcessing ──► [RunEdge async] ──► append Result
         │                                              │
         │                              ├─ more edges ──► stepSyncEdgeConfirm (index++)
         │                              └─ done ───────► stepSyncSummary
         │
         ├─ n ──► append skipped Result ──► (next edge or stepSyncSummary)
         │
         └─ Esc ──► stepSyncSummary (partial results)

stepSyncSummary
   │
   ├─ HasCreated() ──► stepSyncBrowser
   └─ else ──────────► [Done → tree]

stepSyncBrowser
   ├─ y ──► OpenURLs(CreatedURLs) ──► [Done → tree]
   └─ n/Enter ──► [Done → tree]
```

**Cancel**: From `stepSyncSummary` or `stepSyncBrowser`, Enter/Esc returns to tree view.

---

## State Transitions (CLI sync)

```text
[branchy sync --from X]
   │
   ├─ auth / validation
   ├─ print sync plan (unchanged)
   │
   └─ for each edge (DFS):
         ├─ -y: auto-confirm
         └─ else: stdin "Create MR parent → child? [y/N]"
         └─ append Result

[after loop]
   │
   ├─ HasCreated() ──► stdin "Open created MRs in browser? [y/N]"
   │                      └─ y: OpenURLs
   └─ print per-edge summary + totals
```

---

## Relationships

```text
Project 1──1 Document
Document.CollectEdges(from) ──► []Edge (DFS)
SyncSession 1──* Result
Summary aggregates * Result
BrowserBatch derived from Summary (created only)
Each confirmed Edge ──creates──► Result via mr.Create (through sync.RunEdge)
```

---

## Unchanged entities

- **MergeRequest** (GitLab external) — same as `001-mr-command`
- **BranchTreeBranch** — no schema changes
- **MRRequest / MRFlowModel** — manual MR flow out of scope
