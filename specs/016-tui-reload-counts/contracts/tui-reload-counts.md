# Contract: Manual Remote Count Reload

**Feature**: `016-tui-reload-counts` | **Version**: 1.0 (draft)

## Overview

Defines the user-triggered reload of file-change counts via `r` on the main tree and sync flow. Builds on [007 lazy-fetch](../../007-lazy-auto-fetch/contracts/lazy-fetch.md) (fetch + refresh pipeline) and [006 inbound-count](../../006-parity-diff-display/contracts/inbound-count.md) (display rules).

---

## Keybinding contract

| Key | Surfaces | Action |
|-----|----------|--------|
| `r` | Main tree | Schedule default-remote update + count refresh |
| `r` | Sync root picker | Same |
| `r` | Sync edge confirm | Same |
| `r` | All other TUI screens | Ignored (no binding) |

**Must**:

- Return a `tea.Cmd` immediately (non-blocking)
- Use `remoteUpdateCmd` / `FetchDefaultRemote` (same as auto-fetch)
- Recompute inbound **and** outbound counts on success

**Must not**:

- Block keyboard input
- Show loading spinner or status line
- Show fetch-failure message
- Modify working tree or check out branches

---

## Help text contract

| Surface | Required help fragment |
|---------|------------------------|
| Main tree footer | `r: reload` |
| Sync root picker footer | `r: reload` |
| Sync edge confirm | No `r` help required |

Other help keys unchanged (`s`, `m`, `l`, `u`, direction chords, etc.).

---

## Refresh contract

On `remoteUpdateMsg` for the **current** project (same rules as 007):

| `Err` | Main tree | Sync picker | Sync confirm |
|-------|-----------|-------------|--------------|
| nil | Refresh inbound+outbound badges per active direction | Rebuild inbound badges | Rewrite confirm count line |
| non-nil | No change | No change | No change |

Display rules after refresh: exactly 006 + 008 (hide zero on browse; `?` if unknown; confirm shows known zero; sync surfaces always inbound).

Wrong project id: ignore.

---

## Coalescing contract

- At most one `git fetch` per repo path at a time (007 `FetchDefaultRemote` single-flight).
- Manual `r` while auto-fetch in flight: schedule `remoteUpdateCmd`; second caller waits on shared flight — **no second parallel fetch**.
- Multiple completed `remoteUpdateMsg` deliveries may refresh counts more than once — allowed.

---

## Non-goals

- Scripted CLI reload flag
- MR / link / unlink / projects picker
- Sync summary, browser, or processing steps
- Re-read `branch-tree.yaml` from disk
- Periodic polling
