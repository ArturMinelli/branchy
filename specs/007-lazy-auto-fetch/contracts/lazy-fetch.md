# Contract: Lazy Default-Remote Update

**Feature**: `007-lazy-auto-fetch` | **Version**: 1.0 (draft)

## Overview

Defines when branchy updates remote-tracking refs and how the TUI may react. Complements [006 inbound-count](../../006-parity-diff-display/contracts/inbound-count.md): this contract refreshes the data; 006 still owns how numbers are shown.

---

## Fetch contract

### Functions

```text
DefaultRemote(dir) → (name string, ok bool)
FetchDefaultRemote(dir) → error
```

| Result | Meaning |
|--------|---------|
| `DefaultRemote` `ok == false` | No remotes; caller should not treat this as a user error |
| `FetchDefaultRemote` `nil` | Fetch succeeded, or there was nothing to fetch (no remote) |
| `FetchDefaultRemote` `err` | `git fetch` failed (network, auth, invalid URL) |

### Semantics

- Remote name: `origin` if listed by `git remote`, else the first name, else skip.
- Command: update that remote’s tracking refs only (normal fetch of one remote). Does not change the working tree or current branch.
- Concurrency: callers for the same absolute `dir` share one in-flight fetch and observe the same error.
- No prune requirement. No GitLab API.

### Non-goals

- Scripted CLI (`UseTUI == false`, any flag, non-TTY)
- MR / link / unlink / init / projects flows
- Periodic polling

---

## TUI scheduling contract

| Surface | First paint | Then |
|---------|-------------|------|
| Main tree (project selected) | Local inbound counts | Background `FetchDefaultRemote` |
| Sync picker / edge confirm | Local inbound counts | Same fetch, or join in-flight for that path |

**Must not**:

- Wait for fetch before the first View of tree or sync
- Block keyboard input while fetch runs
- Show a loading panel or fetch-failure message

---

## Refresh contract

On `RemoteUpdate` for the **current** project:

| `Err` | Tree | Sync picker | Sync confirm |
|-------|------|-------------|--------------|
| nil | Recompute inbound; set badges | Rebuild badges | Rewrite “files would change” line |
| non-nil | No change | No change | No change |

Display rules after refresh are exactly 006 (hide zero on browse; show known zero on confirm; `?` if still unresolvable).

On `RemoteUpdate` for a **different** project id: ignore (no overwrite).

---

## Error / degradation

- Offline / auth / timeout: local snapshot remains; no status line.
- No remotes: same as a quiet skip.
- Quit or switch project mid-fetch: do not block; discard the result if it no longer matches.
