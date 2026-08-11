# Contract: TUI Confirm & Loading Chrome

**Feature**: `005-sync-tui-polish` | **Version**: 1.0 (draft)

## Overview

Extends the unified command TUI contract (`004-command-tui`) with shared **Confirm** and **Loading** components. Applies to all interactive TUI flows with yes/no steps. Scripted CLI paths are unchanged.

**Scope**: sync (embedded + standalone), MR, link, unlink, init, embedded tree unlink.

---

## Confirm Panel

### Visual structure

```text
┌─ branchy sync ─────────────────────────────────────┐  ← optional; parent flow renders title
│                                                    │
│  Sync from develop                                 │  ← context lines (0..n)
│                                                    │
│  Create MR develop → feature-a?                    │  ← question (no [y/N] suffix)
│  Edge 2 of 5                                       │  ← optional progress context
│                                                    │
│     [ No ]    [ Yes ]                              │  ← focused option highlighted
│                                                    │
│  ←/→: select  enter: confirm  y/n: shortcut  esc: decline
└────────────────────────────────────────────────────┘
```

### Behavior contract

| Input | When focus = No | When focus = Yes |
|-------|-----------------|------------------|
| `left` / `right` / `tab` | Toggle focus | Toggle focus |
| `enter` | Select No | Select Yes |
| `y` | — | Select Yes |
| `n` | Select No | — |
| `esc` | Select No | Select No |

**Defaults**: Focus starts on **No**.

**Rendering rules**:

- Focused button: bold + accent foreground (match `titleStyle` accent or inverted)
- Unfocused button: muted (`helpStyle`)
- Panel border: lipgloss rounded border, neutral color
- MUST NOT contain literal `[y/N]` in question text

### Per-flow confirm screens

| Surface | Context lines | Question |
|---------|---------------|----------|
| Sync edge | `Sync from {root}`, `Edge N of M` | `Create MR {parent} → {child}?` |
| Sync browser | MR URL list (optional, below panel) | `Open created MRs in browser?` |
| MR create | `Title: {title}` | `Create MR {source} → {target}?` |
| MR browser | URL if present | `Open in browser?` |
| Link | — | `Link {parent} → {child}?` |
| Unlink | — | `Remove "{branch}" and {n} branch(es) from tree?` |
| Init | `Repository: {path}`, `Project ID: {id}` | `Register this repository with branchy?` |
| Tree unlink | — | Same as unlink flow |

### Semantic mapping (unchanged from pre-feature)

| Choice | Sync edge | Sync browser | MR create | MR browser | Link | Unlink | Init | Tree unlink |
|--------|-----------|--------------|-----------|------------|------|--------|------|-------------|
| Yes | Create MR | Open tabs DFS | Create MR | Open tab | Save edge | Remove subtree | Register | Remove subtree |
| No | Skip edge | Done | Cancel flow | Done | Cancel | Cancel | Cancel | Cancel |
| Esc | = No (sync edge: cancel remaining) | = No | = No | = No | = No | = No | = No | = No |

---

## Loading Panel

### Visual structure

```text
  Sync from develop

  ⣾ Creating MR develop → feature-a
  Edge 2 of 5

  (input blocked)
```

### Behavior contract

| Rule | Requirement |
|------|-------------|
| Animation | Spinner ticks via `bubbles/spinner` (not static text) |
| Input | ALL `tea.KeyMsg` ignored until completion msg received |
| Auto-advance | On success, parent flow transitions without user input |
| Failure | Loading clears; parent flow applies existing error/skip rules |
| Appear latency | Spinner visible within 200ms of entering loading step |

### Loading operations

| Trigger | Message template | Progress |
|---------|------------------|----------|
| Sync edge MR | `Creating MR {parent} → {child}` | `Edge {n} of {m}` |
| Sync browser | `Opening MRs in browser…` | `{i} of {total}` optional |
| MR create | `Creating MR {source} → {target}` | — |
| MR browser | `Opening in browser…` | — |
| Init register | `Registering project…` | — |

**Out of scope**: Link/unlink save (no loading step).

---

## Shared component API (implementation contract)

Location: `internal/tui/confirm.go`, `internal/tui/loading.go`

```go
// ConfirmChoice is the outcome of a resolved confirm interaction.
type ConfirmChoice int // None, Yes, No

type ConfirmOptions struct {
    Context []string
    Question string
    Progress string // optional, e.g. "Edge 2 of 5"
    Width    int
}

func NewConfirm(opts ConfirmOptions) ConfirmModel
func (m ConfirmModel) Update(msg tea.KeyMsg) (ConfirmModel, ConfirmChoice)
func (m ConfirmModel) View() string

type LoadingOptions struct {
    Message  string
    Progress string
}

func NewLoading(opts LoadingOptions) LoadingModel
func (m LoadingModel) Init() tea.Cmd
func (m LoadingModel) Update(msg tea.Msg) (LoadingModel, tea.Cmd)
func (m LoadingModel) View() string
func (m LoadingModel) Active() bool // true while awaiting completion
```

Parent flows MUST:

1. Embed `ConfirmModel` or `LoadingModel` as struct fields
2. Delegate confirm key handling to `ConfirmModel.Update` when on confirm step
3. On `ConfirmYes`, transition to loading step (if async) or execute action via `tea.Cmd`
4. On loading step, only handle spinner ticks, resize, and typed result msgs

---

## Scripted CLI (unchanged)

| Condition | Confirm | Loading |
|-----------|---------|---------|
| Any flag set | stdin `[y/N]` text | none |
| Non-TTY | stdin / fail fast | none |
| Positional args (link/unlink) | N/A | none |

No changes to `internal/sync` plain CLI prompts.

---

## Exit codes & semantics

No change to process exit codes or domain outcomes. Presentation-only contract.
