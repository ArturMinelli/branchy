package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
)

type unlinkStep int

const (
	stepUnlinkPick unlinkStep = iota
	stepUnlinkConfirm
	stepUnlinkSuccess
	stepUnlinkError
)

// UnlinkFlowModel is a multi-step Bubble Tea model for interactive unlink.
type UnlinkFlowModel struct {
	project      *project.Project
	step         unlinkStep
	target       string
	subtreeCount int
	branchList   list.Model
	errMsg       string
	confirm      ConfirmModel
	finished     bool
	cancelled    bool
	flowWindow
}

// RunUnlink starts a standalone unlink TUI for the given project.
func RunUnlink(p *project.Project) error {
	m := newUnlinkFlowModel(p)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	_, err := prog.Run()
	return err
}

func newUnlinkFlowModel(p *project.Project) UnlinkFlowModel {
	names := p.Tree.Names()
	m := UnlinkFlowModel{project: p}
	if len(names) == 0 {
		m.step = stepUnlinkError
		m.errMsg = "branch tree is empty"
		return m
	}
	m.step = stepUnlinkPick
	m.branchList = newBranchList(names, "Select branch to unlink")
	return m
}

func (m UnlinkFlowModel) Init() tea.Cmd { return nil }

func (m UnlinkFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case stepUnlinkPick:
			return m.updatePick(msg)
		case stepUnlinkConfirm:
			return m.updateConfirm(msg)
		case stepUnlinkSuccess, stepUnlinkError:
			if keyMatchesDone(msg) {
				m.finished = true
				return m, tea.Quit
			}
		}
	}

	if m.step == stepUnlinkPick {
		var cmd tea.Cmd
		m.branchList, cmd = m.branchList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m UnlinkFlowModel) View() string {
	if m.tooSmall {
		return m.wrap("")
	}

	var b strings.Builder
	b.WriteString(RenderTitle("branchy unlink"))
	b.WriteString("\n\n")

	switch m.step {
	case stepUnlinkPick:
		b.WriteString(m.branchList.View())
		b.WriteString("\n")
		b.WriteString(RenderHelp("↑/↓: navigate  enter: select  esc: cancel  q: quit"))
	case stepUnlinkConfirm:
		b.WriteString(m.confirm.View())
	case stepUnlinkSuccess:
		b.WriteString(okStyle.Render(fmt.Sprintf("Removed %s and %d branches from tree", m.target, m.subtreeCount)))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter: done"))
	case stepUnlinkError:
		b.WriteString(errStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter/esc: exit"))
	}

	return b.String()
}

func (m UnlinkFlowModel) updatePick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if keyMatchesQuit(msg) || keyMatchesBack(msg) {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if keyMatchesEnter(msg) {
		item, ok := m.branchList.SelectedItem().(branchItem)
		if !ok {
			return m, nil
		}
		m.target = item.name
		m.subtreeCount = len(m.project.Tree.SubtreeNames(item.name))
		m.step = stepUnlinkConfirm
		label := "branch"
		if m.subtreeCount != 1 {
			label = "branches"
		}
		m.confirm = NewConfirm(ConfirmOptions{
			Question: fmt.Sprintf(`Remove "%s" and %d %s from tree?`, m.target, m.subtreeCount, label),
			Width:    m.width,
		})
		return m, nil
	}
	var cmd tea.Cmd
	m.branchList, cmd = m.branchList.Update(msg)
	return m, cmd
}

func (m UnlinkFlowModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var choice ConfirmChoice
	m.confirm, choice = m.confirm.Update(msg)
	if choice == ConfirmNo {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if choice == ConfirmYes {
		if err := m.project.Tree.UnlinkSubtree(m.target); err != nil {
			m.errMsg = err.Error()
			m.step = stepUnlinkError
			return m, nil
		}
		if err := m.project.SaveTree(); err != nil {
			m.errMsg = err.Error()
			m.step = stepUnlinkError
			return m, nil
		}
		m.step = stepUnlinkSuccess
		return m, nil
	}
	return m, nil
}
