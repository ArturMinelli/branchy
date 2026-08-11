# Data Model: TUI Native Confirm & Loading

**Feature**: `005-sync-tui-polish` | **Date**: 2026-08-11

## Entities

### ConfirmModel (shared component, in-memory)

Binary-choice TUI panel reused across all yes/no steps.

| Attribute | Type | Description |
|-----------|------|-------------|
| `ContextLines` | []string | Optional lines above question (repo path, edge pair, subtree count) |
| `Question` | string | Primary prompt text (without `[y/N]` suffix) |
| `FocusIndex` | int | `0` = No (default), `1` = Yes |
| `Width` | int | Terminal width for border wrapping (from parent flow) |

**Validation**:

- `Question` MUST be non-empty when rendered
- `FocusIndex` MUST be 0 or 1

**State transitions** (on key input):

```text
Pending + left/right/tab  → toggle FocusIndex
Pending + Enter           → choice = focused option
Pending + y               → choice = Yes
Pending + n OR Esc        → choice = No
```

**Output**: `ConfirmChoice` enum: `None` (still pending), `Yes`, `No`

---

### LoadingModel (shared component, in-memory)

Animated wait state for async Bubble Tea commands.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Message` | string | Primary status (e.g. "Creating MR develop → feature-a") |
| `Progress` | string | Optional secondary line (e.g. "Edge 2 of 5") |
| `Spinner` | spinner.Model | Charm spinner instance |

**Lifecycle**:

```text
Init()           → spinner.Tick cmd
Tick msgs        → update spinner frame
Result msg       → parent flow advances; LoadingModel discarded
Key msgs         → ignored (no transition)
```

---

### ConfirmStep (ephemeral, per flow)

Binds a `ConfirmModel` to a flow-specific step and outcome handler.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Flow` | string | `sync`, `mr`, `link`, `unlink`, `init`, `app_unlink` |
| `StepID` | string | e.g. `edge_confirm`, `browser_confirm`, `link_confirm` |
| `OnYes` | action | Flow-specific: start loading cmd, advance step, etc. |
| `OnNo` | action | Flow-specific: skip, cancel, back |

**In-scope confirm steps**:

| Flow | Step | Question pattern |
|------|------|------------------|
| sync (embedded + standalone) | per-edge | Create MR `{parent}` → `{child}`? |
| sync | browser | Open created MRs in browser? |
| mr | create | Create MR `{source}` → `{target}`? |
| mr | browser | Open in browser? |
| link | confirm | Link `{parent}` → `{child}`? |
| unlink | confirm | Remove `{branch}` and N branches? |
| init | confirm | Register this repository with branchy? |
| app (tree) | unlink | Remove `{branch}` and N branches? |

---

### LoadingStep (ephemeral, per flow)

Binds a `LoadingModel` to an in-flight `tea.Cmd`.

| Attribute | Type | Description |
|-----------|------|-------------|
| `Flow` | string | Parent flow name |
| `Operation` | enum | `create_mr`, `open_browser`, `register_project` |
| `Cmd` | tea.Cmd | Async work function |
| `ResultType` | string | Expected completion msg type for routing |

**In-scope loading operations**:

| Flow | Trigger | Operation | Result handling |
|------|---------|-----------|-----------------|
| sync | edge Yes | `create_mr` | Append result; next edge or summary |
| sync | browser Yes | `open_browser` | Warn on failure; finish |
| mr | confirm Yes | `create_mr` | Result or error screen |
| mr | browser Yes | `open_browser` | Warn; finish |
| init | confirm Yes | `register_project` | Success or error screen |

---

### FlowStep extensions (existing models)

New step constants added to existing enums:

**syncStep** (existing `stepSyncProcessing` enhanced):

| Step | Change |
|------|--------|
| `stepSyncEdgeConfirm` | Renders `ConfirmModel` instead of text |
| `stepSyncProcessing` | Renders `LoadingModel` + holds spinner |
| `stepSyncBrowser` | Renders `ConfirmModel` |
| `stepSyncBrowserLoading` | **New** — browser batch in flight |

**mrStep**:

| Step | Change |
|------|--------|
| `stepMRConfirm` | `ConfirmModel` |
| `stepMRLoading` | **New** — async `mr.Create` |
| `stepMRBrowser` | `ConfirmModel` |
| `stepMRBrowserLoading` | **New** — async browser open |

**initStep**:

| Step | Change |
|------|--------|
| `stepInitConfirm` | `ConfirmModel` |
| `stepInitLoading` | **New** — async `project.Init` |

**linkStep / unlinkStep**:

| Step | Change |
|------|--------|
| `stepLinkConfirm` / `stepUnlinkConfirm` | `ConfirmModel` only (no loading) |

**AppModel screen**:

| Screen | Change |
|--------|--------|
| `screenUnlinkConfirm` | Embed `ConfirmModel`; remove `[y/N]` string |

---

## Relationships

```text
DedicatedCommandFlow / AppModel
  ├── ConfirmStep ──uses──> ConfirmModel
  └── LoadingStep  ──uses──> LoadingModel ──wraps──> tea.Cmd ──> domain (sync/mr/project/browser)
```

---

## Unchanged entities

- `InvocationMode` / `UseTUI` — no change; TUI-only feature
- `Sync Edge`, `Edge Result`, `Browser Batch` — domain semantics unchanged
- `SharedTUIChrome` — extended with confirm border + button focus styles + spinner placement tokens
