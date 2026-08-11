# Contract: `branchy mr` CLI

**Feature**: `001-mr-command` | **Version**: 1.0 (draft)

## Command

```text
branchy mr [flags]
```

Creates a GitLab merge request between two branches in the current project's branch tree.

**Prerequisites**: Registered project (`branchy init`), GitLab auth (`glab auth login`), at least two branches in branch tree.

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--source` | | string | `""` | Source branch name (tree branch) |
| `--target` | | string | `""` | Target branch name (tree branch) |
| `--title` | | string | `""` | MR title; auto-generated if omitted |
| `--yes` | `-y` | bool | `false` | Skip confirmation prompt in flag mode |

### Flag mode rules

| `--source` | `--target` | Behavior |
|------------|------------|----------|
| set | set | Non-TUI: validate → confirm (unless `--yes`) → create → print result |
| set | unset | TUI opens; source pre-filled |
| unset | set | Error: `--source` required when `--target` provided |
| unset | unset | Full TUI flow from source selection |

---

## Modes

### Interactive TUI (default)

Triggered when `--source` or `--target` is missing.

**Steps**:
1. Source branch picker (sorted flat list from branch tree)
2. Target branch picker (excludes selected source)
3. Title editor (default: `MR: <source> → <target>`)
4. Confirm: `Create MR <source> → <target>? [y/N]`
5. Result: display URL + action (`created` / `skipped`)
6. Browser prompt: `Open in browser? [y/N]`

**Keybindings** (consistent with main TUI):

| Key | Action |
|-----|--------|
| `↑` / `k`, `↓` / `j` | Navigate list |
| `Enter` | Select / confirm |
| `Esc` / `b` | Back / cancel |
| `y` | Yes (confirm / open browser) |
| `n` | No |
| `q` / `Ctrl+C` | Quit |

### Flag mode (non-TUI)

```bash
branchy mr --source feature-x --target develop --yes
branchy mr --source feature-x --target develop --title "My MR"
branchy mr --source feature-x --target develop   # prompts [y/N] on stdin
```

**Stdout on success (created)**:
```text
Created: feature-x → develop
  https://gitlab.example.com/group/repo/-/merge_requests/123
```

**Stdout on skipped (existing MR)**:
```text
Skipped: feature-x → develop (open MR already exists)
  https://gitlab.example.com/group/repo/-/merge_requests/99
```

**Stdout on failure**:
```text
Error: <human-readable message>
```
(exit code non-zero)

**Flag mode does NOT** prompt for browser open or auto-open browser (v1).

---

## Main TUI integration

From `branchy` tree view:

| Key | Action |
|-----|--------|
| `m` | Enter MR flow |

- If a branch is selected in tree view → source pre-filled, start at target picker.
- If no branch selected → start at source picker.
- `Esc` from MR flow → return to tree view (no MR created).

Help line on tree view updated to include `m: mr`.

---

## Validation errors (exit code 1)

| Condition | Message pattern |
|-----------|-----------------|
| Not in registered project | `no branchy project registered for ...` |
| GitLab auth missing | `glab auth: ... (run: glab auth login)` |
| Branch not in tree | `branch "<name>" not in tree` |
| Source equals target | `source and target must differ` |
| Fewer than 2 tree branches (TUI) | `need at least 2 branches in tree to create an MR` |
| MR creation failed | glab error message surfaced |

---

## Internal service contract (`internal/mr`)

```go
type CreateRequest struct {
    Source      string
    Target      string
    Title       string
    Description string // optional; auto-generated if empty
}

type CreateResult struct {
    Source  string
    Target  string
    Action  string // "created", "skipped", "failed"
    URL     string
    Message string
}

func Create(p *project.Project, req CreateRequest) (*CreateResult, error)
```

**Guarantees**:
- Validates tree membership and source ≠ target before calling glab
- Does NOT open browser (caller's responsibility)
- Returns `Action=skipped` with existing URL when open MR found
- Auto-generates title/description when empty

---

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success (created or skipped) |
| 1 | Error (validation, auth, GitLab failure, user cancel in flag mode) |
