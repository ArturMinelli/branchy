# Contract: Domain / UI Split

**Feature**: `014-domain-ui-split` | **Version**: 1.0 (draft)

## Overview

The stored tree exposes an unstyled display walk. Drawing stays in TUI. Sync exposes openable URLs and does not open tabs. Surfaces ask, then call `browser.OpenURLs`. Interactive code does not construct GitLab sessions.

---

## 1. Display walk

```go
// internal/tree
type DisplayNode struct {
    Name        string
    Depth       int
    IsRoot      bool
    IsLast      bool
    LastAtDepth []bool
}

func (d *Document) WalkDisplay() []DisplayNode
```

| MUST | MUST NOT |
|------|----------|
| DFS pre-order from `Roots()` | Contain `└──`, `├──`, `│`, or lipgloss |
| Sort children with `sort.Strings` | Live in `internal/tui` as a second parent/child walk |
| Be the source of `flattenTree` row order | Replace `CollectEdges` / `CollectEdgesUpward` |

`internal/tui/treeview.go` maps `[]DisplayNode` → `[]branchRow` (glyphs only at this layer).

---

## 2. Browser open after sync

| Symbol | Package | Role |
|--------|---------|------|
| `OpenableURLs(*Summary) []string` | `sync` | Which URLs may open |
| `OpenURLs([]string) string` | `browser` | Sequential open + warnings |
| Ask + print/confirm | `cli/sync.go`, `tui/syncflow.go` | Surface only |

Call sites after the move:

- `internal/cli/sync.go` — on yes: `browser.OpenURLs(urls)` (prompt copy unchanged)
- `internal/tui/syncflow.go` — on accept: `browser.OpenURLs(urls)` (loading copy unchanged)
- `internal/tui/mrflow.go` — still `browser.Open` for one URL

`internal/sync` MUST NOT import `internal/browser`.

Pause: 200ms between URLs. Warning format unchanged: `could not open %s: %v` joined by `"; "`. Failures are non-fatal.

---

## 3. GitLab (verification)

Interactive flows MUST NOT construct a GitLab session. Auth and create remain in `sync.Begin` / `mr.Create`.

File-change badges MAY call `internal/git`; they MUST NOT call GitLab.

---

## Out of scope

- Unifying direction / action types (spec 015)
- New packages or ports/adapters
- Changing prompt wording, keys, or badge formatting
- Teaching MR to use `OpenURLs`

---

## Grep audit (merge gate)

```text
rg 'RenderASCII|lipgloss' internal/tree
```

**Expected**: no matches (`render.go` deleted).

```text
rg 'func OpenURLs' internal/sync
```

**Expected**: no matches.

```text
rg 'sync\.OpenURLs' internal/
```

**Expected**: no matches.

```text
rg 'browser\.OpenURLs' internal/cli/sync.go internal/tui/syncflow.go
```

**Expected**: both files match.

```text
rg 'gitlab' internal/tui
```

**Expected**: no matches (US3).

```text
rg 'WalkDisplay' internal/tui/treeview.go
```

**Expected**: `flattenTree` (or successor) uses it; no parallel `Roots` + children recursion in that file.
