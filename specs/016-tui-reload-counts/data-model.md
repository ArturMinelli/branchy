# Data Model: Manual Remote Count Reload

**Feature**: `016-tui-reload-counts` | **Date**: 2026-09-04

## Entities

### ManualReload (ephemeral, user-triggered)

A user action that schedules the same pipeline as automatic background fetch (007).

| Attribute | Type | Description |
|-----------|------|-------------|
| `Trigger` | key `r` | User input on allowed surfaces |
| `Project` | `*project.Project` | Current project when key pressed |
| `Surface` | enum | `tree`, `sync_picker`, `sync_confirm` |

**Lifecycle**:

```text
User presses r
    → remoteUpdateCmd(project)     [non-blocking tea.Cmd]
    → git.FetchDefaultRemote(path)   [may join in-flight flight]
    → remoteUpdateMsg{projectID, path, err}
    → apply if projectID matches current
    → on nil err: recompute inbound + outbound; refresh views
```

No new persisted state. No loading step.

---

### RemoteUpdate (existing, 007)

Unchanged message type. Manual and automatic reloads produce identical messages.

| Attribute | Type | Description |
|-----------|------|-------------|
| `ProjectID` | string | Branchy project id |
| `Path` | string | Git work-tree path |
| `Err` | error | Nil on success/skip; set on fetch failure |

Manual reload does not add fields (e.g. no `Manual bool`) — apply logic is the same.

---

### FetchFlight (existing, 007)

Unchanged in-process single-flight per absolute repo path inside `git.FetchDefaultRemote`.

**Rules relevant to manual reload**:

- Auto-fetch on open + manual `r` during same flight → one `git fetch`, waiters share result.
- After flight completes, another `r` may start a new fetch.
- Rapid `r` after completion → sequential fetches (not parallel); counts may refresh multiple times — acceptable per spec edge case.

---

### FileChangeCount maps (existing, 006/008)

Unchanged entities `inbound` and `outbound` maps on `BranchTreeView` and `SyncFlowModel`.

Manual reload **recomputes** both maps after successful fetch; display rules unchanged:

| Surface | Direction shown | Zero rule |
|---------|-----------------|-----------|
| Main tree | Active direction (inbound/outbound) | Hide zero |
| Sync picker | Inbound only | Hide zero |
| Sync confirm | Inbound only | Show known zero |

---

## State transitions

### Main tree

```text
User on tree, counts visible
    → press r
    → fetch scheduled (input still live)
         ├─ success + same project → inbound+outbound recomputed; badges refresh; direction preserved
         ├─ failure → counts unchanged; silent
         └─ project switched → result dropped
```

### Sync picker / confirm

```text
User on picker or edge confirm
    → press r
    → same fetch schedule
         └─ success → picker badges and/or confirm line refresh (inbound for sync surfaces)
```

### Interaction with auto-fetch

```text
Tree opened → auto remoteUpdateCmd in Init
User presses r before auto-fetch completes
    → second remoteUpdateCmd joins git single-flight
    → up to two remoteUpdateMsgs may apply (harmless duplicate refresh)
```

---

## Relationships

```text
User key r
    └── remoteUpdateCmd(project)
            └── git.FetchDefaultRemote(path)  [FetchFlight]
                    └── remoteUpdateMsg
                            ├── Model.applyRemoteUpdate → BranchTreeView
                            └── SyncFlowModel.applyFileCounts (if sync open)
```

---

## Unchanged entities

- `project.Project`, `tree.Document`
- `sync.Direction` session state
- Confirm / Loading models — not used for reload
- Scripted CLI invocation mode — no manual reload
