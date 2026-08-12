# Research: Diff Direction Toggle

**Feature**: `008-diff-direction-toggle` | **Date**: 2026-08-12

## 1. Outbound comparison semantics

**Decision**: Outbound file-change count is the three-dot reverse of inbound:

```text
# inbound (existing):  files parent would bring into child
git -C <repo> diff --name-only <child>...<parent>

# outbound (new):      files child would bring into parent
git -C <repo> diff --name-only <parent>...<child>
```

Expose `OutboundFiles(dir, parent, child string) (int, error)` with the **same argument order** as `InboundFiles` (parent, then child). Internally it is the merge-base..child side — equivalent to calling `InboundFiles(dir, child, parent)`, but callers must not swap args themselves.

**Rationale**: Spec FR-003 / SC-003 require the same files-changed metric GitLab would show on an MR **child → parent**. Three-dot already proved correct for inbound in 006; flipping the ends is the dual. Keeping `(parent, child)` argument order matches the tree edge vocabulary and prevents accidental double-swaps in TUI code.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Two-dot `git diff parent child` | Counts both sides; grilling locked one-sided files-changed |
| Reuse `InboundFiles` with swapped call sites only | Easy to invert wrong at a call site; named wrapper documents intent |
| Commit ahead count | Grilling rejected; metric stays files changed |
| GitLab compare API | Local-only; same constraints as 006 |

---

## 2. Shared file-count type and loaders

**Decision**:

- Keep a single ephemeral count value type (rename `inboundCount` → `fileChangeCount` in `internal/tui` so both directions share it without lying).
- Add `loadOutboundCounts(dir, doc)` mirroring `loadInboundCounts`, calling `git.OutboundFiles` per non-root child.
- Reuse `formatInboundBadge` (or rename to `formatFileChangeBadge`) for both directions — hide zero, `?` for unknown.
- Sync continues to call only `loadInboundCounts`.

**Rationale**: Struct shape is identical (`files`, `ok`). Dual loaders keep the ParentOf walk and error→unknown mapping in one place per direction. Sync must never see outbound (FR-007).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| One loader returning both maps | Couples sync’s inbound-only path to outbound work |
| Separate badge formatters | Display rules are identical (FR-004) |
| Compute outbound by swapping inbound map keys | Wrong — inbound N ≠ outbound M for diverged edges |

---

## 3. When outbound counts are computed

**Decision**: Eagerly load **both** inbound and outbound maps whenever the main tree refreshes counts (`selectProject` and successful `remoteUpdateMsg` apply). Do not wait for the first toggle. Cache both on `BranchTreeView`. Toggle only flips which map `renderRow` reads.

**Rationale**: Spec trees are small (same ≤50-edge budget as 006). Doubling local diffs at project open is still sub-noticeable and makes the first toggle instantaneous (FR-009). Lazy-on-toggle would risk a hitch and complicate refresh (007 already recomputes on fetch success — both maps must refresh in the **current** direction).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Lazy load outbound on first toggle | First press may stall; refresh path must still know whether outbound is warm |
| Recompute active direction only on remote update | Other map goes stale; next toggle shows old numbers |
| Async tea.Cmd for outbound | Overkill for local git; 005 reserved loading chrome for network |

---

## 4. Session-scoped direction ownership

**Decision**: Store `diffDirection` on the root `Model` (default inbound). Persist for the Bubble Tea process only — not in project YAML, not on disk. Survives leaving the tree (sync, MR, project picker) and switching projects because `Model` outlives those screens. `selectProject` rebuilds the tree view and reloads both maps but **does not** reset direction.

**Rationale**: FR-006 requires session-wide mode across project switches and returns from sync. Putting direction on `BranchTreeView` would lose it when `selectProject` replaces `treeView`. Root model is the natural session owner.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Direction on `BranchTreeView` | Lost on every `newBranchTreeView` in `selectProject` |
| Per-project remembered mode | Spec / grilling: session only |
| Config / env flag | Spec: no CLI flag; TUI keystroke only |

---

## 5. Toggle key and footer copy

**Decision**:

- Bind **`d`** (“direction” / “diff direction”) on the main tree only.
- Footer always includes the toggle and names the **current** mode, for example:
  - Inbound: `…  d: show outbound  …` plus a short mode cue `counts: inbound (parent→child)`
  - Outbound: `…  d: show inbound  …` plus `counts: outbound (child→parent)`
- Exact string lives in one helper (`treeHelpFooter(direction)`) so View and tests share it.
- Sync / other screens do not bind `d` and do not mention direction.

**Rationale**: `d` is unused on the main tree (`s/m/l/u/q/esc/↑↓` taken). Footer-only mode cue was grilled (no badge arrows). Naming the *other* action on the key (`show outbound`) plus an explicit current-mode fragment satisfies FR-005 and SC-006 without cluttering every row.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| `o` for outbound | Asymmetric; after flip users think of “direction”, not “outbound” |
| `t` for toggle | Vague; conflicts with mental model of “tab” |
| Arrow/prefix on every badge | Rejected in grilling |
| Footer key only, no mode label | Fails “tell which direction is active” |

---

## 6. Tree rendering with two maps

**Decision**: Extend `BranchTreeView` to hold `inbound` and `outbound` maps plus a `direction` field (mirrored from `Model` on each toggle / project select). `renderRow` selects the active map; badge formatting unchanged. `setInbound` becomes `setFileCounts(inbound, outbound)` (or paired setters). Remote refresh updates both maps without changing direction.

**Rationale**: View already owns badge placement. Keeping both maps on the view avoids re-pushing the active slice on every frame. Direction mirrored onto the view keeps `View()` free of reaching into `Model`.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Single active map swapped on toggle | Remote refresh must recompute both anyway; easy to drop the inactive cache |
| Model picks map in View() | Leaks badge policy out of treeview |

---

## 7. Sync isolation

**Decision**: No changes to sync’s count loading or display. `SyncFlowModel` keeps `inbound` only. Embedded sync started while the tree is outbound still loads inbound for picker/confirm. Direction key is handled only in `updateTree`.

**Rationale**: FR-007 / User Story 4 — sync creates parent→child MRs; outbound numbers would contradict the confirm question.

---

## 8. Testing strategy

**Decision**:

- **`internal/git`**: temp-repo tests for `OutboundFiles` — child-only files → N; identical → 0; parent-only after diverge → 0 outbound; missing ref → error; symmetry check that `OutboundFiles(dir,p,c) == InboundFiles(dir,c,p)`.
- **`internal/tui`**: inject dual maps; assert inbound default, toggle swaps badges, zero-hide per active direction, footer strings, direction survives simulated leave/return (Model field unchanged), sync path still inbound-only (existing sync tests + one regression that tree outbound does not affect sync helpers).

**Rationale**: Same split as 006 — real git for semantics, injected maps for presentation.

---

## 9. README / discoverability

**Decision**: Update the main TUI keys line in `README.md` to mention `d` and inbound/outbound counts (same pattern as unlink’s footer + README update in 003).

**Rationale**: Footer is the primary cue (FR-005); README stays the external key reference.
