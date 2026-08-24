# Data Model: Embed TUI Flows

**Feature**: `013-embed-tui-flows` | **Date**: 2026-08-24

No new on-disk schema. This feature adds **session routing** so the main tree hosts the existing link/unlink wizard models instead of a second editor.

## Entities

### Main session (`tui.Model`)

| Field (after) | Role |
|---------------|------|
| `screen` | Includes `screenLink` and `screenUnlink` (rename from `screenUnlinkConfirm` if clearer) |
| `linkFlow` | Embedded `LinkFlowModel` while `screenLink` |
| `unlinkFlow` | Embedded `UnlinkFlowModel` while `screenUnlink` |
| `syncFlow` / `mrFlow` | Unchanged |

**Removed**: `linkParent`, `linkChild`, `linkInput`, `unlinkTarget`, `unlinkCount`, `unlinkConfirm`.

**Relationships**:

```text
tui.Run (main session)
  └── Model
        ├── screenTree  --l--> linkFlow  (Embedded, PrefillParent)
        ├── screenTree  --u--> unlinkFlow (Embedded, PrefillTarget)
        ├── screenTree  --s--> syncFlow   (unchanged)
        └── screenTree  --m--> mrFlow     (unchanged)

tui.RunLink  → LinkFlowModel  (Embedded: false) → process exit
tui.RunUnlink → UnlinkFlowModel (Embedded: false) → process exit
```

---

### Link flow

| Attribute | Description |
|-----------|-------------|
| `opts.Embedded` | Hosted in main session; `finish`/`cancel` must not `tea.Quit` |
| `opts.PrefillParent` | If non-empty and in the tree, start at `stepLinkChild` |
| `step` | parent picker → child name → confirm → success \| error |
| `parent` / `child` | Inputs; persist only via `project.Link` on confirm Yes |
| `finished` / `cancelled` | Host watches `finished` (MR also watches `cancelled`; either is fine if both imply leave) |

**Validation** (unchanged, in `project.Link`): required names, must differ, duplicate edge. Empty child on enter stays on the child step (today’s standalone). Prefill MUST NOT copy those rules.

---

### Unlink flow

| Attribute | Description |
|-----------|-------------|
| `opts.Embedded` | Same quit policy as link |
| `opts.PrefillTarget` | If non-empty and in the tree, start at `stepUnlinkConfirm` with subtree size |
| `target` / `subtreeCount` | Confirm copy uses the same wording as today |
| Persist | `parent, _ := Tree.ParentOf(target)` then `project.Unlink(parent, target)` |

**Validation** (unchanged, in `project.Unlink`): not in tree, edge not found, root-with-empty-parent rules. Host does not reimplement them.

---

### Embedded vs standalone (same models)

| Mode | Program | `q` / cancel / done |
|------|---------|---------------------|
| Standalone | `tea.NewProgram(flow)` | `tea.Quit` → shell |
| Embedded | child of `Model` | `finished` → `screenTree` (success also `selectProject`) |

---

## State transitions

### Link (embedded, parent prefilled)

```text
tree --l--> child name
              | esc → parent picker
              | enter (non-empty child) → confirm
confirm Yes → project.Link → success --done--> tree (refreshed)
confirm No / esc → cancelled → tree (unchanged)
Link error → error --done--> tree (unchanged)
```

### Unlink (embedded, target prefilled)

```text
tree --u--> confirm
confirm Yes → project.Unlink → success --done--> tree (refreshed)
confirm No / esc → cancelled → tree (unchanged)
Unlink error → error --done--> tree (unchanged)
```

### Standalone

Unchanged steps; `tea.Quit` on cancel/done instead of return-to-tree.

---

## Unchanged entities

- `project.Project`, `project.Link`, `project.Unlink` — spec 011
- On-disk branch tree YAML
- Sync/MR flow models and their `Embedded` options
- File-change counts and direction arrows (refreshed via existing `selectProject` / remote update after successful link/unlink)
