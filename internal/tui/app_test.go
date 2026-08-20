package tui

import (
	"errors"
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
	if model.screen != screenUnlinkConfirm {
		t.Fatalf("expected screenUnlinkConfirm, got %d", model.screen)
	}
	if model.unlinkTarget != "feature-a" {
		t.Fatalf("expected unlink target feature-a, got %q", model.unlinkTarget)
	}
	if model.unlinkCount != 2 {
		t.Fatalf("expected unlink count 2, got %d", model.unlinkCount)
	}
	view := model.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("embedded unlink confirm must not contain [y/N]")
	}
}

func TestUpdateUnlinkCancel(t *testing.T) {
	p := unlinkTestProject()
	m := Model{
		screen:       screenUnlinkConfirm,
		current:      p,
		unlinkTarget: "feature-a",
		unlinkCount:  2,
		treeView:     newBranchTreeView(p.Tree),
	}

	updated, _ := m.updateUnlink(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model := updated.(Model)
	if model.screen != screenTree {
		t.Fatalf("expected screenTree after cancel, got %d", model.screen)
	}
	if len(model.current.Tree.Branches) != 3 {
		t.Fatalf("expected tree unchanged, got %d branches", len(model.current.Tree.Branches))
	}
}

func TestUpdateUnlinkConfirmRemovesSubtree(t *testing.T) {
	p := unlinkTestProject()
	m := Model{
		screen:       screenUnlinkConfirm,
		current:      p,
		unlinkTarget: "feature-a",
		unlinkCount:  2,
		treeView:     newBranchTreeView(p.Tree),
	}

	// SaveTree will fail without a real path; confirm still mutates in memory first.
	// We test mutation by calling UnlinkSubtree path up to save failure handling.
	updated, _ := m.updateUnlink(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model := updated.(Model)
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

func TestTreePaintsLocalInboundBeforeRemoteMsg(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	if m.treeView.inbound == nil {
		t.Fatal("expected local inbound snapshot before remote update")
	}
	if cmd := m.Init(); cmd == nil {
		t.Fatal("expected remote update cmd after project selected")
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
	m.treeView.setDirection(diffInbound)

	view := stripANSI(m.View())
	if !strings.Contains(view, "feature-a  5") {
		t.Fatalf("expected inbound badge:\n%s", view)
	}
	if !strings.Contains(view, "↓ feature-a") {
		t.Fatalf("expected inbound arrow:\n%s", view)
	}
	if !strings.Contains(view, "counts: inbound (parent→child)") || !strings.Contains(view, "ctrl+↑: outbound") || !strings.Contains(view, "ctrl+↓: inbound") {
		t.Fatalf("expected inbound footer:\n%s", view)
	}
	if strings.Contains(view, "d: show") {
		t.Fatalf("footer must not list d:\n%s", view)
	}

	updated, _ := m.updateTree(tea.KeyMsg{Type: tea.KeyCtrlUp})
	model := updated.(Model)
	if model.diffDirection != diffOutbound {
		t.Fatal("expected outbound direction after ctrl+up")
	}
	view = stripANSI(model.View())
	if !strings.Contains(view, "feature-a  2") {
		t.Fatalf("expected outbound badge:\n%s", view)
	}
	if !strings.Contains(view, "↑ feature-a") {
		t.Fatalf("expected outbound arrow:\n%s", view)
	}
	if !strings.Contains(view, "counts: outbound (child→parent)") || !strings.Contains(view, "ctrl+↑: outbound") {
		t.Fatalf("expected outbound footer:\n%s", view)
	}

	cursor := model.treeView.cursor
	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyCtrlUp})
	model = updated.(Model)
	if model.diffDirection != diffOutbound {
		t.Fatal("second ctrl+up must stay outbound")
	}
	if model.treeView.cursor != cursor {
		t.Fatal("ctrl+up must not move the cursor")
	}

	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyCtrlDown})
	model = updated.(Model)
	if model.diffDirection != diffInbound {
		t.Fatal("ctrl+down must set inbound")
	}

	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(Model)
	if model.diffDirection != diffInbound {
		t.Fatal("d must not change direction")
	}

	updated, _ = model.updateTree(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if model.diffDirection != diffInbound {
		t.Fatal("plain down must not change direction")
	}
	if model.treeView.cursor == cursor {
		t.Fatal("plain down must move the cursor")
	}
}

func TestRemoteUpdateKeepsDirection(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.diffDirection = diffOutbound
	m.treeView.setDirection(diffOutbound)
	m.treeView.setFileCounts(
		map[string]fileChangeCount{"feature-a": {files: 99, ok: true}},
		map[string]fileChangeCount{"feature-a": {files: 88, ok: true}},
	)

	updated, _ := m.Update(remoteUpdateMsg{projectID: p.ID, path: p.Path})
	model := updated.(Model)
	if model.diffDirection != diffOutbound {
		t.Fatal("remote refresh must not reset direction")
	}
	if model.treeView.direction != diffOutbound {
		t.Fatal("tree direction must stay outbound after refresh")
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
	if model.diffDirection != diffOutbound {
		t.Fatal("expected outbound after ctrl+up")
	}

	// Simulate embedded sync finishing without running SyncFlowModel.Update.
	model.screen = screenTree
	model.syncFlow = SyncFlowModel{}
	if model.diffDirection != diffOutbound {
		t.Fatal("direction must survive sync return")
	}

	_ = model.selectProject(other)
	if model.diffDirection != diffOutbound {
		t.Fatal("direction must survive project switch")
	}
	if model.treeView.direction != diffOutbound {
		t.Fatal("tree must mirror outbound after project switch")
	}

	fresh := newModel([]*project.Project{p}, p)
	if fresh.diffDirection != diffInbound {
		t.Fatal("new session must start inbound")
	}
}

func TestSyncFollowsTreeOutbound(t *testing.T) {
	p := unlinkTestProject()
	m := newModel([]*project.Project{p}, p)
	m.diffDirection = diffOutbound
	m.treeView.setDirection(diffOutbound)
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
