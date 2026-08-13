# Feature Specification: Bidirectional Sync

**Feature Branch**: `009-bidirectional-sync`

**Created**: 2026-08-13

**Status**: Draft

**Input**: User description: "The sync command now has to work in both directions. The downward direction and the upward direction, respecting what we already have. If we press sync in the new opposite direction, it will create MRs from all childs and so on up until the branch we pressed sync."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cascade child work up to the selected branch (Priority: P1)

A developer flips the main tree to outbound and presses sync on a branch that has descendants. Instead of creating parent → child merge requests down the subtree, sync offers a child → parent merge request for every edge in that same subtree, stopping at the branch they pressed sync on. Deeper children are offered first, then their parents, so work can land upward in stack order. Each edge is still confirmed one at a time; existing open merge requests in that same child → parent direction are skipped, not failed.

**Why this priority**: This is the new product behavior — without the upward cascade, bidirectional sync does not exist.

**Independent Test**: On a tree with at least two levels below the selected branch, flip to outbound, press sync, accept every edge; verify merge requests are created child → parent only, none above the selected branch, and deeper edges are offered before shallower ones.

**Acceptance Scenarios**:

1. **Given** the main tree is in outbound mode and the selected branch has descendants, **When** the developer presses sync, **Then** the flow offers one child → parent merge request per descendant edge in that subtree and does not offer any edge whose parent is above the selected branch.
2. **Given** a grandchild under a child under the selected branch, **When** upward sync runs, **Then** the grandchild → child edge is offered before the child → selected-branch edge.
3. **Given** the developer confirms an upward edge, **When** no open merge request already exists from that child into that parent, **Then** that merge request is created before the next edge is offered.
4. **Given** an open merge request already exists from a child into its parent, **When** that upward edge is reached, **Then** creation is skipped, the existing address is shown, and the flow continues.
5. **Given** the developer declines one upward edge, **When** more edges remain, **Then** that edge is skipped and the remaining edges are still offered.

---

### User Story 2 - Downward sync stays as it is today (Priority: P1)

A developer leaves the tree in inbound mode (the default) and presses sync. The walk, confirmations, skip rules, summary, and end-of-session browser question are unchanged: parent → child merge requests, top-down tree order, inbound file-change numbers. Scripted sync with flags stays this downward path only.

**Why this priority**: The request is to add the opposite direction while respecting what already works; breaking downward sync would be a regression of the primary workflow.

**Independent Test**: With the tree inbound, run the same multi-edge sync used today; compare offer order, source → target, counts, skip-of-existing, and browser prompt to current downward behavior. Repeat with flagged scripted sync and confirm it is still downward-only.

**Acceptance Scenarios**:

1. **Given** the main tree is in inbound mode, **When** the developer presses sync on a branch with descendants, **Then** edges are offered parent → child in the same top-down tree order as today.
2. **Given** inbound sync, **When** an edge is confirmed, **Then** the merge request is from parent into child (not the reverse).
3. **Given** a flagged scripted sync (for example a starting-branch flag), **When** it runs, **Then** it remains downward-only and does not gain a direction choice.
4. **Given** downward sync, **When** an open parent → child merge request already exists, **Then** that edge is still skipped as today; an open child → parent merge request on the same pair does not cause that skip.

---

### User Story 3 - Confirm numbers match the merge request being created (Priority: P1)

While syncing upward, the root picker (when shown) and each edge confirm show how many files the parent would receive from the child — the same outbound number the tree shows in outbound mode. While syncing downward, those screens keep today’s inbound numbers. The developer is never asked to confirm a merge request using the opposite direction’s count.

**Why this priority**: A mismatched count would make accept/decline unsafe, which is why sync previously stayed inbound-only.

**Independent Test**: Use a pair whose inbound and outbound file-change sizes differ. Run downward sync and confirm the inbound number; flip to outbound, run upward sync, and confirm the outbound number on picker and confirm.

**Acceptance Scenarios**:

1. **Given** upward sync is active, **When** an edge confirm is shown, **Then** it reports the outbound file-change count (files the parent would receive), including a known zero.
2. **Given** downward sync is active, **When** an edge confirm is shown, **Then** it still reports the inbound file-change count (files the child would receive).
3. **Given** the standalone sync root picker is in upward mode, **When** a non-root branch is listed, **Then** its badge uses the outbound count (hide zero / unknown placeholder, same visibility rules as the tree).
4. **Given** that picker is in downward mode, **When** the same branch is listed, **Then** its badge uses the inbound count.

---

### User Story 4 - Standalone sync can go either way (Priority: P2)

A developer runs interactive sync from the shell (no flags) and never opens the main tree. The root picker starts in downward mode and offers the same direction toggle as the tree. Flipping it switches the pending sync to upward, flips picker badges, and after a root is chosen the cascade uses that direction. Confirm screens do not offer another toggle — direction is already chosen.

**Why this priority**: Interactive sync from the shell is a first-class surface; without a picker toggle, upward would exist only inside the main tree.

**Independent Test**: Launch standalone interactive sync; verify it starts downward; toggle direction; pick a root with descendants; verify the cascade is upward (child → parent, deepest first) with outbound counts.

**Acceptance Scenarios**:

1. **Given** standalone interactive sync has just opened, **When** the developer reads the root picker, **Then** it is in downward mode and the help text names the current direction and the toggle key.
2. **Given** the picker is showing, **When** the developer activates the direction toggle, **Then** the pending direction becomes upward and listed badges switch to outbound counts.
3. **Given** the picker is in upward mode, **When** the developer chooses a root, **Then** the following edge confirms are child → parent, deepest first, using outbound counts.
4. **Given** a root has been chosen, **When** edge confirms are showing, **Then** there is no direction toggle on those screens.

---

### Edge Cases

- Selected or picked branch has no descendants: sync ends immediately with the same kind of informative empty message used today; no confirms and no browser question.
- An open merge request already exists in the direction being synced: skip and show its address; do not treat it as a failure.
- An open merge request exists only in the opposite direction on the same parent/child pair: it does not skip the current edge.
- One edge fails to create: record the failure, continue remaining edges, exclude it from the browser batch.
- Developer cancels mid-cascade: no further edges are processed; results so far remain visible.
- Every edge declined: summary shows all skipped; no browser question.
- Deepest-first upward order: a three-level chain offers grandchild → child, then child → ceiling; siblings at the same depth follow existing tree sibling order, still bottom-up.
- Browser open after upward sync: only merge requests created in this session, in the same deepest-first order they were offered.
- Main-tree sync does not show a direction toggle; it uses the tree’s current direction at the moment sync starts.
- Standalone picker toggle is for that run only; the next standalone launch starts downward again.
- Scripted / flagged sync never reads the tree direction and never offers upward.
- Manual single merge-request flow is unchanged and out of scope.
- Credentials missing: fail early with the same clear auth message as today.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When the main tree is inbound, sync from the selected branch MUST perform today’s downward cascade: parent → child merge requests, top-down tree order, inbound file-change counts.
- **FR-002**: When the main tree is outbound, sync from the selected branch MUST perform an upward cascade for that branch’s descendant edges only, stopping at the selected branch.
- **FR-003**: Upward sync MUST use the same parent/child edges as downward sync from that branch; it MUST NOT add ancestor edges above the selected or picked branch and MUST NOT include edges outside that subtree.
- **FR-004**: For each upward edge, the merge request source MUST be the child and the target MUST be the parent.
- **FR-005**: Upward edges MUST be offered deepest-first (a child is offered before its parent), then remaining same-depth edges in existing tree sibling order.
- **FR-006**: Per-edge confirm, decline-and-continue, existing-open skip, per-edge summary, cancel, and end-of-session browser question MUST keep the same rules as today’s sync, applied to the edges and direction of the current run.
- **FR-007**: The end browser question MUST include only merge requests created in this run; skipped (existing), declined, and failed edges MUST be excluded. Tabs MUST open in the order those created items were offered (top-down when downward, deepest-first when upward).
- **FR-008**: While upward sync is active, the root picker (when shown) and each edge confirm MUST show outbound file-change counts — files the parent would receive from the child — including a known zero on confirm and hide-zero / unknown placeholder on the picker.
- **FR-009**: While downward sync is active, those screens MUST show inbound file-change counts, including a known zero on confirm.
- **FR-010**: Standalone interactive sync MUST start in downward mode and MUST offer the same direction toggle on the root picker; the picker help MUST name the current direction and the toggle key.
- **FR-011**: Toggling direction on the standalone picker MUST flip pending direction and picker badges immediately, before a root is chosen.
- **FR-012**: After a root is chosen (standalone) or sync is started from the main tree, confirm screens MUST NOT offer a direction toggle.
- **FR-013**: Scripted / flagged sync MUST remain downward-only and MUST NOT gain a direction flag or read the main-tree direction.
- **FR-014**: An open merge request MUST skip an edge only when it matches that edge’s source and target for the current direction; the opposite-direction open request on the same pair MUST NOT cause a skip.
- **FR-015**: Each edge prompt MUST clearly name source and target for the current direction (parent → child when downward, child → parent when upward).
- **FR-016**: If the chosen branch has no descendant edges, sync MUST finish immediately with an informative empty result and MUST NOT show edge confirms or a browser question.

### Key Entities

- **Sync direction**: Downward (parent → child, top-down) or upward (child → parent, deepest-first). Chosen by the main-tree direction when starting from the tree, or by the standalone picker toggle (default downward).
- **Sync ceiling**: The selected or picked branch. All edges in the run stay inside its descendant subtree; it is the highest target of an upward run and the highest source of a downward run.
- **Sync edge**: A parent/child pair from that subtree. Direction decides which end is source and which is target, and in which order the pair is offered.
- **Sync session**: One sync run (main tree or standalone), with one direction, a list of per-edge results, and an optional browser-open decision at the end.
- **Edge result**: Created, skipped (declined or existing matching-direction request), or failed — with optional address and message.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From a ceiling with at least two descendant levels, a developer can complete an all-accept upward sync and get a child → parent request on 100% of those descendant edges and 0% of edges above the ceiling.
- **SC-002**: In a three-level chain, 100% of upward runs offer the deeper edge before the shallower edge.
- **SC-003**: With the tree inbound, 100% of main-tree sync runs still create parent → child requests in today’s top-down order.
- **SC-004**: For a pair whose inbound and outbound sizes differ, 100% of downward confirms show the inbound number and 100% of upward confirms show the outbound number.
- **SC-005**: A developer who only uses standalone interactive sync can switch to upward and start a cascade without opening the main tree, in under 15 seconds of orientation (excluding network waits).
- **SC-006**: Flagged scripted sync produces 0 upward merge requests and requires no new direction choice.
- **SC-007**: Users can decline individual upward edges without aborting the run; 100% of declined edges are skipped while remaining edges are still offered.
- **SC-008**: Zero automatic browser opens occur; when the user agrees to open created requests after an upward run, 100% of those tabs follow deepest-first offer order.

## Assumptions

- “Respecting what we already have” means keep per-edge confirm, decline-and-continue, skip-existing, summary, cancel, auth failure, and the end browser question; only direction, offer order, counts, and source/target change when going up.
- The main-tree direction toggle from the existing count-direction feature is the switch for embedded sync; this feature does not add a second sync-specific key on the tree.
- Standalone interactive sync is in scope; flagged / scripted sync is not.
- The manual single merge-request flow is out of scope.
- Opposite-direction open requests are independent; they never count as “already synced” for the current run.
- Standalone picker direction is not remembered after the process exits.
- Same-depth sibling order stays the existing tree order; only the vertical walk reverses for upward.
- This feature replaces the earlier rule that sync screens always show inbound counts: counts now follow the sync direction.

## Decisions (Grilling Session 2026-08-13)

| Topic                 | Decision                                                                                          |
|-----------------------|---------------------------------------------------------------------------------------------------|
| Direction trigger     | Main-tree `d` toggle: inbound sync = downward, outbound sync = upward                             |
| Upward walk           | Same descendant edges as downward; each request is child → parent; ceiling is the starting branch |
| Surfaces              | Main-tree sync and standalone interactive sync; flagged scripted sync stays downward-only         |
| Confirm counts        | Match the request being created (outbound when upward, inbound when downward)                     |
| Upward offer order    | Deepest children first, then their parents                                                        |
| Standalone direction  | Same toggle on the root picker; starts downward; badges flip with direction                       |
