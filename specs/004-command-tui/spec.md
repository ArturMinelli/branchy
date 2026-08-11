# Feature Specification: Unified Command TUI

**Feature Branch**: `004-command-tui`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "We need also a very nice TUI for the `branchy sync` command, in fact all commands and subcommand should be in TUI format when possible. The UI/UX of the commands have to be simple and elegant"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Standalone sync TUI from CLI (Priority: P1)

A developer runs `branchy sync` without flags from a registered project directory. Instead of numbered text prompts and line-based yes/no questions, branchy launches a dedicated full-screen sync TUI (focused wizard, then exit) that guides them through root-branch selection, per-edge MR confirmation, a results summary, and an optional browser-open step — matching the semantics of the embedded sync flow triggered by `s` in the main tree view.

**Why this priority**: Sync is the most complex interactive workflow and the explicit ask; plain-text CLI is the biggest UX gap today.

**Independent Test**: Run `branchy sync` in a project with a multi-edge tree; complete the flow entirely via keyboard in the TUI; verify MRs are created/skipped per user choices and the summary matches outcomes.

**Acceptance Scenarios**:

1. **Given** a registered project and no flags, **When** the user runs `branchy sync`, **Then** a dedicated sync TUI opens with a searchable/selectable list of root branches (not a numbered text menu).
2. **Given** a root branch is chosen in the sync TUI, **When** sync begins, **Then** the TUI walks each parent→child edge with clear per-edge confirmation, progress indication, and the same semantics as the embedded tree sync flow.
3. **Given** all edges are processed, **When** sync completes, **Then** the TUI shows a readable per-edge summary (created, skipped, failed) with URLs where applicable.
4. **Given** one or more MRs were created, **When** the summary is shown, **Then** the user is prompted whether to open created MRs in the browser (created-only, tree depth-first order).
5. **Given** the user cancels mid-flow (e.g. Esc), **When** cancellation is accepted, **Then** the TUI exits cleanly with partial results preserved in the summary where applicable.
6. **Given** any flag is passed (e.g. `--from`, `-y`), **When** the user runs `branchy sync`, **Then** plain CLI mode is used with no full-screen TUI.

---

### User Story 2 - TUI for projects, link, unlink, and init (Priority: P1)

A developer runs `branchy projects`, `branchy link`, `branchy unlink`, or `branchy init` without flags or positional arguments (where applicable). Each launches a dedicated, simple, elegant TUI flow — list pickers, confirmations, and clear status messages — instead of unstructured terminal output.

**Why this priority**: User confirmed all interactive commands including init are in scope; dedicated mini-flows match the existing `branchy mr` pattern.

**Independent Test**: Run each command in interactive mode; verify keyboard-driven flows complete successfully with visually consistent layout, help hints, and error presentation.

**Acceptance Scenarios**:

1. **Given** multiple registered projects, **When** the user runs `branchy projects`, **Then** a TUI list shows project id and path with keyboard navigation (not tab-separated plain text).
2. **Given** a registered project, **When** the user runs `branchy link` without positional args, **Then** a TUI guides parent branch selection, child name entry, and confirmation before persisting the edge.
3. **Given** a registered project, **When** the user runs `branchy unlink` without positional args, **Then** a TUI guides branch/subtree selection and shows a confirmation with subtree size before removal.
4. **Given** an unregistered git repo, **When** the user runs `branchy init` without flags, **Then** a TUI wizard confirms registration, shows import/scaffold outcome, and presents a clear success state.
5. **Given** any flag is passed to these commands (e.g. `init --force`), **When** the command runs, **Then** plain CLI mode is used with no full-screen TUI.
6. **Given** positional args are provided for link/unlink (e.g. `branchy link parent child`), **When** the command runs, **Then** it executes immediately with plain one-line output (current scripted behavior).

---

### User Story 3 - Scriptable non-TUI paths preserved (Priority: P1)

A developer or CI pipeline runs branchy with any flag or with fully specified positional arguments. Execution stays plain-text: predictable output, exit codes, and no full-screen TUI takeover.

**Why this priority**: Any-flag rule locks scripting safety; TUI is the default only for fully interactive invocations.

**Independent Test**: Run `branchy sync --from develop -y`, `branchy init --force`, `branchy link parent child`, `branchy unlink parent child`, and `branchy mr --source A --target B -y`; verify no TUI launches and behavior matches current CLI semantics.

**Acceptance Scenarios**:

1. **Given** any command flag is present, **When** the command runs, **Then** no full-screen TUI is shown regardless of which flags are used.
2. **Given** positional args fully specify link or unlink, **When** the command runs, **Then** no TUI is shown and a one-line success or error message is printed.
3. **Given** a scripted sync with flags, **When** sync completes, **Then** results print as plain text with the same end browser prompt rules as today.
4. **Given** stdin is not a TTY (CI, pipes), **When** an interactive command runs, **Then** plain CLI mode is used automatically and the process does not hang.

---

### User Story 4 - Visual consistency and simplicity (Priority: P2)

Across all branchy TUIs (main tree, sync, MR, projects, link, unlink, init), the user experiences a unified visual language: consistent typography emphasis, muted help text, color semantics for success/warning/error, keyboard hints, and minimal screen clutter.

**Why this priority**: "Simple and elegant" is an explicit quality bar; inconsistency undermines the goal.

**Independent Test**: Open main TUI sync (`s`), standalone `branchy sync`, `branchy mr`, and `branchy projects`; verify shared patterns for headers, help footer, confirm dialogs, and result screens.

**Acceptance Scenarios**:

1. **Given** any branchy TUI screen, **When** rendered in a standard terminal (≥80×24), **Then** primary action, context title, and keyboard help are visible without horizontal scrolling.
2. **Given** a confirmation step, **When** displayed, **Then** it uses the same yes/no key bindings and visual treatment as existing sync and MR flows.
3. **Given** a multi-step dedicated flow, **When** the user completes or quits, **Then** the terminal returns to the shell (flow does not leave the user inside the main tree app unless they ran `branchy` with no subcommand).

---

### Edge Cases

- What happens when the terminal is too small? Show a clear minimum-size message rather than a broken layout.
- What happens when not in a TTY? Non-interactive plain CLI is used automatically; no TUI launch.
- What happens when GitLab auth is missing during sync TUI? Flow stops with a clear auth error; user can exit without corrupting state.
- What happens when the branch tree is empty during `branchy sync`? Informative empty state in TUI; no crash.
- What happens when the user runs a command outside a registered project (except `init`)? Clear error in TUI or plain CLI matching entry mode.
- What happens when `branchy` is run with no subcommand? Main tree TUI continues to work unchanged.
- What happens when `branchy mr` is run with any single flag? Plain CLI mode per the any-flag rule (consistent with other commands).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `branchy sync` without flags MUST launch a dedicated full-screen sync TUI (focused wizard using the same sync semantics as the embedded tree flow, then exit to shell).
- **FR-002**: Standalone sync TUI MUST provide equivalent functional semantics to embedded tree sync: per-edge confirm, summary, optional browser batch for created MRs only, depth-first order.
- **FR-003**: Root branch selection in standalone sync TUI MUST use an interactive picker (not numbered stdin prompts).
- **FR-004**: `branchy projects`, `branchy link`, `branchy unlink`, and `branchy init` without flags or positional args (where args would fully specify the operation) MUST launch dedicated TUI flows.
- **FR-005**: Any command flag on any subcommand MUST disable TUI and use plain CLI output.
- **FR-006**: Link and unlink with explicit positional arguments MUST execute as one-shot plain CLI without TUI.
- **FR-007**: System MUST detect non-TTY environments and use plain CLI output instead of TUI.
- **FR-008**: All TUI flows MUST support keyboard-only operation with documented key hints on each screen.
- **FR-009**: All TUI flows MUST present errors in a user-visible, non-fatal way where recovery is possible (back, retry, or quit).
- **FR-010**: Visual presentation across all TUI flows MUST share consistent title, help, success, warning, and error styling.
- **FR-011**: `branchy mr` interactive TUI (no flags) MUST remain available; any-flag rule applies to `mr` as well.
- **FR-012**: Dedicated command TUIs MUST exit to the shell on completion or quit; they MUST NOT silently drop the user into the main tree application.
- **FR-013**: TUI flows MUST exit with appropriate process exit codes on failure (matching CLI conventions).
- **FR-014**: Minimum terminal dimensions MUST be enforced or gracefully degraded with a readable message.

### Key Entities

- **Dedicated Command Flow**: A focused multi-step TUI session for one subcommand (sync, projects, link, unlink, init, mr) that runs independently and returns to the shell — the standard pattern for all interactive subcommands.
- **Interactive Mode**: Invocation with no flags, required positional args not fully provided, and a TTY — triggers dedicated TUI.
- **Scripted Mode**: Invocation with any flag, fully specified positional args for link/unlink, or non-TTY — plain text I/O only.
- **Shared TUI Chrome**: Common visual and interaction patterns (header, help bar, confirm dialog, result panel) reused across all flows.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can complete `branchy sync` interactively entirely via keyboard with zero free-text stdin prompts.
- **SC-002**: 100% of invocations with any flag avoid launching full-screen TUI.
- **SC-003**: Interactive standalone sync TUI produces equivalent outcomes to embedded tree sync for the same tree and user choices.
- **SC-004**: First-time users complete the standalone sync TUI flow in under 3 minutes for trees with up to 10 edges (excluding network latency).
- **SC-005**: All dedicated command TUIs (sync, projects, link, unlink, init) use the same confirm-dialog interaction pattern as existing MR/sync flows.
- **SC-006**: Non-TTY and CI runs never hang waiting for TUI input.
- **SC-007**: Users can complete init, projects browse, and interactive link/unlink flows via keyboard only without typing branch names at a raw prompt (pickers and guided entry instead).

## Assumptions

- Embedded sync in the main tree (`s` key) remains unchanged; standalone `branchy sync` reuses the same sync flow semantics in a dedicated `RunSync`-style entry point.
- `branchy mr` without flags already meets the dedicated-flow pattern; this feature extends that pattern to sync, projects, link, unlink, and init.
- The any-flag rule applies uniformly: even a single optional flag (e.g. `sync --from develop`) forces plain CLI for scriptability.
- Link/unlink with explicit `<parent> <child>` positional args remain one-shot CLI.
- Projects TUI is a browsable list; selecting a project may show details but need not launch the main tree app (user runs `branchy` separately for tree work).
- "Simple and elegant" means minimal steps, clear labels, consistent keys — not animations or heavy decoration.

## Decisions (Grilling Session 2026-08-11)

| Topic | Decision |
|-------|----------|
| Command scope | All interactive commands including init wizard |
| TUI entry pattern | Dedicated mini-flow per command (like `branchy mr`), exit to shell |
| Scriptable escape hatch | Any flag present → plain CLI; positional args for link/unlink → plain CLI |
