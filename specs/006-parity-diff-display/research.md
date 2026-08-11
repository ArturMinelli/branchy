# Research: Parent Diff Display

**Feature**: `006-parity-diff-display` | **Date**: 2026-08-11

## 1. GitLab-comparable file-change metric

**Decision**: Count files with a local three-dot diff from child to parent:

```text
git -C <repo> diff --name-only <child>...<parent>
```

The inbound count is the number of non-empty lines. This is the files-changed number GitLab shows on an MR whose **source** is parent and **target** is child (changes on parent since the merge-base with child).

**Rationale**: Spec FR-004 / SC-003 require the GitLab MR files-changed number, not commit count and not additions/deletions. Spec FR-009 forbids requiring a fetch or GitLab API. Three-dot (`A...B`) is merge-base..B — the same comparison GitLab’s MR “Changes” tab uses. Child-only commits after divergence do not inflate the count.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Two-dot `git diff child parent` | Includes the reverse side; not what GitLab MRs show |
| `git rev-list --count child..parent` | Commit count; user chose files changed |
| GitLab compare / MR API | Network + auth; contradicts local-only FR-009; fails when no MR exists yet |
| Additions/deletions (`--shortstat` +/−) | User locked files-changed only |
| `git diff --stat` parse | Fragile text; `--name-only` line count is exact |

**Notes**: Default Git rename detection (on since Git 2.9) matches GitLab’s usual rename handling — a rename is one file. Binary files still appear in `--name-only` and are counted, as on GitLab.

---

## 2. Comparison API placement

**Decision**: Add `InboundFiles(dir, parent, child string) (int, error)` on the existing `internal/git` package. TUI never shells out to git directly.

**Rationale**: `internal/git` is already the `os/exec` git seam (`Root`). A single function hides three-dot argument order so callers cannot invert parent/child. Independently testable with temp repos. No new package (`internal/diff`) — YAGNI.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Compute inside `treeview.go` | Couples presentation to git; untestable without a repo; duplicated in syncflow |
| `internal/tree` owns counts | Tree is YAML hierarchy, not git state |
| go-git library | New dependency; existing code uses `os/exec` |

---

## 3. When counts are computed

**Decision**: Compute once when a project tree is shown (`selectProject`) and once when a sync flow starts (picker / first confirm). Cache in memory on `BranchTreeView` and `SyncFlowModel`. Do not recompute in `View()`.

**Rationale**: Typical branchy trees are small (tens of edges). Sequential local diffs are milliseconds each. Per-frame git would freeze the TUI. Sync creates MRs and does not update local refs, so session cache stays valid. Spec does not require live refresh after fetch.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Async `tea.Cmd` + spinner | Overkill for local git; 005 reserved loading for network |
| Recompute every View | Performance; violates “no stall on paint” |
| Refresh after sync | Local refs unchanged; YAGNI |

---

## 4. Display format

**Decision**:

- **Tree / picker**: compact muted suffix `N` after the branch name. Hide when count is 0. Unknown: `?` in warn style. Roots: no suffix.
- **Edge confirm**: extra context line `{N} files would change on {child}` (including `0`). Unknown: `File count unavailable`.

Keep picker items single-line by appending the badge to `branchItem.Title()`, not `Description()` (default list delegate would become two lines and change link/MR/unlink pickers).

**Rationale**: Spec assumes compact badges. Hide-zero on browse surfaces was grilled. Confirm always shows a known zero so empty MRs are not a surprise. `?` is recognizable in under 2 seconds (SC-004) and is not a fabricated zero.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Two-line list Description | Changes height of every `newBranchList` picker |
| Always show `0` on tree | Rejected in grilling |
| `synced` label | Rejected in grilling |
| Ahead/behind pair `↓3 ↑2` | Rejected in grilling |

---

## 5. Shared load + format helpers

**Decision**: `internal/tui/inbound.go` walks `doc` via `ParentOf`, calls `git.InboundFiles` per child, and exposes:

- `loadInboundCounts(dir, doc) map[string]inboundCount`
- `formatInboundBadge(count, present) string` — empty / `N` / `?`
- `formatInboundConfirm(child, count, ok) string` — confirm sentence

Tree, picker, and confirm all consume the same map so FR-005 (identical N) holds.

**Rationale**: Display policy in one place prevents tree vs picker drift. Git package stays free of `tree.Document`.

---

## 6. Testing strategy

**Decision**:

- **`internal/git`**: real temp repos — N files only on parent → N; identical trees → 0; missing ref → error; files only on child after diverge → 0 inbound.
- **`internal/tui`**: inject a prebuilt count map into the tree/picker/confirm (no git). Assert hide-zero, root omission, `?`, and confirm-zero text.

**Rationale**: Semantically important cases need a real git merge-base. Rendering rules should not require fixtures. Matches 004/005 pattern (domain tests + view-string tests).

---

## 7. Scripted CLI unchanged

**Decision**: No changes to `internal/cli` or `internal/sync` stdin prompts. Counts are TUI-only (FR-010).

**Rationale**: Grilling and spec bound the feature to main tree + sync TUI.

---

## 8. Shared `branchItem` compatibility

**Decision**: Add an optional badge string to `branchItem`. `newBranchList(names, title)` keeps empty badges. Sync picker uses a variant (or optional counts argument) that sets badges from the inbound map. Link, unlink, and MR pickers stay visually identical.

**Rationale**: `branchItem` / `newBranchList` live in `mrflow.go` and are reused. Changing `Description()` would regress those pickers.
