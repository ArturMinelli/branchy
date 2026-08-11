package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/project"
)

type projectsStep int

const (
	stepProjectsList projectsStep = iota
	stepProjectsDetail
	stepProjectsEmpty
	stepProjectsError
)

// ProjectsFlowModel is a read-only Bubble Tea model for browsing registered projects.
type ProjectsFlowModel struct {
	step        projectsStep
	projects    []*project.Project
	projectList list.Model
	selected    *project.Project
	errMsg      string
	finished    bool
	flowWindow
}

// RunProjects starts a standalone projects browser TUI.
func RunProjects() error {
	m, err := newProjectsFlowModel()
	if err != nil {
		return err
	}
	prog := tea.NewProgram(m, tea.WithAltScreen())
	_, err = prog.Run()
	return err
}

func newProjectsFlowModel() (ProjectsFlowModel, error) {
	projects, err := project.ListAll()
	if err != nil {
		return ProjectsFlowModel{step: stepProjectsError, errMsg: err.Error()}, nil
	}
	m := ProjectsFlowModel{projects: projects}
	if len(projects) == 0 {
		m.step = stepProjectsEmpty
		return m, nil
	}
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = projectItem{id: p.ID, path: p.Path}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Registered projects"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	m.projectList = l
	m.step = stepProjectsList
	return m, nil
}

func (m ProjectsFlowModel) Init() tea.Cmd { return nil }

func (m ProjectsFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)
		m.projectList.SetWidth(msg.Width)
		m.projectList.SetHeight(msg.Height - 6)
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
		case stepProjectsList:
			return m.updateList(msg)
		case stepProjectsDetail:
			if keyMatchesBack(msg) || keyMatchesQuit(msg) {
				m.step = stepProjectsList
				m.selected = nil
				return m, nil
			}
		case stepProjectsEmpty, stepProjectsError:
			if keyMatchesDone(msg) {
				m.finished = true
				return m, tea.Quit
			}
		}
	}

	if m.step == stepProjectsList {
		var cmd tea.Cmd
		m.projectList, cmd = m.projectList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ProjectsFlowModel) View() string {
	if m.tooSmall {
		return m.wrap("")
	}

	var b strings.Builder
	b.WriteString(RenderTitle("branchy projects"))
	b.WriteString("\n\n")

	switch m.step {
	case stepProjectsList:
		b.WriteString(m.projectList.View())
		b.WriteString("\n")
		b.WriteString(RenderHelp("↑/↓: navigate  enter: details  q: quit"))
	case stepProjectsDetail:
		if m.selected != nil {
			b.WriteString(fmt.Sprintf("ID:       %s\n", m.selected.ID))
			b.WriteString(fmt.Sprintf("Path:     %s\n", m.selected.Path))
			b.WriteString(fmt.Sprintf("Branches: %d\n\n", len(m.selected.Tree.Names())))
		}
		b.WriteString(RenderHelp("esc: back  q: quit"))
	case stepProjectsEmpty:
		b.WriteString(warnStyle.Render("No projects registered."))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter/esc: exit"))
	case stepProjectsError:
		b.WriteString(errStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter/esc: exit"))
	}

	return b.String()
}

func (m ProjectsFlowModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if keyMatchesQuit(msg) {
		m.finished = true
		return m, tea.Quit
	}
	if keyMatchesEnter(msg) {
		item, ok := m.projectList.SelectedItem().(projectItem)
		if !ok {
			return m, nil
		}
		for _, p := range m.projects {
			if p.ID == item.id {
				m.selected = p
				break
			}
		}
		m.step = stepProjectsDetail
		return m, nil
	}
	var cmd tea.Cmd
	m.projectList, cmd = m.projectList.Update(msg)
	return m, cmd
}
