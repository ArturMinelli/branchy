# Data Model: Unified Command TUI

**Feature**: `004-command-tui` | **Date**: 2026-08-11

## Entities

### InvocationMode (derived, per command run)

Determines TUI vs plain CLI for a single process invocation.

| Attribute | Type | Description |
|-----------|------|-------------|
| `IsTTY` | bool | stdout is a terminal |
| `FlagsChanged` | bool | any cobra flag explicitly set |
| `ArgCount` | int | positional argument count |
| `Mode` | enum | `interactive_tui` \| `scripted_cli` |

**Rules**:

```text
scripted_cli  IF NOT IsTTY
             OR FlagsChanged
             OR (command IN link, unlink AND ArgCount == 2)

interactive_tui  IF IsTTY AND NOT FlagsChanged AND NOT scripted by args
```

---

### DedicatedCommandFlow (ephemeral, session)

A standalone Bubble Tea program for one subcommand.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Command` | string | `sync`, `projects`, `link`, `unlink`, `init`, `mr` |
| `Steps` | []FlowStep | Ordered screens in the flow |
| `Embedded` | bool | `false` for CLI-launched flows (always quit to shell) |
| `ExitCode` | int | 0 success, 1 error (set before quit) |

**Lifecycle**: `Run*()` → `tea.NewProgram(m, tea.WithAltScreen())` → user completes/cancels → `tea.Quit` → return to shell.

---

### FlowStep (ephemeral)

| Attribute | Type | Description |
|-----------|------|-------------|
| `Name` | string | e.g. `pick_root`, `edge_confirm`, `confirm_link` |
| `Cancelled` | bool | user aborted |
| `Finished` | bool | terminal step reached |

---

### SharedTUIChrome (presentation, in-memory)

Consistent visual tokens reused across flows.

| Token | Usage |
|-------|-------|
| `Title` | Bold accent — command name header |
| `Help` | Muted footer — key bindings |
| `OK` | Success lines (created) |
| `Warn` | Skipped / caution |
| `Error` | Failures |

**Minimum viewport**: width ≥ 80, height ≥ 24; otherwise `too_small` step.

---

### SyncFlowSession (extends existing SyncFlowModel)

| Field | Type | Description |
|-------|------|-------------|
| `FromBranch` | string | Root branch; empty triggers `stepSyncPickRoot` |
| `Edges` | []Edge | DFS-collected edges below root |
| `Results` | []Result | Per-edge outcomes |
| `Step` | syncStep | Current step including new `stepSyncPickRoot` |

**New step** `stepSyncPickRoot`: searchable list of `Tree.Names()` (root candidates).

---

### LinkFlowSession (ephemeral)

| Field | Type | Description |
|-------|------|-------------|
| `Parent` | string | Selected parent branch |
| `Child` | string | Typed child branch name |
| `Confirmed` | bool | User accepted confirm screen |

**Mutation**: `Tree.Link(parent, child)` → `SaveTree()`.

---

### UnlinkFlowSession (ephemeral)

| Field | Type | Description |
|-------|------|-------------|
| `Target` | string | Subtree root to remove |
| `SubtreeCount` | int | `len(SubtreeNames(target))` |
| `Confirmed` | bool | User accepted confirm |

**Mutation**: `Tree.UnlinkSubtree(target)` → `SaveTree()`.

---

### InitFlowSession (ephemeral)

| Field | Type | Description |
|-------|------|-------------|
| `RepoPath` | string | Git root detected |
| `ProjectID` | string | Derived slug for registration |
| `Outcome` | enum | `registered`, `already_exists`, `error` |
| `BranchCount` | int | Branches in tree after init |

**Mutation**: `project.Init({Force: false})` on confirm.

---

### ProjectsFlowSession (ephemeral, read-only)

| Field | Type | Description |
|-------|------|-------------|
| `Projects` | []*Project | From `project.ListAll()` |
| `Selected` | *Project | Highlighted list item |
| `DetailVisible` | bool | Enter pressed on item |

No persistence mutations.

---

## State Transitions

### Sync (standalone `branchy sync`)

```text
[RunSync, fromBranch empty]
   │
   ├─ stepSyncPickRoot ──enter──► stepSyncEdgeConfirm (or stepSyncEmpty / stepSyncError)
   │
   ├─ stepSyncEdgeConfirm ──y──► stepSyncProcessing ──► next edge or stepSyncSummary
   │                      ──n──► skip edge, continue
   │                      ──esc──► stepSyncSummary (partial)
   │
   ├─ stepSyncSummary ──enter (created > 0)──► stepSyncBrowser
   │                ──enter (no created)──► quit
   │
   └─ stepSyncBrowser ──y/n──► quit (shell)
```

### Link (standalone `branchy link`)

```text
stepPickParent ──► stepEnterChild ──► stepConfirm ──► stepSuccess ──► quit
       │                  │               │
       └──── esc ─────────┴───────────────┴──► quit (cancelled)
```

### Unlink (standalone `branchy unlink`)

```text
stepPickBranch ──► stepConfirm ──► stepSuccess ──► quit
```

### Init (standalone `branchy init`)

```text
stepConfirmRegister ──y──► stepRunning ──► stepSuccess | stepError ──► quit
                    ──n──► quit (cancelled)
```

### Projects (standalone `branchy projects`)

```text
stepList ──enter──► stepDetail ──esc──► stepList ──q──► quit
```

---

## Validation Rules

| Rule | Command | Behavior |
|------|---------|----------|
| Registered project required | sync, link, unlink, mr | Error screen or stderr |
| Git repo required | init | Error screen |
| `len(branches) >= 1` for sync pick | sync | Empty tree message |
| `glab auth` | sync | `stepSyncError` (existing) |
| Parent exists in tree | link | Picker only lists tree names |
| Child non-empty | link | Block confirm |
| Target in tree | unlink | Picker lists tree names |
| Not TTY + no flags | sync, link, unlink, mr | Actionable error message |
| Terminal < 80×24 | all TUI | `too_small` screen |

---

## Relationships

```text
InvocationMode ──selects──► DedicatedCommandFlow | scripted CLI handler
DedicatedCommandFlow ──uses──► SharedTUIChrome
SyncFlowSession ──reads──► Project.Tree ──calls──► sync.RunEdge / sync.CreatedURLs
LinkFlowSession / UnlinkFlowSession ──mutates──► Project.Tree ──persists──► SaveTree
InitFlowSession ──calls──► project.Init
ProjectsFlowSession ──reads──► config index (via ListAll)
```

---

## Unchanged Entities

- `BranchTreeDocument`, `BranchNode`, `Edge`, `Sync.Result`, `Summary` — no schema changes
- Embedded main-app flows (`s`, `m`, `l`, `u` keys) — behavior unchanged
- GitLab MR creation semantics — unchanged
