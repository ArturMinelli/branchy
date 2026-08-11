# Research: Interactive Sync Confirmations

**Feature**: `002-interactive-sync-confirm` | **Date**: 2026-08-11

## R1: TUI sync architecture

**Decision**: Implement `SyncFlowModel` in `internal/tui/syncflow.go` as a self-contained Bubble Tea model, embedded in main `Model` via `screenSync` (same pattern as `MRFlowModel` / `screenMR`).

**Rationale**: Per-edge confirmation requires incremental UI updates between GitLab calls. The current `screenSync` bulk `y` → `sync.Run()` blocks the TUI and auto-confirms all edges. A step-based model (`stepSyncEdgeConfirm` → `stepSyncProcessing` → … → `stepSyncBrowser`) supports async MR creation via `tea.Cmd` and matches the proven MR flow architecture.

**Alternatives considered**:
- *Extend `updateSync` in `app.go` only* — rejected; state machine grows unwieldy; harder to test in isolation.
- *Blocking `sync.Run` with synchronous Confirm callback* — rejected; Bubble Tea cannot render per-edge prompts inside a blocking callback without freezing the UI.
- *One `sync.Run` call per edge from TUI* — rejected; re-runs auth/edge collection each time; awkward partial summaries.

---

## R2: Sync package API changes

**Decision**:
1. Remove auto-browser loop from `sync.Run`.
2. Export `RunEdge(p, edge) Result` for single-edge MR creation (wraps existing `processEdge` without confirm).
3. Add `CreatedURLs(summary *Summary) []string` — filters `Action == "created"` preserving `Results` slice order (DFS).
4. Add `OpenURLs(urls []string) (warn string)` — sequential `browser.Open`, returns non-fatal warning message if any open fails.

**Rationale**: TUI drives the edge loop; CLI keeps `sync.Run` with `Confirm` callback for stdin prompts. Shared helpers ensure identical browser-batch semantics. `Results` order already matches `CollectEdges` DFS because edges are appended in walk order.

**Alternatives considered**:
- *Return URLs from Run and let callers filter* — rejected; duplicates filter logic in CLI and TUI.
- *Browser prompt inside sync.Run via new callback* — rejected; mixes orchestration with I/O; TUI needs different prompt UX than stdin.

---

## R3: CLI end browser prompt

**Decision**: After `sync.Run` completes in `root.go`, call `CreatedURLs(summary)`; if non-empty, prompt `Open created MRs in browser? [y/N]` on stdin; on yes, call `sync.OpenURLs`.

**Rationale**: Matches spec FR-007/FR-008. Per-edge prompts already exist; only auto-browser removal + end prompt are new. `-y` continues to pass `confirm := always true` for per-edge only.

**Alternatives considered**:
- *New `--open-browser` flag* — rejected; spec requires explicit end prompt, not a flag bypass.
- *Reuse MR flow browser prompt component in CLI* — rejected; stdin is appropriate for CLI; no TUI needed.

---

## R4: TUI per-edge prompt UX

**Decision**: On entering sync (`s` key), initialize `SyncFlowModel` with `CollectEdges(syncFrom)` and immediately show first edge prompt:

```text
Sync from <root>

Create MR <parent> → <child>? [y/N]
(edge 1 of N)
```

Keys: `y` confirm, `n` skip, `Esc` cancel remaining (show partial summary if any edges processed).

**Rationale**: Spec FR-002 removes bulk confirm. Edge counter gives progress context without a separate plan screen (grilling decision: remove bulk, go straight to per-edge).

**Alternatives considered**:
- *Plan review screen then per-edge* — rejected per user grilling choice.
- *Reuse CLI-style blocking prompts in TUI* — rejected; inconsistent with MR TUI polish.

---

## R5: Browser tab ordering (Chrome)

**Decision**: Open tabs sequentially via `browser.Open` in DFS order (iteration order of `CreatedURLs`). Add a short fixed delay (50ms) between opens when count > 1.

**Rationale**: Rapid-fire `xdg-open`/`open` calls can cause Chrome to open tabs out of order. Sequential calls with a brief pause is a well-known workaround; 50ms is imperceptible to users but sufficient for the shell to dispatch each open separately.

**Alternatives considered**:
- *Single `browser.Open` with no delay* — rejected; does not reliably preserve order in Chrome.
- *Open all in one browser command with tab syntax* — rejected; platform/browser-specific, fragile across Linux/macOS.
- *Sort by URL string* — rejected; does not match DFS edge order when URLs are unrelated strings.

---

## R6: Cancel / partial sync behavior

**Decision**: `Esc` during per-edge prompt aborts remaining edges; TUI transitions to summary showing results collected so far (created/skipped/failed/declined). Browser prompt appears only if at least one `created` in partial summary.

**Rationale**: Matches spec acceptance scenario for cancellation. User keeps work done on confirmed edges.

**Alternatives considered**:
- *Esc returns to tree with no summary* — rejected; loses visibility into MRs already created.
- *Esc asks "abort remaining?"* — rejected; YAGNI; Esc is sufficient cancel signal.

---

## R7: Skipped vs declined semantics

**Decision**: Keep existing action strings: user decline → `skipped` + message `skipped by user`; existing open MR → `skipped` + glab message. Only `created` enters browser batch.

**Rationale**: No schema change; summary UI already styles `skipped` distinctly. Browser batch filter is action-based.

**Alternatives considered**:
- *New action `declined`* — rejected; unnecessary churn in summary rendering and tests.

---

## R8: Empty edge list

**Decision**: If `CollectEdges(from)` returns empty, show message `No child branches below "<from>".` and return to tree on Enter/Esc (no per-edge or browser prompts).

**Rationale**: Matches spec edge case; avoids empty sync flow.

**Alternatives considered**:
- *Treat as error* — rejected; valid tree state, not a failure.
