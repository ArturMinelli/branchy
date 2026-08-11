package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestProjectsFlowEmptyOrList(t *testing.T) {
	m, err := newProjectsFlowModel()
	if err != nil {
		t.Fatal(err)
	}
	if m.step != stepProjectsList && m.step != stepProjectsEmpty && m.step != stepProjectsError {
		t.Fatalf("unexpected step %d", m.step)
	}
}

func TestProjectsFlowListToDetail(t *testing.T) {
	m, err := newProjectsFlowModel()
	if err != nil {
		t.Fatal(err)
	}
	if m.step != stepProjectsList {
		t.Skip("no projects to test detail view")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(ProjectsFlowModel)
	if flow.step != stepProjectsDetail {
		t.Fatalf("expected stepProjectsDetail, got %d", flow.step)
	}
	if flow.selected == nil {
		t.Fatal("expected selected project")
	}
}

func TestProjectsFlowDetailBack(t *testing.T) {
	m, err := newProjectsFlowModel()
	if err != nil {
		t.Fatal(err)
	}
	if m.step != stepProjectsList {
		t.Skip("no projects to test")
	}
	m.step = stepProjectsDetail

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(ProjectsFlowModel)
	if flow.step != stepProjectsList {
		t.Fatalf("expected stepProjectsList, got %d", flow.step)
	}
}
