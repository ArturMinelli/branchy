# Feature Specification: Shared Use Cases

**Feature Branch**: `011-shared-use-cases`

**Created**: 2026-08-21

**Status**: Draft

**Input**: User description: "Extract shared Link, Unlink, Sync, CreateMR, and Init operations so the plain CLI and both interactive surfaces call the same owner. Deepen existing packages; no new application layer. Behavior freeze (outcomes). Independently mergeable."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One link operation for every surface (Priority: P1)

A contributor changes the rules for adding a parent→child edge (required names, no self-link, duplicate-edge rejection, persist). After that single change, scripted `branchy link`, standalone interactive link, and main-tree link all follow the new rule. End users still get the same successful tree, the same error text for invalid pairs, and the same persisted file.

**Why this priority**: Link is duplicated in three places today; that is the clearest proof that operations have no single owner.

**Independent Test**: Introduce a validation change in the shared link operation only. Run scripted link, standalone interactive link, and main-tree link against the same invalid and valid pairs. All three accept or reject identically and persist the same tree.

**Acceptance Scenarios**:

1. **Given** a registered project, **When** the user links parent→child via scripted command, standalone interactive link, or the main tree, **Then** the resulting tree and on-disk store match for the same inputs.
2. **Given** an invalid pair (empty name, same parent and child, or an edge that already exists), **When** any of those three surfaces attempts the link, **Then** they report the same rejection and do not persist a partial change.
3. **Given** a successful link, **When** the user reopens the project, **Then** the new edge is present regardless of which surface created it.

---

### User Story 2 - One unlink operation, including validation (Priority: P1)

A contributor changes unlink rules (missing branch, missing edge, subtree size reporting, persist). Scripted unlink and both interactive unlink paths use that same rule. The extra edge checks that currently live only on the scripted command are not a second source of truth.

**Why this priority**: Unlink already drifts: the scripted path re-validates the edge; interactive paths do not. Unifying the operation is what stops that drift.

**Independent Test**: Unlink the same parent→child subtree from scripted CLI and from each interactive path. Confirm identical success output meaning (including how many branches were removed) and identical rejection for a missing edge or unknown branch.

**Acceptance Scenarios**:

1. **Given** an existing parent→child edge with a subtree, **When** unlink runs from any surface, **Then** the same members are removed and the same parent children list remains.
2. **Given** a child that is not in the tree, or a parent→child pair that is not an edge, **When** unlink is attempted from any surface, **Then** the attempt fails with the same meaning and the tree is unchanged.
3. **Given** a successful unlink, **When** the tree is reloaded, **Then** the subtree is gone regardless of which surface performed it.

---

### User Story 3 - Sync and merge-request create own GitLab auth and edge work (Priority: P1)

A developer runs sync or a manual merge-request from either the plain command path or an interactive screen. Authentication against GitLab is performed inside those operations, not as a separate interactive-only pre-check. Interactive screens display the error the operation returns. Creating one merge request and walking a sync cascade use the same create-MR outcome model (created, skipped because already open, failed).

**Why this priority**: Auth and create currently happen in more than one caller; interactive sync constructs its own GitLab client before the cascade. That is the leak this spec closes for sync/MR.

**Independent Test**: With GitLab unauthenticated, run scripted sync/MR and interactive sync/MR. Both fail with the same auth meaning and create nothing. With auth OK, a single-edge sync and a manual MR with the same source and target produce the same created-or-skipped outcome.

**Acceptance Scenarios**:

1. **Given** GitLab is not authenticated, **When** the user starts sync or a manual merge request from any surface, **Then** the attempt fails with an auth error and no merge request is created.
2. **Given** GitLab is authenticated and an open MR already exists for the pair, **When** sync or manual create runs, **Then** the pair is skipped (not failed) and the existing URL is returned on every surface.
3. **Given** a confirmed sync edge, **When** it is processed, **Then** the outcome (created, skipped, or failed) matches a manual create of the same source and target.
4. **Given** the user declines a sync edge, **When** the cascade continues, **Then** that edge is skipped as user-declined and later edges still run — same as today.

---

### User Story 4 - Init stays one registration path (Priority: P2)

Interactive init and scripted init (`--force` or not) share the same registration operation: detect the git root, import or scaffold the tree, write the store, update the index. The interactive wizard only confirms and then calls that operation.

**Why this priority**: Init is already closer to a single function than link/unlink; locking it in the same pass prevents a new split.

**Independent Test**: Register a repo with interactive init and with scripted init on equivalent fixtures. Compare project id, path, and imported tree.

**Acceptance Scenarios**:

1. **Given** an unregistered git repo, **When** the user completes interactive init or scripted init, **Then** the project is indexed and the tree is stored the same way.
2. **Given** an already registered repo without force, **When** either surface tries to init, **Then** both refuse with the same already-registered meaning.
3. **Given** `--force` on the scripted path, **When** init re-imports, **Then** the existing project id is reused (current force behavior). Interactive init without force is unchanged.

---

### Edge Cases

- Cancelling an interactive confirm still must not call the operation (no persist, no GitLab create).
- Partial sync (some edges declined or failed) still returns the same per-edge outcomes as today; sharing the operation must not introduce bulk all-or-nothing.
- Scripted `-y` still means “confirm every edge” and still prompts to open browsers at the end when URLs exist.
- Surfaces may still collect input differently (flags vs pickers vs inline fields in this spec). This spec does not restyle the main tree.
- Existing error messages that users rely on keep the same meaning; wording may be unified only when two surfaces currently disagree, and then both must use the stricter, more complete rule (unlink edge checks).
- Browser opening after sync remains a surface concern in this spec; the shared operation returns URLs, it does not open tabs. (Moving open-out of the sync package is spec 014.)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: There MUST be exactly one link operation that validates, mutates the tree, and persists. Every user-facing link path MUST call it.
- **FR-002**: There MUST be exactly one unlink operation that validates (including “edge exists”), removes the subtree, and persists. Every user-facing unlink path MUST call it.
- **FR-003**: There MUST be exactly one merge-request create operation that validates branches, checks GitLab authentication, and creates or skips. Manual MR and sync MUST both use it for the actual create.
- **FR-004**: There MUST be exactly one sync operation that resolves edges for the requested direction, processes each edge (optional per-edge confirm), and returns a per-edge summary. Scripted and interactive sync MUST use it for processing; interactive MUST NOT perform a separate GitLab auth check before calling it.
- **FR-005**: There MUST be exactly one init/register operation used by interactive and scripted init.
- **FR-006**: Shared operations MUST live as deepenings of the existing project, tree, sync, and merge-request areas — not a new orchestration layer.
- **FR-007**: After this change, 100% of previously specified user-visible outcomes for link, unlink, sync, MR, and init MUST still hold (flags, keys, copy, cascade order, skip-if-open, browser prompt timing).
- **FR-008**: A change to a validation or persist rule in the shared operation MUST be observable from every surface that offers that action, without editing each surface’s copy of the rule.
- **FR-009**: Interactive screens MUST treat operation errors as display data (show the returned message). They MUST NOT construct a GitLab session of their own for auth or create.
- **FR-010**: This spec MUST leave the product fully usable: every command and the main tree still run. It MUST NOT require the later file-split or TUI-embed specs.

### Key Entities

- **Use case / operation**: The single owner of one user action (link, unlink, sync, create merge request, init). Input in, result or error out, including persist and remote calls that belong to that action.
- **Surface**: A way the user invokes an operation — scripted command, dedicated interactive command, or the main tree. Surfaces collect input and render results; they do not own rules.
- **Outcome**: Created, skipped (already open or user-declined), or failed — already used by merge requests and sync; this spec does not rename them (that is spec 015).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For link, unlink, sync (one edge), manual MR, and init, a reviewer can point to a single operation that all surfaces call; zero surfaces retain a private copy of validate-then-persist or validate-then-create.
- **SC-002**: Existing product scenarios (scripted and interactive) pass with the same outcomes as before this spec: 0 intended behavior changes.
- **SC-003**: With GitLab logged out, 100% of sync and MR surfaces fail on auth and create 0 merge requests.
- **SC-004**: An invalid unlink that only the scripted path used to catch is now rejected on interactive unlink as well (0 extra “success then corrupt tree” paths).
- **SC-005**: A contributor can change duplicate-edge rejection once and demonstrate the new behavior on scripted link and on interactive link in one test pass, without editing two rule implementations.

## Assumptions

- End users of branchy are not the primary beneficiaries; contributors are. User-visible behavior stays frozen except where two surfaces currently disagree, in which case the complete validation wins.
- “Deepen existing packages” means adding or tightening functions next to `project`, `tree`, `sync`, and merge-request create — not introducing `app` or `usecase` packages.
- Sync still accepts a per-edge confirm callback (or equivalent) so interactive and `-y` can share one walk.
- Returning URLs from sync without opening the browser is in scope; relocating the open helper is spec 014.
- Main-tree link/unlink screens stay as they are in this spec; embedding the standalone wizards is spec 013.
- Command files stay in one CLI module; splitting them is spec 012.

## Decisions (Grilling Session 2026-08-21)

| Topic | Decision |
|-------|----------|
| Shipping | Five specs, this first; each merge leaves CLI/TUI fully working |
| Behavior | Outcome freeze; no flag/key/copy redesign |
| Operation home | Deepen existing packages; no new application layer |
| GitLab auth | Owned by sync and MR create; surfaces only display returned errors |
| Later specs | Cobra split, TUI embed, presentation split, type unify are out of this spec |
| Link/unlink persist | `project.Link` / `project.Unlink` wrap tree mutate + save; `tree.Link` / `UnlinkSubtree` stay exported in-memory primitives |
| Unlink shape | Parent + child; require the edge when parent is given; TUI passes `ParentOf(child)`; unlinking a root is allowed (empty parent) |
| Sync sharing | Keep `sync.Run` (CLI batch) and `sync.RunEdge` (TUI per confirm); TUI skip-by-user goes through a shared helper; TUI does not import GitLab |
| Interactive auth timing | After the root is chosen, via a sync-package begin (auth + collect edges) — same moment as today, without a TUI-owned GitLab client |
