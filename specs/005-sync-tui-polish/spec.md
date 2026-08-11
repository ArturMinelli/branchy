# Feature Specification: TUI Native Confirm & Loading

**Feature Branch**: `005-sync-tui-polish`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "We need more interactive and TUI native flow in the sync command. We need yes or no steps to use nice TUI confirm actions, and we also need a nicer loading state for long running actions"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Native confirm dialogs across all TUI flows (Priority: P1)

A developer uses any branchy TUI flow that requires a yes/no decision — sync (embedded `s` and standalone `branchy sync`), MR (`branchy mr` / `m`), link, unlink, init, and embedded tree unlink — and sees a proper TUI confirm panel with visually distinct Yes and No options, clear context for the decision, and keyboard hints. The experience feels consistent with polished terminal apps rather than inline `[y/N]` text prompts.

**Why this priority**: Confirm steps are repeated across every interactive command; plain text prompts undermine the TUI-native goal and create visual inconsistency.

**Independent Test**: Run sync, MR, link, unlink, and init TUIs; at each confirm screen verify Yes/No are rendered as a framed TUI control with arrow-key focus navigation (not raw `[y/N]` text), default focus on No, and Enter confirms the focused option.

**Acceptance Scenarios**:

1. **Given** any TUI flow reaches a yes/no decision, **When** the confirm screen renders, **Then** a bordered confirm panel shows the question, relevant context, and highlighted Yes/No options with No focused by default.
2. **Given** the confirm screen is shown, **When** the user presses left/right or tab to move focus and presses Enter, **Then** the focused option (Yes or No) is selected.
3. **Given** the confirm screen is shown, **When** the user presses `y`, **Then** Yes is selected (same outcome as focusing Yes and pressing Enter).
4. **Given** the confirm screen is shown, **When** the user presses `n` or Esc, **Then** No is selected (cancel/decline semantics unchanged per flow).
5. **Given** embedded sync in the main tree and standalone `branchy sync`, **When** confirm screens appear, **Then** they use identical confirm component and styling.
6. **Given** embedded tree unlink (`u` key), **When** the unlink confirm appears, **Then** it uses the same native confirm panel as dedicated command flows.

---

### User Story 2 - Loading feedback during long-running TUI actions (Priority: P1)

While network-bound or browser work runs in any TUI flow — MR creation during sync, MR creation in the MR wizard, browser tab opening, init registration — the user sees an animated loading indicator with a descriptive status message. The UI does not appear frozen on a static line. Input is blocked until the operation completes.

**Why this priority**: GitLab and registration operations can take several seconds; without feedback users assume the app hung.

**Independent Test**: Confirm an edge in sync TUI that triggers MR creation; verify an animated spinner appears with edge context; verify keys are ignored until completion; verify transition to next step is automatic.

**Acceptance Scenarios**:

1. **Given** the user confirmed MR creation for a sync edge, **When** the network call is in progress, **Then** the UI shows an animated loading indicator and message naming the edge being processed.
2. **Given** loading is active, **When** the user presses any key (including Esc, q, y, n), **Then** input is ignored until the operation completes.
3. **Given** loading completes successfully, **When** the operation finishes, **Then** the flow advances automatically to the next step without user input.
4. **Given** loading completes with failure, **When** the operation fails, **Then** loading clears, the error is surfaced per existing flow rules, and the user can continue or exit.
5. **Given** sync processes multiple confirmed edges, **When** each runs, **Then** loading shows current edge progress (e.g. edge N of M).
6. **Given** the user confirms browser-open at end of sync or MR, **When** tabs open in sequence, **Then** a loading state shows progress while URLs are opened.
7. **Given** the user confirms project registration in init TUI, **When** registration runs, **Then** a loading state shows until init completes.

---

### User Story 3 - Visual consistency via shared TUI chrome (Priority: P1)

All branchy TUI flows share the same confirm panel and loading panel components. Headers, help footers, color semantics, and keyboard hints remain consistent with existing chrome (title, muted help, ok/warn/err).

**Why this priority**: A single shared implementation prevents drift and satisfies the "simple and elegant" bar from unified command TUI work.

**Independent Test**: Step through confirm and loading screens in sync, MR, link, unlink, init, and embedded unlink; verify matching frame style, button highlight, spinner placement, and help footer patterns.

**Acceptance Scenarios**:

1. **Given** any two TUI flows at a confirm step, **When** rendered side by side, **Then** confirm panels share the same border, button layout, focus highlight, and help hint format.
2. **Given** any two TUI flows during loading, **When** rendered, **Then** loading panels share the same spinner style and message layout.
3. **Given** a terminal below minimum size, **When** confirm or loading would render, **Then** the existing too-small guard applies without broken layout.

---

### User Story 4 - Scripted CLI paths unchanged (Priority: P2)

A developer or CI runs branchy with flags, positional args, or non-TTY stdin. Execution stays plain-text with existing stdin yes/no prompts; no TUI confirm panels or spinners appear.

**Why this priority**: Scripting safety from 004-command-tui must be preserved.

**Independent Test**: Run `branchy sync --from develop`, `branchy link parent child`, and sync in non-TTY; verify no full-screen TUI and no spinner output.

**Acceptance Scenarios**:

1. **Given** any command flag or non-TTY environment, **When** the command runs, **Then** no native confirm panel or TUI loading state is shown.
2. **Given** scripted sync with per-edge stdin prompts, **When** sync runs, **Then** behavior and output format match pre-feature CLI semantics.

---

### Edge Cases

- What happens when the terminal is too small? Existing minimum-size guard applies; confirm and loading screens degrade inside the flow window wrapper.
- What happens when the user presses keys during loading? All input is ignored until the async operation completes; no cancel, no double-submit.
- What happens when loading takes longer than 30 seconds? Spinner continues with unchanged message; no timeout change to underlying operation semantics.
- What happens when the user cancels at confirm (No/Esc)? Unchanged per-flow semantics; loading state is never entered.
- What happens when MR creation fails during loading? Loading clears; failure recorded per flow rules; user proceeds to next step or error screen.
- What happens when zero edges exist in sync? Empty state unchanged; no confirm or loading screens.
- What happens when init is run with `--force`? Plain CLI only; no TUI confirm or loading.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All interactive TUI flows with yes/no steps — sync (embedded and standalone), MR, link, unlink, init, and embedded tree unlink — MUST replace inline `[y/N]` text prompts with a shared native TUI confirm panel.
- **FR-002**: Native confirm panels MUST show the decision question, relevant context (branches, edge N of M, subtree count, etc.), and visually distinct Yes/No options.
- **FR-003**: Native confirm panels MUST default focus to No (safe default).
- **FR-004**: Native confirm panels MUST support arrow-key (or tab) focus navigation between Yes and No, with Enter confirming the focused option.
- **FR-005**: Native confirm panels MUST also accept direct `y` (Yes) and `n`/Esc (No) shortcuts matching focused selection behavior.
- **FR-006**: Sync TUI MUST show a loading panel with animated indicator while MR creation runs for a confirmed edge, including edge progress (N of M).
- **FR-007**: MR TUI MUST show a loading panel while MR creation runs after confirm.
- **FR-008**: Sync and MR TUI MUST show a loading panel while browser URLs open after browser confirm.
- **FR-009**: Init TUI MUST show a loading panel while project registration runs after confirm.
- **FR-010**: While any loading panel is active, the TUI MUST ignore all keyboard input until the in-flight operation completes.
- **FR-011**: Confirm and loading UI MUST be implemented as shared TUI chrome (single confirm component, single loading component) used by all in-scope flows.
- **FR-012**: All in-scope flows MUST retain existing functional semantics: skip/confirm outcomes, cancellation, summary content, browser batch rules (created-only, DFS order), and exit behavior.
- **FR-013**: Scripted CLI mode (any flag, full positional args, non-TTY) MUST be unaffected; no TUI confirm or loading in plain CLI mode.
- **FR-014**: Visual treatment MUST align with existing branchy chrome: title header, muted help footer, ok/warn/err color semantics.

### Key Entities

- **Confirm Panel**: A framed TUI decision screen with question text, context, Yes/No options, arrow-key focus, default No focus, and y/n/enter/esc bindings.
- **Loading Panel**: A TUI state with animated activity indicator, primary status message, and optional progress context (edge N of M, action description).
- **TUI Decision Point**: Any step in an interactive flow requiring user consent before a mutating or network action.
- **In-Flight Operation**: A network, filesystem, or browser action triggered after confirmation whose duration warrants loading feedback.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of yes/no steps in in-scope TUI flows use the shared native confirm panel (zero inline `[y/N]` prompt strings in TUI views).
- **SC-002**: 100% of network-bound operations in in-scope TUI flows (MR creation, init registration, browser batch open) show an animated loading indicator before advancing.
- **SC-003**: Users can distinguish confirm, loading, and result states by visual structure alone without relying on raw key-hint suffixes.
- **SC-004**: Functional parity: same outcomes for equivalent user choices across all flows before and after this feature.
- **SC-005**: Embedded and standalone instances of the same flow (e.g. sync) render identical confirm and loading treatments.
- **SC-006**: Loading indicator appears within 200ms of confirming an action that triggers a network call.
- **SC-007**: All six in-scope TUI surfaces (sync embedded, sync standalone, MR, link, unlink, init, tree unlink confirm) pass a visual consistency checklist for confirm and loading chrome.

## Assumptions

- "All TUI flows" means every interactive branchy flow that currently has a yes/no confirm step; projects list is out of scope (no yes/no confirm).
- Default confirm focus remains No, matching prior `[y/N]` safe-default convention.
- Long-running actions remain async via Bubble Tea commands; loading wraps existing command patterns.
- Animations are subtle (spinner tick); no heavy decoration per "simple and elegant" quality bar.
- Link and unlink confirms retain their current question text and semantics; only presentation changes.
- No change to CLI stdin yes/no prompts for scripted mode.

## Decisions (Grilling Session 2026-08-11)

| Topic | Decision |
|-------|----------|
| Scope | All TUI flows with yes/no steps: sync (embedded + standalone), MR, link, unlink, init, embedded tree unlink |
| Confirm interaction | Arrow-key focus between styled Yes/No buttons; Enter confirms focused option; y/n shortcuts preserved |
| Cancel during loading | Block all input until operation completes; no cancel during in-flight work |
