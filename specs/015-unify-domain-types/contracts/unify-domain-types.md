# Contract: Unify Domain Types

**Feature**: `015-unify-domain-types` | **Version**: 1.0 (draft)

## Overview

CLI and interactive UI share one merge-direction type and one create/sync outcome vocabulary. Surfaces classify results from that vocabulary. Merge-request create has one error policy.

---

## 1. Direction

```go
// internal/sync — existing names, now consumed by TUI as well
type Direction int

const (
    Downward Direction = iota // parent → child
    Upward                    // child → parent
)

func Ends(edge tree.Edge, dir Direction) (source, target string)
func EdgesBelow(doc *tree.Document, root string, dir Direction) []tree.Edge
```

| MUST | MUST NOT |
|------|----------|
| Main tree session field is `sync.Direction` | `type diffDirection` in `internal/tui` |
| Treeview, help footer, arrows, and sync picker use that type | `syncDirection(` conversion helper |
| Scripted sync omit Direction (zero = Downward) | New `--direction` flag |
| Help strings stay inbound/outbound and downward/upward as today | Rename user-visible keys or arrows |

Call sites after unify:

- `internal/tui/app.go` — store `sync.Direction`; pass it to `SyncFlowOptions.Direction` with no remap
- `internal/tui/treeview.go` / `inbound.go` — `setDirection(sync.Direction)`, `treeHelpFooter(sync.Direction)`
- `internal/tui/syncflow.go` — already `sync.Direction`; delete `syncDirection`

---

## 2. Action and skip reason

```go
// internal/mr
type Action string

const (
    ActionCreated Action = "created"
    ActionSkipped Action = "skipped"
    ActionFailed  Action = "failed"
)

type SkipReason string

const (
    SkipNone         SkipReason = ""
    SkipAlreadyOpen  SkipReason = "already_open"
    SkipUserDeclined SkipReason = "user_declined"
)
```

`mr.CreateResult` and `sync.Result` both carry `Action` and `SkipReason`.

`sync.OpenableURLs`:

1. Ignore entries with empty URL
2. Include `ActionCreated`
3. Include `ActionSkipped` when `SkipReason != SkipUserDeclined`
4. MUST NOT parse `Message`

Printers MUST switch on `mr.Action*` constants (`internal/cli/sync.go`, `internal/cli/mr.go`, `internal/tui/syncflow.go`, `internal/tui/mrflow.go`).

---

## 3. Create error policy

`mr.Create(p, req) (*CreateResult, error)`:

| Input / GitLab outcome | Return |
|------------------------|--------|
| Invalid source/target (empty, equal, not in tree) | `nil, err` — no GitLab |
| Auth failure | `nil, err` |
| Open MR exists or recovered after create race | `result{ActionSkipped, SkipAlreadyOpen, URL}, nil` |
| Create fails and no URL recovered | `result{ActionFailed, Message}, nil` |
| Create succeeds | `result{ActionCreated, URL}, nil` |

Surface mapping (unchanged outcomes):

- Scripted `mr`: `err != nil` → return err; `ActionFailed` → `fmt.Errorf` (non-zero); skipped/created → exit 0
- TUI `mr`: show result; auth/validation error is the operation error string
- Sync: skipped/failed results continue the cascade; unknown from-branch still aborts in `Run`/`Begin`

---

## Out of scope

- New packages, GitLab interfaces, or shared printers
- Moving `OpenURLs` / `WalkDisplay` (014)
- Renaming `git.InboundFiles` / `OutboundFiles`
- User-visible copy, keys, arrows, or exit-code changes

---

## Grep audit (merge gate)

```text
rg 'type diffDirection' internal/tui
```

**Expected**: no matches.

```text
rg 'syncDirection' internal/
```

**Expected**: no matches.

```text
rg 'diffInbound|diffOutbound' internal/
```

**Expected**: no matches.

```text
rg 'case "created"|case "skipped"|case "failed"' internal/cli internal/tui
```

**Expected**: no matches (use `mr.Action*` / typed `Action`).

```text
rg 'skipped by user' internal/sync/sync.go
```

**Expected**: assignment to `Message` (and/or tests) only — not the `OpenableURLs` condition.

```text
rg 'r\.Message == "skipped by user"' internal/
```

**Expected**: no matches.
