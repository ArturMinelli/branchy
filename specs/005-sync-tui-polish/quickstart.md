# Quickstart: TUI Native Confirm & Loading

**Feature**: `005-sync-tui-polish` | **Date**: 2026-08-11

Validation guide for proving shared confirm panels and loading states work across all in-scope TUI flows.

## Prerequisites

- Go 1.26+ installed
- [glab](https://gitlab.com/gitlab-org/cli) authenticated: `glab auth login`
- A registered branchy project with a multi-edge branch tree (≥2 edges below one root)
- Terminal ≥ 80×24, true-color or 256-color recommended

## Setup

```bash
go build -o branchy ./cmd/branchy
go test ./internal/tui/... -count=1
```

---

## Scenario 1: Sync confirm panel (standalone)

```bash
./branchy sync
```

**Steps**:

1. Select a root branch with multiple child edges
2. At first edge prompt, observe framed confirm with `[ No ]` / `[ Yes ]` buttons (no `[y/N]` text)
3. Press `→` to focus Yes, press `Enter`
4. Observe animated spinner with edge name and "Edge 1 of N"
5. Press random keys during spinner — nothing should happen
6. After completion, next edge confirm appears automatically

**Expected**:

- Default focus on **No** (highlighted)
- Spinner animates during MR creation
- Functional outcomes match pre-feature sync for same y/n choices

---

## Scenario 2: Sync browser confirm + loading

Complete a sync creating ≥1 MR. At summary, press `Enter` to reach browser step.

**Steps**:

1. Confirm panel asks to open MRs (not `[y/N]` text)
2. Select Yes
3. Brief loading state while tabs open
4. Flow exits cleanly

**Expected**:

- Tabs open in DFS order among created MRs only
- Declined/skipped edges excluded

---

## Scenario 3: Embedded sync (`s` key)

```bash
./branchy
```

**Steps**:

1. Select branch with children, press `s`
2. Verify confirm/loading UI matches Scenario 1

**Expected**:

- Identical confirm and loading chrome to standalone `branchy sync`
- Returns to tree view on completion (not shell)

---

## Scenario 4: MR wizard confirm + loading

```bash
./branchy mr
```

**Steps**:

1. Pick source and target branches, enter title
2. At confirm screen, use arrow keys + Enter
3. Observe spinner during MR creation
4. At browser prompt, confirm panel (not text prompt)

**Expected**:

- No UI freeze during `mr.Create` network call
- Same confirm styling as sync

---

## Scenario 5: Link and unlink confirms

```bash
./branchy link
./branchy unlink
```

**Steps**:

1. Complete each flow to the confirm step
2. Verify shared confirm panel (no loading on save — instant)

**Expected**:

- Native confirm panel on both flows
- Save/remove completes immediately after Yes (no spinner)

---

## Scenario 6: Init confirm + loading

From an unregistered git repo:

```bash
cd /path/to/unregistered-repo
/path/to/branchy init
```

**Steps**:

1. Confirm panel for registration (not `[y/N]`)
2. Select Yes
3. Spinner during registration
4. Success screen

**Expected**:

- Loading during `project.Init`
- No spinner on No/cancel

---

## Scenario 7: Tree unlink confirm

```bash
./branchy
```

**Steps**:

1. Select a branch with children, press `u`
2. Verify shared confirm panel (same as `branchy unlink`)

**Expected**:

- No `[y/N]` in unlink warning text
- Yes removes subtree; No returns to tree

---

## Scenario 8: Scripted CLI unchanged

```bash
./branchy sync --from develop
./branchy link parent child
echo "n" | ./branchy sync --from develop
```

**Expected**:

- Plain stdin `[y/N]` prompts (sync)
- One-line output (link)
- No spinner output in non-TTY pipe
- No full-screen TUI

---

## Scenario 9: Visual consistency checklist

Run Scenarios 1, 4, 5, 6, 7 in sequence. For each confirm screen verify:

- [ ] Same border style
- [ ] Same button layout (No left, Yes right)
- [ ] Same focus highlight
- [ ] Same help footer pattern
- [ ] No `[y/N]` literal in question

For each loading screen verify:

- [ ] Same spinner style
- [ ] Message above or beside spinner consistently
- [ ] Input blocked during load

---

## Automated tests

```bash
go test ./internal/tui/ -run 'Confirm|Loading|Sync|MR|Init|Link|Unlink' -v -count=1
```

**Expected**: All tests pass; confirm tests cover focus navigation and shortcuts; flow tests assert views lack `[y/N]`.
