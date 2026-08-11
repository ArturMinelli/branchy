package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/config"
	"branchy/internal/git"
	"branchy/internal/project"
)

type initStep int

const (
	stepInitConfirm initStep = iota
	stepInitSuccess
	stepInitError
)

// InitFlowModel is a multi-step Bubble Tea model for interactive project registration.
type InitFlowModel struct {
	step        initStep
	repoPath    string
	projectID   string
	branchCount int
	errMsg      string
	finished    bool
	cancelled   bool
	flowWindow
}

// RunInit starts a standalone init TUI wizard.
func RunInit() error {
	m := newInitFlowModel()
	prog := tea.NewProgram(m, tea.WithAltScreen())
	_, err := prog.Run()
	return err
}

func newInitFlowModel() InitFlowModel {
	m := InitFlowModel{}
	root, err := git.Root("")
	if err != nil {
		m.step = stepInitError
		m.errMsg = err.Error()
		return m
	}
	m.repoPath = root
	m.projectID = config.Slugify(git.SlugFromPath(root))
	m.step = stepInitConfirm
	return m
}

func (m InitFlowModel) Init() tea.Cmd { return nil }

func (m InitFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)
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
		case stepInitConfirm:
			return m.updateConfirm(msg)
		case stepInitSuccess, stepInitError:
			if keyMatchesDone(msg) {
				m.finished = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m InitFlowModel) View() string {
	if m.tooSmall {
		return m.wrap("")
	}

	var b strings.Builder
	b.WriteString(RenderTitle("branchy init"))
	b.WriteString("\n\n")

	switch m.step {
	case stepInitConfirm:
		b.WriteString(fmt.Sprintf("Repository: %s\n", m.repoPath))
		b.WriteString(fmt.Sprintf("Project ID:   %s\n\n", m.projectID))
		b.WriteString("Register this repository with branchy? [y/N]\n\n")
		b.WriteString(RenderHelp("y: register  n/esc: cancel"))
	case stepInitSuccess:
		b.WriteString(okStyle.Render(fmt.Sprintf("Registered %q at %s", m.projectID, m.repoPath)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Branches in tree: %d\n\n", m.branchCount))
		b.WriteString(RenderHelp("enter: done"))
	case stepInitError:
		b.WriteString(errStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter/esc: exit"))
	}

	return b.String()
}

func (m InitFlowModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if keyMatchesNo(msg) || keyMatchesBack(msg) {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if keyMatchesQuit(msg) {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if keyMatchesYes(msg) || keyMatchesEnter(msg) {
		p, err := project.Init(project.InitOptions{Force: false})
		if err != nil {
			m.errMsg = err.Error()
			m.step = stepInitError
			return m, nil
		}
		m.projectID = p.ID
		m.repoPath = p.Path
		m.branchCount = len(p.Tree.Names())
		m.step = stepInitSuccess
		return m, nil
	}
	return m, nil
}
