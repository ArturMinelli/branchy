# Data Model: Unify Domain Types

**Feature**: `015-unify-domain-types` | **Date**: 2026-08-24

No on-disk schema change. This feature unifies **in-memory vocabularies** for merge direction and create/sync outcomes.

## Entities

### Direction (existing, now the only type)

Owner: `internal/sync`.

| Value | Iota | Walk / MR ends | TUI counts | TUI arrow |
|-------|------|----------------|------------|-----------|
| `Downward` | 0 | parent → child, `CollectEdges` | inbound | `↓` |
| `Upward` | 1 | child → parent, `CollectEdgesUpward` | outbound | `↑` |

**Session scope** (unchanged):

- Main tree: one `sync.Direction` on the app model, copied into `BranchTreeView`
- Standalone sync: run-scoped on `SyncFlowModel` (Control+Up / Control+Down)
- Scripted sync: omitted → zero value `Downward`

**Deleted**: TUI `diffDirection`, `diffInbound`, `diffOutbound`, `syncDirection()`.

**Invariant**: Interactive code MUST NOT define a second direction enum or convert at the `sync.Begin` / `Run` / `RunEdge` boundary.

---

### Action (typed)

Owner: `internal/mr`.

```text
created | skipped | failed
```

Used by:

- `mr.CreateResult.Action`
- `sync.Result.Action`

Printers (CLI sync/mr, TUI sync summary, TUI mr result) MUST `switch` on `mr.ActionCreated` / `ActionSkipped` / `ActionFailed`. MUST NOT switch on untyped `"created"` literals.

User-visible interpolation of `Action` still prints those three words.

---

### SkipReason (new, in-memory)

Owner: `internal/mr`. Contributor-facing; not printed.

| Value | When | Openable? |
|-------|------|-----------|
| `SkipNone` (`""`) | created, failed, or skipped without a subtype | created with URL: yes; failed without URL: no |
| `SkipAlreadyOpen` | skip because an open MR exists | yes, if URL present |
| `SkipUserDeclined` | sync confirm declined | no |

`Message` stays the human sentence (`"open MR already exists"`, `"skipped by user"`). `OpenableURLs` MUST use `SkipReason`, not `Message`.

**Relationships**:

```text
mr.Create  -->  CreateResult { Action, SkipReason, URL, Message }
                    │
                    ▼  (processEdge copy)
sync.Result { Action, SkipReason, URL, Message, Parent, Child, Source, Target }
                    │
                    ▼
            OpenableURLs(summary)  -->  []string
```

---

### Create error vs result (policy)

No new entity. The pair `(result, error)` is the contract:

```text
validation / auth     -->  (nil, error)          abort the call
already-open          -->  (skipped result, nil)
unrecovered failure   -->  (failed result, nil)
created               -->  (created result, nil)
```

Scripted `mr` is a **surface** mapping: failed result → non-zero exit. That mapping is not an operation-level error.

Sync cascade: skipped and failed results continue; unknown **from-branch** still aborts in `Run`/`Begin` before edges.

---

## Validation rules

- After the change, `rg 'type diffDirection' internal/tui` is empty.
- `rg 'syncDirection' internal/` is empty.
- `OpenableURLs` tests MUST exclude user-declined via `SkipReason`, including a fixture that has a URL **and** `SkipUserDeclined` (proves Message is not the filter).
- `mr.Create` on unknown/same branch returns error and does not construct a skipped/failed result.
- CLI `mr` on `ActionFailed` still returns an error (non-zero); on `ActionSkipped` does not.

---

## Unchanged entities

- On-disk YAML / `tree.Document`
- `WalkDisplay` / `DisplayNode` (014)
- `browser.Open` / `OpenURLs` (014)
- `git.InboundFiles` / `OutboundFiles` (names and signatures)
- Prompt copy, keys, arrows, badge formatting
