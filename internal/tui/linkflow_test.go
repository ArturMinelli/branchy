package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
	"branchy/internal/tree"
)

func testLinkProject() *project.Project {
	return &project.Project{
		ID:   "test",
		Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"main":    {Children: []string{"develop"}},
			"develop": {},
		}},
	}
}

func TestLinkFlowStartsAtParentPicker(t *testing.T) {
	m := newLinkFlowModel(testLinkProject())
	if m.step != stepLinkParent {
		t.Fatalf("expected stepLinkParent, got %d", m.step)
	}
}

func TestLinkFlowParentToChild(t *testing.T) {
	m := newLinkFlowModel(testLinkProject())
	m.branchList.Select(0)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(LinkFlowModel)
	if flow.step != stepLinkChild {
		t.Fatalf("expected stepLinkChild, got %d", flow.step)
	}
	if flow.parent == "" {
		t.Fatal("expected parent set")
	}
}

func TestLinkFlowConfirmViewUsesPanel(t *testing.T) {
	m := newLinkFlowModel(testLinkProject())
	m.parent = "main"
	m.child = "feature"
	m.step = stepLinkConfirm
	m.confirm = NewConfirm(ConfirmOptions{Question: "Link main → feature?", Width: 80})
	view := m.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("link confirm view must not contain [y/N]")
	}
}

func TestLinkFlowConfirmCancel(t *testing.T) {
	m := newLinkFlowModel(testLinkProject())
	m.parent = "main"
	m.child = "feature"
	m.step = stepLinkConfirm
	m.confirm = NewConfirm(ConfirmOptions{Question: "Link main → feature?", Width: 80})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(LinkFlowModel)
	if !flow.cancelled {
		t.Fatal("expected cancelled")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
