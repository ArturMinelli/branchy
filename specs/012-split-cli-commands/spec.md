# Feature Specification: Split CLI Commands

**Feature Branch**: `012-split-cli-commands`

**Created**: 2026-08-21

**Status**: Draft

**Input**: User description: "Split the command-line module so each command lives in its own file; keep handlers thin; keep a thin program entry. Depends on shared use cases already merged. Behavior freeze. Independently mergeable."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Edit one command without opening the others (Priority: P1)

A contributor needs to change `sync` flags, help text, or how results are printed. They open the sync command definition and do not have to scroll through init, mr, link, unlink, or projects. A mistake in the sync file cannot be “because I edited the giant command file.”

**Why this priority**: One command file for every subcommand is the readability problem this spec exists to fix.

**Independent Test**: Change a sync-only help string or flag description. Confirm no other command’s help text changes. Run `branchy --help` and each subcommand `--help` to verify isolation.

**Acceptance Scenarios**:

1. **Given** the command-line module after this spec, **When** a contributor looks up `sync`, **Then** its use line, flags, and run path are in a dedicated command definition, not mixed with other commands’ bodies.
2. **Given** a change to `projects` listing format, **When** that change is made, **Then** it does not require editing the `sync` or `mr` command definitions.
3. **Given** `branchy` with no subcommand, **When** it runs on a TTY, **Then** the main interactive tree still launches as today.

---

### User Story 2 - Command handlers only parse, call, and print (Priority: P1)

A contributor reading a command’s run path sees: read flags/args, choose interactive vs scripted, call the shared operation from spec 011, print the result. They do not see tree mutation, GitLab calls, or unlink edge walks inlined in the command.

**Why this priority**: File split without thin handlers just scatters the same mess. Spec 011 made operations; this spec makes commands use them and nothing else.

**Independent Test**: Read each command’s run path. Confirm there is no direct tree mutate-and-save and no GitLab client construction. Scripted and interactive behavior still matches pre-spec 012 (and spec 011) scenarios.

**Acceptance Scenarios**:

1. **Given** scripted `link parent child`, **When** it runs, **Then** it calls the shared link operation and prints the same one-line success or error as today.
2. **Given** scripted `unlink parent child`, **When** it runs, **Then** it calls the shared unlink operation (no extra private validation block) and prints the same summary.
3. **Given** scripted `sync` / `mr` / `init` / `projects`, **When** they run, **Then** outcomes, prompts, and exit meaning match the previous spec.
4. **Given** interactive mode (TTY, no flags, no full positionals), **When** a command runs, **Then** it still launches the same dedicated interactive flow as today.

---

### User Story 3 - Shared TTY and prompt rules stay in one place (Priority: P2)

Interactive-vs-scripted detection (TTY and “any flag set”) and yes/no stdin prompts remain shared. Each command file does not reimplement “is this a terminal” or “read y/N.”

**Why this priority**: Splitting files would otherwise copy the mode helper six times.

**Independent Test**: Pipe a command to a non-TTY and pass a flag; confirm scripted mode. Run with no flags on a TTY; confirm interactive mode. Both still use the same rule as today.

**Acceptance Scenarios**:

1. **Given** stdout is not a TTY, **When** the user runs a command that would otherwise be interactive, **Then** the scripted path is used (no full-screen interactive UI).
2. **Given** any flag is set, **When** the command runs, **Then** the scripted path is used even on a TTY.
3. **Given** two scripted commands that ask y/N, **When** the user answers `y` or `n`, **Then** interpretation is the same (`y`/`yes` vs anything else).

---

### Edge Cases

- Program entry stays a few lines: invoke the command tree, print errors to stderr, non-zero exit on failure.
- Command registration still exposes the same subcommand names: `init`, `sync`, `mr`, `link`, `unlink`, `projects`.
- Flag names, shorthands (`-y`), and positional arity do not change.
- Tests that invoke commands must not depend on a single mutable global command object shared across tests if that already causes order-dependent failures; constructing the root command per test is allowed and preferred.
- This spec does not move commands out of the existing command-line module into the program directory.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each user-facing subcommand MUST be defined in its own command file inside the existing command-line module.
- **FR-002**: The program entry MUST remain a thin wrapper that runs the command tree and reports errors.
- **FR-003**: A command run path MUST NOT mutate the branch tree, persist the store, or talk to GitLab except by calling the shared operations from spec 011.
- **FR-004**: Interactive-vs-scripted selection MUST remain: TTY and no explicit flags (and no full positionals where that rule already applies).
- **FR-005**: Shared prompt and TTY helpers MUST NOT be copy-pasted into each command file.
- **FR-006**: User-visible command names, flags, help purpose, output lines, and exit behavior MUST match the previous revision.
- **FR-007**: After merge, every command and the main tree MUST still run without requiring spec 013–015.

### Key Entities

- **Command definition**: One subcommand’s name, flags, arguments, help, and run path.
- **Command tree**: The assembled `branchy` program including the default (no-subcommand) interactive tree.
- **Scripted vs interactive mode**: Existing rule — interactive only when the terminal is interactive and the user did not pass flags (or full positionals).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A contributor can name the file to open for each of the six subcommands without opening a combined “all commands” source file.
- **SC-002**: 100% of existing command-line product scenarios still pass (help, flags, scripted output, interactive launch conditions).
- **SC-003**: Review of each command run path shows 0 inline tree persist or GitLab session constructions.
- **SC-004**: Changing sync-only help does not alter the `projects` or `mr` help text (0 unintended string coupling).

## Assumptions

- Spec 011 is already merged: shared operations exist for commands to call.
- Layout stays in the existing command-line module (`internal/cli` equivalent); the program directory is not a new home for command files.
- “Thin handler” means parse → operation → print, including the existing stdin prompts for scripted sync/MR confirmations and the end-of-sync browser question.
- No new flags or command aliases.

## Decisions (Grilling Session 2026-08-21)

| Topic | Decision |
|-------|----------|
| Shipping | Second of five; merge leaves the product working |
| Behavior | Full outcome freeze |
| Layout | Keep command-line module, one file per command; program entry stays thin |
| Operations | Already owned by spec 011; this spec only relocates command wiring |
| File shape | Flat `internal/cli/`: `root.go`, `mode.go`, `init.go`, `sync.go`, `mr.go`, `link.go`, `unlink.go`, `projects.go` |
| Scripted helpers | Private in the same file as the command (`runSyncCLI` in `sync.go`, `runMRFlags` in `mr.go`); no shared `prompt.go` |
| Feature context | `.specify/feature.json` → `specs/012-split-cli-commands` for plan/tasks |
| Sync tree reads | Read-only `p.Tree.Names()` / `p.Tree.CollectEdges()` allowed in `sync.go` only (outcome freeze for picker + plan print) |
| Grep audit | Narrow forbidden patterns; document `sync.go` read-only exception explicitly |
| Plan artifacts | Confirm existing Phase 0–1 docs; patch only for grilling deltas |
| CLI tests | `mode_test.go` + grep audit only; no per-command test files |
| tasks.md | Unchecked executable checklist for implement/verify (not a completed audit log) |
| sync.go `IsTTY()` | Direct call in `runSyncCLI` for interactive `--from` picker allowed (outcome freeze; not duplicated mode logic) |
