# Quickstart: Shared Use Cases

**Feature**: `011-shared-use-cases` | **Date**: 2026-08-21

Validation guide. Rules: [contracts/shared-operations.md](./contracts/shared-operations.md). Entities: [data-model.md](./data-model.md).

## Prerequisites

- Go 1.26+ installed
- A throwaway git repo you can register (or `HOME` pointed at a temp dir for automated tests)
- `glab` available for scenarios 4–5; logged out for the auth check
- Terminal ≥ 80×24 for TUI scenarios

## Setup

```bash
go test ./internal/project/... ./internal/sync/... ./internal/mr/... ./internal/tui/... ./internal/cli/... -count=1
go build -o branchy ./cmd/branchy
```

Automated tests should cover: `project.Link` persist + duplicate edge; `project.Unlink` edge miss / root ok / count; `sync.Begin` empty vs auth; `SkippedByUser` shape; TUI/CLI no longer calling `Tree.Link`+save. Scenarios below are the cross-surface check.

---

## Scenario 1: Link is one operation

From a registered repo:

```bash
./branchy link main feature-shared-uc
# then, in another checkout of the same project id, or after unlink, repeat via TUI:
./branchy link          # standalone wizard, same pair
# and from the main tree: press l, enter the same pair
```

**Expected**:

- All three persist the same `main → feature-shared-uc` edge
- Second attempt from any surface: `edge already exists` (same meaning), store unchanged
- Cancel on a TUI confirm: no persist

---

## Scenario 2: Unlink edge check on every surface

Tree with `main → feature-a → feature-b`.

```bash
./branchy unlink main feature-missing    # unknown child
./branchy unlink other feature-a         # not an edge
./branchy unlink main feature-a          # success
```

Then rebuild the edge and unlink the same pair from standalone `./branchy unlink` and from main-tree `u`.

**Expected**:

- Missing child / missing edge fail with the same meaning on CLI **and** on both interactive paths (this is the intended unification)
- Successful unlink removes the same subtree members; reload shows them gone
- Unlinking a **root** from TUI still works (empty parent path)

---

## Scenario 3: Init still one register

```bash
./branchy init --force     # scripted, already registered
# vs interactive: ./branchy init  on a TTY, confirm
```

**Expected**:

- Without force, both refuse already-registered with the same meaning
- With `--force`, scripted reuses the project id
- Interactive cancel does not write

---

## Scenario 4: Sync/MR auth is not TUI-owned

Log out of `glab`. Pick a root that **has** descendants.

```bash
./branchy sync --from main -y
./branchy sync              # TUI: choose that root
./branchy mr --source main --target develop -y
```

**Expected**:

- Scripted sync/MR fail with `glab auth` meaning and create 0 MRs
- Interactive sync: error **after the root is chosen**, before the first edge Yes/No — not a TUI-constructed client
- Interactive empty subtree (root with no edges): still the empty screen, no auth (same as today)

---

## Scenario 5: Create/skip outcomes match

With auth OK and an open MR already on a pair:

```bash
./branchy mr --source <s> --target <t> -y
./branchy sync --from <parent> -y
```

**Expected**:

- Both report skipped (already open) + URL, not failed
- Decline one TUI edge: skipped by user; later edges still run; that URL is not in the open-in-browser list
- `-y` still confirms every edge; browser prompt still at the end when openable URLs exist

---

## Done when

- [ ] Scenarios 1–5 pass
- [ ] `go test` packages listed in Setup are green
- [ ] `internal/tui` has no `gitlab` import
- [ ] `internal/cli` scripted link/unlink have no private validate+`SaveTree` block
