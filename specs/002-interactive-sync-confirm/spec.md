# Feature Specification: Interactive Sync Confirmations

**Feature Branch**: `002-interactive-sync-confirm`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "We need to make it so the sync command that we run via \"s\" key shortcut in the TUI, does not automatically create all MRs in one shot, we need to ask and confirm for each child branch, and then in the end we should also ask if they want us to open all MRs in their browser. We need to make sure the chrome tabs get opened in order."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Per-edge sync confirmation in TUI (Priority: P1)

A developer selects a branch in the branchy tree view and presses `s` to sync. Instead of a single bulk confirmation that creates every merge request at once, the TUI walks each parent→child edge in tree order and asks whether to create an MR for that edge. The user can accept or decline each edge independently; declined edges are skipped and the flow continues to the next edge.

**Why this priority**: This is the core problem — the TUI currently creates all MRs in one shot after a single `y` press, which is risky and inconsistent with user intent.

**Independent Test**: Open branchy TUI, select a branch with at least three child edges, press `s`, decline the second edge and accept the others; verify only accepted edges result in created MRs and the summary reflects created/skipped/failed per edge.

**Acceptance Scenarios**:

1. **Given** a registered project with multiple parent→child edges below the selected branch, **When** the user presses `s`, **Then** the TUI immediately prompts for the first edge (no bulk "confirm all" step).
2. **Given** the TUI is prompting for edge `parent → child`, **When** the user confirms, **Then** an MR is created (or an existing open MR is reported as skipped) before moving to the next edge.
3. **Given** the TUI is prompting for edge `parent → child`, **When** the user declines, **Then** that edge is marked skipped and the TUI continues to the next edge without aborting the sync.
4. **Given** all edges have been processed, **When** the sync completes, **Then** the TUI shows a per-edge summary (created, skipped, failed) with URLs where applicable.
5. **Given** the user cancels during sync (e.g. Esc), **When** cancellation is accepted, **Then** no further edges are processed and the user returns to the tree view with results for edges processed so far.

---

### User Story 2 - Optional browser open at end of sync (Priority: P1)

After all edges are processed, the system asks whether to open the newly created MRs in the browser. Browser tabs open only if the user confirms. Tabs open in tree depth-first order (the same order edges are walked), including only MRs created during the current session — not skipped existing MRs or declined edges.

**Why this priority**: Users want control over browser noise; ordered tabs matter when reviewing a cascade of sync MRs.

**Independent Test**: Complete a TUI sync that creates two MRs; at the end answer yes to browser open; verify exactly two tabs open in DFS edge order.

**Acceptance Scenarios**:

1. **Given** sync created one or more MRs in the session, **When** all edges are processed, **Then** the TUI prompts "Open created MRs in browser? [y/N]".
2. **Given** the user answers yes to the browser prompt, **When** the browser is available, **Then** one tab opens per created MR, in tree depth-first order among created MRs only.
3. **Given** the user answers no to the browser prompt, **When** the flow completes, **Then** no browser tabs open and MR URLs remain visible in the summary.
4. **Given** sync created zero MRs (all declined, skipped, or failed), **When** all edges are processed, **Then** no browser prompt is shown.
5. **Given** the browser cannot be opened, **When** the user confirmed browser open, **Then** MRs remain created successfully, URLs stay in the summary, and a non-fatal warning is shown.

---

### User Story 3 - Unified sync behavior in CLI (Priority: P2)

A developer runs `branchy sync --from <branch>` from the terminal. Per-edge confirmation already exists; the change is to stop auto-opening browser tabs and instead prompt at the end whether to open created MRs, using the same rules as the TUI (created-only, DFS order). The `--yes` / `-y` flag continues to skip per-edge prompts but does not auto-open the browser.

**Why this priority**: Keeps TUI and CLI sync semantics aligned so users are not surprised by different browser behavior depending on entry point.

**Independent Test**: Run `branchy sync --from develop`, confirm two edges, answer yes at the end browser prompt; verify two tabs open in DFS order. Repeat with `-y` and verify per-edge prompts are skipped but browser still requires end confirmation.

**Acceptance Scenarios**:

1. **Given** `branchy sync --from <branch>` without `-y`, **When** the user completes per-edge prompts, **Then** a final prompt asks whether to open created MRs in the browser.
2. **Given** `branchy sync --from <branch> -y`, **When** all edges are processed without per-edge prompts, **Then** the end browser prompt still appears if any MRs were created.
3. **Given** the user confirms browser open at the end, **When** tabs open, **Then** order matches tree depth-first order among created MRs only.
4. **Given** no MRs were created, **When** sync completes, **Then** no browser prompt is shown.

---

### Edge Cases

- What happens when GitLab credentials are not configured? Sync fails early with a clear auth message (unchanged from current behavior).
- What happens when an open MR already exists for an edge? Creation is skipped, the existing URL is shown in the summary, and that URL is excluded from the end browser batch.
- What happens when MR creation fails on one edge? The failure is recorded, sync continues to remaining edges, and the failed edge is excluded from the browser batch.
- What happens when the selected branch has no child edges? Sync completes immediately with an informative message; no prompts are shown.
- What happens when the user declines every edge? Sync completes with all edges marked skipped; no browser prompt is shown.
- What happens when only one edge exists? A single per-edge prompt is shown, then the end summary and optional browser prompt behave the same as multi-edge sync.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: TUI sync (`s` shortcut) MUST prompt for confirmation on each parent→child edge individually; it MUST NOT use a single bulk confirmation that creates all MRs at once.
- **FR-002**: TUI sync MUST NOT show a separate bulk "confirm entire sync" step before per-edge prompts; the flow starts with the first edge prompt immediately.
- **FR-003**: When the user declines an edge during sync, the system MUST skip that edge and continue processing remaining edges.
- **FR-004**: After all edges are processed, the system MUST prompt the user whether to open created MRs in the browser; the browser MUST NOT open automatically.
- **FR-005**: The end browser prompt MUST include only MRs created during the current sync session; skipped (existing MR), declined, and failed edges MUST be excluded.
- **FR-006**: When the user confirms browser open, tabs MUST open in tree depth-first order (matching edge walk order) among the created MRs.
- **FR-007**: CLI `branchy sync` MUST adopt the same end browser prompt behavior as TUI sync; auto-opening browser tabs at the end of sync MUST be removed.
- **FR-008**: CLI flag `-y` / `--yes` MUST continue to skip per-edge confirmation prompts but MUST NOT bypass the end browser prompt.
- **FR-009**: If zero MRs were created in the session, the system MUST NOT show the browser prompt.
- **FR-010**: Sync MUST display a per-edge summary after completion showing action (created, skipped, failed), message, and URL where applicable.
- **FR-011**: Per-edge prompts MUST clearly identify the parent and child branch for the edge being considered.
- **FR-012**: Browser open failures MUST be non-fatal; created MRs and URLs MUST remain available in the summary.

### Key Entities

- **Sync Edge**: A parent→child pair from the branch tree below the chosen root branch; processed in depth-first tree order.
- **Sync Session**: A single sync invocation from TUI or CLI, producing a collection of per-edge results and an optional browser-open decision at the end.
- **Edge Result**: Outcome for one edge — created, skipped (user declined or existing MR), or failed — with optional MR URL and message.
- **Browser Batch**: The ordered list of MR URLs eligible for opening at session end (created MRs only, DFS order).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can decline individual sync edges without aborting the full sync; 100% of declined edges are skipped while remaining edges are still offered.
- **SC-002**: Zero automatic browser opens occur during sync in both TUI and CLI; browser opens only after explicit end-of-session confirmation.
- **SC-003**: When multiple MRs are opened in one session, 100% of tabs appear in tree depth-first order relative to the created MR set.
- **SC-004**: End browser prompt appears only when at least one MR was created in the session.
- **SC-005**: TUI and CLI sync produce equivalent browser-open semantics (same inclusion rules and ordering).
- **SC-006**: Users can complete a selective sync (accept some edges, decline others) and see an accurate per-edge summary in under 2 minutes for trees with up to 10 edges (excluding network latency).

## Assumptions

- Edge walk order remains depth-first, consistent with the existing branch tree `CollectEdges` behavior.
- Per-edge confirmation in CLI (`Create MR parent → child? [y/N]`) is the reference interaction pattern; TUI per-edge prompts should be equivalent in meaning with TUI-appropriate presentation.
- The `--yes` flag meaning is unchanged for per-edge prompts: skip individual confirmations, not browser consent.
- Existing MR detection and skip behavior per edge is unchanged; only browser batch inclusion and bulk TUI confirm change.
- Manual MR flow (`branchy mr` / `m` shortcut) is out of scope; it already prompts per MR and per browser open.
- Tab ordering refers to the sequence tabs are opened in the browser, matching DFS edge order among created MRs (gaps from skipped/declined edges are omitted, not re-sorted by creation time).
