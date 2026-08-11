# Quickstart: Manual MR Command

**Feature**: `001-mr-command` | **Date**: 2026-08-11

Validation guide for proving the manual MR feature works end-to-end after implementation.

## Prerequisites

- Go 1.22+ installed
- [glab](https://gitlab.com/gitlab-org/cli) authenticated: `glab auth login`
- A git repo registered with branchy containing at least two branches in `branch-tree.yaml`
- Network access to GitLab

## Setup

```bash
# From repo root
go build -o branchy ./cmd/branchy

# Inside a registered git repo (or register one)
branchy init
```

Ensure `branch-tree.yaml` has ≥2 branches, e.g.:

```yaml
branches:
  develop:
    children: []
  feature-x:
    children: []
```

## Scenario 1: Full TUI flow (P1)

```bash
./branchy mr
```

**Steps**:
1. Select source branch → Enter
2. Select target branch → Enter
3. Edit title (or accept default) → Enter
4. Confirm with `y`
5. At browser prompt, press `y` or `n`

**Expected**:
- MR created on GitLab with chosen source/target/title
- URL displayed in terminal
- Browser opens only if you answered `y`
- Exit code 0

## Scenario 2: Flag mode (P2)

```bash
./branchy mr --source feature-x --target develop --yes
```

**Expected**:
- No TUI launched
- stdout shows `Created: feature-x → develop` and MR URL
- No browser opened
- Exit code 0

## Scenario 3: Flag mode with confirmation

```bash
./branchy mr --source feature-x --target develop
# Answer y at prompt
```

**Expected**:
- stdin prompt `Create MR feature-x → develop? [y/N]`
- MR created on `y`
- Exit code 0

## Scenario 4: Main TUI shortcut (P2)

```bash
./branchy
```

**Steps**:
1. Navigate tree to select a branch
2. Press `m`
3. Complete target → title → confirm → browser prompt

**Expected**:
- Source pre-filled from selected branch
- `Esc` during flow returns to tree view without creating MR
- Help line shows `m: mr`

## Scenario 5: Existing MR skip

```bash
# Run twice with same source/target
./branchy mr --source feature-x --target develop --yes
./branchy mr --source feature-x --target develop --yes
```

**Expected (second run)**:
- `Skipped: feature-x → develop (open MR already exists)`
- Existing MR URL printed
- Exit code 0

## Scenario 6: Validation errors

```bash
# Same branch
./branchy mr --source develop --target develop --yes
# Expected: error "source and target must differ", exit 1

# Unknown branch
./branchy mr --source nonexistent --target develop --yes
# Expected: error "branch \"nonexistent\" not in tree", exit 1
```

## Scenario 7: Unit tests

```bash
go test ./internal/mr/...
go test ./internal/tui/... -run MR
```

**Expected**: All tests pass.

## References

- CLI contract: [contracts/mr-cli.md](./contracts/mr-cli.md)
- Data model: [data-model.md](./data-model.md)
- Feature spec: [spec.md](./spec.md)
