package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type flowKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Yes      key.Binding
	No       key.Binding
}

var flowKeys = flowKeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k")),
	Down:     key.NewBinding(key.WithKeys("down", "j")),
	Left:     key.NewBinding(key.WithKeys("left", "h")),
	Right:    key.NewBinding(key.WithKeys("right", "l")),
	Tab:      key.NewBinding(key.WithKeys("tab")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
	Enter:    key.NewBinding(key.WithKeys("enter")),
	Back:     key.NewBinding(key.WithKeys("esc", "b")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Yes:      key.NewBinding(key.WithKeys("y")),
	No:       key.NewBinding(key.WithKeys("n")),
}

type flowWindow struct {
	width    int
	height   int
	tooSmall bool
}

func (w *flowWindow) onResize(msg tea.WindowSizeMsg) {
	w.width = msg.Width
	w.height = msg.Height
	w.tooSmall = !MinSizeOK(msg.Width, msg.Height)
}

func (w flowWindow) wrap(content string) string {
	if w.tooSmall {
		return RenderTooSmall() + "\n\n" + RenderHelp("q: quit")
	}
	return content
}
