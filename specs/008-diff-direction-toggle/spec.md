# Feature Specification: Diff Direction Toggle

**Feature Branch**: `008-diff-direction-toggle`

**Created**: 2026-08-12

**Status**: Draft

**Input**: User description: "As of today, the branchy TUI displays a change count from parent to child, however i think it would be interesting to have some sort of toggle to change the direction of the change count diffs, so then instead of counting the diffs from parent to child, we would count from child to parent."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Flip the main tree to outbound counts (Priority: P1)

A developer is looking at the main branch tree. By default each child still shows how many file changes it would receive from its parent (inbound). They press the direction toggle once. Every child row now shows how many file changes the parent would receive from that child (outbound) — the same files-changed number a merge request from child into parent would show. Pressing the toggle again restores inbound counts. Roots still show no count. Zeros stay hidden. Unknown comparisons still show a clear placeholder.

**Why this priority**: Seeing the reverse of today’s inbound number is the whole feature; without this flip the rest of the work has no user value.

**Independent Test**: Open the main tree on a project whose children have known, different inbound and outbound file-change counts (including at least one child that is inbound-only and one that is outbound-only). Verify default inbound badges, then toggle and verify outbound badges, then toggle back.

**Acceptance Scenarios**:

1. **Given** the main tree is showing inbound counts, **When** the developer activates the direction toggle, **Then** each non-root child that has a non-zero outbound file-change count shows that outbound number instead of the inbound number.
2. **Given** a child that would receive N file changes from its parent and would send M file changes to its parent, **When** the developer toggles direction, **Then** the row shows M (not N) while outbound is active, and N again after a second toggle.
3. **Given** a child with zero outbound file changes (nothing unique to send to its parent), **When** outbound is active, **Then** that row shows no count (the zero is hidden), even if inbound for that child is non-zero.
4. **Given** a root branch with no parent, **When** either direction is active, **Then** that row still shows no change count.
5. **Given** a child that cannot be compared to its parent, **When** either direction is active, **Then** that row still shows the unknown placeholder, not a fabricated zero.

---

### User Story 2 - Tell which direction is active (Priority: P1)

A developer glancing at the main tree can tell whether the numbers mean inbound (parent → child) or outbound (child → parent) without guessing. The help footer names the current direction and the key that flips it, so the toggle is discoverable and the mode is obvious after a flip.

**Why this priority**: The same digit can mean opposite things; without a visible mode cue, toggling is unsafe.

**Independent Test**: Open the main tree; read the footer before any toggle; flip direction; confirm the footer now names the other direction and still names the toggle key.

**Acceptance Scenarios**:

1. **Given** the main tree is in the default inbound mode, **When** the developer reads the help footer, **Then** it indicates that counts are inbound (parent → child) and lists the toggle key.
2. **Given** the developer has flipped to outbound, **When** they read the help footer, **Then** it indicates that counts are outbound (child → parent) and still lists the toggle key.
3. **Given** the developer has never used the toggle before, **When** they look at the main-tree footer, **Then** they can discover how to change direction without leaving the TUI or reading external docs.

---

### User Story 3 - Direction lasts for this session only (Priority: P2)

A developer flips to outbound, leaves the tree (project list, sync, another flow), and comes back in the same session: outbound is still active. They quit branchy and open it again: counts are inbound once more. The choice is never stored as a lasting preference.

**Why this priority**: Session scope keeps the default glanceable (inbound = what sync will do) while still letting someone inspect outbound without fighting the tool every screen change.

**Independent Test**: Toggle outbound, open sync or the project list and return to the tree (mode still outbound); quit and relaunch (mode inbound again).

**Acceptance Scenarios**:

1. **Given** outbound is active on the main tree, **When** the developer starts sync, finishes or cancels, and returns to the tree in the same session, **Then** the tree is still in outbound mode.
2. **Given** outbound is active, **When** the developer switches project within the same session and opens that project’s tree, **Then** that tree also shows outbound counts.
3. **Given** outbound was active when the developer quit, **When** they start a new session, **Then** the main tree opens in inbound mode.

---

### User Story 4 - Sync still shows inbound (Priority: P2)

A developer who has flipped the main tree to outbound starts sync. The sync root picker and each edge confirm still show inbound file-change counts — how many files the child would receive if the parent were synced into it — because sync still creates parent → child merge requests. The tree’s direction toggle does not exist on those screens and does not change those numbers.

**Why this priority**: Mixing outbound numbers into a parent → child confirm would contradict the action the user is about to take.

**Independent Test**: Toggle the main tree to outbound, start sync, and verify picker badges and confirm lines still match inbound (parent → child) counts, including hide-zero on the picker and known zero on confirm.

**Acceptance Scenarios**:

1. **Given** the main tree is in outbound mode, **When** the developer opens the sync root picker, **Then** listed non-root branches show inbound counts (same hide-zero / unknown rules as today), not outbound counts.
2. **Given** the main tree is in outbound mode, **When** sync shows an edge confirm, **Then** the confirm still reports how many files would change on the child (inbound), including a known zero.
3. **Given** a parent→child pair visible on both the outbound tree and the sync picker, **When** inbound and outbound sizes differ, **Then** the picker shows the inbound number and the tree shows the outbound number.

---

### Edge Cases

- Root branches have no parent — no count in either direction.
- Zero in the active direction is hidden on the tree; a row can therefore appear “quiet” in one direction and numbered in the other.
- Unknown / incomparable pairs stay unknown in both directions; toggling does not invent a number.
- Diverged branches: outbound is only the child’s unique file changes toward the parent, not the parent’s unique changes toward the child.
- Very large outbound counts stay readable in a narrow terminal, same compactness as today’s inbound badges.
- If tree counts refresh in place during the session (for example after a background update), they refresh in the currently selected direction.
- Scripted / non-TUI paths stay count-free; this feature does not add a CLI flag for direction.
- Sync results summary, MR / link / unlink pickers, and other non-tree screens do not gain the toggle or outbound badges.
- Multiple parents — not applicable; the tree is still a single-parent hierarchy.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The main tree MUST start each session in inbound mode (file changes the child would receive from its parent).
- **FR-002**: From the main tree, the user MUST be able to flip count direction with a single keystroke, switching all child-row counts between inbound and outbound.
- **FR-003**: While outbound is active, each non-root row MUST show how many file changes the parent would receive from that child — the same files-changed number a merge request from child to parent would show — not commit count and not a two-way combined total.
- **FR-004**: Outbound display on the main tree MUST reuse inbound’s visibility rules: hide zero, no count on roots, unknown placeholder when comparison fails, compact badge next to the branch name.
- **FR-005**: The main-tree help footer MUST name the current direction (inbound vs outbound) and MUST list the key that toggles it, in both modes.
- **FR-006**: Count direction MUST be session-scoped: it survives leaving and returning to the tree (including switching projects) until the user quits; a new session MUST open inbound.
- **FR-007**: The sync root picker and sync edge confirm MUST continue to show inbound counts only, regardless of the main tree’s current direction. Those screens MUST NOT offer the direction toggle.
- **FR-008**: For a given parent/child pair and comparable local state, the outbound number on the tree MUST match the files-changed number that same pair would show on a child → parent merge request.
- **FR-009**: Toggling direction MUST update visible tree counts immediately; the user MUST NOT need to restart the TUI or re-open the project.
- **FR-010**: Scripted/plain CLI behavior MUST remain unchanged.

### Key Entities

- **Inbound file-change count**: How many files would change on the child if the parent were merged into it — today’s default, matching a parent → child merge request.
- **Outbound file-change count**: How many files would change on the parent if the child were merged into it — matching a child → parent merge request.
- **Count direction**: Session-wide mode on the main tree, either inbound or outbound. Default inbound. Not stored after quit. Does not apply to sync.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a project with at least three children whose outbound sizes differ, a user can switch to outbound and identify the largest outbound child from the main tree alone in under 10 seconds.
- **SC-002**: After one toggle, 100% of comparable non-root rows show the outbound count (or hide it when that outbound count is zero); after a second toggle, 100% of those rows match the inbound counts shown at session start.
- **SC-003**: For a given parent/child pair, the outbound number on the tree matches the files-changed number a child → parent merge request would show.
- **SC-004**: 100% of new sessions open with inbound counts; 0% of quits restore outbound on the next launch.
- **SC-005**: With the tree in outbound mode, 100% of sync picker rows and edge confirms still show inbound counts (same hide-zero / show-zero-on-confirm rules as today).
- **SC-006**: A first-time user can tell which direction is active, and how to flip it, from the main-tree footer in under 2 seconds.

## Assumptions

- The unit of “diffs” stays files changed, consistent with the existing inbound-count feature; only the direction flips.
- Default remains inbound because that is what parent → child sync will apply.
- One session-wide direction is enough; there is no per-project remembered mode.
- The exact toggle key is chosen at planning time from an unused main-tree key; the spec only requires that it is a single keystroke and is listed in the footer.
- Badge styling stays the same as inbound (muted number, warning placeholder); direction is communicated in the footer, not by a second glyph on every row.
- Existing inbound behavior on the tree (when inbound is active) does not change.

## Decisions (Grilling Session 2026-08-12)

| Topic              | Decision                                                                 |
|--------------------|--------------------------------------------------------------------------|
| Surfaces           | Main tree only; sync picker and edge confirm stay inbound                |
| Reverse meaning    | Files the parent would receive from the child (same files-changed metric) |
| Persistence        | This session only; always start inbound                                  |
| Mode cue           | Help footer shows current direction and the toggle key                   |
