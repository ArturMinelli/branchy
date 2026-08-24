# Quickstart: Embed TUI Flows

**Feature**: `013-embed-tui-flows` | **Date**: 2026-08-24

Validation guide. Contract: [contracts/embedded-flows.md](./contracts/embedded-flows.md). Entities: [data-model.md](./data-model.md).

## Prerequisites

- Specs 011 and 012 merged (shared operations; thin CLI files)
- Go 1.26+
- Registered test project (temp HOME fixture from 011 quickstart is enough for persist checks)

## Setup

```bash
go test ./internal/tui/... -count=1
go build -o branchy ./cmd/branchy
```

---

## Scenario 1: Main-tree link is the dedicated wizard

On a TTY, from a registered repo with at least one branch selected:

```bash
./branchy
# press l
```

**Expected**:

- Parent picker is **skipped**; screen is the child-name step with the selected branch as parent (not the old two-field editor)
- Complete child + confirm Yes → success step `Linked parent → child`, then enter → tree shows the new child
- Same pair via `./branchy link` (standalone) produces the same on-disk tree (SC-002)

---

## Scenario 2: Cancel link leaves the tree unchanged

From the main tree, `l`, type a child, `esc` on confirm (or No).

**Expected**: back on the tree; YAML unchanged (SC-003). Esc on the child step opens the parent picker; esc on the picker returns to the tree without persist.

---

## Scenario 3: Main-tree unlink is the dedicated confirm

Select a branch with a subtree, press `u`.

**Expected**:

- No branch picker; dedicated unlink confirm (subtree size, yes/no) — not a host-only panel
- Yes → success step, then enter → those names gone from the tree and from disk
- No / esc → tree unchanged
- Same members removed as `./branchy unlink parent child` for that edge (SC-002)

---

## Scenario 4: Standalone still exits to the shell

```bash
./branchy link    # complete or q
./branchy unlink  # complete or q
```

**Expected**: process returns to the prompt; main tree does not open (FR-006).

---

## Scenario 5: Sync, MR, picker, direction unchanged

From the main tree: `s`, `m`, ctrl+up / ctrl+down, project picker when multiple projects.

**Expected**: existing embedded sync/MR and navigation still work (SC-004). `go test ./internal/tui/...` includes `embedded_sync_test.go`.

---

## Scenario 6: No inline copies remain

```bash
rg 'linkParent|linkChild|linkInput|unlinkTarget|unlinkConfirm' internal/tui/app.go
rg 'func \(m Model\) updateLink|func \(m Model\) updateUnlink' internal/tui/
rg 'project\.Link|project\.Unlink' internal/tui/app.go
```

**Expected**: first two greps — no matches (SC-001). Third — no matches; persist only in `linkflow.go` / `unlinkflow.go`.

---

## Done when

- [ ] Scenarios 1–6 pass
- [ ] `go test ./... -count=1` green
- [ ] Standalone link/unlink tests still pass
- [ ] `app.go` still a single file; sync/MR flow files not split
