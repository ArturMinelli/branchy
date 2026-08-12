# Contract: Diff Direction Toggle

**Feature**: `008-diff-direction-toggle` | **Version**: 1.0 (draft)

## Overview

Defines outbound file-change comparison and how the main tree toggles between inbound and outbound badges. Complements [006 inbound-count](../../006-parity-diff-display/contracts/inbound-count.md): inbound semantics and sync display are unchanged; this contract adds the reverse metric and tree-only direction mode.

---

## Comparison contract

### Function

```text
OutboundFiles(dir, parent, child) → (files int, err error)
```

| Argument | Meaning |
|----------|---------|
| `dir` | Git working copy (`git -C`) |
| `parent` | Tree parent; MR **target** for outbound |
| `child` | Tree child; MR **source** for outbound |

### Semantics

Equivalent to GitLab’s files-changed count on MR `child` → `parent`:

```text
git -C <dir> diff --name-only <parent>...<child>
```

- **Success**: `files` = number of non-empty output lines (≥ 0).
- **Failure**: non-nil error if either ref is missing, not a repo, or git exits non-zero. Callers MUST treat this as **unknown**, never as `0`.
- **Diverged branches**: only files on the child side of the merge-base count. Files unique to the parent do not.
- **Symmetry**: for the same resolved refs, `OutboundFiles(dir, parent, child)` MUST equal `InboundFiles(dir, child, parent)`.

### Non-goals

- No change to `InboundFiles` semantics
- No two-dot combined diff
- No commit counts, no +/− totals
- No GitLab HTTP compare

### Shared numeric identity

For the same `dir` + parent/child names + unchanged refs:

| Surface | Number |
|---------|--------|
| Main tree inbound mode | `InboundFiles` |
| Main tree outbound mode | `OutboundFiles` |
| Sync picker / edge confirm | `InboundFiles` only |

---

## Direction mode contract

### States

| State | Tree badges | Default |
|-------|-------------|---------|
| `inbound` | inbound map | yes (session start) |
| `outbound` | outbound map | no |

### Scope

| In scope | Out of scope |
|----------|--------------|
| Main tree screen | Sync picker, sync confirm, sync summary |
| Session (root Model) | Disk / project config / CLI flags |
| Survives project switch & return from flows | Survives process quit |

### Key binding

| Key | Screen | Action |
|-----|--------|--------|
| `d` | Main tree only | Flip `inbound` ↔ `outbound`; refresh badges from cache |

Other screens MUST ignore `d` for this purpose (no direction flip).

---

## Display contract

### Badge (main tree)

Same rules as 006 inbound badges, applied to the **active** map:

| Condition | Visible text | Style |
|-----------|--------------|-------|
| Root | *(none)* | — |
| `ok` and files = 0 | *(none)* | — |
| `ok` and files = N > 0 | `N` | muted (`helpStyle`) |
| `unknown` | `?` | warn (`warnStyle`) |

No directional glyph on the badge itself.

### Help footer (main tree)

MUST include:

1. The toggle key with the action toward the **other** mode (`d: show outbound` or `d: show inbound`)
2. A clear cue for the **current** mode (`counts: inbound (parent→child)` or `counts: outbound (child→parent)`)

Exact layout may be one or two muted lines; both pieces MUST be visible without scrolling off a normal 80×24 tree view.

### Sync surfaces

Unchanged from 006:

| Surface | Direction awareness |
|---------|---------------------|
| Sync root picker | Always inbound; no `d` hint |
| Sync edge confirm | Always inbound; question stays `Create MR {parent} → {child}?` |

---

## Refresh contract

When tree counts are recomputed (project select or successful background remote update):

1. Reload **both** inbound and outbound maps
2. Keep the current `CountDirection`
3. Paint badges from the active map

Toggle MUST NOT run git; it only flips which cached map is shown.

---

## Error / degradation

- Same as 006: a failed comparison shows `?` for that child in that direction’s map independently
- Failure to compute outbound MUST NOT block inbound badges or navigation
- Toggle remains available even if every outbound comparison is unknown
