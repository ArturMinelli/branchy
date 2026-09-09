# Feature Specification: Manual Remote Count Reload

**Feature Branch**: `016-tui-reload-counts`

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "We need to add a new r command in the TUI so that we reload the counts from remote, and we get the correct amount of file diffs between the branches"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Refresh stale counts on the main tree (Priority: P1)

A developer is viewing the main branch tree and notices that file-change counts look wrong or outdated — for example after a teammate pushed to the default remote, or after creating merge requests outside branchy. They press `r`. Branchy updates from the default remote in the background, then recomputes and refreshes the visible file-change counts in place. The active direction (inbound or outbound) stays the same. Navigation and other keys keep working while the update runs.

**Why this priority**: The main tree is the default workspace; on-demand refresh is the core reason for this feature.

**Independent Test**: Open the main tree with known stale counts (local refs behind the remote). Press `r`. After the update completes, verify counts match what GitLab would show for the same parent→child or child→parent pairs, without restarting branchy.

**Acceptance Scenarios**:

1. **Given** the main tree is showing file-change counts for the current project, **When** the developer presses `r`, **Then** branchy starts a default-remote update and, on success, refreshes inbound and outbound counts on every visible child row using the same display rules as today (hide zero, `?` for unknown, arrow matches active direction).
2. **Given** outbound direction is active on the main tree, **When** a manual reload completes successfully, **Then** outbound badges update in place and inbound direction mode is unchanged.
3. **Given** a manual reload is in progress, **When** the developer navigates with ↑/↓, toggles direction, or starts sync, **Then** those actions are not blocked.

---

### User Story 2 - Refresh counts during sync (Priority: P1)

A developer is in the sync flow (root picker or edge confirm) and wants up-to-date file-change counts before choosing a root or accepting an edge. They press `r`. The same default-remote update runs; when it finishes, the sync picker badges and the current edge confirm line refresh in place with recomputed counts.

**Why this priority**: Sync is where counts directly inform yes/no decisions; stale numbers there are as harmful as on the tree.

**Independent Test**: Open embedded sync with stale counts on the picker or confirm. Press `r`. Verify counts update in place after the remote update without leaving the flow.

**Acceptance Scenarios**:

1. **Given** the sync root picker is visible, **When** the developer presses `r` and the update succeeds, **Then** listed branches show refreshed inbound file-change badges (same hide-zero / unknown rules as today).
2. **Given** an edge confirm is on screen, **When** the developer presses `r` and the update succeeds, **Then** the “files would change” context line updates to the new count (including a newly known zero).
3. **Given** sync is showing counts for a parent→child edge, **When** manual reload completes, **Then** the confirm still shows inbound (parent→child) size regardless of the main tree’s direction mode.

---

### User Story 3 - Stay usable when reload fails (Priority: P2)

A developer presses `r` while offline or when the default remote is unreachable. The TUI does not stall. Existing counts stay on screen. No error banner appears for the failed reload (consistent with automatic background fetch).

**Why this priority**: Manual reload is an enhancement; local workflow must remain trustworthy when the network fails.

**Independent Test**: Disconnect the default remote, press `r` on the main tree and in sync; verify counts unchanged, no fetch-failure message, and the UI remains operable.

**Acceptance Scenarios**:

1. **Given** the default remote cannot be updated, **When** a manual reload attempt finishes unsuccessfully, **Then** the UI keeps the previous count snapshot and does not show a fetch-failure message.
2. **Given** some branches showed `?` before reload, **When** reload fails, **Then** they remain `?` (not a fabricated zero).

---

### Edge Cases

- Project has no default remote — manual reload is a no-op; counts stay as-is; no error.
- Manual reload requested while an automatic background fetch is already in flight — share the in-flight update; do not start a second fetch; refresh counts once when that update completes.
- User switches project or quits while manual reload is running — do not block exit; do not apply a stale refresh to the wrong project.
- User presses `r` repeatedly in quick succession — behavior must not corrupt counts or spawn unbounded parallel network work.
- MR, link, unlink, and project-picker screens — out of scope; `r` does not apply there.
- Scripted/plain CLI — out of scope; no manual reload key.
- Manual reload updates remote-tracking refs only; it does not checkout branches or modify the working tree.
- Very slow remotes — input must remain responsive; counts may update seconds after `r`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The main tree TUI MUST bind `r` to trigger a manual default-remote update followed by recomputation of inbound and outbound file-change counts for the current project.
- **FR-002**: When a manual reload succeeds, the main tree MUST refresh visible file-change badges in place using the same rules as the inbound-count and direction-toggle features (hide zero on browse surfaces, `?` for unknown, preserve active inbound/outbound direction).
- **FR-003**: The sync TUI (embedded and standalone) MUST bind `r` on the root picker and edge confirm steps to the same manual reload behavior, refreshing picker badges and the current confirm line on success.
- **FR-004**: Manual reload MUST use the project’s default remote (same source as automatic background fetch).
- **FR-005**: Manual reload MUST NOT block navigation, sync progression, direction toggle, or quitting while the update is in progress.
- **FR-006**: When manual reload fails or the project has no default remote, the TUI MUST keep the previous count snapshot and MUST NOT show a fetch-failure message.
- **FR-007**: A successful manual reload MUST only update counts for the project that requested it (no cross-project overwrite if the user switched projects).
- **FR-008**: The main-tree help footer MUST list `r` so the reload action is discoverable.
- **FR-009**: Manual reload MUST recompute both inbound and outbound counts even though only one direction is visible on the main tree, so direction toggles immediately reflect fresh data.
- **FR-010**: Scripted/plain CLI invocations MUST NOT gain a manual reload command.
- **FR-011**: At most one default-remote update per project MUST run at a time; if manual reload is requested while a fetch is already in flight for that project, the request MUST coalesce with the in-flight update (no second fetch) and counts MUST refresh once when that update completes.

### Key Entities

- **Manual reload**: A user-initiated default-remote update plus recomputation of file-change counts, triggered by `r`.
- **Count snapshot**: The inbound/outbound file-change maps currently shown before reload completes.
- **In-place refresh**: Updating tree rows, sync picker badges, or confirm text without restarting the flow or losing cursor position.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After a successful manual reload, file-change counts on the main tree match the files-changed numbers GitLab would show for the same branch pairs, for 100% of resolvable parent→child edges in a test project with known remote updates.
- **SC-002**: A developer can trigger reload and continue navigating the tree within 1 second of pressing `r` (no keyboard capture during the network call).
- **SC-003**: After manual reload from stale local refs, at least one previously wrong or `?` count becomes correct without restarting branchy.
- **SC-004**: When the default remote is unreachable, 100% of manual reload attempts leave the prior counts visible and show no fetch error UI.
- **SC-005**: The same parent→child pair shows the same refreshed inbound count on the main tree (inbound mode), sync picker, and edge confirm after a successful manual reload.

## Assumptions

- “Reload counts from remote” means fetch the default remote, then recompute file diffs from updated remote-tracking refs — not re-read the branch-tree config file from disk.
- Display rules from features 006 (parity diff display), 007 (lazy auto-fetch), and 008 (direction toggle) remain in force; this feature only adds an explicit user trigger.
- Automatic background fetch on tree/sync open continues to exist; manual reload complements it rather than replacing it.
- No periodic polling; `r` is strictly on-demand.
- During reload, browse surfaces show `?` badges and sync confirm shows "File count unavailable"; no spinner, status line, or error banner (same global constraints as automatic background fetch).

## Decisions (Grilling Session 2026-09-04)

| Topic | Decision |
|-------|----------|
| Surfaces | Main tree and sync flow (root picker + edge confirm) |
| In-flight fetch | Coalesce with existing update; one fetch, one refresh |
| In-progress feedback | `?` badges on browse surfaces and "File count unavailable" on sync confirm while reload is in flight |
