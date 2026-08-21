# Quickstart: Split CLI Commands

**Feature**: `012-split-cli-commands` | **Date**: 2026-08-21

Validation guide. Layout rules: [contracts/cli-commands.md](./contracts/cli-commands.md). Entities: [data-model.md](./data-model.md).

## Prerequisites

- Spec 011 merged (shared operations in project/sync/mr)
- Go 1.26+
- Registered test project or temp HOME fixture from 011 quickstart

## Setup

```bash
go test ./internal/cli/... -count=1
go build -o branchy ./cmd/branchy
```

Post-split file check:

```bash
ls internal/cli/{root,mode,init,sync,mr,link,unlink,projects}.go
test ! -f internal/cli/root.go.bak
wc -l internal/cli/root.go   # expect ~40 lines, not ~340
```

---

## Scenario 1: Help isolation

```bash
./branchy --help
./branchy sync --help
./branchy mr --help
./branchy link --help
./branchy unlink --help
./branchy init --help
./branchy projects --help
```

**Expected**:

- Same subcommand list and purpose as before split
- Changing help in `sync.go` does not require editing `mr.go` (SC-001, SC-004)

---

## Scenario 2: Default TUI still launches

On a TTY, from a registered repo:

```bash
./branchy
```

**Expected**: Main interactive tree opens (root command `RunE` unchanged)

---

## Scenario 3: Scripted paths unchanged

```bash
./branchy link main feature-split-test
./branchy unlink main feature-split-test
./branchy projects
./branchy init --force    # if already registered
./branchy sync --from main -y   # with glab auth
./branchy mr --source a --target b -y
```

**Expected**:

- Same success/error lines and exit codes as pre-split
- Handlers call `project.Link`, `project.Unlink`, `sync.Run`, `mr.Create` — grep audit clean

---

## Scenario 4: Mode rules unchanged

```bash
./branchy sync --from main 2>&1 | head -1    # non-TTY pipe → scripted, no full-screen TUI
echo n | ./branchy sync --from main          # per-edge prompts on TTY
```

**Expected**:

- `--from` forces scripted even on TTY (`UseTUI` false)
- Piped stdout uses scripted path
- `go test ./internal/cli/...` green including `TestSyncCmdFlagForcesCLI`

---

## Scenario 5: Handler thinness audit

```bash
rg 'Tree\.Link|SaveTree|UnlinkSubtree|HasEdge|gitlab' internal/cli/
rg 'p\.Tree\.(Names|CollectEdges)' internal/cli/
```

**Expected**: First grep — no matches (SC-003). Second grep — matches only in `sync.go` (read-only picker/plan print; see contracts exception).

---

## Done when

- [ ] Scenarios 1–5 pass
- [ ] `go test ./... -count=1` green
- [ ] Eight CLI source files exist; `root.go` is assembly-only
- [ ] `cmd/branchy/main.go` unchanged except if import path accidentally broken (should not change)
