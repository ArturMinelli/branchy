package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitFlowStartsAtConfirm(t *testing.T) {
	m := newInitFlowModel()
	if m.step != stepInitConfirm && m.step != stepInitError {
		t.Fatalf("expected stepInitConfirm or stepInitError, got %d", m.step)
	}
	if m.step == stepInitConfirm && m.repoPath == "" {
		t.Fatal("expected repo path when in git repo")
	}
}

func TestInitFlowConfirmCancel(t *testing.T) {
	m := newInitFlowModel()
	if m.step == stepInitError {
		t.Skip("not in a git repo")
	}
	m.step = stepInitConfirm

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(InitFlowModel)
	if !flow.cancelled {
		t.Fatal("expected cancelled")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
