# Contract: `branchy sync` CLI (updated)

**Feature**: `002-interactive-sync-confirm` | **Version**: 2.0 (draft)

## Command

```text
branchy sync [--from <branch>] [-y|--yes]
```

Creates GitLab merge requests for each parent→child edge below the chosen root branch, with per-edge confirmation and optional end-of-session browser open.

**Prerequisites**: Registered project, GitLab auth (`glab auth login`), at least one edge below root (or command exits gracefully with message).

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--from` | | string | `""` | Root branch to sync from; interactive picker if omitted |
| `--yes` | `-y` | bool | `false` | Skip per-edge confirmation prompts |

### Flag semantics (v2 changes)

| Flag | Per-edge confirm | End browser prompt |
|------|------------------|-------------------|
| (default) | stdin `[y/N]` per edge | stdin `[y/N]` if any MR created |
| `-y` | skipped (all edges auto-confirmed) | stdin `[y/N]` if any MR created |

**Breaking change from v1**: Browser tabs are **no longer** opened automatically after sync. User must confirm at end prompt.

---

## Modes

### Interactive (default)

**Flow**:
1. Select `--from` branch (flag or numbered picker)
2. Print sync plan listing all edges in DFS order
3. For each edge: `Create MR <parent> → <child>? [y/N]` (skipped with `-y`)
4. Print per-edge results as they complete
5. If any MR was **created** in this session: `Open created MRs in browser? [y/N]`
6. Print summary totals

**Stdout on end browser confirm (yes)**:
```text
Opening 2 MR(s) in browser...
```

Tabs open in DFS order among created MRs only.

**Stdout on browser open failure** (non-fatal):
```text
Warning: could not open browser for <url>
```

### TUI (`branchy` → `s`)

Delegated to `SyncFlowModel` — see [data-model.md](../data-model.md) state transitions.

| Key | Per-edge step | Browser step |
|-----|---------------|--------------|
| `y` | Create MR for edge | Open created MRs in browser |
| `n` | Skip edge, continue | Skip browser open |
| `Enter` | — | Skip browser open (same as `n`) |
| `Esc` | Cancel remaining edges → partial summary | Return to tree |
| `q` | Quit app | Quit app |

**No bulk confirm**: pressing `s` goes directly to first per-edge prompt.

---

## Per-edge result output

**Created**:
```text
Created: develop → feature-x
  https://gitlab.example.com/.../merge_requests/123
```

**Skipped (user)**:
```text
Skipped: develop → feature-x (skipped by user)
```

**Skipped (existing MR)**:
```text
Skipped: develop → feature-x (open MR already exists)
  https://gitlab.example.com/.../merge_requests/99
```

**Failed**:
```text
Failed: develop → feature-x — <error message>
```

**Final totals**:
```text
Done — created: 2, skipped: 1, failed: 0
```

---

## Browser batch rules

| Rule | Behavior |
|------|----------|
| When prompted | Only if ≥1 `created` result in session |
| URLs included | `Action == "created"` only |
| URLs excluded | User-declined, existing MR skip, failed |
| Tab order | DFS edge order among included URLs |
| Open mechanism | Sequential with 50ms delay between tabs |

---

## Validation / edge cases

| Condition | Behavior |
|-----------|----------|
| No child edges | `No child branches below "<from>".` — exit 0 |
| Auth missing | `glab auth: ... (run: glab auth login)` — exit 1 |
| Branch not in tree | `branch "<name>" not in tree` — exit 1 |
| All edges declined | Summary shows all skipped; no browser prompt |
| `-y` with failures | Continues remaining edges; browser prompt only for successful creates |

---

## Internal service contract (`internal/sync`)

```go
// Run — batch sync with optional per-edge Confirm callback.
// Does NOT open browser (v2).
func Run(p *project.Project, opts Options) (*Summary, error)

// RunEdge — create MR for a single edge (no confirm, no browser).
func RunEdge(p *project.Project, edge tree.Edge) Result

// CreatedURLs — DFS-ordered URLs for created MRs only.
func CreatedURLs(summary *Summary) []string

// OpenURLs — sequential browser open; returns warning message if any fail.
func OpenURLs(urls []string) string
```

**Guarantees**:
- `Results` order matches `CollectEdges` DFS order
- `CreatedURLs` preserves that order filtered to `created`
- `OpenURLs` never fails the sync command; warnings only

---

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Sync completed (including all-declined or no edges) |
| 1 | Error (validation, auth, fatal read error on stdin) |

---

## Out of scope

- `branchy mr` / `m` shortcut — unchanged
- `--open-browser` flag — not added; end prompt is mandatory when creates exist
- Editing MR title/description during sync — unchanged auto-generated values
