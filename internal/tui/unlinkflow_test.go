package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
	"branchy/internal/tree"
)

func testUnlinkProject() *project.Project {
	return &project.Project{
		ID:   "test",
		Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"main":      {Children: []string{"develop"}},
			"develop":   {Children: []string{"feature-a"}},
			"feature-a": {},
		}},
	}
}

func TestUnlinkFlowStartsAtPick(t *testing.T) {
	m := newUnlinkFlowModel(testUnlinkProject(), UnlinkFlowOptions{})
	if m.step != stepUnlinkPick {
		t.Fatalf("expected stepUnlinkPick, got %d", m.step)
	}
}

func TestUnlinkFlowPickToConfirm(t *testing.T) {
	m := newUnlinkFlowModel(testUnlinkProject(), UnlinkFlowOptions{})
	for i, name := range m.project.Tree.Names() {
		if name == "develop" {
			m.branchList.Select(i)
			break
		}
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(UnlinkFlowModel)
	if flow.step != stepUnlinkConfirm {
		t.Fatalf("expected stepUnlinkConfirm, got %d", flow.step)
	}
	if flow.subtreeCount != 2 {
		t.Fatalf("expected subtree count 2, got %d", flow.subtreeCount)
	}
}

func TestUnlinkFlowConfirmViewUsesPanel(t *testing.T) {
	m := newUnlinkFlowModel(testUnlinkProject(), UnlinkFlowOptions{})
	m.target = "develop"
	m.subtreeCount = 2
	m.step = stepUnlinkConfirm
	m.confirm = NewConfirm(ConfirmOptions{Question: `Remove "develop" and 2 branches from tree?`, Width: 80})
	view := m.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("unlink confirm view must not contain [y/N]")
	}
}

func TestUnlinkFlowConfirmCancel(t *testing.T) {
	m := newUnlinkFlowModel(testUnlinkProject(), UnlinkFlowOptions{})
	m.target = "develop"
	m.subtreeCount = 2
	m.step = stepUnlinkConfirm

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	flow := updated.(UnlinkFlowModel)
	if !flow.cancelled {
		t.Fatal("expected cancelled")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestUnlinkFlowPrefillStartsAtConfirm(t *testing.T) {
	m := newUnlinkFlowModel(testUnlinkProject(), UnlinkFlowOptions{PrefillTarget: "develop"})
	if m.step != stepUnlinkConfirm {
		t.Fatalf("expected stepUnlinkConfirm, got %d", m.step)
	}
	if m.target != "develop" {
		t.Fatalf("expected target develop, got %q", m.target)
	}
	if m.subtreeCount != 2 {
		t.Fatalf("expected subtree count 2, got %d", m.subtreeCount)
	}
	view := m.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("prefilled confirm must not contain [y/N]")
	}
}

func TestUnlinkFlowCancelEmbedded(t *testing.T) {
	m := newUnlinkFlowModel(testUnlinkProject(), UnlinkFlowOptions{Embedded: true, PrefillTarget: "develop"})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	flow := updated.(UnlinkFlowModel)
	if !flow.cancelled {
		t.Fatal("expected cancelled")
	}
	if cmd != nil {
		t.Fatal("embedded cancel should not quit program")
	}
}
