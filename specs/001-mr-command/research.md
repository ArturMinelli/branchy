# Research: Manual MR Command

**Feature**: `001-mr-command` | **Date**: 2026-08-11

## R1: MR business logic placement

**Decision**: Create `internal/mr` package with `CreateRequest`, `CreateResult`, and `Create()` function.

**Rationale**: Sync's `processEdge` already implements find-existing → create → return URL. Manual MR needs the same logic with different confirmation/browser semantics. A shared package avoids duplication and gives CLI/TUI a single entry point.

**Alternatives considered**:
- *Inline in CLI only* — rejected; TUI and sync both need the logic.
- *Extend `internal/sync`* — rejected; sync implies tree-walking batch semantics; manual MR is a distinct use case.
- *Put everything in `internal/gitlab`* — rejected; validation (tree membership, same-branch check) is domain logic, not glab transport.

---

## R2: TUI architecture for multi-step MR flow

**Decision**: Implement `MRFlowModel` in `internal/tui/mrflow.go` as a self-contained Bubble Tea model with explicit step enum (`stepSource`, `stepTarget`, `stepTitle`, `stepConfirm`, `stepResult`, `stepBrowserPrompt`).

**Rationale**: The main `Model` in `app.go` already uses screen-based navigation. Embedding `MRFlowModel` as a nested model (or dedicated screen with delegated Update/View) keeps MR flow testable in isolation and runnable standalone via `tui.RunMR(project, opts)`.

**Alternatives considered**:
- *Separate screens in `app.go` only* — rejected; `branchy mr` CLI needs the same flow without the project/tree screens.
- *New top-level package `internal/mrtui`* — rejected; over-abstracted for one flow; belongs alongside existing TUI code.
- *Use `huh` form library* — rejected; adds dependency; bubbles `list` + text input already used in codebase.

---

## R3: Branch picker presentation

**Decision**: Flat sorted list using `bubbles/list` (same delegate as project picker), not the tree view widget.

**Rationale**: Source/target selection is a pick-one-from-many task. A flat filterable list is faster to scan than navigating a hierarchy. Tree view is better for browsing relationships; MR creation needs explicit source vs target labels.

**Alternatives considered**:
- *Reuse `BranchTreeView`* — rejected; conflates hierarchy browsing with directional branch selection.
- *Two-column side-by-side pickers* — rejected; cramped in narrow terminals; harder to implement with bubbles primitives.

---

## R4: Standalone `branchy mr` TUI entry

**Decision**: `tui.RunMR(p *project.Project, opts MROptions)` starts a minimal Bubble Tea program (alt-screen) containing only `MRFlowModel`.

**Rationale**: Matches `tui.Run()` pattern. Avoids booting project picker when user already invoked `branchy mr` from a registered repo cwd.

**Alternatives considered**:
- *Fall back to plain stdin prompts (like sync CLI)* — rejected; spec requires polished TUI UX.
- *Reuse full `tui.Run` and auto-navigate to MR screen* — rejected; extra steps and confusing when cwd resolves to project.

---

## R5: Browser open behavior divergence from sync

**Decision**: MR flow prompts "Open in browser? [y/N]" after success; sync keeps auto-open unchanged.

**Rationale**: Spec explicitly requires user opt-in for manual MR. Sync is batch-oriented where auto-open is expected. `internal/mr.Create()` returns URL without opening browser; callers (TUI or CLI) decide.

**Alternatives considered**:
- *Unify sync to also prompt* — rejected; out of scope, breaking change.
- *Add `--open-browser` flag to CLI mode in v1* — deferred per spec assumptions.

---

## R6: Title editing in TUI

**Decision**: Single-line text input on `stepTitle` screen; default `MR: <source> → <target>`; backspace/type to edit; Enter advances to confirm.

**Rationale**: Matches existing `screenLink` text-input pattern in `app.go`. Full description auto-generated at creation time (not editable).

**Alternatives considered**:
- *Modal popup with `textinput` bubble* — acceptable variant; same UX outcome.
- *Skip title step, edit on confirm screen* — rejected; spec requires dedicated edit step before confirm.

---

## R7: Refactoring sync to use `internal/mr`

**Decision**: Extract shared `mr.Create()` call from `sync.processEdge`; sync retains batch loop, auto-browser, and per-edge Confirm callback.

**Rationale**: DRY without changing sync's external behavior. `mr.Create()` accepts `OpenBrowser bool` defaulting false; sync sets true after collecting URLs (existing behavior).

**Alternatives considered**:
- *Copy-paste MR logic* — rejected; maintenance burden.
- *Full sync rewrite* — rejected; unnecessary scope.

---

## R8: CLI flag-mode confirmation

**Decision**: When `--source` and `--target` provided without `--yes`, use stdin `Create MR <source> → <target>? [y/N]` (same pattern as sync CLI).

**Rationale**: Consistent with existing `branchy sync` non-interactive/interactive hybrid. No TUI launched when both branch flags present.

**Alternatives considered**:
- *Require `--yes` always in flag mode* — rejected; less ergonomic for interactive terminal use.
