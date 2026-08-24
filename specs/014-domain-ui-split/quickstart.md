# Quickstart: Domain / UI Split

**Feature**: `014-domain-ui-split` | **Date**: 2026-08-24

Validation guide. Contract: [contracts/domain-ui-split.md](./contracts/domain-ui-split.md). Entities: [data-model.md](./data-model.md).

## Prerequisites

- Specs 011–013 merged
- Go 1.26+
- Registered test project (temp HOME fixture from earlier quickstarts)

## Setup

```bash
go test ./internal/tree/... ./internal/browser/... ./internal/sync/... ./internal/tui/... ./internal/cli/... -count=1
go build -o branchy ./cmd/branchy
```

---

## Scenario 1: Shared walk, styling in TUI

```bash
go test ./internal/tree/... ./internal/tui/... -count=1 -run 'WalkDisplay|FlattenTree|TreeView'
```

**Expected**:

- `WalkDisplay` on a fixture with two roots and nested children returns every name once, roots sorted, children sorted, correct `IsLast` / `LastAtDepth`
- Interactive tree still shows `└── ` / `├── ` and the same membership/order (SC-004)
- Changing a color in `treeview.go` does not require editing `internal/tree/tree.go` (SC-001)

---

## Scenario 2: Decline open — no tabs, URLs still listed

Scripted sync that produces at least one openable URL:

```bash
# after a successful dry run / fixture sync that prints MR URLs
# at "Open MRs in browser? [y/N]" press Enter
```

**Expected**: no browser helper invoked; summary still contains the URLs (SC-003). Same for TUI confirm No.

---

## Scenario 3: Accept open — same order and pause

At the same prompt, accept (`y`).

**Expected**: tabs open in OpenableURLs order; 200ms between tabs; a failed open prints the existing non-fatal warning and does not fail the command (FR-004).

Unit stand-in:

```bash
go test ./internal/browser/... -count=1 -run OpenURLs
```

---

## Scenario 4: Sync operation does not open a browser

```bash
rg 'func OpenURLs' internal/sync
rg 'internal/browser' internal/sync
```

**Expected**: no matches. `OpenableURLs` still lives in `internal/sync` and is what both surfaces filter on.

---

## Scenario 5: No GitLab in interactive flows

```bash
rg 'gitlab' internal/tui
```

**Expected**: no matches (SC-002). Logged-out interactive sync still surfaces the operation’s auth error (spec 011), not a screen-local client.

---

## Scenario 6: Unused styled renderer gone

```bash
rg 'RenderASCII|lipgloss' internal/tree
test ! -f internal/tree/render.go
```

**Expected**: no matches; `render.go` deleted.

---

## Done when

- [ ] Scenarios 1–6 pass
- [ ] `go test ./... -count=1` green
- [ ] Contract grep audit clean
- [ ] No intended visual or prompt changes (SC-005)
