# Research: Shared Use Cases

**Feature**: `011-shared-use-cases` | **Date**: 2026-08-21

## 1. Where persist lives (link / unlink)

**Decision**: `project.Link(parent, child) error` and `project.Unlink(parent, child) (*UnlinkResult, error)` are the surface-facing operations. Each calls the existing tree primitive, then `SaveTree`. `tree.Document.Link` and `UnlinkSubtree` stay exported in-memory primitives for tree tests and for the project ops to call.

**Rationale**: FR-001/002 require validate + mutate + persist in one place. Persist needs the project id (`config.TreePath`). Grilling locked “deepen project, keep tree primitives.” Putting SaveTree on every surface leaves a second copy of the sequence (today’s bug). Unexporting tree mutate would force every tree unit test through the store and is out of scope.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Surfaces keep `Tree.Link` + `SaveTree` | Violates FR-001; the persist sequence stays copied |
| Unexport `tree.Link` / `UnlinkSubtree` | Grilling kept primitives; tree tests would need a `Project` |
| New `internal/usecase` package | Spec FR-006 / grilling: no application layer |

---

## 2. Unlink signature and roots

**Decision**: `Unlink(parent, child)`:

| Parent | Child | Rule |
|--------|-------|------|
| non-empty | any | Require `HasEdge(parent, child)`, then `UnlinkSubtree(child)` |
| empty | a root (`ParentOf` missing) | Allow; `UnlinkSubtree(child)` |
| empty | not a root | Error — caller must pass the real parent |
| any | missing / empty / same as parent | Same errors as today (`branch %q not in tree`, required, must differ) |

Interactive paths pass `ParentOf(child)` (empty string when the target is a root). Scripted CLI keeps requiring two args and therefore the edge check.

Return `UnlinkResult{Parent, Child, Removed}` so CLI can print `Unlinked %s → %s (%d branches removed)` without recounting after the delete.

**Rationale**: Grilling: complete scripted validation when parent is given; roots remain unlinkable from TUI. Today’s TUI never asked for parent, so `ParentOf` is the only extra input, not a second rule implementation.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Always require an edge (roots fail everywhere) | Grilling: TUI root unlink stays |
| Child-only operation | Drops the scripted edge check the spec said must win |
| Two functions (`UnlinkEdge` / `UnlinkRoot`) | Extra surface; one function with the empty-parent root path is enough |

---

## 3. Sync: one owner without a blocking Confirm in the TUI

**Decision**: Keep both entry points:

- `sync.Run` — CLI batch walk with optional `Confirm` callback (`-y` or stdin). Unchanged control flow.
- `sync.RunEdge` — TUI “Yes” on one edge (already used).
- `sync.SkippedByUser(edge, dir) Result` — TUI “No”; same `Action` / `Message` (`skipped by user`) as `processEdge` when Confirm returns false.
- `sync.Begin(p, from, dir) ([]tree.Edge, error)` — TUI `prepareSyncFrom`: validate `from`, collect `EdgesBelow`, and **if there is at least one edge** run the same GitLab auth check `Run` uses. Empty list → `nil` error, no auth (preserves today’s empty-tree TUI screen).

TUI must not import `internal/gitlab` or construct `gitlab.Client`. Decline must not hand-build a `sync.Result`.

**Rationale**: Bubble Tea cannot block inside `sync.Run`’s Confirm without a goroutine/channel pair. Grilling kept the TUI confirm loop. The leak is TUI-owned auth and a duplicate skip struct, not the fact that the TUI advances one edge per key.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Force TUI through `sync.Run` + channel Confirm | Grilling rejected; fights the event loop |
| Delete `sync.Run`; CLI loops on `RunEdge` | Moves the cascade into the CLI module; spec wants it in sync |
| Auth only inside `mr.Create` (first Yes) | Grilling: fail after root pick, before first confirm, matching today |

---

## 4. Auth ownership and timing

**Decision**:

- `mr.Create` keeps its `AuthOK` (manual MR and each `RunEdge` still go through it).
- `sync.Run` keeps its start-of-run `AuthOK` (scripted path unchanged, including “empty subtree still auths”).
- Interactive sync calls `sync.Begin` after the root is chosen. Auth error text stays `glab auth: %v (run: glab auth login)`. TUI only displays `err.Error()`.

Empty interactive plan (no edges) still skips auth, as today. That CLI/TUI difference is existing and not in this spec’s disagreement list (unlink validation is).

**Rationale**: FR-004/009: surfaces do not own GitLab. Grilling: same *moment* as today’s TUI pre-check, moved into sync.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| TUI keeps `gitlab.Client` in `prepareSyncFrom` | Spec FR-009 |
| Auth on first confirmed create only | Later than today; grilling said keep timing |
| `Begin` always auths, even with zero edges | Would change the TUI empty screen to an auth error |

---

## 5. Init and create-MR are already the owner

**Decision**: No new init or create function. `project.Init` is already called by scripted `init` and `initflow.runInitCmd`. `mr.Create` is already called by scripted `mr` and `mrflow`. Work is: add missing `project` tests so the shared init/link/unlink path is the test surface, and keep TUI/CLI from growing a second register/create.

**Rationale**: Spec US4 is “lock the path we already have.” Deleting or wrapping `Init` would be churn without a second caller to collapse.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `project.Register` alias | Extra name; `Init` is already the operation |
| Move GitLab client construction behind an interface | Spec 015 / 014 out of scope; one `gitlab.Client` inside `mr.Create` is enough |

---

## 6. Save failure and in-memory mutation

**Decision**: Keep today’s order: mutate the in-memory `Document`, then `SaveTree`. If save fails, the in-memory tree is already changed and the error is returned. Do not roll back. Main-tree unlink tests already assert mutation when save cannot write (no real project path).

**Rationale**: Outcome freeze. A two-phase commit would be new behavior.

---

## 7. What this spec does not move

**Decision**: Browser opening stays `sync.OpenURLs` called from CLI and TUI (014). Main-tree link/unlink stay inline screens in `app.go` (013). Command definitions stay in `internal/cli/root.go` (012). `Result.Action` stays a string; direction types stay as they are (015).

**Rationale**: Independently mergeable; FR-010.
