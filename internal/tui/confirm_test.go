package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmDefaultFocusNo(t *testing.T) {
	m := NewConfirm(ConfirmOptions{Question: "Proceed?"})
	if m.FocusIndex() != 0 {
		t.Fatalf("expected default focus on No, got %d", m.FocusIndex())
	}
	view := m.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("confirm view must not contain [y/N]")
	}
	if !strings.Contains(view, " No ") {
		t.Fatal("expected No button in view")
	}
}

func TestConfirmArrowNavigation(t *testing.T) {
	m := NewConfirm(ConfirmOptions{Question: "Proceed?"})
	m, choice := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if choice != ConfirmNone {
		t.Fatalf("expected no choice on arrow, got %d", choice)
	}
	if m.FocusIndex() != 1 {
		t.Fatal("expected focus on Yes after right arrow")
	}
	m, choice = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if choice != ConfirmNone || m.FocusIndex() != 0 {
		t.Fatal("expected focus back on No after left arrow")
	}
}

func TestConfirmShortcuts(t *testing.T) {
	m := NewConfirm(ConfirmOptions{Question: "Proceed?"})
	_, choice := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if choice != ConfirmYes {
		t.Fatal("expected y to select Yes")
	}
	m = NewConfirm(ConfirmOptions{Question: "Proceed?"})
	_, choice = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if choice != ConfirmNo {
		t.Fatal("expected n to select No")
	}
	m = NewConfirm(ConfirmOptions{Question: "Proceed?"})
	_, choice = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if choice != ConfirmNo {
		t.Fatal("expected esc to select No")
	}
}

func TestConfirmEnterSelectsFocus(t *testing.T) {
	m := NewConfirm(ConfirmOptions{Question: "Proceed?"})
	_, choice := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if choice != ConfirmNo {
		t.Fatal("expected enter on default No focus to select No")
	}
	m = NewConfirm(ConfirmOptions{Question: "Proceed?"})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	_, choice = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if choice != ConfirmYes {
		t.Fatal("expected enter on Yes focus to select Yes")
	}
}
