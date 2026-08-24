# Research: Domain / UI Split

**Feature**: `014-domain-ui-split` | **Date**: 2026-08-24

## 1. Unused `RenderASCII`

**Decision**: Delete `internal/tree/render.go` (`Document.RenderASCII` and its lipgloss styles). Do not move it into TUI.

**Rationale**: Grep finds no callers outside the definition. Spec assumption: unused styled drawing is removed rather than left on the document. Grilling locked delete-not-move. After deletion, `internal/tree` has no lipgloss import (FR-001).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Move `RenderASCII` into TUI as a plain listing | Grilling chose delete; nothing consumes a plain listing today |
| Keep it on `Document` but drop colors | Still presentation on the stored tree; unused |

---

## 2. Display walk home and shape

**Decision**: Add an unstyled walk on `Document`:

```go
type DisplayNode struct {
    Name        string
    Depth       int
    IsRoot      bool
    IsLast      bool
    LastAtDepth []bool // ancestor last-sibling flags; len == Depth
}

func (d *Document) WalkDisplay() []DisplayNode
```

Place it in `internal/tree/tree.go` next to `Roots` (same DFS family as `CollectEdges`, different payload). Child order matches today’s display: `sort.Strings` on each node’s children; roots via existing `Roots()` (already sorted). Empty document → empty slice.

`internal/tui/treeview.go` `flattenTree` maps each `DisplayNode` to `branchRow` and **paints glyphs there**:

| Structural field | Glyph (TUI only) |
|------------------|------------------|
| `IsRoot` | no connector |
| `!IsRoot && IsLast` | `└── ` |
| `!IsRoot && !IsLast` | `├── ` |
| ancestor `LastAtDepth[i] == true` | `"    "` |
| ancestor `LastAtDepth[i] == false` | `"│   "` |

`LastAtDepth` is structural (which ancestors were last siblings), not a drawing string. Box-drawing stays in TUI so a second consumer can choose different glyphs without a second tree walk.

**Rationale**: Grilling locked “delete renderer; extract unstyled walk; treeview consumes it.” A name-only list would force each consumer to re-walk for last-child spines. `CollectEdges` is the sync-edge walk (parent/child pairs, up or down); merging it with display would mix two reasons to change.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Leave `flattenTree` as the only walk after deleting `RenderASCII` | Grilling required a `Document` walk; FR-002 wants one reusable traversal |
| Yield prefix/connector strings from `Document` | Leaks presentation into the document (FR-001) |
| Visitor callback instead of a slice | Trees are small; a slice matches `flattenTree` and is easier to test |
| Unify with `CollectEdges` | Different order/payload (edges vs nodes); sync must not depend on display |

---

## 3. Sequential browser open

**Decision**: Move `sync.OpenURLs` to `internal/browser.OpenURLs`. Keep behavior:

- Open in given order
- `200 * time.Millisecond` pause **between** tabs (not before the first)
- Non-fatal warnings: `could not open %s: %v`, joined with `"; "`
- Empty/nil input → `""`

Test hook moves with it (`var openURL = Open` in `browser`, same role as today’s `sync.browserOpen`).

`sync.OpenableURLs` stays in `sync` (created + skipped-already-open, exclude user-declined and URL-less failures). CLI and TUI still: compute URLs from the summary → ask → on yes call `browser.OpenURLs`.

`internal/tui/mrflow.go` keeps single-URL `browser.Open` (not this spec’s sequential helper).

**Rationale**: Grilling locked sequential open in `browser`. Spec FR-003: sync must not open a browser; it must still expose openable URLs. Spec 011 explicitly deferred this move to 014. Deepening `browser` next to `Open` is the natural home (no new package).

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Keep `OpenURLs` on `sync` | Violates FR-003; grilling rejected |
| Duplicate the loop in CLI and TUI | Two copies of pause/warning wording |
| Force MR through `OpenURLs` | Outcome freeze; MR is one tab and already uses `Open` |

---

## 4. GitLab in interactive flows (US3)

**Decision**: Verification only. No implementation tasks beyond a merge-gate grep. Today `internal/tui` has zero `gitlab` imports; auth/create stay inside `sync.Begin` / `mr.Create` (spec 011).

**Rationale**: Grilling locked audit-only. FR-005 is already satisfied; keep it as a regression contract so it cannot return.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Full US3 implementation story | No leftover session construction to move |
| Drop US3 from 014 | Still a required invariant; grep is cheap |

---

## 5. What this spec does not change

**Decision**: File-change badges stay TUI + local git (`internal/git`). Direction types and created/skipped/failed strings stay as they are (015). No new `app`/`usecase` package. No GitLab interface. Prompt copy, [y/N] default, and TUI “Open MRs in browser?” confirm stay.

**Rationale**: Spec edge cases and FR-006–008. Git is infrastructure for both domain and display.
