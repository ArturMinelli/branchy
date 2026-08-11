package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
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
