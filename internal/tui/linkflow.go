package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
)

type linkStep int

const (
	stepLinkParent linkStep = iota
	stepLinkChild
	stepLinkConfirm
	stepLinkSuccess
	stepLinkError
)

// LinkFlowModel is a multi-step Bubble Tea model for interactive link.
type LinkFlowModel struct {
	project    *project.Project
	step       linkStep
	parent     string
	child      string
	branchList list.Model
	errMsg     string
	confirm    ConfirmModel
	finished   bool
	cancelled  bool
	flowWindow
}

// RunLink starts a standalone link TUI for the given project.
func RunLink(p *project.Project) error {
	m := newLinkFlowModel(p)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	_, err := prog.Run()
	return err
}

func newLinkFlowModel(p *project.Project) LinkFlowModel {
	names := p.Tree.Names()
	m := LinkFlowModel{project: p}
	if len(names) == 0 {
		m.step = stepLinkError
		m.errMsg = "branch tree is empty"
		return m
	}
	m.step = stepLinkParent
	m.branchList = newBranchList(names, "Select parent branch")
	return m
}

func (m LinkFlowModel) Init() tea.Cmd { return nil }

func (m LinkFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)
		m.branchList.SetWidth(msg.Width)
		m.branchList.SetHeight(msg.Height - 6)
		return m, nil

	case tea.KeyMsg:
		if m.tooSmall {
			if keyMatchesQuit(msg) {
				m.finished = true
				return m, tea.Quit
			}
			return m, nil
		}

		switch m.step {
		case stepLinkParent:
			return m.updateParentPicker(msg)
		case stepLinkChild:
			return m.updateChildInput(msg)
		case stepLinkConfirm:
			return m.updateConfirm(msg)
		case stepLinkSuccess, stepLinkError:
			if keyMatchesDone(msg) {
				m.finished = true
				return m, tea.Quit
			}
		}
	}

	if m.step == stepLinkParent {
		var cmd tea.Cmd
		m.branchList, cmd = m.branchList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m LinkFlowModel) View() string {
	if m.tooSmall {
		return m.wrap("")
	}

	var b strings.Builder
	b.WriteString(RenderTitle("branchy link"))
	b.WriteString("\n\n")

	switch m.step {
	case stepLinkParent:
		b.WriteString(m.branchList.View())
		b.WriteString("\n")
		b.WriteString(RenderHelp("↑/↓: navigate  enter: select  esc: cancel  q: quit"))
	case stepLinkChild:
		b.WriteString(fmt.Sprintf("Parent: %s\n\n", m.parent))
		b.WriteString("Child branch name:\n")
		b.WriteString(m.child)
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("type name  enter: continue  esc: cancel"))
	case stepLinkConfirm:
		b.WriteString(m.confirm.View())
	case stepLinkSuccess:
		b.WriteString(okStyle.Render(fmt.Sprintf("Linked %s → %s", m.parent, m.child)))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter: done"))
	case stepLinkError:
		b.WriteString(errStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter/esc: exit"))
	}

	return b.String()
}

func (m LinkFlowModel) updateParentPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if keyMatchesQuit(msg) {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if keyMatchesBack(msg) {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if keyMatchesEnter(msg) {
		item, ok := m.branchList.SelectedItem().(branchItem)
		if !ok {
			return m, nil
		}
		m.parent = item.name
		m.step = stepLinkChild
		return m, nil
	}
	var cmd tea.Cmd
	m.branchList, cmd = m.branchList.Update(msg)
	return m, cmd
}

func (m LinkFlowModel) updateChildInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if keyMatchesBack(msg) {
		m.step = stepLinkParent
		m.branchList = newBranchList(m.project.Tree.Names(), "Select parent branch")
		m.branchList.SetWidth(m.width)
		m.branchList.SetHeight(m.height - 6)
		return m, nil
	}
	if keyMatchesQuit(msg) {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if keyMatchesEnter(msg) {
		if strings.TrimSpace(m.child) == "" {
			return m, nil
		}
		m.step = stepLinkConfirm
		m.confirm = NewConfirm(ConfirmOptions{
			Question: fmt.Sprintf("Link %s → %s?", m.parent, m.child),
			Width:    m.width,
		})
		return m, nil
	}
	ch := msg.String()
	if ch == "backspace" {
		m.child = trimLast(m.child)
		return m, nil
	}
	if len(ch) != 1 || ch < " " || ch > "~" {
		return m, nil
	}
	m.child += ch
	return m, nil
}

func (m LinkFlowModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var choice ConfirmChoice
	m.confirm, choice = m.confirm.Update(msg)
	if choice == ConfirmNo {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if choice == ConfirmYes {
		if err := m.project.Link(m.parent, m.child); err != nil {
			m.errMsg = err.Error()
			m.step = stepLinkError
			return m, nil
		}
		m.step = stepLinkSuccess
		return m, nil
	}
	return m, nil
}

func keyMatchesQuit(msg tea.KeyMsg) bool {
	return key.Matches(msg, flowKeys.Quit)
}

func keyMatchesBack(msg tea.KeyMsg) bool {
	return key.Matches(msg, flowKeys.Back)
}

func keyMatchesEnter(msg tea.KeyMsg) bool {
	return key.Matches(msg, flowKeys.Enter)
}

func keyMatchesYes(msg tea.KeyMsg) bool {
	return key.Matches(msg, flowKeys.Yes)
}

func keyMatchesNo(msg tea.KeyMsg) bool {
	return key.Matches(msg, flowKeys.No)
}

func keyMatchesDone(msg tea.KeyMsg) bool {
	return keyMatchesEnter(msg) || keyMatchesBack(msg) || keyMatchesQuit(msg)
}
