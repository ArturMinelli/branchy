# Feature Specification: Manual MR Command

**Feature Branch**: `001-mr-command`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "We need to create a new mr command to this repo, so that users can create an mr explicitly and manually between 2 branches, and then be redirected to browser if they choose to. We need a nice TUI UI/UX for the branch picker"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create MR via TUI branch picker (Priority: P1)

A developer wants to open a merge request between any two branches in the project's branch tree without walking the parent→child sync flow. They run `branchy mr` (or press `m` from the main tree view) and are guided through a polished TUI to pick a source branch, a target branch, optionally edit the MR title, confirm creation, and optionally open the result in their browser.

**Why this priority**: This is the core value of the feature — explicit, manual MR creation with a clear, pleasant branch-selection experience.

**Independent Test**: Run `branchy mr` in a registered project with at least two branches in the tree; complete source/target selection, confirm, and verify an MR is created on GitLab with the chosen branches and title.

**Acceptance Scenarios**:

1. **Given** a registered project with multiple branches in the branch tree, **When** the user runs `branchy mr`, **Then** a TUI opens showing all branches from the branch tree for selection.
2. **Given** the user has selected a source branch, **When** they proceed to target selection, **Then** the TUI shows remaining tree branches (excluding the source) and clearly labels source vs target.
3. **Given** source and target are selected, **When** the user reaches the confirm step, **Then** a default title is pre-filled (e.g. `MR: <source> → <target>`) and the user can edit it before confirming.
4. **Given** the user confirms MR creation, **When** creation succeeds, **Then** the TUI shows the MR URL and prompts "Open in browser? [y/N]".
5. **Given** the user answers yes to the browser prompt, **When** the system browser is available, **Then** the MR page opens in the default browser.
6. **Given** the user answers no to the browser prompt, **When** the flow completes, **Then** the MR URL remains visible in the TUI summary and the command exits successfully.

---

### User Story 2 - Create MR via CLI flags (Priority: P2)

A developer who already knows both branches wants to create an MR without navigating the TUI picker. They pass `--source` and `--target` flags (and optionally `--title` and `--yes`) to create the MR directly from the terminal.

**Why this priority**: Power users and scripting/automation need a fast non-interactive path; flags reuse the same MR creation logic as the TUI flow.

**Independent Test**: Run `branchy mr --source feature-x --target develop --yes` and verify the MR is created without TUI interaction.

**Acceptance Scenarios**:

1. **Given** valid `--source` and `--target` branch names in the tree, **When** the user runs `branchy mr --source A --target B --yes`, **Then** an MR is created without opening the TUI picker.
2. **Given** `--title` is provided, **When** the MR is created, **Then** the supplied title is used instead of the auto-generated default.
3. **Given** `--yes` is not provided in flag mode, **When** the user runs `branchy mr --source A --target B`, **Then** the CLI prompts for confirmation before creating the MR.
4. **Given** `--source` equals `--target`, **When** the command runs, **Then** it fails with a clear error message and does not create an MR.

---

### User Story 3 - Launch MR flow from main TUI (Priority: P2)

A developer is already browsing the branch tree in the default `branchy` TUI and wants to create a manual MR without leaving the app. They press `m` (or equivalent shortcut) to enter the MR flow, with the currently selected tree branch pre-filled as the source branch.

**Why this priority**: Keeps the manual MR workflow discoverable and integrated with the existing tree-navigation experience.

**Independent Test**: Open `branchy` TUI, select a branch, press `m`, and complete the MR flow through target selection and confirmation.

**Acceptance Scenarios**:

1. **Given** the user is on the tree view with a branch selected, **When** they press `m`, **Then** the MR TUI opens with the selected branch pre-set as source.
2. **Given** no branch is selected in the tree view, **When** the user presses `m`, **Then** the MR flow starts at source-branch selection (same as `branchy mr`).
3. **Given** the user is in the MR flow launched from the main TUI, **When** they cancel (e.g. Esc), **Then** they return to the tree view without side effects.

---

### Edge Cases

- What happens when source or target branch is not in the branch tree? The command fails with a clear error listing valid tree branches.
- What happens when an open MR already exists for the same source→target pair? The flow reports the existing MR URL, marks the action as skipped (not failed), and still offers the browser prompt for the existing URL.
- What happens when GitLab credentials are not configured? The command fails early with a message directing the user to authenticate (consistent with existing sync behavior).
- What happens when the branch tree has fewer than two branches? The TUI shows an informative message and does not proceed to MR creation.
- What happens when the browser cannot be opened? The MR is still created successfully; the user sees the URL in the TUI/CLI output and a non-fatal warning that the browser could not be opened.
- What happens when MR creation fails on GitLab (permissions, protected branch, etc.)? The user sees a clear error with the failure reason; no browser prompt is shown.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a `branchy mr` subcommand for manual merge request creation between two branches.
- **FR-002**: System MUST list only branches defined in the project's branch tree (`branch-tree.yaml`) in the TUI branch picker.
- **FR-003**: System MUST present a TUI flow for selecting source branch, target branch, editing title, confirming creation, and optionally opening the browser.
- **FR-004**: System MUST pre-fill a default MR title when the user reaches the confirm step; the user MUST be able to edit the title before confirming in TUI mode.
- **FR-005**: System MUST prompt the user "Open in browser?" after successful MR creation (or when an existing open MR is found); browser opens only on explicit user confirmation.
- **FR-006**: System MUST support `--source`, `--target`, optional `--title`, and `--yes` flags to create MRs without the TUI picker.
- **FR-007**: System MUST reject creation when source and target are the same branch, with a clear error message.
- **FR-008**: System MUST reject branches not present in the branch tree, with a clear error message.
- **FR-009**: System MUST detect existing open MRs for the same source→target pair and skip creation, surfacing the existing MR URL instead.
- **FR-010**: System MUST integrate the MR flow into the main branchy TUI via a keyboard shortcut (`m`), pre-filling source from the currently selected tree branch when available.
- **FR-011**: System MUST allow the user to cancel the MR TUI flow and return to the previous screen without creating an MR.
- **FR-012**: System MUST display the MR URL in the success summary regardless of whether the browser was opened.
- **FR-013**: MR description MUST be auto-generated (timestamped, similar to sync) and is not user-editable in v1.

### Key Entities

- **Merge Request**: A GitLab merge request defined by source branch, target branch, title, and description; identified by a web URL.
- **Branch Tree Branch**: A branch name registered in the project's `branch-tree.yaml`; the sole source of selectable branches for this feature.
- **MR Request**: The user's in-progress or submitted intent to create an MR, comprising source, target, title, and browser-open preference.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a manual MR between two tree branches via the TUI in under 60 seconds (excluding GitLab network latency).
- **SC-002**: 100% of successful MR creations display the MR URL before the command exits.
- **SC-003**: Browser opens only when the user explicitly confirms; zero automatic browser opens in the MR flow.
- **SC-004**: Flag-based MR creation (`--source`, `--target`, `--yes`) completes without TUI interaction in a single command invocation.
- **SC-005**: Users launching from the main TUI (`m` shortcut) reach target selection in one fewer step when a branch is already selected.
- **SC-006**: All error cases (invalid branch, same source/target, auth failure, GitLab rejection) produce a human-readable message without a stack trace in normal usage.

## Assumptions

- Merge requests are created through the same GitLab integration already used by the existing sync command.
- Branch direction follows existing sync convention: first selected branch is source, second is target.
- MR description is auto-generated and not editable in v1; only title is user-editable in TUI mode.
- The MR TUI matches the visual style and interaction patterns of the existing branchy TUI for consistency.
- Non-interactive flag mode does not prompt for browser open; browser open in flag mode is out of scope for v1.
- Remote-only branches not in the branch tree are excluded from the picker by design.
