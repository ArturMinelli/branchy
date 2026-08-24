# Contract: Embedded Link / Unlink Flows

**Feature**: `013-embed-tui-flows` | **Version**: 1.0 (draft)

## Overview

Main-tree `l` and `u` MUST run `LinkFlowModel` / `UnlinkFlowModel` inside the main session. Standalone `branchy link` / `unlink` MUST keep using `RunLink` / `RunUnlink` and exit to the shell. Persist stays `project.Link` / `project.Unlink`.

---

## Construction

| Surface | Constructor | Options |
|---------|-------------|---------|
| `branchy link` (TUI) | `RunLink(p)` | `Embedded: false`, no prefill |
| Main tree `l` | `newLinkFlowModel(p, LinkFlowOptions{Embedded: true, PrefillParent: selected})` | empty selection → no prefill |
| `branchy unlink` (TUI) | `RunUnlink(p)` | `Embedded: false`, no prefill |
| Main tree `u` | `newUnlinkFlowModel(p, UnlinkFlowOptions{Embedded: true, PrefillTarget: selected})` | empty selection → **do not launch** |

`RunLink` / `RunUnlink` MUST set `Embedded: false` even if a caller passed true (same as `RunSync`).

---

## Host forwarding (`app.go`)

While `screenLink` or `screenUnlink`:

1. Forward every `tea.Msg` to the child `Update` (including resize), same as `screenSync` / `screenMR`.
2. `View()` is the child’s view only (wizard chrome, not the tree “Link branch” editor).
3. When the child reports finished (and cancelled, matching MR):
   - If the last step was **success**: `selectProject(m.current)` then clear the child.
   - Otherwise: `screenTree`, clear the child, no persist.

MUST NOT: parse parent/child keystrokes, own a second `ConfirmModel` for unlink, or call `project.Link` / `Unlink` from `app.go`.

---

## Quit vs cancel

| Context | `esc` / back | `q` |
|---------|--------------|-----|
| Standalone, any leave | `tea.Quit` | `tea.Quit` |
| Embedded, picker / child / confirm decline | `finished` + `cancelled`, no quit | same (return to tree), matching embedded MR/sync |
| Embedded, success/error done keys | `finished`, no quit; host returns to tree | same |

Done keys on success/error remain `enter` / `esc` / `q` (`keyMatchesDone`).

---

## Prefill (grilling)

1. **Link + PrefillParent in tree** → start at child-name step; parent already set.
2. **Esc on child** → parent picker (rebuild list), even if entry skipped the picker.
3. **Unlink + PrefillTarget in tree** → start at dedicated confirm (subtree size, yes/no). Confirm No/esc leaves the flow; it does not reopen the picker.
4. Prefill MUST NOT duplicate `project.Link` / `Unlink` validation.

---

## Persist and copy

- Confirm Yes link → `project.Link(parent, child)` then success or error step. Success copy unchanged: `Linked %s → %s`.
- Confirm Yes unlink → `ParentOf` + `project.Unlink` then success or error. Success copy unchanged: `Removed %s and %d branches from tree` (existing pluralization).
- Cancel / back before Yes → no `Link` / `Unlink` call.
- Error meanings unchanged (spec 011).

**Forbidden in `internal/tui`**: `Tree.Link` / `UnlinkSubtree` + `SaveTree` (already forbidden by 011).

---

## Standalone process boundary

After `RunLink` / `RunUnlink` returns, the process MUST be at the shell. MUST NOT open the main tree.

---

## Out of scope

- Splitting `app.go` into picker/tree/chrome files
- Splitting `syncflow.go` / `mrflow.go`
- Relocating `confirm.go` / list widgets
- Spec 014 presentation split; spec 015 type unify

---

## Grep audit (merge gate)

```text
rg 'linkParent|linkChild|linkInput|unlinkTarget|unlinkConfirm' internal/tui/app.go
```

**Expected**: no matches.

```text
rg 'func \(m Model\) updateLink|func \(m Model\) updateUnlink' internal/tui/
```

**Expected**: no matches.

`l` / `u` in `updateTree` only construct options and switch screen.
