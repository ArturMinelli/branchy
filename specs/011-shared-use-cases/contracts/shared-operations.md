# Contract: Shared Operations

**Feature**: `011-shared-use-cases` | **Version**: 1.0 (draft)

## Overview

User-facing link, unlink, sync processing, merge-request create, and init have one owner each. Surfaces collect input and render; they do not copy validate-then-persist or validate-then-create. Behavior matches the previous revision except interactive unlink, which MUST apply the scripted edge check.

---

## Link

**Owner**: `project.Link(parent, child) error`

**Invariants**:

1. Empty parent or child → `parent and child are required`
2. `parent == child` → `parent and child must differ`
3. Duplicate edge → `edge already exists: P → C` (creates missing nodes otherwise, same as `tree.Link`)
4. On success the edge is in `p.Tree` and in `branch-tree.yaml` for `p.ID`
5. Every user-facing path MUST call this: scripted `branchy link P C`, standalone link confirm, main-tree Enter on the link screen

**Out of owner**: pickers, flags, success printf / TUI success step.

**Forbidden**: `Tree.Link` + `SaveTree` in `internal/cli` or `internal/tui`.

---

## Unlink

**Owner**: `project.Unlink(parent, child) (*UnlinkResult, error)`

**Invariants**:

| Input | Result |
|-------|--------|
| empty child | error, tree unchanged on disk |
| non-empty parent, `parent == child` | `parent and child must differ` |
| child not in tree | `branch %q not in tree` |
| parent set, `!HasEdge(parent, child)` | `edge not found: P → C`, tree unchanged |
| parent empty, child is a root | success; remove that subtree |
| parent empty, child has a parent | error (not a root path) |
| parent set, edge exists | `UnlinkSubtree(child)` + save; `Removed` is pre-delete subtree size |

**Callers**:

- Scripted `branchy unlink P C` — pass the two args (no extra validation block in the command)
- Standalone unlink and main-tree unlink — `parent, _ := Tree.ParentOf(target)`; then `Unlink(parent, target)`

**Success copy** (unchanged wording):

- CLI: `Unlinked %s → %s (%d branches removed)` using the result
- TUI: existing “Removed … and N branches” using `Removed` / target name

**Forbidden**: `UnlinkSubtree` + `SaveTree` in CLI or TUI.

---

## Init

**Owner**: `project.Init(InitOptions) (*Project, error)` (existing)

**Invariants** (lock, do not fork):

1. Detect git root, import or scaffold tree, write store, update index
2. Already registered without `Force` → `already registered as %q (use --force to re-import)`
3. `Force` reuses the existing id
4. Interactive init calls this with `Force: false` after confirm; cancel MUST NOT call it

---

## Merge-request create

**Owner**: `mr.Create(p, CreateRequest) (*CreateResult, error)` (existing)

**Invariants**:

1. Validates branches (error, no GitLab) then `AuthOK` then find-or-create
2. Already-open → skipped + URL, not a Go error
3. Unrecovered create failure → `ActionFailed` + message, `err == nil` (existing)
4. Manual CLI/TUI and `sync.processEdge` MUST both call this for the create

**Forbidden**: `gitlab.Client` construction in `internal/tui`.

---

## Sync

### Begin (interactive start)

**Owner**: `sync.Begin(p, from, dir) ([]tree.Edge, error)`

| Condition | Contract |
|-----------|----------|
| empty / unknown `from` | same error as `sync.Run` |
| zero edges | `nil, nil` or empty slice + nil; **no** GitLab call |
| one or more edges, auth fail | error `glab auth: … (run: glab auth login)`; no edge processing |
| one or more edges, auth ok | edges in `EdgesBelow` order |

TUI `prepareSyncFrom` MUST call Begin instead of `gitlab.Client{}.AuthOK`.

### Run (scripted batch)

**Owner**: `sync.Run` (existing)

Unchanged: auth first (even if the subtree is empty), then walk with `Confirm`, then per-edge `processEdge` → `mr.Create`. `-y` still means Confirm always true.

### RunEdge / SkippedByUser (interactive step)

| Call | When | Result |
|------|------|--------|
| `RunEdge(p, edge, dir)` | user Yes | same as `processEdge` without Confirm |
| `SkippedByUser(edge, dir)` | user No | `Action=skipped`, `Message=skipped by user`, ends from `Ends(edge, dir)`, no GitLab |

TUI MUST NOT assemble that skip struct itself.

Cancel-remaining (esc) is **not** skip-by-user for unprocessed edges; keep today’s jump to summary or quit.

### Browser

Still a surface step using `sync.OpenableURLs` / `OpenURLs`. The shared processing functions MUST NOT open tabs. Relocating OpenURLs out of `sync` is spec 014.

---

## Surface rules

| Surface | May do | Must not do |
|---------|--------|-------------|
| `internal/cli` | parse, prompt, call owner, print | tree mutate, `SaveTree` for link/unlink, GitLab client, extra unlink edge walk |
| `internal/tui` | pickers, confirms, display `err.Error()` | `gitlab` import, `Tree.Link`/`UnlinkSubtree`+save, hand-built skip results |
| Main tree | same owners; keep inline screens | a second persist implementation |

---

## Compatibility

- Flags, keys, help, cascade order, skip-if-open, browser prompt timing: unchanged
- Interactive unlink of a **non-edge** pair now fails (intended unification)
- Interactive unlink of a **root** still succeeds
- Product usable without specs 012–015
