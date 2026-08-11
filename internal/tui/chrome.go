package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

const minTermWidth = 80
const minTermHeight = 24

// RenderTitle renders a command header (e.g. "branchy sync").
func RenderTitle(command string) string {
	return titleStyle.Render(command)
}

// RenderHelp renders muted help footer text.
func RenderHelp(text string) string {
	return helpStyle.Render(text)
}

// MinSizeOK reports whether the terminal meets minimum dimensions.
func MinSizeOK(width, height int) bool {
	return width >= minTermWidth && height >= minTermHeight
}

// RenderTooSmall renders a message when the terminal is too small.
func RenderTooSmall() string {
	return errStyle.Render("Terminal too small — resize to at least 80×24 (press q to quit)")
}
