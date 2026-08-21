# Feature Specification: Embed TUI Flows

**Feature Branch**: `013-embed-tui-flows`

**Created**: 2026-08-21

**Status**: Draft

**Input**: User description: "Main tree link and unlink should reuse the dedicated command flows instead of inline copies. Freeze outcomes (saved trees, errors, keys that still exist). Wizard screens on the main tree are allowed. Do not split the main app file into extra screens or split sync/MR flows in this pass."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Main-tree link is the same flow as standalone link (Priority: P1)

A developer presses `l` on the main branch tree. They go through the same link wizard they would see from `branchy link` (parent selection, child name, confirm, result), not a one-screen two-field editor unique to the main tree. Saving still uses the shared link operation. Cancelling returns them to the tree with no persist. Starting from a selected branch may still prefill the parent, matching today’s convenience if the dedicated flow already allows a prefilled parent; if not, prefill is allowed as the only extra input, not a second implementation of link rules.

**Why this priority**: The main tree is the only remaining duplicate UI for link. Spec 011 unified the operation; this spec unifies the interactive path.

**Independent Test**: Link the same parent→child from `branchy link` and from `l` on the main tree. Same validation, same persisted tree. Cancel from the main-tree wizard; tree unchanged. After save, the main tree shows the new child.

**Acceptance Scenarios**:

1. **Given** the main tree with a branch selected, **When** the developer presses `l`, **Then** they enter the link wizard (not an inline two-field editor on the tree screen).
2. **Given** they complete the wizard with a valid pair, **When** they confirm, **Then** the edge is persisted via the shared link operation and the tree view refreshes to include it.
3. **Given** they cancel or go back from the wizard, **When** they return to the tree, **Then** no edge was added.
4. **Given** an invalid pair, **When** they try to confirm, **Then** they see the same rejection meaning as standalone `branchy link`.

---

### User Story 2 - Main-tree unlink is the same flow as standalone unlink (Priority: P1)

A developer presses `u` on a selected branch. They get the dedicated unlink confirmation (subtree size, yes/no), not a one-off confirm wired only in the main app. The shared unlink operation performs the delete. Cancel returns to the tree.

**Why this priority**: Same duplication as link, including the historical validation gap that spec 011 closed.

**Independent Test**: Unlink a subtree from `branchy unlink` and from `u` on the main tree. Same members removed. Cancel leaves the tree intact. After confirm, the main tree no longer lists those branches.

**Acceptance Scenarios**:

1. **Given** a selected branch that exists in the tree, **When** the developer presses `u`, **Then** they enter the unlink wizard/confirm used by standalone unlink.
2. **Given** they confirm, **When** unlink completes, **Then** the subtree is gone on disk and the main tree refreshes.
3. **Given** they decline, **When** they return to the tree, **Then** the branch is still there.
4. **Given** the selection is not a valid unlink target, **When** they try, **Then** the error matches standalone unlink.

---

### User Story 3 - Sync and MR stay embedded as they are; main app stays a router (Priority: P2)

A developer pressing `s` or `m` still gets the existing embedded sync and MR flows. The main session still hosts project picker, tree, and those flows. This spec does not break those embeddings and does not split the main session into extra source files for picker/tree/chrome, and does not split the sync or MR wizards into more files.

**Why this priority**: The goal is to delete the inline link/unlink copies, not to re-architect the whole interactive app.

**Independent Test**: From the main tree, run sync and manual MR as today. Run standalone `branchy link` / `unlink` and confirm they still exit to the shell rather than the main tree.

**Acceptance Scenarios**:

1. **Given** the main tree, **When** the developer presses `s` or `m`, **Then** the existing embedded sync/MR behavior is unchanged.
2. **Given** standalone `branchy link` or `unlink`, **When** the wizard finishes or quits, **Then** the process returns to the shell (does not open the main tree).
3. **Given** project picker, direction chords, and tree navigation, **When** the user uses them, **Then** they behave as in the previous specs.

---

### Edge Cases

- Outcome freeze: persisted trees, error meanings, `l` / `u` / `esc` / `q`, and standalone vs main-tree process boundaries stay. The main-tree *screens* for link/unlink may look like the standalone wizards.
- Keys that still exist keep their roles (`l` link, `u` unlink, `esc` back/cancel, `q` quit). Wizard-internal keys match the standalone flows (tab, enter, y/n) even if the old inline editor handled typing differently.
- After a successful link or unlink from the main tree, inbound/outbound counts and direction arrows refresh as they do after today’s inline save.
- Quit (`q`) from an embedded wizard follows the same quit-vs-cancel rules as embedded sync/MR today (quit the application vs return to tree). Match whichever those flows already do for `q` vs `esc`, and apply it consistently to embedded link/unlink.
- Prefill: starting link from a selected branch should still offer that branch as parent when it is a valid parent; the wizard must not reimplement link validation to do so.
- Out of scope: splitting the main session file into picker/tree/chrome files; splitting sync/MR wizard files; extracting shared list widgets into new files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Main-tree link MUST run the same interactive link flow as standalone `branchy link`, embedded in the main session.
- **FR-002**: Main-tree unlink MUST run the same interactive unlink flow as standalone `branchy unlink`, embedded in the main session.
- **FR-003**: The main session MUST NOT keep a second inline link editor or a second unlink-only confirm implementation.
- **FR-004**: Both embeddings MUST call the shared link/unlink operations (spec 011), not mutate-and-save locally.
- **FR-005**: Cancel/back from an embedded wizard MUST return to the tree with no persist; success MUST refresh the tree.
- **FR-006**: Standalone `branchy link` / `unlink` MUST continue to be dedicated sessions that exit to the shell.
- **FR-007**: Sync, MR, project picker, direction chords, and tree navigation MUST keep their current behavior.
- **FR-008**: This spec MUST NOT include splitting the main session into additional screen files or splitting sync/MR flow files.
- **FR-009**: After merge, the product MUST remain fully usable without spec 014–015.

### Key Entities

- **Main session**: The interactive program launched by `branchy` with no subcommand (project picker + tree + embedded flows).
- **Dedicated flow**: The interactive wizard launched by a subcommand; also the implementation reused when embedded.
- **Embedded vs standalone**: Same flow, different host — return to tree vs exit to shell.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 0 remaining inline link/unlink implementations in the main session; 100% of main-tree `l`/`u` paths go through the dedicated flows.
- **SC-002**: Linking or unlinking the same pair from standalone and from the main tree produces the same on-disk tree in 100% of fixture cases used in tests.
- **SC-003**: Cancel from embedded link or unlink leaves the tree unchanged in 100% of trials.
- **SC-004**: Existing sync-from-tree, MR-from-tree, direction, and picker scenarios still pass (0 regressions).
- **SC-005**: A first-time user who has used `branchy link` can complete main-tree link without learning a second editor, in one attempt.

## Assumptions

- Specs 011 and 012 are merged (shared operations; command files already thin).
- Outcome freeze, not pixel freeze: wizard chrome on the main tree is an allowed visible change.
- Prefilling parent from the current tree selection is desirable and does not count as a second link implementation.
- Shared list/confirm widgets may be referenced by multiple flows as they already are; this spec does not relocate them into new files.

## Decisions (Grilling Session 2026-08-21)

| Topic | Decision |
|-------|----------|
| Shipping | Third of five; merge leaves the product working |
| Freeze vs embed | Freeze outcomes; main-tree link/unlink may use standalone wizard screens |
| TUI scope | Embed link/unlink only; do not split the main app file or sync/MR flows |
| Operations | Still spec 011; this spec only removes duplicate interactive paths |
