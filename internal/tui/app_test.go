package tui

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
	"branchy/internal/sync"
	"branchy/internal/tree"
)

func unlinkTestProject() *project.Project {
	return &project.Project{
		ID:   "test",
		Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"develop":   {Children: []string{"feature-a"}},
			"feature-a": {Children: []string{"feature-b"}},
			"feature-b": {},
		}},
	}
}

func persistableProject(t *testing.T, doc *tree.Document) *project.Project {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := &project.Project{
		ID:   "test",
		Path: filepath.Join(home, "repo"),
		Tree: doc,
	}
	if err := p.SaveTree(); err != nil {
		t.Fatalf("SaveTree: %v", err)
	}
	return p
}

func TestUpdateTreeUnlinkOpensConfirm(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	for i, row := range m.treeView.rows {
		if row.name == "feature-a" {
			m.treeView.cursor = i
			break
		}
	}

	updated, _ := m.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	model := updated.(Model)
	if model.screen != screenUnlink {
		t.Fatalf("expected screenUnlink, got %d", model.screen)
	}
	if model.unlinkFlow.step != stepUnlinkConfirm {
		t.Fatalf("expected stepUnlinkConfirm, got %d", model.unlinkFlow.step)
	}
	if model.unlinkFlow.target != "feature-a" {
		t.Fatalf("expected unlink target feature-a, got %q", model.unlinkFlow.target)
	}
	if model.unlinkFlow.subtreeCount != 2 {
		t.Fatalf("expected unlink count 2, got %d", model.unlinkFlow.subtreeCount)
	}
	view := model.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("embedded unlink confirm must not contain [y/N]")
	}
	if strings.Contains(view, "Link branch") {
		t.Fatal("view must not be the old inline editor")
	}
}

func TestUpdateUnlinkCancel(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.screen = screenUnlink
	m.unlinkFlow = newUnlinkFlowModel(p, UnlinkFlowOptions{Embedded: true, PrefillTarget: "feature-a"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model := updated.(Model)
	if model.screen != screenTree {
		t.Fatalf("expected screenTree after cancel, got %d", model.screen)
	}
	if len(model.current.Tree.Branches) != 3 {
		t.Fatalf("expected tree unchanged, got %d branches", len(model.current.Tree.Branches))
	}
}

func TestUpdateUnlinkConfirmRemovesSubtree(t *testing.T) {
	p := persistableProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"develop":   {Children: []string{"feature-a"}},
		"feature-a": {Children: []string{"feature-b"}},
		"feature-b": {},
	}})
	m := newModel([]*project.Project{p}, p)
	m.screen = screenUnlink
	m.unlinkFlow = newUnlinkFlowModel(p, UnlinkFlowOptions{Embedded: true, PrefillTarget: "feature-a"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model := updated.(Model)
	if model.screen != screenUnlink {
		t.Fatalf("expected stay on unlink success step, got %d", model.screen)
	}
	if model.unlinkFlow.step != stepUnlinkSuccess {
		t.Fatalf("expected stepUnlinkSuccess, got %d", model.unlinkFlow.step)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.screen != screenTree {
		t.Fatalf("expected return to screenTree, got %d", model.screen)
	}
	if _, ok := model.current.Tree.Branches["feature-a"]; ok {
		t.Fatal("feature-a should be removed after confirm")
	}
	if _, ok := model.current.Tree.Branches["feature-b"]; ok {
		t.Fatal("feature-b should be removed after confirm")
	}
}

func TestUpdateTreeLinkOpensChildStep(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	for i, row := range m.treeView.rows {
		if row.name == "feature-a" {
			m.treeView.cursor = i
			break
		}
	}

	updated, _ := m.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model := updated.(Model)
	if model.screen != screenLink {
		t.Fatalf("expected screenLink, got %d", model.screen)
	}
	if model.linkFlow.step != stepLinkChild {
		t.Fatalf("expected stepLinkChild, got %d", model.linkFlow.step)
	}
	if model.linkFlow.parent != "feature-a" {
		t.Fatalf("expected parent feature-a, got %q", model.linkFlow.parent)
	}
	view := stripANSI(model.View())
	if strings.Contains(view, "Link branch") || strings.Contains(view, "tab: switch field") {
		t.Fatalf("must not show the old two-field editor:\n%s", view)
	}
	if !strings.Contains(view, "Parent: feature-a") {
		t.Fatalf("expected prefilled parent:\n%s", view)
	}
}

func TestUpdateLinkCancelLeavesTree(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.screen = screenLink
	m.linkFlow = newLinkFlowModel(p, LinkFlowOptions{Embedded: true, PrefillParent: "develop"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)
	if model.screen != screenLink {
		t.Fatalf("esc from child should open picker, got screen %d", model.screen)
	}
	if model.linkFlow.step != stepLinkParent {
		t.Fatalf("expected parent picker, got %d", model.linkFlow.step)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.screen != screenTree {
		t.Fatalf("expected screenTree after picker cancel, got %d", model.screen)
	}
	if len(model.current.Tree.Branches) != 3 {
		t.Fatalf("expected tree unchanged, got %d branches", len(model.current.Tree.Branches))
	}
}

func TestUpdateLinkSuccessRefreshesTree(t *testing.T) {
	p := persistableProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"develop":   {Children: []string{"feature-a"}},
		"feature-a": {},
	}})
	m := newModel([]*project.Project{p}, p)
	m.screen = screenLink
	m.linkFlow = newLinkFlowModel(p, LinkFlowOptions{Embedded: true, PrefillParent: "develop"})
	m.linkFlow.child = "feat"
	m.linkFlow.step = stepLinkConfirm
	m.linkFlow.confirm = NewConfirm(ConfirmOptions{Question: "Link develop → feat?", Width: 80})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model := updated.(Model)
	if model.linkFlow.step != stepLinkSuccess {
		t.Fatalf("expected stepLinkSuccess, got %d", model.linkFlow.step)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.screen != screenTree {
		t.Fatalf("expected screenTree after done, got %d", model.screen)
	}
	if _, ok := model.current.Tree.Branches["feat"]; !ok {
		t.Fatal("expected feat in tree after link")
	}
}

func TestTreeDefersCountsUntilRemoteUpdate(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	if m.treeView.inbound != nil {
		t.Fatal("expected no inbound snapshot before remote update")
	}
	if cmd := m.Init(); cmd == nil {
		t.Fatal("expected remote update cmd after project selected")
	}
	view := stripANSI(m.View())
	if !strings.Contains(view, "feature-a  ? ↓") {
		t.Fatalf("expected unknown badges before fetch:\n%s", view)
	}
}

func TestRemoteUpdateRefreshesCurrentProject(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(map[string]fileChangeCount{"feature-a": {files: 99, ok: true}}, nil)

	updated, cmd := m.Update(remoteUpdateMsg{projectID: p.ID, path: p.Path})
	if cmd != nil {
		t.Fatal("refresh must not start another cmd")
	}
	model := updated.(Model)
	c := model.treeView.inbound["feature-a"]
	if c.files == 99 && c.ok {
		t.Fatal("expected inbound counts to be recomputed")
	}
}

func TestRemoteUpdateIgnoresOtherProject(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(map[string]fileChangeCount{"feature-a": {files: 99, ok: true}}, nil)

	updated, _ := m.Update(remoteUpdateMsg{projectID: "other", path: p.Path})
	model := updated.(Model)
	if model.treeView.inbound["feature-a"].files != 99 {
		t.Fatal("foreign project must not refresh inbound")
	}
}

func TestRemoteUpdateFailureKeepsSnapshot(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(map[string]fileChangeCount{"feature-a": {files: 99, ok: true}}, nil)

	updated, _ := m.Update(remoteUpdateMsg{projectID: p.ID, path: p.Path, err: errors.New("offline")})
	model := updated.(Model)
	if model.treeView.inbound["feature-a"].files != 99 {
		t.Fatal("failed update must keep local snapshot")
	}
	view := strings.ToLower(model.View())
	if strings.Contains(view, "fetch") {
		t.Fatalf("view must not mention fetch failure:\n%s", model.View())
	}
}

func TestDirectionChordsSetBadgesAndFooter(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 5, ok: true}},
		map[string]fileChangeCount{"feature-a": {files: 2, ok: true}},
	)
	m.treeView.setDirection(sync.Downward)

	view := stripANSI(m.View())
	if !strings.Contains(view, "feature-a  5 ↓") {
		t.Fatalf("expected inbound badge:\n%s", view)
	}
	if !strings.Contains(view, "counts: inbound (parent→child)") || !strings.Contains(view, "ctrl+↑: outbound") || !strings.Contains(view, "ctrl+↓: inbound") {
		t.Fatalf("expected inbound footer:\n%s", view)
	}
	if strings.Contains(view, "d: show") {
		t.Fatalf("footer must not list d:\n%s", view)
	}

	updated, _ := m.updateTree(tea.KeyMsg{Type: tea.KeyCtrlUp})
	model := updated.(Model)
	if model.direction != sync.Upward {
		t.Fatal("expected outbound direction after ctrl+up")
	}
	view = stripANSI(model.View())
	if !strings.Contains(view, "feature-a  2 ↑") {
		t.Fatalf("expected outbound badge:\n%s", view)
	}
	if !strings.Contains(view, "counts: outbound (child→parent)") || !strings.Contains(view, "ctrl+↑: outbound") {
		t.Fatalf("expected outbound footer:\n%s", view)
	}

	cursor := model.treeView.cursor
	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyCtrlUp})
	model = updated.(Model)
	if model.direction != sync.Upward {
		t.Fatal("second ctrl+up must stay outbound")
	}
	if model.treeView.cursor != cursor {
		t.Fatal("ctrl+up must not move the cursor")
	}

	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyCtrlDown})
	model = updated.(Model)
	if model.direction != sync.Downward {
		t.Fatal("ctrl+down must set inbound")
	}

	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(Model)
	if model.direction != sync.Downward {
		t.Fatal("d must not change direction")
	}

	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if model.direction != sync.Downward {
		t.Fatal("plain down must not change direction")
	}
	if model.treeView.cursor == cursor {
		t.Fatal("plain down must move the cursor")
	}
}

func TestRemoteUpdateKeepsDirection(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.direction = sync.Upward
	m.treeView.setDirection(sync.Upward)
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 99, ok: true}},
		map[string]fileChangeCount{"feature-a": {files: 88, ok: true}},
	)

	updated, _ := m.Update(remoteUpdateMsg{projectID: p.ID, path: p.Path})
	model := updated.(Model)
	if model.direction != sync.Upward {
		t.Fatal("remote refresh must not reset direction")
	}
	if model.treeView.direction != sync.Upward {
		t.Fatal("tree direction must stay outbound after refresh")
	}
}

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

func TestUpdateTreeReloadSchedulesRemoteUpdateCmd(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)

	updated, cmd := m.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected remote update cmd for r")
	}
	msg := cmd()
	ru, ok := msg.(remoteUpdateMsg)
	if !ok {
		t.Fatalf("expected remoteUpdateMsg, got %T", msg)
	}
	if ru.projectID != p.ID {
		t.Fatalf("projectID %q, want %q", ru.projectID, p.ID)
	}
	model := updated.(Model)
	if model.current == nil || model.current.ID != p.ID {
		t.Fatal("reload must not change current project")
	}
}

func TestUpdateTreeReloadDoesNotBlockNavigation(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	cursor := m.treeView.cursor

	updated, cmd := m.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected remote update cmd for r")
	}
	model := updated.(Model)

	updated, cmd = model.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd != nil {
		t.Fatal("navigation must not wait for remote update")
	}
	model = updated.(Model)
	if model.treeView.cursor == cursor {
		t.Fatal("j must move cursor while reload is scheduled")
	}
}

func TestDirectionSurvivesSyncReturnAndProjectSwitch(t *testing.T) {
	p := unlinkTestProject()
	other := &project.Project{
		ID:   "other",
		Path: "/tmp/other",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"main":  {Children: []string{"child"}},
			"child": {},
		}},
	}
	m := newModel([]*project.Project{p, other}, p)
	updated, _ := m.updateTree(tea.KeyMsg{Type: tea.KeyCtrlUp})
	model := updated.(Model)
	if model.direction != sync.Upward {
		t.Fatal("expected outbound after ctrl+up")
	}

	// Simulate embedded sync finishing without running SyncFlowModel.Update.
	model.screen = screenTree
	model.syncFlow = SyncFlowModel{}
	if model.direction != sync.Upward {
		t.Fatal("direction must survive sync return")
	}

	_ = model.selectProject(other)
	if model.direction != sync.Upward {
		t.Fatal("direction must survive project switch")
	}
	if model.treeView.direction != sync.Upward {
		t.Fatal("tree must mirror outbound after project switch")
	}

	fresh := newModel([]*project.Project{p}, p)
	if fresh.direction != sync.Downward {
		t.Fatal("new session must start inbound")
	}
}

func TestSyncFollowsTreeOutbound(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.direction = sync.Upward
	m.treeView.setDirection(sync.Upward)
	for i, row := range m.treeView.rows {
		if row.name == "develop" {
			m.treeView.cursor = i
			break
		}
	}

	updated, _ := m.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model := updated.(Model)
	if model.screen != screenSync {
		t.Fatalf("expected sync screen, got %d", model.screen)
	}
	if model.syncFlow.direction != sync.Upward {
		t.Fatal("outbound tree s must start upward sync")
	}
	if model.syncFlow.fromBranch != "develop" {
		t.Fatalf("ceiling should be develop, got %q", model.syncFlow.fromBranch)
	}
	view := stripANSI(model.View())
	if strings.Contains(view, "d: show") || strings.Contains(view, "counts: outbound") || strings.Contains(view, "ctrl+↑") {
		t.Fatalf("embedded confirm must not show direction controls:\n%s", view)
	}
}

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

func TestSyncFollowsTreeInbound(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	for i, row := range m.treeView.rows {
		if row.name == "develop" {
			m.treeView.cursor = i
			break
		}
	}

	updated, _ := m.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model := updated.(Model)
	if model.syncFlow.direction != sync.Downward {
		t.Fatal("inbound tree s must start downward sync")
	}
	if model.syncFlow.fromBranch != "develop" {
		t.Fatalf("ceiling should be develop, got %q", model.syncFlow.fromBranch)
	}
}
