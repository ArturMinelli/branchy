# Quickstart: Interactive Sync Confirmations

**Feature**: `002-interactive-sync-confirm` | **Date**: 2026-08-11

Validation guide for proving per-edge sync confirmation and optional ordered browser open work end-to-end after implementation.

## Prerequisites

- Go 1.22+ installed
- [glab](https://gitlab.com/gitlab-org/cli) authenticated: `glab auth login`
- A registered branchy project with a branch tree containing at least 3 parent→child edges below one root (for selective sync testing)
- Network access to GitLab

## Setup

```bash
go build -o branchy ./cmd/branchy
```

Example tree (3 edges below `main`):

```yaml
branches:
  main:
    children: [develop]
  develop:
    children: [release, feature-a]
  release:
    children: []
  feature-a:
    children: []
```

## Scenario 1: TUI per-edge confirm (P1)

```bash
./branchy
```

**Steps**:
1. Select project → navigate tree to `main`
2. Press `s`
3. At first edge prompt, press `y`
4. At second edge prompt, press `n` (decline)
5. At third edge prompt, press `y`
6. Review summary — only accepted edges show `created` or `skipped` (existing MR)
7. At browser prompt, press `y`

**Expected**:
- No bulk "confirm entire sync" screen before first edge
- Declined edge shows `skipped (skipped by user)`; sync continues
- Browser prompt appears only if ≥1 MR was created
- Tabs open in DFS order for created MRs only (declined edge excluded)
- Press Enter to return to tree

## Scenario 2: TUI cancel mid-sync (P1)

**Steps**:
1. Press `s` on a branch with multiple edges
2. Confirm first edge (`y`)
3. Press `Esc` on second edge prompt

**Expected**:
- Partial summary shows first edge result
- No further edges processed
- Browser prompt if first edge created an MR

## Scenario 3: CLI per-edge + end browser (P2)

```bash
./branchy sync --from main
```

**Steps**:
1. Answer `y` / `n` per edge at stdin prompts
2. At end: `Open created MRs in browser? [y/N]` → answer `y`

**Expected**:
- No automatic browser open during edge loop
- Browser opens only after end confirmation
- Created-only URLs, DFS order

## Scenario 4: CLI `-y` skips per-edge only (P2)

```bash
./branchy sync --from main -y
```

**Expected**:
- No per-edge `[y/N]` prompts
- End browser prompt still appears if any MR created
- No automatic browser open without end confirmation

## Scenario 5: All declined — no browser prompt (P1)

**TUI or CLI**: decline/skip every edge.

**Expected**:
- Summary shows all edges skipped
- No browser prompt
- Exit code 0

## Scenario 6: No child branches

```bash
./branchy sync --from feature-a   # leaf branch with no children
```

**Expected**:
- `No child branches below "feature-a".`
- No prompts, exit 0

## Scenario 7: Regression — manual MR unchanged

```bash
./branchy mr
# and
./branchy   # press m
```

**Expected**:
- Manual MR flow unchanged (per-MR browser prompt, not batch)
- No sync bulk confirm introduced in MR flow

## Scenario 8: Unit tests

```bash
go test ./internal/sync/... -run 'CreatedURLs|OpenURLs'
go test ./internal/tui/... -run Sync
```

**Expected**: All tests pass.

## References

- CLI contract: [contracts/sync-cli.md](./contracts/sync-cli.md)
- Data model: [data-model.md](./data-model.md)
- Feature spec: [spec.md](./spec.md)
