# Feature Specification: Parent Diff Display

**Feature Branch**: `006-parity-diff-display`

**Created**: 2026-08-11

**Status**: Draft

**Input**: User description: "We need to update the main TUI, and the sync TUI, so that it displays how many diffs each branch has in relation to its parent, so that we always know how many changes that branch is going to get when we sync the parent to it"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See pending sync size on the main tree (Priority: P1)

A developer opens the main branch tree and, for each child that is behind its parent, sees how many file changes that child would receive if the parent were synced into it. The number matches what GitLab would show as files changed on the parent→child merge request. Branches already in sync stay visually quiet (no count). Roots show no inbound count.

**Why this priority**: The main tree is the default workspace; glanceable inbound size is the primary value of the feature.

**Independent Test**: Open the main TUI on a project whose children have known, different inbound file-change counts relative to their parents; verify each behind child shows the expected count, in-sync children show no count, and roots show no inbound count.

**Acceptance Scenarios**:

1. **Given** a branch tree where child `feature` would receive N file changes from parent `develop` (the same N GitLab would show on an MR `develop` → `feature`), **When** the developer views the main tree, **Then** the `feature` row shows N as the pending inbound change count.
2. **Given** a root branch with no parent, **When** the developer views the main tree, **Then** that row does not show an inbound change count.
3. **Given** a child that would receive zero file changes from its parent, **When** the developer views the main tree, **Then** that row shows no inbound count (the zero is hidden).

---

### User Story 2 - See pending sync size while choosing a sync root (Priority: P1)

A developer starts sync (embedded `s` or standalone `branchy sync`) and lands on the root picker. Each listed non-root branch shows the same inbound-from-parent file-change count as the main tree, so they can see how stale each branch is relative to its parent before choosing where to sync from.

**Why this priority**: The picker is the first sync decision; without counts there, the user must leave sync or rely on memory from the tree.

**Independent Test**: Open the sync root picker on a tree with mixed in-sync and behind children; verify listed non-root branches show the same inbound counts as the main tree, zeros are hidden, and roots show no inbound count.

**Acceptance Scenarios**:

1. **Given** the sync root picker lists branches that also appear on the main tree, **When** the user scans the list, **Then** each non-root branch that has a non-zero inbound file-change count shows that same count.
2. **Given** a listed branch is a tree root or is already in sync with its parent, **When** the picker is shown, **Then** that row shows no inbound count.

---

### User Story 3 - See pending sync size on each edge confirm (Priority: P1)

A developer is confirming whether to create the merge request for a parent→child edge. The confirm panel includes how many file changes that child will receive from that parent — the same number GitLab would show on that MR — so they can accept or skip with that context.

**Why this priority**: Confirm is the moment the count most directly informs a yes/no decision.

**Independent Test**: Walk sync on a subtree with known parent→child file-change counts; verify each edge confirm shows the inbound file-change count for that edge before the user answers.

**Acceptance Scenarios**:

1. **Given** sync is confirming edge `parent` → `child` and that MR would show N files changed, **When** the confirm panel is shown, **Then** the panel includes N so the user knows the inbound size for that edge.
2. **Given** the inbound file-change count for the edge is zero, **When** the confirm panel is shown, **Then** the panel still includes the known count (including zero) so the user is not surprised by an empty change set.
3. **Given** the same parent→child pair is visible on the main tree and in the confirm panel, **When** both are shown for comparable local state, **Then** the numeric count matches.

---

### User Story 4 - Stay usable when comparison is unavailable (Priority: P2)

A developer opens the tree or sync TUI when a branch is missing locally, refs are stale, or comparison fails. The TUI remains usable and communicates that the count is unknown rather than inventing a number or crashing.

**Why this priority**: Real repos often have incomplete local refs; the feature must degrade safely.

**Independent Test**: Remove or rename a local branch that still appears in the tree config; open the main TUI and sync; verify a clear unknown/unavailable indicator appears for that branch’s count and that navigation and sync can continue.

**Acceptance Scenarios**:

1. **Given** a child listed in the tree cannot be compared to its parent, **When** the main tree, sync root picker, or edge confirm is shown, **Then** the UI shows a clear placeholder for that count (not a fake zero) and remains operable.
2. **Given** comparison succeeds for some edges and fails for others, **When** viewing the tree or walking sync, **Then** successful counts still display while failed ones show the placeholder independently.

---

### Edge Cases

- Root branches have no parent — no inbound count on the tree or in the sync root picker.
- Child already matching its parent (zero inbound file changes) — hide the count on the tree and picker; still show the known zero on the edge confirm.
- Local branch missing, not fetched, or otherwise incomparable — unknown placeholder, not a fabricated zero.
- Diverged branches — the count is still the file-change size GitLab would show on the parent→child MR (changes the child would receive), not the child’s unique commits going the other way.
- Very large counts — display must remain readable in narrow terminals.
- Scripted/non-TUI sync and CLI flags — out of scope; interactive TUI only.
- Multiple parents — not applicable; the tree model is a single-parent hierarchy.
- Sync results summary — out of scope; counts are not required there.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The main tree TUI MUST display, for each non-root branch with a non-zero inbound change count, how many file changes that branch would receive from its parent if the parent were synced into it.
- **FR-002**: The sync TUI root picker MUST display the same inbound-from-parent file-change count on each listed non-root branch, using the same hide-when-zero rule as the main tree.
- **FR-003**: The sync TUI MUST display the inbound file-change count on each parent→child edge confirmation before the user accepts or declines creating the merge request, including when that count is zero.
- **FR-004**: The unit of “diffs” MUST be the number of files changed that GitLab would show on the merge request from parent to child (not commit count, not a two-way ahead/behind pair).
- **FR-005**: The same parent→child pair MUST show the same numeric count on the main tree, the sync root picker, and the edge confirm when local comparison state is the same.
- **FR-006**: Root branches MUST NOT show an inbound-from-parent count.
- **FR-007**: When the inbound file-change count is zero, the main tree and sync root picker MUST hide the count rather than showing `0` or an in-sync label.
- **FR-008**: When a comparison cannot be computed, the TUI MUST show a clear unavailable/unknown indicator for that count and MUST remain usable.
- **FR-009**: Counts MUST reflect local repository state available at display time. This feature does not require auto-fetching remotes.
- **FR-010**: Scripted/plain CLI sync prompts are out of scope; their behavior remains unchanged.

### Key Entities

- **Parent**: The immediate parent of a branch in the configured branch tree; the source side of a sync edge into that branch.
- **Inbound file-change count**: How many files would change on the child if the parent were merged into it — the same files-changed number GitLab would show on the parent→child merge request.
- **Unknown count**: A non-numeric placeholder used when parent and child cannot be compared from local state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a project with at least three parent→child pairs of known different inbound file-change sizes, a user can identify the largest pending sync target from the main tree alone in under 10 seconds without leaving the TUI.
- **SC-002**: During sync, 100% of root-picker rows for comparable non-root, non-zero branches show the inbound file-change count, and 100% of edge confirmations for comparable branches show the inbound file-change count before the yes/no decision.
- **SC-003**: For a given parent→child pair, the number shown in branchy matches the files-changed number GitLab would show on that merge request.
- **SC-004**: When comparison fails for a branch, users never see a fabricated zero; the unknown state is recognizable in under 2 seconds.
- **SC-005**: Existing sync and tree navigation flows remain completable; adding counts does not block starting sync, confirming edges, or quitting.

## Assumptions

- “In relation to its parent” means inbound parent→child file changes only, not the child’s unique commits in the other direction.
- The GitLab-comparable number is files changed on the parent→child merge request, not commit count and not additions/deletions totals.
- Counts use local repository state already present; auto-fetch is out of scope.
- Display is TUI-only (main tree, sync root picker, sync edge confirm). Plain CLI is unchanged. Sync results summary does not need counts in this feature.
- Formatting stays compact (a short suffix or badge on tree/picker rows; a context line on the confirm panel) so narrow terminals remain usable.
- Hiding zero on browse surfaces (tree, picker) keeps the UI quiet; showing zero on the confirm panel is intentional because that is a commit-or-skip decision.

## Decisions (Grilling Session 2026-08-11)

| Topic | Decision |
|-------|----------|
| Metric | Files changed, matching the number GitLab would show on the parent→child MR |
| Zero inbound on tree/picker | Hide the count |
| Sync surfaces | Root picker and edge confirm (not the results summary) |
