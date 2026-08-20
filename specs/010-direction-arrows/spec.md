# Feature Specification: Merge Direction Arrows

**Feature Branch**: `010-direction-arrows`

**Created**: 2026-08-20

**Status**: Draft

**Input**: User description: "We need to improve the way we deal with the merge direction in this cli, it is not very intuitive as of now, and we need to make it clearer in the TUI, where the changes will come from and where they will go (we need to do that in the main TUI, we need to show little arrows in the branches to clearly show direction). Also, lets change the keys, lets use control + arrows instead of the "d" key."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read merge direction from the tree itself (Priority: P1)

A developer opens the main branch tree. Every branch that takes part in a parent/child merge shows a small arrow next to its name: down when changes flow parent → child, up when they flow child → parent. They do not need the footer to know which way a sync would send work. Roots with no children stay unlabeled. File-change counts stay where they are; the arrow is an extra cue, not a replacement.

**Why this priority**: The current mode lives only in the footer, which is why direction feels unintuitive. Visible arrows on the tree are the product change.

**Independent Test**: Open the main tree on a project with at least one root that has children and one childless root. Confirm down-arrows on every participating row in the default (inbound) mode, none on the childless root, and counts still present as today.

**Acceptance Scenarios**:

1. **Given** the main tree is in inbound mode (parent → child), **When** the developer looks at any branch that has a parent or has children, **Then** that row shows a down arrow next to the branch name.
2. **Given** the main tree is in outbound mode (child → parent), **When** the developer looks at those same rows, **Then** each shows an up arrow next to the branch name instead of a down arrow.
3. **Given** a root with no children, **When** either direction is active, **Then** that row shows no direction arrow.
4. **Given** a child whose file-change count is hidden (zero) or unknown, **When** either direction is active, **Then** the direction arrow is still shown; missing or hidden counts do not hide direction.
5. **Given** a selected row, **When** the tree is rendered, **Then** the selection marker and the direction arrow are distinct, so the developer can tell which branch is selected and which way merges flow.

---

### User Story 2 - Set direction with Control+Up and Control+Down (Priority: P1)

A developer wants outbound (child → parent, flow up the tree): they press Control+Up. The tree switches to outbound counts and up-arrows, whether it was already outbound or not. They want inbound (parent → child, flow down): they press Control+Down. Plain Up/Down still move the cursor. There is no toggle key.

**Why this priority**: Replacing a letter toggle with keys that match tree geometry is the other half of making direction intuitive.

**Independent Test**: From inbound, press Control+Up and verify outbound counts and up-arrows; press Control+Up again and verify it stays outbound; press Control+Down and verify inbound counts and down-arrows. Confirm plain Up/Down still only move the cursor.

**Acceptance Scenarios**:

1. **Given** the main tree is inbound, **When** the developer presses Control+Up, **Then** the tree becomes outbound: counts switch to outbound, and participating rows show up-arrows.
2. **Given** the main tree is already outbound, **When** the developer presses Control+Up, **Then** it stays outbound (not a toggle).
3. **Given** the main tree is outbound, **When** the developer presses Control+Down, **Then** the tree becomes inbound: counts switch to inbound, and participating rows show down-arrows.
4. **Given** the main tree is already inbound, **When** the developer presses Control+Down, **Then** it stays inbound.
5. **Given** the main tree has focus, **When** the developer presses plain Up or Down (without Control), **Then** the cursor moves as today and direction does not change.
6. **Given** either direction is active, **When** the developer looks for the old `d` key, **Then** `d` does not change direction (it is no longer a direction control).

---

### User Story 3 - Help names the chords, not `d` (Priority: P1)

A developer who has never used direction controls reads the main-tree help. It names the current direction (inbound parent → child vs outbound child → parent) and lists Control+Up / Control+Down as the way to set it. `d` is not listed as a direction key.

**Why this priority**: Discoverability must move with the new chords, or the old toggle remains the mental model.

**Independent Test**: Open the main tree; read the help before any chord; press Control+Up; confirm the help now names outbound and still lists Control+Up / Control+Down, never `d` as a direction action.

**Acceptance Scenarios**:

1. **Given** the main tree is inbound, **When** the developer reads the help, **Then** it indicates inbound (parent → child) and lists Control+Up as the way to set outbound and Control+Down as the way to set inbound.
2. **Given** the developer has set outbound, **When** they read the help, **Then** it indicates outbound (child → parent) and still lists those same chords.
3. **Given** a first-time user, **When** they look only at the main-tree help, **Then** they can discover how to change direction without reading external docs, and they are not told to press `d`.

---

### User Story 4 - Standalone sync picker uses the same chords (Priority: P2)

A developer runs interactive sync from the shell (no flags) and never opens the main tree. The root picker still starts inbound/downward. Control+Up sets pending upward (outbound counts); Control+Down sets pending downward (inbound counts). The picker is a flat list, so it does not show tree arrows. After a root is chosen, confirms still do not offer a direction change. `d` does not change picker direction.

**Why this priority**: The same unintuitive `d` exists on the picker; matching keys keeps one mental model. Arrows stay a main-tree concern.

**Independent Test**: Launch standalone interactive sync; verify it starts downward; press Control+Up; verify pending direction and badges become outbound; pick a root; verify confirms have no direction controls and no `d`.

**Acceptance Scenarios**:

1. **Given** standalone interactive sync has just opened, **When** the developer reads the root picker, **Then** it is downward, the help names the current direction and Control+Up / Control+Down, and listed rows have no tree-style direction arrows.
2. **Given** the picker is showing, **When** the developer presses Control+Up, **Then** pending direction becomes upward and listed badges switch to outbound counts.
3. **Given** the picker is already upward, **When** the developer presses Control+Up, **Then** it stays upward; Control+Down returns it to downward.
4. **Given** a root has been chosen, **When** edge confirms are showing, **Then** Control+Up / Control+Down and `d` do not change direction.
5. **Given** the picker or a later confirm, **When** the developer presses `d`, **Then** direction does not change.

---

### Edge Cases

- Childless roots show no direction arrow in either mode.
- Roots that have children show the same up/down arrow as other participating rows.
- Leaves (children with no children of their own) still show the arrow; they always have a parent.
- Zero and unknown file-change counts do not hide the arrow.
- Pressing the chord for the already-active direction is a no-op, not a flip.
- Plain Up/Down (and the existing vim navigation keys) never change direction.
- Direction remains session-scoped on the main tree: it survives leaving and returning to the tree (including switching projects) until the user quits; a new session opens inbound with down-arrows.
- Standalone picker direction is for that run only; the next standalone launch starts downward again.
- Sync started from the main tree still uses the tree’s current direction at the moment sync starts; confirm screens do not offer a direction change and do not show tree arrows.
- Scripted / flagged sync stays downward-only and gains no direction keys.
- Other screens (project list, link, unlink, manual merge-request flow) do not gain direction arrows or Control+Up / Control+Down direction controls.
- If the terminal does not deliver Control+Up / Control+Down, direction stays as last set (inbound at a new session); the tree arrows still show that current direction.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The main tree MUST show a down arrow next to the name of every participating branch while inbound (parent → child) is active, and an up arrow while outbound (child → parent) is active.
- **FR-002**: A participating branch is any row that has a parent or has at least one child. Childless roots MUST NOT show a direction arrow.
- **FR-003**: Direction arrows MUST appear on participating rows regardless of whether a file-change count is shown, hidden, or unknown.
- **FR-004**: Direction arrows MUST sit next to the branch name and MUST remain visually distinct from the selection marker and from the file-change count.
- **FR-005**: Control+Up on the main tree MUST set outbound (child → parent). If outbound is already active, the tree MUST stay outbound.
- **FR-006**: Control+Down on the main tree MUST set inbound (parent → child). If inbound is already active, the tree MUST stay inbound.
- **FR-007**: Plain Up and Down MUST continue to move the cursor only and MUST NOT change direction.
- **FR-008**: The `d` key MUST NOT change merge direction on the main tree, the standalone sync picker, or any other screen.
- **FR-009**: The main-tree help MUST name the current direction (inbound parent → child vs outbound child → parent) and MUST list Control+Up and Control+Down as the keys that set it. It MUST NOT list `d` as a direction control.
- **FR-010**: Setting direction on the main tree MUST update arrows and visible counts immediately, without restarting the session or re-opening the project.
- **FR-011**: Main-tree direction MUST remain session-scoped as today: it survives leaving and returning to the tree (including switching projects) until quit; a new session MUST open inbound.
- **FR-012**: Sync started from the main tree MUST still follow the tree’s current direction at the moment sync starts (inbound → downward parent → child, outbound → upward child → parent).
- **FR-013**: The standalone interactive sync root picker MUST use Control+Up to set pending upward/outbound and Control+Down to set pending downward/inbound, with the same set-not-toggle behavior. It MUST NOT show tree direction arrows.
- **FR-014**: After a sync root is chosen (standalone) or sync is started from the main tree, confirm and summary screens MUST NOT change direction in response to Control+Up, Control+Down, or `d`.
- **FR-015**: Scripted / flagged sync MUST remain downward-only and MUST NOT gain direction keys.
- **FR-016**: Project list, link, unlink, and the manual single merge-request flow MUST NOT gain direction arrows or these direction chords.

### Key Entities

- **Merge direction**: Session-wide mode on the main tree (and pending mode on the standalone picker): inbound / downward (parent → child) or outbound / upward (child → parent). Default inbound. Not stored after quit.
- **Direction arrow**: A small up or down mark next to a participating branch name on the main tree. Down means changes flow toward children; up means changes flow toward parents.
- **Participating branch**: A row that has a parent or has children — any branch that can be a source or a destination of a tree merge.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A first-time user looking only at the main tree (not the help line) can correctly state whether merges would flow parent → child or child → parent in under 3 seconds.
- **SC-002**: In a tree with at least one childless root and three participating branches, 100% of participating rows show the matching arrow and 0% of childless roots show an arrow, in both directions.
- **SC-003**: From inbound, one Control+Up makes 100% of participating rows show up-arrows and outbound counts; a second Control+Up leaves that state unchanged; one Control+Down restores inbound counts and down-arrows.
- **SC-004**: 100% of new sessions open inbound with down-arrows; 0% of quits restore outbound on the next launch.
- **SC-005**: After the change, 0% of main-tree or standalone-picker help text names `d` as a direction control, and pressing `d` changes direction in 0% of those screens.
- **SC-006**: A developer using only standalone interactive sync can set upward with Control+Up and start a cascade without opening the main tree, in under 15 seconds of orientation (excluding network waits).
- **SC-007**: With the tree in outbound mode, 100% of main-tree sync runs still create child → parent requests for the selected subtree; with the tree inbound, 100% still create parent → child requests.

## Assumptions

- File-change count meaning, hide-zero rules, unknown placeholders, and session-scoped direction persistence stay as they are; this feature adds arrows and replaces the direction key, it does not change what the numbers mean.
- Default remains inbound (parent → child, down-arrows) because that is still the starting sync direction.
- The direction arrow is a single up or down mark placed immediately before the branch name, after the tree connector, so it reads as part of the row rather than as a second badge after the count.
- Selection continues to use a distinct marker; it is not reused as the direction arrow.
- Standalone picker stays a flat list: same chords and help, no arrows.
- Sync confirm copy already names source → target; this feature does not restyle those confirms.
- Help may abbreviate the chords (for example `ctrl+↑` / `ctrl+↓`) as long as both directions are named and `d` is gone.
- Documentation that currently teaches `d` as the direction key is expected to follow this spec when the feature is implemented.

## Decisions (Grilling Session 2026-08-20)

| Topic              | Decision                                                                                          |
|--------------------|---------------------------------------------------------------------------------------------------|
| Key mapping        | Control+Up sets outbound (child → parent); Control+Down sets inbound (parent → child); not a toggle |
| Arrow placement    | ↑ / ↓ next to every participating branch name on the main tree, including roots that have children |
| Old `d` key        | Removed as a direction control everywhere                                                          |
| Standalone picker  | Same Control+arrow keys; no tree arrows (flat list)                                                |
| Arrow chrome       | Same muted color as the tree connectors                                                            |
| Name alignment     | Pad rows without an arrow so branch names stay in one column                                       |
| Broken terminals   | No fallback chord; arrows still show the current direction                                         |
