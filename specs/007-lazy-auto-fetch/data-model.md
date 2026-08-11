# Data Model: Lazy Auto-Fetch

**Feature**: `007-lazy-auto-fetch` | **Date**: 2026-08-11

## Entities

### DefaultRemote (derived, not persisted)

The remote used for the background update.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Name` | string | `origin` if configured, else the first `git remote` name |
| `Present` | bool | False when the repo has no remotes |

**Validation**: Empty name ⇔ `Present == false` ⇔ skip fetch (not a failure).

---

### RemoteUpdate (ephemeral, one per tea.Cmd)

A scheduled background fetch for one project.

| Attribute | Type | Description |
|-----------|------|-------------|
| `ProjectID` | string | Branchy project id (for FR-009) |
| `Path` | string | Absolute git work-tree path |
| `Err` | error | Nil on skip or success; set on `git fetch` failure |

**Lifecycle**:

```text
selectProject / SyncFlow.Init
    → local inbound snapshot (existing InboundCountMap)
    → tea.Cmd FetchDefaultRemote(Path)
            → RemoteUpdate{ProjectID, Path, Err}
                    → apply if ProjectID matches current
                    → ignore Err (keep snapshot)
                    → on nil Err: recompute InboundCountMap, refresh views
```

---

### FetchFlight (in-process, per directory)

Single-flight coordinator inside `git.FetchDefaultRemote`.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Dir` | string | Absolute repo path (map key) |
| `Done` | signal | Closed when the shared fetch finishes |
| `Err` | error | Result visible to waiters |

**Rules**:

- Second caller for the same `Dir` waits on the existing flight and receives the same `Err`.
- After the flight is removed, a later call may start a new fetch.
- Does not persist across process restarts.

---

### Local snapshot / InboundCountMap (existing, 006)

Unchanged entity. This feature only **recomputes** it after a successful update and pushes it into:

- `BranchTreeView.inbound`
- `SyncFlowModel.inbound` (+ picker badges / confirm context)

---

## State transitions

### Main tree

```text
Project selected
    → counts = local snapshot          [first paint]
    → fetch in flight
         ├─ success + still this project → counts = new snapshot; View updates
         ├─ failure / no remote          → counts unchanged; View unchanged
         └─ user switched project        → drop this result
```

### Sync (embedded or standalone)

```text
Sync opened
    → counts = local snapshot          [picker / confirm immediately]
    → join in-flight fetch or start one
         └─ success + same project → applyInbound (picker and/or confirm)
```

### Input

Fetch never enters a loading step. Key handling stays on tree/sync as today.

---

## Relationships

```text
project.Project
    └── Path ──► git.DefaultRemote / git.FetchDefaultRemote
                      └── RemoteUpdate
                            └── (if ok) loadInboundCounts
                                  ├── BranchTreeView
                                  └── SyncFlowModel (picker + confirm)
```

---

## Unchanged entities

- Inbound display rules (badge / confirm / unknown)
- `InvocationMode` / `UseTUI` — fetch is TUI-only
- Confirm / Loading chrome — not used for fetch
