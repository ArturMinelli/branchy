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
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{})
	if m.step != stepLinkParent {
		t.Fatalf("expected stepLinkParent, got %d", m.step)
	}
}

func TestLinkFlowParentToChild(t *testing.T) {
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{})
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
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{})
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
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{})
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

func TestLinkFlowPrefillStartsAtChild(t *testing.T) {
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{PrefillParent: "develop"})
	if m.step != stepLinkChild {
		t.Fatalf("expected stepLinkChild, got %d", m.step)
	}
	if m.parent != "develop" {
		t.Fatalf("expected parent develop, got %q", m.parent)
	}
}

func TestLinkFlowEscFromChildOpensPicker(t *testing.T) {
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{})
	m.step = stepLinkChild
	m.parent = "main"

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(LinkFlowModel)
	if flow.step != stepLinkParent {
		t.Fatalf("expected stepLinkParent, got %d", flow.step)
	}
	if flow.finished || flow.cancelled {
		t.Fatal("esc from child must not leave the flow")
	}
}

func TestLinkFlowPrefillEscOpensPicker(t *testing.T) {
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{PrefillParent: "main"})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(LinkFlowModel)
	if flow.step != stepLinkParent {
		t.Fatalf("expected parent picker after esc, got %d", flow.step)
	}
}

func TestLinkFlowCancelEmbedded(t *testing.T) {
	m := newLinkFlowModel(testLinkProject(), LinkFlowOptions{Embedded: true})
	m.step = stepLinkConfirm
	m.confirm = NewConfirm(ConfirmOptions{Question: "Link main → feature?", Width: 80})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(LinkFlowModel)
	if !flow.cancelled {
		t.Fatal("expected cancelled")
	}
	if cmd != nil {
		t.Fatal("embedded cancel should not quit program")
	}
}
