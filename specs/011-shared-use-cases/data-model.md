# Data Model: Shared Use Cases

**Feature**: `011-shared-use-cases` | **Date**: 2026-08-21

No on-disk schema change. Entities below are the in-process operation results and the existing documents they mutate.

## Entities

### Project (existing)

| Attribute | Type | Description |
|-----------|------|-------------|
| `ID` | string | Index key |
| `Path` | string | Git root |
| `Tree` | `*tree.Document` | Loaded branch-tree.yaml |

**Operations added on this entity** (the persist owners):

| Operation | Input | Success | Failure |
|-----------|-------|---------|---------|
| `Link` | parent, child | in-memory edge + file saved | validation or I/O; no partial file write beyond today’s `SaveTree` |
| `Unlink` | parent, child | subtree removed + file saved + `UnlinkResult` | validation or I/O |
| `Init` | `InitOptions{Force}` | already the register owner | already-registered without force; git root missing |

`SaveTree` remains the write primitive; surfaces that link/unlink MUST go through `Link`/`Unlink` instead of calling `SaveTree` themselves.

---

### UnlinkResult (new)

| Attribute | Type | Description |
|-----------|------|-------------|
| `Parent` | string | Parent passed in (empty when the target was a root) |
| `Child` | string | Subtree root removed |
| `Removed` | int | `len(SubtreeNames(child))` **before** delete (includes the child) |

**Validation** (see [contracts/shared-operations.md](./contracts/shared-operations.md)):

- Child required; parent and child must differ when parent is non-empty
- Child must exist in `Tree.Branches`
- Non-empty parent → `HasEdge(parent, child)` or `edge not found: P → C`
- Empty parent → child must be a root, else error
- Persist after a successful in-memory `UnlinkSubtree`

**State**: no lifecycle beyond the call. Reload via `project.Load` / `ResolveFromCWD` must see the saved tree (FR success scenario).

---

### tree.Document (existing primitive)

Unchanged YAML shape (`branches` map). `Link` / `UnlinkSubtree` remain in-memory only. They do **not** count as the user-facing operations.

| Primitive | Role after this spec |
|-----------|----------------------|
| `Link` | Called only by `project.Link` and tree tests |
| `UnlinkSubtree` | Called only by `project.Unlink` and tree tests |
| `HasEdge` / `ParentOf` / `SubtreeNames` | Used inside `Unlink` and by TUI for confirm copy / passing parent |

---

### Sync begin snapshot (derived, not persisted)

Returned by `sync.Begin`:

| Attribute | Type | Description |
|-----------|------|-------------|
| `Edges` | `[]tree.Edge` | `EdgesBelow(doc, from, dir)` order |
| `Auth` | side effect | `AuthOK` only when `len(Edges) > 0` |

**Errors**: unknown / empty from-branch (same messages as `sync.Run`); GitLab auth failure with the existing `glab auth:` wrap.

**State transitions** (interactive):

```text
[pick root] --(Begin, 0 edges)→ empty screen (no auth)
[pick root] --(Begin, N>0, auth fail)→ error screen, 0 MRs
[pick root] --(Begin, N>0, auth ok)→ edge confirm 1 of N
confirm Yes → RunEdge → next confirm or summary
confirm No  → SkippedByUser → next confirm or summary
esc remaining → summary if any results, else quit (unchanged)
```

---

### Sync Result (existing)

| Attribute | Type | Description |
|-----------|------|-------------|
| `Parent` / `Child` | string | Tree edge |
| `Source` / `Target` | string | MR ends for direction |
| `Action` | string | `created` \| `skipped` \| `failed` (not renamed — spec 015) |
| `URL` | string | Set for created and skip-already-open |
| `Message` | string | `skipped by user` for decline; `open MR already exists` for skip-open |

`SkippedByUser` MUST fill Parent/Child/Source/Target/Action/Message identically to `processEdge` when Confirm returns false. `OpenableURLs` already excludes `skipped by user`.

---

### MR CreateResult (existing)

Unchanged. Manual MR and `processEdge` both use `mr.Create`. Auth for a single pair lives here. This spec does not add fields.

---

## Relationships

```text
Surface (CLI / linkflow / unlinkflow / app.go)
        │
        ├── project.Link ──► tree.Link ──► SaveTree
        ├── project.Unlink ──► HasEdge/root rule ──► UnlinkSubtree ──► SaveTree
        └── project.Init   (already)

Surface (CLI sync)
        └── sync.Run ──► AuthOK ──► EdgesBelow ──► processEdge ──► mr.Create

Surface (TUI sync)
        ├── sync.Begin ──► EdgesBelow [──► AuthOK if N>0]
        ├── sync.RunEdge ──► processEdge ──► mr.Create
        └── sync.SkippedByUser ──► Result (no GitLab)

Surface (CLI mr / mrflow)
        └── mr.Create ──► AuthOK ──► find-or-create
```

---

## Unchanged entities

- On-disk `branch-tree.yaml` and index.yaml
- `sync.Options.Confirm`, `-y`, browser prompt timing
- Direction types (`sync.Direction` vs TUI `diffDirection`) — spec 015
- Main-tree inline link/unlink *screens* — spec 013
- Command file layout — spec 012
