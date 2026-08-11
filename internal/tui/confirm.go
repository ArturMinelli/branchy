// Confirm and loading components are TUI-only. Scripted CLI uses stdin prompts
// in internal/sync and other plain CLI paths.
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmChoice is the outcome of a resolved confirm interaction.
type ConfirmChoice int

const (
	ConfirmNone ConfirmChoice = iota
	ConfirmYes
	ConfirmNo
)

// ConfirmOptions configures a native yes/no confirm panel.
type ConfirmOptions struct {
	Context  []string
	Question string
	Progress string
	Width    int
}

// ConfirmModel is a framed TUI yes/no decision panel.
type ConfirmModel struct {
	opts  ConfirmOptions
	focus int // 0 = No, 1 = Yes
}

// NewConfirm creates a confirm panel with focus on No.
func NewConfirm(opts ConfirmOptions) ConfirmModel {
	return ConfirmModel{opts: opts, focus: 0}
}

// FocusIndex returns 0 for No and 1 for Yes.
func (m ConfirmModel) FocusIndex() int {
	return m.focus
}

// Update handles keyboard input for the confirm panel.
func (m ConfirmModel) Update(msg tea.KeyMsg) (ConfirmModel, ConfirmChoice) {
	if key.Matches(msg, flowKeys.Yes) {
		return m, ConfirmYes
	}
	if key.Matches(msg, flowKeys.No) {
		return m, ConfirmNo
	}
	if key.Matches(msg, flowKeys.Left) || key.Matches(msg, flowKeys.ShiftTab) {
		if m.focus > 0 {
			m.focus--
		}
		return m, ConfirmNone
	}
	if key.Matches(msg, flowKeys.Right) || key.Matches(msg, flowKeys.Tab) {
		if m.focus < 1 {
			m.focus++
		}
		return m, ConfirmNone
	}
	if key.Matches(msg, flowKeys.Enter) {
		if m.focus == 1 {
			return m, ConfirmYes
		}
		return m, ConfirmNo
	}
	if key.Matches(msg, flowKeys.Back) {
		return m, ConfirmNo
	}
	return m, ConfirmNone
}

// View renders the confirm panel.
func (m ConfirmModel) View() string {
	var b strings.Builder
	for _, line := range m.opts.Context {
		b.WriteString(line)
		b.WriteString("\n")
	}
	if m.opts.Question != "" {
		b.WriteString(m.opts.Question)
		b.WriteString("\n")
	}
	if m.opts.Progress != "" {
		b.WriteString("\n")
		b.WriteString(m.opts.Progress)
	}
	b.WriteString("\n\n")

	noLabel := confirmButtonStyle.Render(" No ")
	yesLabel := confirmButtonStyle.Render(" Yes ")
	if m.focus == 0 {
		noLabel = confirmButtonFocusStyle.Render(" No ")
	} else {
		yesLabel = confirmButtonFocusStyle.Render(" Yes ")
	}
	b.WriteString("   ")
	b.WriteString(noLabel)
	b.WriteString("    ")
	b.WriteString(yesLabel)
	b.WriteString("\n\n")
	b.WriteString(RenderConfirmHelp())

	content := b.String()
	if m.opts.Width > 0 {
		return confirmBorderStyle.Width(m.opts.Width - 4).Render(content)
	}
	return confirmBorderStyle.Render(content)
}

// RenderConfirmHelp renders the standard confirm key hints.
func RenderConfirmHelp() string {
	return RenderHelp("←/→: select  enter: confirm  y/n: shortcut  esc: decline")
}
