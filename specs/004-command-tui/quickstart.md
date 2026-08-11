# Quickstart: Unified Command TUI

**Feature**: `004-command-tui` | **Date**: 2026-08-11

Validation guide for proving dedicated command TUIs work end-to-end after implementation.

## Prerequisites

- Go 1.22+ installed
- [glab](https://gitlab.com/gitlab-org/cli) authenticated: `glab auth login`
- A registered branchy project with a multi-edge branch tree
- Terminal ≥ 80×24

## Setup

```bash
go build -o branchy ./cmd/branchy
```

## Scenario 1: Standalone sync TUI (P1)

```bash
./branchy sync
```

**Steps**:
1. TUI opens (not numbered text menu)
2. Pick root branch from searchable list
3. Confirm/decline edges with `y`/`n`
4. Review summary; optional browser prompt

**Expected**:
- Full-screen TUI with `branchy sync` title
- Returns to shell on completion (not main tree app)
- Same outcomes as TUI `s` key for same choices

## Scenario 2: Sync with flag stays plain CLI (P1)

```bash
./branchy sync --from develop
```

**Expected**:
- No TUI; numbered/plain stdin flow (existing behavior)
- Per-edge prompts on stdin

## Scenario 3: Link TUI (P1)

```bash
./branchy link
```

**Steps**:
1. Pick parent from branch list
2. Type child branch name
3. Confirm `y`
4. Exit to shell

**Expected**:
- Edge appears in tree when re-opening `branchy`
- No raw dual-field typing screen

## Scenario 4: Link scripted (P1)

```bash
./branchy link develop feature-new
```

**Expected**:
- One-line `Linked develop → feature-new`
- No TUI

## Scenario 5: Unlink TUI (P1)

```bash
./branchy unlink
```

**Steps**:
1. Pick branch from list
2. Confirm subtree count message
3. Press `y`

**Expected**:
- Subtree removed; success screen; exit to shell

## Scenario 6: Init wizard (P1)

Inside an unregistered git repo:

```bash
./branchy init
```

**Expected**:
- Confirm registration screen
- Success shows project id and branch count
- `branchy init --force` remains plain one-liner (no TUI)

## Scenario 7: Projects browser (P1)

```bash
./branchy projects
```

**Expected**:
- Searchable list with id + path
- Enter shows detail; does not open main tree app

## Scenario 8: Non-TTY no hang (P1)

```bash
./branchy sync 2>&1 | cat
# expect error about --from required

./branchy link 2>&1 | cat
# expect error about parent child required

./branchy projects | cat
# expect plain tab-separated list
```

**Expected**: Process exits immediately; no TUI wait.

## Scenario 9: MR any-flag rule (P1)

```bash
./branchy mr --source develop
```

**Expected**:
- Plain CLI error or usage (not TUI target picker)

```bash
./branchy mr
```

**Expected**:
- TUI flow (unchanged)

## Scenario 10: Visual consistency (P2)

Run in sequence: `./branchy sync`, `./branchy mr`, `./branchy link`.

**Expected**:
- Same title/help/confirm styling across flows
- Confirm screens use `y`/`n`/`esc` consistently

## Scenario 11: Terminal too small (edge)

Resize terminal below 80×24 before running `./branchy sync`.

**Expected**:
- Readable "too small" message; `q` exits

## Scenario 12: Unit tests

```bash
go test ./internal/cli/... -run 'UseTUI|IsTTY'
go test ./internal/tui/... -run 'Sync|Link|Unlink|Init|Projects'
```

**Expected**: All tests pass.

## Scenario 13: Regression — main app unchanged

```bash
./branchy
```

**Expected**:
- Main tree TUI works; `s`, `m`, `l`, `u` shortcuts unchanged

## References

- Contract: [contracts/command-tui.md](./contracts/command-tui.md)
- Data model: [data-model.md](./data-model.md)
- Feature spec: [spec.md](./spec.md)
