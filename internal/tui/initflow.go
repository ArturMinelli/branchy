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
	stepInitLoading
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
	confirm     ConfirmModel
	loading     LoadingModel
	finished    bool
	cancelled   bool
	flowWindow
}

type initResultMsg struct {
	project *project.Project
	err     error
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
	return m.resetInitConfirm()
}

func (m InitFlowModel) resetInitConfirm() InitFlowModel {
	m.confirm = NewConfirm(ConfirmOptions{
		Context: []string{
			fmt.Sprintf("Repository: %s", m.repoPath),
			fmt.Sprintf("Project ID:   %s", m.projectID),
		},
		Question: "Register this repository with branchy?",
		Width:    m.width,
	})
	return m
}

func runInitCmd() tea.Cmd {
	return func() tea.Msg {
		p, err := project.Init(project.InitOptions{Force: false})
		return initResultMsg{project: p, err: err}
	}
}

func (m InitFlowModel) Init() tea.Cmd { return nil }

func (m InitFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)
		if m.step == stepInitConfirm {
			m = m.resetInitConfirm()
		}
		return m, nil

	case initResultMsg:
		m.loading = m.loading.Clear()
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.step = stepInitError
			return m, nil
		}
		m.projectID = msg.project.ID
		m.repoPath = msg.project.Path
		m.branchCount = len(msg.project.Tree.Names())
		m.step = stepInitSuccess
		return m, nil

	case tea.KeyMsg:
		if m.step == stepInitLoading {
			return m, nil
		}
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

	if m.step == stepInitLoading {
		var cmd tea.Cmd
		m.loading, cmd = m.loading.Update(msg)
		return m, cmd
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
		b.WriteString(m.confirm.View())
	case stepInitLoading:
		b.WriteString(m.loading.View())
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
	if keyMatchesQuit(msg) {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}

	var choice ConfirmChoice
	m.confirm, choice = m.confirm.Update(msg)
	if choice == ConfirmNo {
		m.cancelled = true
		m.finished = true
		return m, tea.Quit
	}
	if choice == ConfirmYes {
		m.step = stepInitLoading
		m.loading = NewLoading(LoadingOptions{Message: "Registering project…"})
		return m, tea.Batch(m.loading.Init(), runInitCmd())
	}
	return m, nil
}
