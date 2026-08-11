package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
	"branchy/internal/tree"
)

func testProject() *project.Project {
	return &project.Project{
		ID:   "test",
		Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"develop":   {},
			"feature-x": {},
			"release":   {},
		}},
	}
}

func TestMRFlowPrefilledSourceNotInTree(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{PrefilledSource: "missing"})
	if m.step != stepMRError {
		t.Fatalf("expected stepMRError, got %d", m.step)
	}
	if m.errMsg == "" {
		t.Fatal("expected error message")
	}
}

func TestMRFlowPrefilledSourceStartsAtTarget(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{PrefilledSource: "develop"})
	if m.step != stepMRTarget {
		t.Fatalf("expected stepMRTarget, got %d", m.step)
	}
	if m.source != "develop" {
		t.Fatalf("expected source develop, got %q", m.source)
	}
}

func TestMRFlowFewBranches(t *testing.T) {
	p := &project.Project{
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"only": {},
		}},
	}
	m := newMRFlowModel(p, MROptions{})
	if m.step != stepMRFewBranches {
		t.Fatalf("expected stepMRFewBranches, got %d", m.step)
	}
}

func TestMRFlowSourceToTargetTransition(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{})
	m.branchList.Select(1) // feature-x sorted after develop

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(MRFlowModel)
	if flow.step != stepMRTarget {
		t.Fatalf("expected stepMRTarget after source select, got %d", flow.step)
	}
	if flow.source == "" {
		t.Fatal("expected source to be set")
	}
}

func TestMRFlowTitleToConfirm(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{})
	m.source = "feature-x"
	m.target = "develop"
	m.title = "MR: feature-x → develop"
	m.step = stepMRTitle

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(MRFlowModel)
	if flow.step != stepMRConfirm {
		t.Fatalf("expected stepMRConfirm, got %d", flow.step)
	}
}

func TestMRFlowConfirmViewUsesPanel(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{})
	m.source = "feature-x"
	m.target = "develop"
	m.title = "MR: feature-x → develop"
	m.step = stepMRConfirm
	m = m.resetMRConfirm()
	view := m.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("mr confirm view must not contain [y/N]")
	}
}

func TestMRFlowYesEntersLoading(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{})
	m.source = "feature-x"
	m.target = "develop"
	m.title = "MR: feature-x → develop"
	m.step = stepMRConfirm
	m = m.resetMRConfirm()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	flow := updated.(MRFlowModel)
	if flow.step != stepMRLoading {
		t.Fatalf("expected stepMRLoading, got %d", flow.step)
	}
	if cmd == nil {
		t.Fatal("expected async create command")
	}
}

func TestMRFlowCancelEmbedded(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{Embedded: true})
	m.step = stepMRConfirm

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(MRFlowModel)
	if !flow.cancelled {
		t.Fatal("expected cancelled")
	}
	if cmd != nil {
		t.Fatal("embedded cancel should not quit program")
	}
}

func TestTargetBranchNamesExcludesSource(t *testing.T) {
	names := []string{"a", "b", "c"}
	got := targetBranchNames(names, "b")
	if len(got) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(got))
	}
	for _, n := range got {
		if n == "b" {
			t.Fatal("source should be excluded")
		}
	}
}

func TestMRFlowTitleTyping(t *testing.T) {
	m := newMRFlowModel(testProject(), MROptions{})
	m.source = "a"
	m.target = "b"
	m.title = "MR: "
	m.step = stepMRTitle

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	flow := updated.(MRFlowModel)
	if flow.title != "MR: x" {
		t.Fatalf("expected title MR: x, got %q", flow.title)
	}

	updated, _ = flow.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	flow = updated.(MRFlowModel)
	if flow.title != "MR: " {
		t.Fatalf("expected backspace to trim, got %q", flow.title)
	}
}
