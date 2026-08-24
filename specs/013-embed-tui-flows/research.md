# Research: Embed TUI Flows

**Feature**: `013-embed-tui-flows` | **Date**: 2026-08-24

## 1. Host pattern for embedded link/unlink

**Decision**: Copy the existing sync/MR embed: add `linkFlow LinkFlowModel` and `unlinkFlow UnlinkFlowModel` on `Model`; screens `screenLink` / `screenUnlink`; when that screen is active, `Update`/`View` forward to the child model; when `finished` (and for unlink, `cancelled` if used the same way as MR), return to `screenTree` and zero the child. Do not extract a generic embed dispatcher. Do not split `app.go`.

**Rationale**: Spec FR-008 forbids splitting the main session. Sync and MR already prove the 10-line forward pattern. A shared helper would be a third abstraction for two more call sites.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Extract `embedFlow(update, view, finished)` helper | YAGNI; spec forbids extra files; four flows is still copy-paste sized |
| Nested `tea.NewProgram` for link/unlink | Cannot return to the tree in the same session |
| Keep inline screens that call `project.Link` | Violates FR-003 (second implementation) |

---

## 2. Options shape (Embedded + prefill)

**Decision**:

```text
type LinkFlowOptions struct {
    Embedded      bool
    PrefillParent string
}

type UnlinkFlowOptions struct {
    Embedded      bool
    PrefillTarget string
}
```

`RunLink` / `RunUnlink` force `Embedded: false` (same as `RunSync`). `newLinkFlowModel(p, opts)` / `newUnlinkFlowModel(p, opts)` replace the current single-arg constructors. Standalone tests pass empty options.

**Rationale**: `MROptions` already has `Embedded` + `PrefilledSource`. Prefill is the only extra input the spec allows; it must not reimplement link/unlink validation (use `p.Tree.Branches` membership / existing `ParentOf` + `project.Unlink` like today).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Separate `NewEmbeddedLink(p, parent)` constructors | Two construction paths; options struct is the existing grain |
| Pass prefill via mutating fields after `new*FlowModel` | Hides the skip-step in the host; MR already skips inside the constructor |

---

## 3. Prefill skip-step rules (grilling 2026-08-24)

**Decision**:

| Launch | Prefill | Starting step |
|--------|---------|---------------|
| Standalone `branchy link` | none | parent picker |
| Main tree `l` with a selected name | `PrefillParent = selection` | **child name** (picker skipped) |
| Main tree `l` with no selection | none | parent picker |
| Standalone `branchy unlink` | none | target picker |
| Main tree `u` with a selected in-tree name | `PrefillTarget = selection` | **confirm** (picker skipped) |
| Main tree `u` with empty selection | do not launch | stay on tree (today’s no-op) |
| Main tree `u` with name not in tree | do not special-case in the host | if launched, flow error step with the same meaning as `project.Unlink` / standalone (“not in tree”) |

Esc from the **child** step always opens the parent picker (rebuild the list, same as standalone), even when the picker was skipped on entry. Esc / No on **confirm** leaves the flow (standalone already exits; embedded returns to tree) — it does not return to the picker.

**Rationale**: Grilling locked skip-picker convenience plus “esc from child opens picker” so the user can still change parent without a second link implementation.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Show picker with selection highlighted | Grilling chose skip |
| Esc from child returns to tree | Grilling chose open picker |
| Unlink always shows picker | Grilling chose skip to confirm |

---

## 4. Embedded `q` / `esc` vs `tea.Quit`

**Decision**: Add `cancelOrQuit` / `finish` helpers on both flow models, copied from `MRFlowModel`:

- Not embedded: `finished` (+ `cancelled` when backing out) and `tea.Quit` (standalone process exits).
- Embedded: same flags, **return `nil` cmd** — the host returns to the tree. Do **not** quit the main session.

`q` and `esc` on pickers follow MR/sync embedded behavior (leave the flow → tree), not the old inline link editor (`q` quit the whole app). Spec edge case: match embedded sync/MR. After success/error, done keys (`enter` / `esc` / `q`, same as today’s `keyMatchesDone`) call `finish`.

**Rationale**: FR-007 keeps sync/MR behavior; the spec tells link/unlink to copy those quit-vs-cancel rules. Wizard chrome is an allowed visible change.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Embedded `q` quits the whole app (old inline link) | Diverges from embedded sync/MR; spec says match those |
| Host intercepts `q` before forwarding | Duplicates key policy; flows already own `q`/`esc` |

---

## 5. Success / error then tree refresh (grilling 2026-08-24)

**Decision**: Keep the wizard success and error steps. After a done key:

- **Success** (`stepLinkSuccess` / `stepUnlinkSuccess`): host calls `selectProject(m.current)` so the tree, inbound/outbound counts, and direction arrows refresh (same as today’s inline save path).
- **Error** or **cancel**: host sets `screenTree` and zeros the child flow; no persist; do not require `selectProject` (tree unchanged). Clearing `errMsg` on the tree matches `selectProject`’s current reset for the success path only.

Persist still happens on confirm Yes via `project.Link` / `project.Unlink` before the success step — not in `app.go`.

**Rationale**: Grilling locked “show success/error, then refreshed tree”. FR-005 requires cancel = no persist and success = refresh. Immediate jump-to-tree would hide the standalone success copy.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Skip success screen; `selectProject` immediately | Grilling rejected; standalone users would learn a second ending |
| Refresh even on cancel | Unnecessary; tree did not change |

---

## 6. What is deleted from `app.go`

**Decision**: Remove `updateLink`, `updateUnlink`, and Model fields `linkParent`, `linkChild`, `linkInput`, `unlinkTarget`, `unlinkCount`, `unlinkConfirm`. `View` must not render the two-field “Link branch” editor or a host-owned unlink confirm; those screens are `m.linkFlow.View()` / `m.unlinkFlow.View()`.

Grep gate after merge: no `linkParent` / `unlinkConfirm` on `Model`; `l`/`u` only construct flow options and switch screen.

**Rationale**: FR-003 — zero remaining inline implementations (SC-001).

---

## 7. What this spec does not change

**Decision**: `syncflow.go`, `mrflow.go`, project picker, direction chords, tree navigation, CLI `RunLink`/`RunUnlink` process boundary, shared widgets (`confirm.go`, list helpers) stay in their current files. Specs 014–015 untouched. No GitLab in flows (already true). No new TUI files.

**Rationale**: FR-007, FR-008, FR-009.
