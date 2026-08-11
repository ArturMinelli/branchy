# Feature Specification: Lazy Auto-Fetch

**Feature Branch**: `007-lazy-auto-fetch`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "we need to make it so this app also auto fetches, but do that lazily so that tui startup is immediate and it displays local info immediatly"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open the tree on local data instantly (Priority: P1)

A developer opens the main branch tree. The tree and inbound change counts appear immediately from whatever is already on disk. Startup does not wait for the network. Shortly after, branchy updates from the default remote in the background. When that finishes, the counts on screen refresh in place — a `?` can become a number without restarting.

**Why this priority**: Instant local paint is the explicit product constraint; background update is what makes counts trustworthy for branches that were never checked out locally.

**Independent Test**: Open the main TUI on a project whose default remote has a tree branch that is missing locally but present on the remote. Confirm the tree is usable immediately with local counts; after the background update, that branch’s inbound count appears (or updates) without quitting.

**Acceptance Scenarios**:

1. **Given** a registered project with local refs already available, **When** the developer opens the main tree, **Then** the tree and inbound counts render from local state without waiting for a remote update.
2. **Given** a child listed in the tree that is missing locally but exists on the default remote, **When** the developer opens the main tree, **Then** they first see the existing unknown/local badge, and after the background update the inbound count refreshes in place.
3. **Given** the background update is still running, **When** the developer navigates, starts sync, or quits, **Then** those actions are not blocked.

---

### User Story 2 - Sync sees the same refreshed counts (Priority: P1)

A developer opens sync (embedded or standalone). The picker and edge confirms show local inbound counts immediately. A background update of the default remote runs if one is not already in flight for this project. When it completes, picker badges and the current edge confirm refresh in place so the yes/no decision uses current remote-tracking data.

**Why this priority**: Sync is where the count informs a decision; tree-only fetch would leave confirm under-informed for never-checked-out children.

**Independent Test**: Start sync on a project with a remote-only child; verify picker/confirm appear immediately with local data; after the background update, the same child’s count updates without leaving the flow.

**Acceptance Scenarios**:

1. **Given** the sync root picker is shown, **When** a background update completes, **Then** listed branches refresh inbound badges in place (same hide-zero / unknown rules as today).
2. **Given** an edge confirm is on screen, **When** a background update completes, **Then** the “files would change” context line updates to the new count (including a newly known zero).
3. **Given** a background update is already running because the main tree started it, **When** the developer opens sync, **Then** sync does not start a second overlapping update; it uses the in-flight one and refreshes when that finishes.

---

### User Story 3 - Stay usable when the remote update fails (Priority: P2)

A developer is offline, unauthenticated, or the default remote is unreachable. The TUI never stalls. Local tree, counts, and sync continue exactly as they were. No extra error or status line is shown for the failed background update.

**Why this priority**: Fetch is an enhancement; local workflow must remain the source of truth when the network fails.

**Independent Test**: Disconnect the default remote (invalid URL or offline); open tree and sync; verify immediate local UI, no fetch error banner, and counts unchanged after the failed attempt.

**Acceptance Scenarios**:

1. **Given** the default remote cannot be updated, **When** the background attempt finishes unsuccessfully, **Then** the UI keeps the local snapshot and does not show a fetch-failure message.
2. **Given** some branches were `?` because they are not local, **When** the update fails, **Then** they remain `?` (not a fabricated zero).

---

### Edge Cases

- Project has no default remote — skip the background update; keep local data; no error.
- Fetch still running when the user switches project or quits — do not block exit; do not apply a stale refresh to the wrong project.
- User starts sync before the tree’s background update finishes — share the in-flight update; do not block sync.
- After a successful update, zeros stay hidden on tree/picker; confirm still shows a known zero.
- Scripted/plain CLI (`branchy sync --from`, flags, non-TTY) — out of scope; no automatic remote update there.
- MR / link / unlink / init / projects TUIs — out of scope for triggering fetch.
- Very slow remotes — first paint must not wait; refresh may arrive seconds later.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The main tree TUI MUST render the branch tree and inbound counts from local repository state before any remote update starts or completes.
- **FR-002**: After the main tree is shown, the system MUST start a background update from the project’s default remote without blocking input.
- **FR-003**: The sync TUI (embedded and standalone) MUST show picker and confirm inbound counts from local state immediately, and MUST run the same kind of background default-remote update if one is not already in flight for that project.
- **FR-004**: At most one background remote update MUST run per project at a time; tree and sync MUST share an in-flight update.
- **FR-005**: When a background update succeeds, inbound counts currently visible on the main tree, sync picker, and sync edge confirm MUST refresh in place using the same display rules as the inbound-count feature (hide zero on browse surfaces; show known zero on confirm; `?` for still-unresolvable names).
- **FR-006**: When a background update fails or the project has no default remote, the TUI MUST keep the local snapshot and MUST NOT show a fetch-failure message.
- **FR-007**: Navigation, starting sync, confirming/skipping edges, and quitting MUST remain possible while a background update is in progress.
- **FR-008**: Scripted/plain CLI invocations MUST NOT start this automatic remote update.
- **FR-009**: A refresh MUST only apply to the project that requested the update (no cross-project overwrite if the user switched projects).

### Key Entities

- **Local snapshot**: Inbound counts computed from refs already on disk (local branches and already-fetched remote-tracking branches).
- **Background remote update**: A non-blocking refresh of the default remote after first paint.
- **In-place refresh**: Recomputing inbound counts and updating the visible tree/picker/confirm without restarting the flow.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From launching the main TUI to seeing the tree and local inbound counts takes no longer than today’s local-only open (no added wait for the network).
- **SC-002**: After a successful background update, 100% of tree branches that are resolvable via the default remote show a real inbound count instead of remaining `?` solely because they were never checked out locally.
- **SC-003**: Users can start sync, move selection, or quit during an in-flight update in 100% of attempts (update never captures the keyboard).
- **SC-004**: When the default remote is unreachable, users see no fetch error UI and can complete tree browsing and sync with local data.
- **SC-005**: After a successful update, the same parent→child pair shows the same refreshed number on the tree, sync picker, and edge confirm.

## Assumptions

- “Default remote” means the repository’s usual upstream remote (typically `origin`).
- “Lazily” means: paint local first, then update in the background — not “fetch only the missing branch names.”
- One background update per project session when the tree or sync is opened is enough; no periodic polling.
- Existing inbound-count display rules (006) stay in force; this feature only refreshes the data behind them.
- No automatic update of working-tree files or checkouts — only remote-tracking information used for comparison.
- Scripted CLI remains unchanged (006 FR-010 still applies).

## Decisions (Grilling Session 2026-08-11)

| Topic | Decision |
|-------|----------|
| Surfaces | Main tree and sync TUI only |
| What to update | Default remote (normal full fetch of that remote) |
| After success | Refresh inbound counts in place |
| On failure | Keep local info; no extra error UI |
