# Feature Specification: Unify Domain Types

**Feature Branch**: `015-unify-domain-types`

**Created**: 2026-08-21

**Status**: Draft

**Input**: User description: "Unify merge direction and created/skipped/failed outcomes so CLI and interactive UI share one vocabulary. One error policy for merge-request create. Behavior freeze. Independently mergeable as the last of five."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One merge direction everywhere (Priority: P1)

A contributor working on “which way does sync send work” sees a single direction concept used by the main tree, the standalone sync picker, edge walking, and MR source/target selection. They do not convert between two private direction types. Users still see inbound/outbound (or down/up) with the same keys and arrows as today.

**Why this priority**: Two direction types are the leftover dual model after features 008–010. Unifying them is what makes later direction bugs fixable once.

**Independent Test**: Set outbound on the main tree, start sync, confirm child→parent edges. Set inbound, confirm parent→child. Standalone picker Control+Up / Control+Down still maps to the same two directions without a translation table in the UI.

**Acceptance Scenarios**:

1. **Given** the main tree in inbound, **When** the user starts sync, **Then** edges are processed parent → child (downward) using the same direction value the tree holds.
2. **Given** the main tree in outbound, **When** the user starts sync, **Then** edges are processed child → parent (upward) with no UI-side remapping type.
3. **Given** standalone interactive sync, **When** the user presses Control+Up or Control+Down, **Then** pending direction is the same concept as the main tree’s inbound/outbound.
4. **Given** scripted sync, **When** it runs, **Then** it remains downward-only and still has no direction flag.

---

### User Story 2 - One outcome vocabulary for create and sync (Priority: P1)

A contributor reading a sync summary and a manual MR result sees the same three outcomes: created, skipped, failed. “Skipped” still distinguishes already-open vs user-declined where the product already does. Scripted printers and interactive summaries switch on that shared vocabulary, not on free-form strings that can typo apart.

**Why this priority**: Sync results currently store actions as loose text while MR create uses named constants; printers duplicate the same three words.

**Independent Test**: Create, skip-already-open, skip-declined, and fail one edge each. Scripted output and interactive summary classify them the same way. A test cannot pass a misspelled action that the other surface would treat differently.

**Acceptance Scenarios**:

1. **Given** a newly created merge request, **When** any surface reports it, **Then** it is classified as created and shows the URL.
2. **Given** an already-open merge request, **When** sync or manual create runs, **Then** it is classified as skipped (already open), not failed.
3. **Given** the user declines a sync edge, **When** the summary is shown, **Then** it is skipped (user declined) and is not offered for browser open.
4. **Given** a remote failure, **When** the summary is shown, **Then** it is failed with the error meaning, on every surface.

---

### User Story 3 - One error policy for merge-request create (Priority: P2)

A contributor calling create knows: invalid input (unknown branch, same source/target) is an error that aborts the call; GitLab “already exists” is a skipped outcome; GitLab transport/create failure is a failed outcome. They do not have to handle both “error return” and “failed outcome” for the same class of problem.

**Why this priority**: Callers currently mix `error` and a failed result for similar cases, which is why CLI and TUI branch differently.

**Independent Test**: Unknown branch → call fails before GitLab. Already-open → skipped result, no error. Create failure → failed result (or a single documented error if the operation cannot produce a result). Scripted MR and interactive MR treat each case the same.

**Acceptance Scenarios**:

1. **Given** source or target missing from the tree, **When** create is attempted, **Then** every surface reports a validation failure and does not call GitLab.
2. **Given** an open MR already exists, **When** create is attempted, **Then** every surface reports skipped with the URL, not an error exit for that pair.
3. **Given** GitLab create fails and no open MR can be recovered, **When** create is attempted, **Then** every surface reports failed with the message, and scripted `mr` still exits non-zero for that failure as today.
4. **Given** sync processing an edge, **When** create returns skipped or failed, **Then** the cascade continues; a validation error on the from-branch still aborts the run as today.

---

### Edge Cases

- User-visible words (inbound, outbound, created, skipped, failed, help lines) stay as they are; this is a contributor-facing unify.
- Direction remains session-scoped on the main tree and run-scoped on standalone sync.
- Openable URL rules stay: created + skipped-already-open; exclude user-declined and failed without URL.
- Do not introduce a GitLab interface layer in this spec.
- Scripted `-y` still auto-confirms edges; declined cannot happen there.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Merge direction MUST be a single shared concept used by the main tree, standalone sync picker, edge collection, and source/target selection.
- **FR-002**: The interactive layer MUST NOT keep a private parallel direction type that has to be converted at the sync boundary.
- **FR-003**: Create and sync MUST share one outcome set: created, skipped, failed, including the existing skipped subtypes (already open vs user declined) where needed.
- **FR-004**: Surfaces MUST classify results using that shared set, not ad-hoc strings that can diverge.
- **FR-005**: Merge-request create MUST use one error policy: validation problems abort with an error; already-open is skipped; unrecovered create problems are failed. Every surface MUST follow that policy.
- **FR-006**: User-visible keys, arrows, copy, cascade order, and exit codes MUST remain as specified in earlier product specs.
- **FR-007**: After merge, the product MUST remain fully usable; this is the last of the five architecture specs.

### Key Entities

- **Direction**: Inbound/downward (parent → child) or outbound/upward (child → parent). One value, two user-facing names already in help.
- **Outcome**: Created, skipped, or failed for one merge-request attempt.
- **Skip reason**: Already-open vs user-declined, required so browser-open filtering stays correct.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 0 conversion helpers between two direction types remain at the tree↔sync boundary.
- **SC-002**: 100% of sync and MR printers/summaries classify outcomes from the shared set; 0 raw unmatched action strings in those paths.
- **SC-003**: Validation vs already-open vs create-failure are each handled one way across scripted MR, interactive MR, and sync (0 mixed error/result for the same class).
- **SC-004**: Existing direction, sync, and MR product scenarios still pass (0 user-visible regressions).
- **SC-005**: A contributor can add a new skip reason in one place and have both scripted and interactive summaries respect browser-open filtering without a second string table.

## Assumptions

- Specs 011–014 are merged.
- User-facing language stays “inbound/outbound” and “created/skipped/failed”; internal names may match either pair as long as there is one type.
- No new `app` package and no GitLab interface; this spec only unifies types and error policy.
- Scripted sync remains downward-only.

## Decisions (Grilling Session 2026-08-21)

| Topic | Decision |
|-------|----------|
| Shipping | Fifth of five; merge leaves the product working |
| Behavior | Outcome freeze; no new direction flags or key changes |
| Operation home | Types live with existing sync / MR / tree areas (deepen, not a new layer) |
| Auth / presentation | Already handled in 011 and 014 |

## Decisions (Grilling Session 2026-08-24)

| Topic | Decision |
|-------|----------|
| Plan target | 015 (014 already planned and implemented) |
| Direction type | Keep `sync.Direction` (`Downward`/`Upward`); TUI drops `diffDirection` and uses `sync.Direction` |
| Outcomes | Typed `mr.Action` on `CreateResult` and `sync.Result`; add `SkipReason`; `OpenableURLs` uses `SkipReason`, not `Message` |
| Create errors | Validation/auth → error; already-open → skipped+nil; GitLab fail → failed result+nil; CLI `mr` maps failed → non-zero exit |
