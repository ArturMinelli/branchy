# TUI Reload Loading State Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show visible reload-in-progress feedback when the user presses `r` by resetting file-change counts to `?` (same pattern as initial main-tree load), then restore or refresh counts when the fetch completes.

**Architecture:** Introduce a small `countSnapshot` helper that saves inbound/outbound maps before reload, clears them to `nil` (which the tree already renders as `?`), and restores the snapshot on fetch failure. Success path reuses existing `withFreshFileCounts` / `applyFileCounts`. Sync picker gets explicit `?` badges during reload via a `reloadingBadges` helper because `fileChangeBadges(nil)` currently renders no badge. Repeat `r` while in-flight coalesces via existing `FetchDefaultRemote` single-flight and does not overwrite the original snapshot.

**Tech Stack:** Go, Bubble Tea (`tea`), Charm (`lipgloss`, `bubbles/list`), existing `remoteUpdateCmd` / `remoteUpdateMsg` pipeline.

## Global Constraints

- Reload MUST remain non-blocking: navigation, sync progression, direction toggle, and quit work while fetch runs.
- Reload MUST NOT show a spinner, status line, or fetch-failure banner (feedback is `?` badges / "File count unavailable" on confirm only).
- On fetch failure, restore the count snapshot that was visible before `r` was pressed.
- On fetch success, recompute both inbound and outbound counts (existing behavior).
- Repeat `r` during in-flight reload: coalesce fetch, keep showing `?`, do not replace the saved snapshot.
- Scope: main tree + sync root picker + sync edge confirm (everywhere `r: reload` exists today).
- Auto-fetch on project open is unchanged (already shows `?` via nil counts; no new snapshot logic needed).

---

## File map

| File | Responsibility |
|------|----------------|
| `internal/tui/reload.go` | `countSnapshot`, `snapshotCounts`, `restoreCounts`, `reloadingBadges` |
| `internal/tui/reload_test.go` | Unit tests for snapshot + badge helpers |
| `internal/tui/app.go` | `beginReload` on tree `r`; snapshot restore in `applyRemoteUpdate` |
| `internal/tui/app_test.go` | Tree reload shows `?`; failure restores; repeat `r` keeps snapshot |
| `internal/tui/syncflow.go` | `beginReload` on sync `r`; picker `?` badges; snapshot field |
| `internal/tui/syncflow_test.go` | Sync picker/confirm reload feedback + failure restore |

---

### Task 1: Shared reload snapshot helpers

**Files:**
- Create: `internal/tui/reload.go`
- Test: `internal/tui/reload_test.go`

**Interfaces:**
- Consumes: `fileChangeCount` from `inbound.go`, `tree.Document` from `internal/tree`
- Produces: `countSnapshot`, `snapshotCounts`, `(*countSnapshot).Restore()`, `reloadingBadges(doc *tree.Document) map[string]string`

- [ ] **Step 1: Write the failing tests**

```go
package tui

import (
	"testing"

	"branchy/internal/tree"
)

func TestSnapshotCountsRoundTrip(t *testing.T) {
	in := map[string]fileChangeCount{"a": {files: 3, ok: true}}
	out := map[string]fileChangeCount{"b": {files: 1, ok: true}}
	snap := snapshotCounts(in, out)

	var restoredIn, restoredOut map[string]fileChangeCount
	snap.Restore(&restoredIn, &restoredOut)
	if restoredIn["a"].files != 3 || restoredOut["b"].files != 1 {
		t.Fatal("restore must return original maps")
	}
}

func TestReloadingBadgesMarksChildrenOnly(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main": {Children: []string{"child"}},
		"child": {},
	}}
	badges := reloadingBadges(doc)
	if badges["child"] != "?" {
		t.Fatalf("expected ? for child, got %q", badges["child"])
	}
	if badges["main"] != "" {
		t.Fatal("root must not get a reload badge")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestSnapshotCountsRoundTrip|TestReloadingBadgesMarksChildrenOnly' -v`
Expected: FAIL with "undefined: snapshotCounts" (or similar compile error)

- [ ] **Step 3: Write minimal implementation**

```go
package tui

import "branchy/internal/tree"

type countSnapshot struct {
	inbound  map[string]fileChangeCount
	outbound map[string]fileChangeCount
}

func snapshotCounts(inbound, outbound map[string]fileChangeCount) *countSnapshot {
	return &countSnapshot{inbound: inbound, outbound: outbound}
}

func (s *countSnapshot) Restore(inbound, outbound *map[string]fileChangeCount) {
	if s == nil {
		return
	}
	*inbound = s.inbound
	*outbound = s.outbound
}

func reloadingBadges(doc *tree.Document) map[string]string {
	if doc == nil {
		return nil
	}
	badges := make(map[string]string)
	for _, name := range doc.Names() {
		parent, hasParent := doc.ParentOf(name)
		if !hasParent {
			continue
		}
		_ = parent
		badges[name] = "?"
	}
	return badges
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run 'TestSnapshotCountsRoundTrip|TestReloadingBadgesMarksChildrenOnly' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/reload.go internal/tui/reload_test.go
git commit -m "feat(tui): add reload snapshot and badge helpers"
```

---

### Task 2: Main tree reload shows `?` and restores on failure

**Files:**
- Modify: `internal/tui/app.go`
- Test: `internal/tui/app_test.go`

**Interfaces:**
- Consumes: `countSnapshot`, `snapshotCounts`, `(*countSnapshot).Restore()` from Task 1
- Produces: `Model.reloadSnapshot *countSnapshot`, `Model.beginReload() Model`, updated `applyRemoteUpdate`, updated `updateTree` reload branch

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/app_test.go`:

```go
func TestUpdateTreeReloadShowsQuestionMarks(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 5, ok: true}},
		map[string]fileChangeCount{"feature-a": {files: 2, ok: true}},
	)

	updated, cmd := m.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected remote update cmd for r")
	}
	model := updated.(Model)
	if model.treeView.inbound != nil || model.treeView.outbound != nil {
		t.Fatal("reload must clear counts to nil")
	}
	view := stripANSI(model.View())
	if !strings.Contains(view, "feature-a  ?") {
		t.Fatalf("expected ? badges during reload:\n%s", view)
	}
}

func TestReloadFailureRestoresSnapshot(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 99, ok: true}},
		nil,
	)
	m = m.beginReload()

	updated, _ := m.Update(remoteUpdateMsg{projectID: p.ID, path: p.Path, err: errors.New("offline")})
	model := updated.(Model)
	if model.treeView.inbound["feature-a"].files != 99 {
		t.Fatal("failed reload must restore prior snapshot")
	}
	if model.reloadSnapshot != nil {
		t.Fatal("snapshot must be cleared after restore")
	}
}

func TestRepeatReloadKeepsOriginalSnapshot(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 42, ok: true}},
		nil,
	)

	m = m.beginReload()
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 0, ok: true}},
		nil,
	)
	m = m.beginReload()

	updated, _ := m.Update(remoteUpdateMsg{projectID: p.ID, path: p.Path, err: errors.New("offline")})
	model := updated.(Model)
	if model.treeView.inbound["feature-a"].files != 42 {
		t.Fatal("second beginReload must not overwrite original snapshot")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestUpdateTreeReloadShowsQuestionMarks|TestReloadFailureRestoresSnapshot|TestRepeatReloadKeepsOriginalSnapshot' -v`
Expected: FAIL (methods/behavior missing)

- [ ] **Step 3: Write minimal implementation**

Add field to `Model` in `app.go`:

```go
reloadSnapshot *countSnapshot
```

Add methods:

```go
func (m Model) beginReload() Model {
	if m.current == nil || m.reloadSnapshot != nil {
		return m
	}
	m.reloadSnapshot = snapshotCounts(m.treeView.inbound, m.treeView.outbound)
	m.treeView.setFileCounts(nil, nil)
	return m
}
```

Update `updateTree` reload branch:

```go
if key.Matches(msg, keys.Reload) {
	if m.current != nil {
		m = m.beginReload()
		return m, remoteUpdateCmd(m.current)
	}
	return m, nil
}
```

Update `applyRemoteUpdate`:

```go
func (m Model) applyRemoteUpdate(msg remoteUpdateMsg) Model {
	if m.current == nil || msg.projectID != m.current.ID {
		return m
	}
	if msg.err != nil {
		if m.reloadSnapshot != nil {
			var in, out map[string]fileChangeCount
			m.reloadSnapshot.Restore(&in, &out)
			m.treeView.setFileCounts(in, out)
			m.reloadSnapshot = nil
			if m.screen == screenSync {
				m.syncFlow = m.syncFlow.restoreReloadSnapshot()
			}
			return m
		}
		if m.treeView.inbound == nil && m.treeView.outbound == nil {
			return m.withFreshFileCounts()
		}
		return m
	}
	m.reloadSnapshot = nil
	m = m.withFreshFileCounts()
	if m.screen == screenSync {
		m.syncFlow = m.syncFlow.applyFileCounts(m.treeView.inbound, m.treeView.outbound)
		m.syncFlow.clearReloadSnapshot()
	}
	return m
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run 'TestUpdateTreeReloadShowsQuestionMarks|TestReloadFailureRestoresSnapshot|TestRepeatReloadKeepsOriginalSnapshot|TestRemoteUpdateFailureKeepsSnapshot|TestTreeDefersCountsUntilRemoteUpdate' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "feat(tui): show ? badges on tree reload with failure restore"
```

---

### Task 3: Sync flow reload shows `?` on picker and confirm

**Files:**
- Modify: `internal/tui/syncflow.go`
- Test: `internal/tui/syncflow_test.go`

**Interfaces:**
- Consumes: `countSnapshot`, `snapshotCounts`, `reloadingBadges` from Task 1; `restoreReloadSnapshot`, `clearReloadSnapshot` produced here
- Produces: `SyncFlowModel.reloadSnapshot *countSnapshot`, `SyncFlowModel.beginReload() SyncFlowModel`, updated `refreshPicker`, updated `updatePickRoot` / `updateEdgeConfirm` reload branches

- [ ] **Step 1: Write the failing tests**

Add to `internal/tui/syncflow_test.go`:

```go
func TestSyncPickRootReloadShowsQuestionMarks(t *testing.T) {
	m := newSyncFlowModel(testSyncProject(), "", SyncFlowOptions{})
	m.inbound = map[string]fileChangeCount{"develop": {files: 7, ok: true}}
	m = m.refreshPicker()

	updated, cmd := m.updatePickRoot(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected remote update cmd")
	}
	flow := updated.(SyncFlowModel)
	if flow.inbound != nil {
		t.Fatal("reload must clear inbound counts")
	}
	byName := pickerByName(t, flow)
	if byName["develop"].badge != "?" {
		t.Fatalf("expected ? badge during reload, got %q", byName["develop"].badge)
	}
}

func TestSyncEdgeConfirmReloadShowsUnavailable(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.inbound = map[string]fileChangeCount{"develop": {files: 7, ok: true}}
	m = m.resetEdgeConfirm()

	updated, cmd := m.updateEdgeConfirm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected remote update cmd")
	}
	flow := updated.(SyncFlowModel)
	view := flow.View()
	if !strings.Contains(view, "File count unavailable") {
		t.Fatalf("expected unavailable confirm line during reload:\n%s", view)
	}
}

func TestSyncReloadFailureRestoresSnapshot(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.inbound = map[string]fileChangeCount{"develop": {files: 7, ok: true}}
	m = m.resetEdgeConfirm()
	m = m.beginReload()

	updated, _ := m.Update(remoteUpdateMsg{projectID: "test", path: "/tmp/test", err: errors.New("offline")})
	flow := updated.(SyncFlowModel)
	if flow.inbound["develop"].files != 7 {
		t.Fatal("failed sync reload must restore snapshot")
	}
	if !strings.Contains(flow.View(), "7 files would change on develop") {
		t.Fatalf("confirm must show restored count:\n%s", flow.View())
	}
}
```

Note: `TestSyncReloadFailureRestoresSnapshot` exercises `SyncFlowModel` directly. Embedded mode restore is covered in Task 4 via App-level `applyRemoteUpdate`.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run 'TestSyncPickRootReloadShowsQuestionMarks|TestSyncEdgeConfirmReloadShowsUnavailable|TestSyncReloadFailureRestoresSnapshot' -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

Add to `SyncFlowModel`:

```go
reloadSnapshot *countSnapshot
```

Add methods:

```go
func (m SyncFlowModel) beginReload() SyncFlowModel {
	if m.project == nil || m.reloadSnapshot != nil {
		return m
	}
	m.reloadSnapshot = snapshotCounts(m.inbound, m.outbound)
	m.inbound = nil
	m.outbound = nil
	return m.refreshCountSurfaces()
}

func (m SyncFlowModel) restoreReloadSnapshot() SyncFlowModel {
	if m.reloadSnapshot == nil {
		return m
	}
	m.reloadSnapshot.Restore(&m.inbound, &m.outbound)
	m.reloadSnapshot = nil
	return m.refreshCountSurfaces()
}

func (m SyncFlowModel) clearReloadSnapshot() SyncFlowModel {
	m.reloadSnapshot = nil
	return m
}
```

Update `refreshPicker`:

```go
func (m SyncFlowModel) refreshPicker() SyncFlowModel {
	idx := m.branchList.Index()
	badges := fileChangeBadges(m.activeCounts())
	if m.reloadSnapshot != nil && m.activeCounts() == nil {
		badges = reloadingBadges(m.project.Tree)
	}
	m.branchList = newBranchListWithBadges(m.project.Tree.Names(), "Select root branch to sync from", badges)
	// ... rest unchanged
	return m
}
```

Update reload handlers in `updatePickRoot` and `updateEdgeConfirm`:

```go
if key.Matches(msg, syncKeys.Reload) {
	if m.project != nil {
		m = m.beginReload()
		return m, remoteUpdateCmd(m.project)
	}
	return m, nil
}
```

Update standalone `SyncFlowModel.Update` `remoteUpdateMsg` case to restore snapshot on error (for non-embedded sync binary):

```go
case remoteUpdateMsg:
	if m.project == nil || msg.projectID != m.project.ID {
		return m, nil
	}
	if msg.err != nil {
		if m.reloadSnapshot != nil {
			return m.restoreReloadSnapshot(), nil
		}
		return m, nil
	}
	m = m.clearReloadSnapshot()
	return m.applyFileCounts(
		loadInboundCounts(m.project.Path, m.project.Tree),
		loadOutboundCounts(m.project.Path, m.project.Tree),
	), nil
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run 'TestSyncPickRootReload|TestSyncEdgeConfirmReload|TestSyncReloadFailure' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/syncflow.go internal/tui/syncflow_test.go
git commit -m "feat(tui): show ? badges on sync reload with failure restore"
```

---

### Task 4: Wire embedded sync reload through App applyRemoteUpdate

**Files:**
- Modify: `internal/tui/app.go` (already partially done in Task 2)
- Test: `internal/tui/app_test.go`

**Interfaces:**
- Consumes: `SyncFlowModel.beginReload`, `restoreReloadSnapshot`, `clearReloadSnapshot` from Task 3
- Produces: embedded sync reload restores snapshot on failure via App `applyRemoteUpdate`

- [ ] **Step 1: Write the failing test**

Add to `internal/tui/app_test.go`:

```go
func TestEmbeddedSyncReloadFailureRestoresViaApp(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 5, ok: true}},
		nil,
	)
	m.syncFlow = newSyncFlowModel(p, "", SyncFlowOptions{Embedded: true, Direction: sync.Downward})
	m.syncFlow.inbound = map[string]fileChangeCount{"feature-a": {files: 5, ok: true}}
	m.syncFlow = m.syncFlow.refreshPicker()
	m.screen = screenSync
	m.syncFlow = m.syncFlow.beginReload()

	updated, _ := m.Update(remoteUpdateMsg{projectID: p.ID, path: p.Path, err: errors.New("offline")})
	model := updated.(Model)
	if model.syncFlow.inbound["feature-a"].files != 5 {
		t.Fatal("embedded sync reload failure must restore sync snapshot")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestEmbeddedSyncReloadFailureRestoresViaApp -v`
Expected: FAIL

- [ ] **Step 3: Verify Task 2 `applyRemoteUpdate` calls `m.syncFlow.restoreReloadSnapshot()` on error and `clearReloadSnapshot()` on success** (adjust if the test reveals a gap).

- [ ] **Step 4: Run full TUI package tests**

Run: `go test ./internal/tui/ -v`
Expected: PASS

- [ ] **Step 5: Commit** (only if Task 2 commit did not already include the embedded wiring)

```bash
git add internal/tui/app.go internal/tui/app_test.go
git commit -m "fix(tui): restore embedded sync counts on reload failure"
```

---

### Task 5: Update reload contract docs

**Files:**
- Modify: `specs/016-tui-reload-counts/contracts/tui-reload-counts.md`
- Modify: `specs/016-tui-reload-counts/spec.md` (Assumptions table row for in-progress feedback)

- [ ] **Step 1: Update contract "Must not" section**

Replace:
```
- Show loading spinner or status line
```
With:
```
- Show loading spinner or status line (badges may show `?` during reload — same as initial tree load)
```

Add under Refresh contract:
```
On manual reload start (`r`):
- Save current inbound/outbound snapshot (if not already reloading)
- Clear counts to nil so browse surfaces show `?` (tree badges, sync picker badges, confirm shows "File count unavailable")
- Repeat `r` while in-flight: keep existing snapshot and `?` display; coalesce fetch
```

- [ ] **Step 2: Update spec assumptions**

Change in-progress feedback from "Silent — no loading indicator" to "`?` badges on browse surfaces and 'File count unavailable' on sync confirm while reload is in flight".

- [ ] **Step 3: Commit**

```bash
git add specs/016-tui-reload-counts/contracts/tui-reload-counts.md specs/016-tui-reload-counts/spec.md
git commit -m "docs: reload in-progress feedback uses ? badges"
```

---

## Self-review

**Spec coverage:**
| Requirement | Task |
|-------------|------|
| `?` feedback during reload (tree) | Task 2 |
| `?` feedback during reload (sync picker/confirm) | Task 3 |
| Failure restores prior snapshot | Tasks 2, 3, 4 |
| Non-blocking navigation | Existing tests retained; no key blocking added |
| Repeat `r` coalesces | `beginReload` idempotent + existing git single-flight |
| No spinner/status/error banner | Global constraints enforced |
| Auto-fetch unchanged | `beginReload` only called from `r` handlers |

**Placeholder scan:** No TBD/TODO/similar-to placeholders.

**Type consistency:** `countSnapshot`, `beginReload`, `restoreReloadSnapshot`, `clearReloadSnapshot` names consistent across tasks.

---

## Confirmed decisions (grilling 2026-09-09)

| Decision | Choice |
|----------|--------|
| Loading UX | Reset counts to `?` (same as initial main-tree load) |
| Scope | Main tree + sync picker + sync edge confirm |
| Fetch failure | Restore previous count snapshot |
| Repeat `r` in-flight | Coalesce fetch; keep `?`; do not overwrite snapshot |
