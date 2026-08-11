# Data Model: Manual MR Command

**Feature**: `001-mr-command` | **Date**: 2026-08-11

## Entities

### MRRequest (in-memory, flow state)

Represents the user's intent to create a merge request. Lives for the duration of a CLI or TUI session.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Source` | string | yes (before create) | GitLab source branch name |
| `Target` | string | yes (before create) | GitLab target branch name |
| `Title` | string | yes (before create) | MR title; default generated if empty |
| `Description` | string | auto | Timestamped auto-generated text; not user-editable in v1 |
| `ProjectID` | string | yes | Registered branchy project |
| `OpenBrowser` | bool | no | Set after user confirms browser prompt (TUI only) |

**Validation rules**:
- `Source` and `Target` MUST differ
- Both MUST exist as keys in `project.Tree.Branches`
- `Title` MUST be non-empty at creation time (default applied if user leaves default)

---

### MRResult (in-memory, outcome)

Returned after attempting MR creation.

| Field | Type | Description |
|-------|------|-------------|
| `Source` | string | Source branch |
| `Target` | string | Target branch |
| `Action` | enum | `created` \| `skipped` \| `failed` |
| `URL` | string | MR web URL (set on created/skipped-with-existing) |
| `Message` | string | Human-readable detail (skip reason, error) |

**Action semantics**:
- `created` — new MR created on GitLab
- `skipped` — open MR already exists for source→target; `URL` points to existing
- `failed` — validation, auth, or GitLab error; `Message` explains why

---

### BranchTreeBranch (existing)

From `internal/tree.Document`. No schema changes.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | Branch name (map key in `Branches`) |
| `Children` | []string | Child branch names (not used for MR picker) |

**Usage in MR flow**: `Document.Names()` provides the selectable branch list for source and target pickers.

---

### MergeRequest (external, GitLab)

Not stored locally. Identified by web URL returned from glab.

| Attribute | Source |
|-----------|--------|
| Source branch | glab `--source-branch` |
| Target branch | glab `--target-branch` |
| Title | glab `--title` |
| Description | glab `--description` |
| Web URL | glab JSON output / parsed stdout |

---

## State Transitions (TUI flow)

```text
[Start]
   │
   ├─ prefilled source? ──yes──► stepTarget
   │                    no ──► stepSource
   │
stepSource ──select branch──► stepTarget
stepTarget ──select branch──► stepTitle
stepTitle  ──edit + enter───► stepConfirm
stepConfirm ──y────────────► [Create via internal/mr]
            ──n/esc────────► [Cancel → previous screen / exit]
[Create]
   ├─ failed ──────────────► stepError (show message)
   └─ created/skipped ─────► stepResult (show URL)
stepResult ────────────────► stepBrowserPrompt
stepBrowserPrompt ──y──────► [browser.Open(URL)]
                    ──n────► [Done]
```

**Cancel transitions**: `Esc` from any step before `stepResult` returns to previous context (tree view if embedded, exit if standalone).

---

## Relationships

```text
Project 1──1 Document (branch tree)
Project 1──* MRRequest (ephemeral, per session)
MRRequest ──creates──► MRResult ──references──► MergeRequest (GitLab)
MRRequest.Source/Target ──must exist in──► Document.Branches
```

---

## Default title generation

```text
MR: {source} → {target}
```

Applied when user reaches `stepTitle` or when `--title` flag is omitted in CLI mode.

## Default description generation

```text
Manual merge request created by branchy on {RFC3339-ish timestamp}.
```

Mirrors sync's timestamped description pattern.
