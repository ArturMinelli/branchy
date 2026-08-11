# Research: Unlink Branch from Tree

**Feature**: `003-unlink-branch` | **Date**: 2026-08-11

## R1: Tree mutation API placement

**Decision**: Add `UnlinkSubtree(root string) error` on `tree.Document`, plus helpers `SubtreeNames(root) []string`, `ParentOf(child) (string, bool)`, and `HasEdge(parent, child) bool`.

**Rationale**: Unlink is the inverse of link at the data layer but removes a subtree, not a single edge. Centralizing in `internal/tree` keeps CLI and TUI thin and testable, matching how `Link` already lives on `Document`.

**Alternatives considered**:
- *CLI/TUI inline map surgery* — rejected; duplicated logic, hard to test.
- *Separate `internal/unlink` package* — rejected; one function family does not warrant a package.
- *Only `Unlink(parent, child)` without subtree helper* — rejected; subtree collection needed for both mutation and TUI count.

---

## R2: Subtree collection algorithm

**Decision**: DFS from `root` (same traversal as `collectEdges`) collecting all reachable branch names into a set; delete each key from `Branches`; then remove `root` from its parent's `Children` slice if a parent exists.

**Rationale**: Reuses proven DFS pattern from `CollectEdges`. Order of deletion does not matter for map keys. Parent cleanup is a separate pass using `ParentOf`.

**Alternatives considered**:
- *BFS* — equivalent outcome; DFS matches existing codebase style.
- *Mark-and-sweep entire document* — rejected; overkill for single-subtree removal.

---

## R3: Parent lookup strategy

**Decision**: `ParentOf(child)` scans all `Branches` entries for `child` in `Children`; returns first match. Assumes single parent per spec.

**Rationale**: Tree is small (typical <50 branches); no reverse-index needed. Spec assumes at most one parent; if multiple existed, first match is acceptable interim behavior.

**Alternatives considered**:
- *Maintain parent pointers in YAML* — rejected; schema change.
- *Build reverse map on every unlink* — acceptable optimization later; YAGNI for v1.

---

## R4: CLI contract shape

**Decision**: `branchy unlink <parent> <child>` — validates edge exists, then calls `UnlinkSubtree(child)`.

**Rationale**: Mirrors `branchy link <parent> <child>` for discoverability and scripting symmetry. Parent argument confirms intent and enables clear error when edge is missing.

**Alternatives considered**:
- *`branchy unlink <branch>` only* — rejected; user chose CLI parity with link during grilling.
- *Interactive stdin parent/child* — rejected; link does not do this; unnecessary for CLI.

---

## R5: TUI interaction model

**Decision**: On tree view, `u` with a selected branch transitions to `screenUnlinkConfirm` showing branch name + subtree count; `y`/Enter confirms, `n`/Esc cancels. No text input.

**Rationale**: User chose "select + confirm" during grilling. Simpler than link's two-field form because target is implicit (selected branch). Reuses existing `Yes`/`No` key bindings.

**Alternatives considered**:
- *Mirror link two-field screen* — rejected per grilling decision.
- *Dedicated `UnlinkFlowModel`* — rejected; single confirm step does not justify nested model (unlike sync/MR flows).

---

## R6: Confirmation copy and subtree count

**Decision**: Display: `Remove "feature-a" and N branch(es) from tree? [y/N]` where N = `len(SubtreeNames(selected))`.

**Rationale**: Meets FR-002; singular/plural handled in message. Includes selected branch in count.

**Alternatives considered**:
- *List all branch names* — rejected; noisy for large subtrees.
- *Warn only when N > 1* — rejected; spec requires count always.

---

## R7: Error handling and atomicity

**Decision**: Validate branch exists (and edge exists for CLI) before mutating in memory; call `SaveTree()` once after mutation; on save failure, reload tree from disk via `selectProject` refresh or re-read to avoid drift.

**Rationale**: Matches link flow: mutate then save; surface save errors to user. No partial disk writes because save is single file write.

**Alternatives considered**:
- *Write-ahead temp file* — rejected; YAGNI; link does not do this.
- *In-memory rollback on save fail* — acceptable; simplest path is reload project from disk on error.

---

## R8: Empty tree and root unlink

**Decision**: Allow removing last branch(es); TUI shows existing `(empty tree)` state from `BranchTreeView`.

**Rationale**: Spec explicitly allows empty tree. `Roots()` returns empty slice; no special-case blocking.

**Alternatives considered**:
- *Block unlink if it would empty tree* — rejected; contradicts spec edge case.

---

## R9: README and help updates

**Decision**: Update README keys section and TUI footer help in same feature; no separate docs task.

**Rationale**: FR-010 requires discoverability; README is user-facing source of truth alongside TUI footer.

**Alternatives considered**:
- *TUI-only help* — rejected; incomplete discoverability for CLI users.
