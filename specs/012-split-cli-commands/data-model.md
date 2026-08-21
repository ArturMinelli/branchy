# Data Model: Split CLI Commands

**Feature**: `012-split-cli-commands` | **Date**: 2026-08-21

No new runtime entities or on-disk schema. This feature reorganizes **command definitions** — cobra wiring that already exists in `internal/cli/root.go`.

## Entities

### Command tree (existing, reorganized)

| Node | File (after split) | Role |
|------|-------------------|------|
| `branchy` (root) | `root.go` | Default: launch main TUI when no subcommand |
| `init` | `init.go` | Register repo via `project.Init` |
| `sync` | `sync.go` | Batch MR cascade via `sync.Run` |
| `mr` | `mr.go` | Single MR via `mr.Create` |
| `link` | `link.go` | Edge add via `project.Link` |
| `unlink` | `unlink.go` | Subtree remove via `project.Unlink` |
| `projects` | `projects.go` | List via `project.ListAll` |

**Relationships**:

```text
cmd/branchy/main.go
        └── cli.Execute()
                └── rootCmd (root.go)
                        ├── initCmd   → project.Init | tui.RunInit
                        ├── syncCmd   → sync.Run | tui.RunSync
                        ├── mrCmd     → mr.Create | tui.RunMR
                        ├── linkCmd   → project.Link | tui.RunLink
                        ├── unlinkCmd → project.Unlink | tui.RunUnlink
                        └── projectsCmd → project.ListAll | tui.RunProjects
```

---

### Command definition (per file)

| Attribute | Description |
|-----------|-------------|
| `Use` / `Short` / `Long` | Unchanged help strings |
| `Args` | Unchanged arity (link/unlink 0–2, others default) |
| `Flags` | Registered in same file’s `init()` |
| `RunE` | Thin: mode check → TUI or scripted handler |

**Validation rules** (unchanged, enforced in handlers before operations):

- Scripted link/unlink: require two args in non-TTY or when not using TUI mode
- Scripted sync: `--from` required in non-interactive mode; `-y` auto-confirms edges
- Scripted mr: `--source` and `--target` required when not TUI; `--source` required if `--target` set
- Init: `--force` passed through to `project.InitOptions`

---

### Scripted vs interactive mode (existing)

| Input | Mode | Owner |
|-------|------|-------|
| TTY + no changed flags (+ positional rules for link/unlink) | Interactive | `UseTUI` → `internal/tui` |
| Non-TTY or any flag changed | Scripted | Command file private helper or inline `RunE` |

Defined in `mode.go`; MUST NOT be duplicated per command file.

---

### Private scripted helpers (derived, not exported)

| Helper | File | Calls |
|--------|------|-------|
| `runSyncCLI` | `sync.go` | `p.Tree.Names`, `p.Tree.CollectEdges` (read-only), `sync.Run`, `sync.OpenableURLs`, `sync.OpenURLs` |
| `runMRFlags` | `mr.go` | `mr.Create` |

---

## State transitions

None at the CLI layer beyond cobra’s parse → run → exit. Operation-level state (tree persist, sync cascade) unchanged from spec 011.

---

## Unchanged entities

- `project.Project`, `sync.Options`, `mr.CreateRequest` — spec 011
- TUI session models — not touched in this spec
- On-disk branch tree and index
