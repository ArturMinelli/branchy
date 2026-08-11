package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// LoadingOptions configures a loading panel.
type LoadingOptions struct {
	Message  string
	Progress string
}

// LoadingModel shows an animated spinner during async work.
type LoadingModel struct {
	opts    LoadingOptions
	spinner spinner.Model
	active  bool
}

// NewLoading creates a loading panel in the active state.
func NewLoading(opts LoadingOptions) LoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return LoadingModel{
		opts:    opts,
		spinner: s,
		active:  true,
	}
}

// Init returns the spinner tick command.
func (m LoadingModel) Init() tea.Cmd {
	if !m.active {
		return nil
	}
	return m.spinner.Tick
}

// Update handles spinner tick messages.
func (m LoadingModel) Update(msg tea.Msg) (LoadingModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the loading panel.
func (m LoadingModel) View() string {
	var b strings.Builder
	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(m.opts.Message)
	b.WriteString("\n")
	if m.opts.Progress != "" {
		b.WriteString("\n")
		b.WriteString(m.opts.Progress)
	}
	return loadingPanelStyle.Render(b.String())
}

// Active reports whether the loading panel is shown.
func (m LoadingModel) Active() bool {
	return m.active
}

// Clear marks the loading panel inactive.
func (m LoadingModel) Clear() LoadingModel {
	m.active = false
	return m
}

// LoadingMessage formats a standard action message.
func LoadingMessage(action, detail string) string {
	if detail == "" {
		return action
	}
	return fmt.Sprintf("%s %s", action, detail)
}
