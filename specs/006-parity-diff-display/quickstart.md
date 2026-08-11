# Quickstart: Parent Diff Display

**Feature**: `006-parity-diff-display` | **Date**: 2026-08-11

Validation guide for inbound file-change counts on the main tree and sync TUI. Comparison rules: [contracts/inbound-count.md](./contracts/inbound-count.md). Entities: [data-model.md](./data-model.md).

## Prerequisites

- Go 1.26+ installed
- A registered branchy project whose git repo has local branches matching the tree
- At least three parent→child pairs with **different** inbound file-change sizes (including one already in sync if possible)
- Terminal ≥ 80×24

## Setup

```bash
go build -o branchy ./cmd/branchy
go test ./internal/git/... ./internal/tui/... -count=1
```

Automated tests cover three-dot semantics, hide-zero, unknown placeholder, and confirm-zero. Scenarios below are the manual visual check.

---

## Scenario 1: Main tree badges

```bash
./branchy
```

Open the project tree.

**Expected**:

- Each non-root branch that would receive files from its parent shows a muted number after its name
- The number matches GitLab’s files-changed count for an MR parent → child (spot-check one pair on GitLab)
- Roots show no inbound number
- Children already matching their parent show no `0`
- Largest pending target is obvious in a few seconds

---

## Scenario 2: Sync root picker

From the tree press `s`, or run:

```bash
./branchy sync
```

**Expected**:

- Listed non-root branches show the same numbers as the main tree
- Roots and in-sync children have no badge
- Filtering still matches bare branch names

---

## Scenario 3: Edge confirm includes count

Select a sync root and reach the first `Create MR {parent} → {child}?` panel.

**Expected**:

- Context includes `{N} files would change on {child}`
- `N` matches the tree/picker badge for that child
- If that pair is already in sync, the line still says `0 files would change on {child}`
- Yes/No behavior is unchanged

---

## Scenario 4: Unknown comparison

Rename or delete a local branch that still appears in the tree, then reopen the main TUI and sync.

**Expected**:

- That child’s row/picker shows `?` (not `0`)
- Edge confirm for that pair says `File count unavailable`
- Other branches still show their numbers
- Navigation, starting sync, skipping, and quitting still work

---

## Scenario 5: Scripted sync unchanged

```bash
./branchy sync --from <root> -y
```

**Expected**:

- Plain CLI only — no badges, no “files would change” lines
- Existing stdin/flag behavior unchanged
