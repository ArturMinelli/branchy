# Feature Specification: Unlink Branch from Tree

**Feature Branch**: `003-unlink-branch`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "We need to make it so that we have a unlink action as well, so that the user can remove a branch from the tree"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unlink subtree from TUI (Priority: P1)

A developer selects a branch in the branchy tree view and presses `u` to unlink it. The system shows a confirmation describing which branch will be removed and how many branches are in its subtree (including the selected branch). If the user confirms, the selected branch and all of its descendants are removed from the project's branch tree configuration. If the user cancels, no changes are made and they return to the tree view.

**Why this priority**: This is the primary interaction path and the symmetric counterpart to the existing `l` (link) shortcut; it addresses the core need to correct or prune the tree without manual YAML editing.

**Independent Test**: Open branchy TUI on a project whose tree has a branch with at least two descendants; select the middle branch, press `u`, confirm; verify that branch and its descendants disappear from the tree while sibling branches and ancestors remain.

**Acceptance Scenarios**:

1. **Given** a registered project with a branch tree and a selected branch, **When** the user presses `u`, **Then** the TUI shows a confirmation prompt naming the selected branch and the total number of branches that will be removed from the subtree.
2. **Given** the unlink confirmation is shown, **When** the user confirms, **Then** the selected branch and every descendant are removed from the branch tree configuration and the updated tree is displayed.
3. **Given** the unlink confirmation is shown, **When** the user cancels (e.g. Esc or `n`), **Then** no changes are saved and the user returns to the tree view with the tree unchanged.
4. **Given** the selected branch is a direct child of a parent, **When** the user confirms unlink, **Then** the child is removed from the parent's children list as well as from the branch registry.
5. **Given** the selected branch is a root with no parent edge, **When** the user confirms unlink, **Then** the root and its entire subtree are removed while other independent roots in the same tree remain.

---

### User Story 2 - Unlink subtree via CLI (Priority: P1)

A developer runs `branchy unlink <parent> <child>` to remove a subtree rooted at `child` from the branch tree. The command validates that the parent→child edge exists, removes `child` and all descendants from the configuration, and persists the updated tree. This mirrors the existing `branchy link <parent> <child>` command for symmetry.

**Why this priority**: CLI parity ensures users and scripts can manage the tree without the TUI, consistent with how link already works.

**Independent Test**: Run `branchy unlink develop feature-a` in a registered project where `feature-a` has children; verify `feature-a` and its descendants are gone, `develop` no longer lists `feature-a` as a child, and other branches are untouched.

**Acceptance Scenarios**:

1. **Given** a valid parent→child edge in the branch tree, **When** the user runs `branchy unlink <parent> <child>`, **Then** `child` and its entire subtree are removed and a success message is printed.
2. **Given** the parent→child edge does not exist, **When** the user runs `branchy unlink <parent> <child>`, **Then** the command fails with a clear error and exit code 1 without modifying the tree.
3. **Given** `child` is not in the tree, **When** the user runs `branchy unlink <parent> <child>`, **Then** the command fails with a clear error and exit code 1.
4. **Given** a successful unlink, **When** the user opens the TUI or re-runs tree commands, **Then** the removed branches no longer appear.

---

### User Story 3 - Discoverability in TUI help (Priority: P2)

A developer viewing the branch tree sees `u` documented alongside existing shortcuts (`s`, `m`, `l`) so they can discover the unlink action without reading external docs.

**Why this priority**: Low effort, prevents the feature from being hidden despite being implemented.

**Independent Test**: Open branchy TUI; verify the footer help line includes an unlink key binding.

**Acceptance Scenarios**:

1. **Given** the user is on the tree view screen, **When** they read the help footer, **Then** `u` is listed with a label indicating unlink/remove.

---

### Edge Cases

- What happens when the user selects a leaf branch (no children)? Only that single branch is removed; confirmation shows count of 1.
- What happens when unlink would remove every branch in the tree? The operation is allowed; the tree becomes empty and the TUI shows the existing empty-tree state.
- What happens when the selected branch name is invalid or missing from the tree? Unlink is not offered or fails with a clear message; no partial writes occur.
- What happens when saving the updated tree fails after confirmation? The user sees an error message and the in-memory tree is not left in an inconsistent state relative to disk.
- What happens when `parent` and `child` are the same in CLI? The command fails with a validation error, consistent with link behavior.
- What happens when `child` has descendants but the edge `parent → child` exists? The entire subtree under `child` is removed, not just the single edge.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The TUI MUST provide an unlink action bound to `u` on the tree view screen when a branch is selected.
- **FR-002**: Before persisting any unlink, the TUI MUST show a confirmation prompt that names the selected branch and states how many branches (selected + descendants) will be removed.
- **FR-003**: On confirmed unlink in the TUI, the system MUST remove the selected branch and all of its descendants from the branch tree configuration.
- **FR-004**: On confirmed unlink, if the selected branch has a parent, the system MUST remove the selected branch from that parent's children list.
- **FR-005**: On cancelled unlink in the TUI, the system MUST make no changes to the branch tree configuration.
- **FR-006**: The CLI MUST support `branchy unlink <parent> <child>` that removes `child` and its entire subtree when the parent→child edge exists.
- **FR-007**: The CLI unlink command MUST validate that `parent` and `child` differ, that both names are non-empty, and that the edge exists before making changes.
- **FR-008**: The CLI unlink command MUST persist the updated tree and print a success message naming the removed root of the subtree (`child`).
- **FR-009**: Failed unlink operations (validation, missing edge, save errors) MUST produce clear error messages and MUST NOT leave a partially updated tree on disk.
- **FR-010**: The tree view help text MUST document the unlink key binding.

### Key Entities

- **Branch Tree**: The project's hierarchical branch configuration; each branch is a node that may have child branches. Unlink removes a subtree rooted at a chosen branch.
- **Subtree**: The selected branch plus all descendants reachable via parent→child edges below it; the unit of removal for this feature.
- **Parent→Child Edge**: The link between a parent branch and its direct child; must be removed when unlinking a non-root child, and must exist for CLI unlink validation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can remove a branch subtree from the TUI in under 10 seconds (select, press `u`, confirm) without editing YAML manually.
- **SC-002**: 100% of confirmed unlink operations result in the selected branch and all descendants absent from the next tree load.
- **SC-003**: CLI unlink completes with correct tree state for valid inputs; invalid inputs fail with a descriptive error and non-zero exit code in 100% of tested cases.
- **SC-004**: Users can discover the unlink action from the TUI help line without consulting external documentation.

## Assumptions

- Each branch has at most one parent (tree structure, not a DAG); unlink always targets a single parent→child relationship for CLI, and the selected branch's direct parent for TUI cleanup.
- Unlink affects only the local branch tree configuration (`branch-tree.yaml`); it does not delete Git branches or close GitLab merge requests.
- Subtree removal is intentional and destructive for configuration purposes; a confirmation step in the TUI is sufficient guardrail (CLI follows the same semantics without an interactive prompt, consistent with `branchy link`).
- An empty branch tree after unlink is valid; no minimum branch count is enforced.
- The unlink key `u` does not conflict with existing tree-view bindings.
