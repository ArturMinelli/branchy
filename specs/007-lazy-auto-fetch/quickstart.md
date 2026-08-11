# Quickstart: Lazy Auto-Fetch

**Feature**: `007-lazy-auto-fetch` | **Date**: 2026-08-11

Validation guide for background default-remote updates. Rules: [contracts/lazy-fetch.md](./contracts/lazy-fetch.md). Counts: [006 inbound-count](../../006-parity-diff-display/contracts/inbound-count.md).

## Prerequisites

- Go 1.26+
- A registered branchy project whose default remote (`origin`) is reachable
- At least one tree child that exists on `origin` but **not** as a local branch (e.g. `develop-1.20.5-…`)
- Terminal ≥ 80×24

## Setup

```bash
go build -o branchy ./cmd/branchy
go test ./internal/git/... ./internal/tui/... -count=1
```

Automated tests cover fetch success/fail/no-remote and in-place TUI refresh. Scenarios below are the manual check.

---

## Scenario 1: Tree paints locally, then refreshes

```bash
./branchy
```

Open the project that has a remote-only child.

**Expected**:

- Tree and local inbound counts appear immediately (no hang)
- The remote-only child may first show `?`
- After the background update, that child shows a file-change number without restarting
- Keyboard still works while the update runs (`j`/`k`, `s`, `q`)

---

## Scenario 2: Sync shares the update

From the tree press `s` quickly (or run `./branchy sync`) so sync opens before or after the tree fetch.

**Expected**:

- Picker/confirm appear immediately with local counts
- No second long hang (one shared update)
- When the update completes, picker badges and the open confirm line refresh in place
- Yes/No still works during the update

---

## Scenario 3: Offline / failed remote

Disconnect the network or point `origin` at an unreachable URL, then open the tree and sync.

**Expected**:

- Local tree and counts still appear immediately
- No fetch error banner or status line
- `?` stays `?` where refs are still missing
- Sync can be started and cancelled as usual

---

## Scenario 4: Scripted CLI does not fetch

```bash
./branchy sync --from <root> -y
```

**Expected**:

- Plain CLI only
- No background fetch started by this feature
- Existing stdin/flag behavior unchanged
