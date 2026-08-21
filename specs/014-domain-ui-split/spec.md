# Feature Specification: Domain / UI Split

**Feature Branch**: `014-domain-ui-split`

**Created**: 2026-08-21

**Status**: Draft

**Input**: User description: "Move presentation and environment side effects out of domain operations: styled tree rendering, opening browsers, GitLab session construction in interactive screens. One tree walk for display. Behavior freeze. Independently mergeable."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Branch tree meaning stays in the tree; drawing stays in the UI (Priority: P1)

A contributor working on hierarchy rules (roots, children, edges, link/unlink) does not have to reason about colors, selection, or terminal styling. A contributor working on how the tree looks does not change the stored document. Walking the hierarchy for display happens once; the interactive tree and any plain drawing both consume that walk rather than each reimplementing parent/child traversal.

**Why this priority**: The stored tree currently also knows how to paint itself. That mixes two reasons to change.

**Independent Test**: Change a display-only attribute (color or connector glyph) without editing tree load/save or link/unlink. Flatten or list the tree for the interactive view using the shared walk. Confirm link/unlink/sync still see the same edges.

**Acceptance Scenarios**:

1. **Given** a stored branch tree, **When** the interactive tree is drawn, **Then** the parent/child order and membership match the document, and styling lives with the interactive drawing, not with the document.
2. **Given** two display consumers (interactive rows vs a plain indented listing, if both exist), **When** they render the same document, **Then** they share one walk of roots and children rather than two copies of the traversal.
3. **Given** a contributor changes link/unlink rules, **When** they do so, **Then** they do not need to edit drawing/styling to keep those rules correct.

---

### User Story 2 - Sync returns URLs; surfaces open the browser (Priority: P1)

A developer finishes sync with openable merge-request URLs. The sync operation still computes which URLs are openable (created or skipped-already-open, not user-declined). Asking “open in browser?” and opening tabs remains a surface step (scripted prompt or interactive confirm), using a shared open helper that is not part of the sync operation itself.

**Why this priority**: Opening tabs is a side effect of the UI, not of computing a sync plan. It currently sits beside edge processing.

**Independent Test**: Run scripted sync with URLs and decline the open prompt — no tabs, summary still lists URLs. Accept the prompt — tabs open in the same order as today. Interactive sync keeps the same end-of-run confirm. The sync operation’s result is unchanged if the open helper is broken (it still returns URLs).

**Acceptance Scenarios**:

1. **Given** a sync summary with openable URLs, **When** the user declines to open, **Then** no browser tabs open and the summary still contains those URLs.
2. **Given** the user accepts, **When** opening proceeds, **Then** tabs open in the same offer order as today, including the existing brief pause between tabs.
3. **Given** user-declined edges or failures without a URL, **When** openable URLs are collected, **Then** those edges are still excluded.

---

### User Story 3 - Interactive screens never start a GitLab session (Priority: P1)

A developer using interactive sync or MR never depends on the screen constructing its own GitLab client for authentication. Auth and create already live in the operations (spec 011). This spec removes any remaining presentation-layer GitLab session and any leftover auth pre-check in the interactive sync screen. Errors still appear as they do after spec 011.

**Why this priority**: Spec 011 required this; this spec is the cleanup pass if any GitLab construction still sits in drawing/flow code, and it also covers other presentation leaks in the same theme.

**Independent Test**: With GitLab logged out, open interactive sync: the user sees an auth failure from the operation, not from a screen-local client. No interactive flow file holds GitLab session setup.

**Acceptance Scenarios**:

1. **Given** interactive sync or MR, **When** GitLab is unavailable or unauthenticated, **Then** the message shown is the operation’s error.
2. **Given** a reviewer inspects interactive flows, **When** they look for GitLab session setup, **Then** they find none outside the sync and MR operations.
3. **Given** file-change counts on the tree, **When** they load, **Then** they still come from local git comparison as today; that remains a display concern fed by git, not GitLab.

---

### Edge Cases

- File-change counts, direction arrows, and help footers stay presentation. They may call git to compare refs; they must not call GitLab.
- A styled tree listing that is unused should not remain in the document module; drawing belongs with the UI or is removed if nothing calls it.
- Browser open failures remain non-fatal warnings, same wording meaning as today.
- Scripted sync still prompts at the end, not per edge, for opening browsers.
- This spec does not rename direction or created/skipped/failed (spec 015).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The stored branch tree MUST NOT own terminal styling or color. Drawing the tree for humans belongs to the interactive/plain presentation layer.
- **FR-002**: Hierarchy traversal for display MUST have a single walk (roots and ordered children) reused by display consumers.
- **FR-003**: The sync operation MUST NOT open a browser. It MUST still expose openable URLs. Surfaces MUST perform the ask-and-open step.
- **FR-004**: Opening URLs MUST preserve order, the pause between tabs, and non-fatal warnings when an open fails.
- **FR-005**: Interactive flows MUST NOT construct a GitLab session. Auth and create remain inside sync and MR operations.
- **FR-006**: File-change badges MAY use local git; they MUST NOT use GitLab.
- **FR-007**: User-visible drawing, prompts, and open behavior MUST match the previous revision (outcome freeze).
- **FR-008**: After merge, the product MUST remain fully usable without spec 015.

### Key Entities

- **Document / tree**: Stored hierarchy (names, children, edges). No styling.
- **Display walk**: Read-only traversal used to build rows or indented text.
- **Openable URL**: A merge-request URL from a confirmed edge that was created or skipped because it already existed.
- **Surface side effect**: Browser open, terminal paint, prompts — not part of tree or sync core.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A contributor can change tree link/unlink without opening a styling file, and can change tree colors without opening persist/validation, in 100% of those change types.
- **SC-002**: 0 GitLab session constructions remain in interactive flows.
- **SC-003**: 0 browser opens occur inside the sync operation; 100% of open-after-sync paths still honor decline vs accept as today.
- **SC-004**: Interactive tree and any plain listing of the same fixture show the same membership and child order (shared walk).
- **SC-005**: Existing product scenarios for sync, tree display, and counts still pass (0 intended visual or prompt changes).

## Assumptions

- Specs 011–013 are merged.
- `RenderASCII`-style drawing, if unused, is moved or deleted rather than left in the document module.
- Git (local refs, fetch, diffs) is infrastructure for both domain and display; this spec only forbids GitLab and browser side effects in the wrong place, and styling on the document.
- No ports/adapters ceremony for git or GitLab.

## Decisions (Grilling Session 2026-08-21)

| Topic | Decision |
|-------|----------|
| Shipping | Fourth of five; merge leaves the product working |
| Behavior | Outcome freeze |
| Auth | Already owned by sync/MR (spec 011); this spec removes leftover UI GitLab sessions |
| Types | Direction and action unification is spec 015 |
