package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadingActiveByDefault(t *testing.T) {
	m := NewLoading(LoadingOptions{Message: "Working"})
	if !m.Active() {
		t.Fatal("expected loading to be active")
	}
	if m.Init() == nil {
		t.Fatal("expected spinner tick command")
	}
}

func TestLoadingSpinnerTickUpdates(t *testing.T) {
	m := NewLoading(LoadingOptions{Message: "Creating MR"})
	before := m.View()
	updated, cmd := m.Update(spinner.TickMsg{})
	if cmd == nil {
		t.Fatal("expected tick to schedule next tick")
	}
	if updated.View() == before && before == "" {
		t.Fatal("expected view to include spinner output")
	}
}

func TestLoadingViewContainsMessage(t *testing.T) {
	m := NewLoading(LoadingOptions{
		Message:  "Creating MR develop → feature-a",
		Progress: "Edge 2 of 5",
	})
	view := m.View()
	if !strings.Contains(view, "Creating MR develop → feature-a") {
		t.Fatalf("expected message in view, got %q", view)
	}
	if !strings.Contains(view, "Edge 2 of 5") {
		t.Fatalf("expected progress in view, got %q", view)
	}
}

func TestLoadingClearInactive(t *testing.T) {
	m := NewLoading(LoadingOptions{Message: "Working"}).Clear()
	if m.Active() {
		t.Fatal("expected loading inactive after clear")
	}
}

func TestLoadingIgnoresKeys(t *testing.T) {
	m := NewLoading(LoadingOptions{Message: "Working"})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatal("keys should not produce commands on loading model")
	}
	if !updated.Active() {
		t.Fatal("loading should remain active")
	}
}
