# Contract: Inbound File-Change Counts

**Feature**: `006-parity-diff-display` | **Version**: 1.0 (draft)

## Overview

Defines how branchy measures “how many changes a child will get when its parent is synced into it,” and how that number appears on the three in-scope TUI surfaces. Scripted CLI is out of scope.

---

## Comparison contract

### Function

```text
InboundFiles(dir, parent, child) → (files int, err error)
```

| Argument | Meaning |
|----------|---------|
| `dir` | Git working copy (`git -C`) |
| `parent` | Tree parent; MR **source** |
| `child` | Tree child; MR **target** |

### Semantics

Equivalent to GitLab’s files-changed count on MR `parent` → `child`:

```text
git -C <dir> diff --name-only <child>...<parent>
```

- **Success**: `files` = number of non-empty output lines (≥ 0).
- **Failure**: non-nil error if either ref is missing, not a repo, or git exits non-zero. Callers MUST treat this as **unknown**, never as `0`.
- **Diverged branches**: only files on the parent side of the merge-base count. Files unique to the child do not.

### Non-goals

- No `git fetch`
- No GitLab HTTP compare
- No commit counts, no +/− line totals

### Shared numeric identity (FR-005)

For the same `dir` + parent/child names + unchanged local refs, the integer returned here is the integer shown on the tree, the sync picker, and the edge confirm.

---

## Display contract

### Badge (main tree row, sync root picker)

| Condition | Visible text | Style |
|-----------|--------------|-------|
| Root (no parent) | *(none)* | — |
| `ok` and files = 0 | *(none)* | — |
| `ok` and files = N > 0 | `N` | muted (`helpStyle`) |
| `unknown` | `?` | warn (`warnStyle`) |

Placement: same line as the branch name, after the name, compact (e.g. two spaces then the badge). Must remain readable at width 80.

Picker `FilterValue` remains the bare branch name. Other command pickers (MR, link, unlink) MUST NOT gain badges unless they pass a count map.

### Confirm line (sync edge only)

Appended to confirm context (not the question):

| Condition | Context line |
|-----------|--------------|
| `ok` | `{N} files would change on {child}` |
| `unknown` | `File count unavailable` |

`N` includes `0`. Question stays `Create MR {parent} → {child}?`.

### Out of scope surfaces

- Sync results summary
- Scripted `branchy sync` prompts
- Main tree help footer (no new keys)

---

## Surface map

| Surface | Shows inbound count | Zero | Unknown |
|---------|---------------------|------|---------|
| Main tree row | yes, non-roots | hide | `?` |
| Sync root picker | yes, non-roots | hide | `?` |
| Sync edge confirm | yes | show `0 files…` | `File count unavailable` |
| Sync summary | no | — | — |
| Plain CLI | no | — | — |

---

## Error / degradation

- A single failed comparison MUST NOT block tree navigation, starting sync, confirming other edges, or quitting.
- Failed edges show the unknown placeholder independently; successful edges still show numbers.
- The TUI MUST remain operable if every comparison fails.
