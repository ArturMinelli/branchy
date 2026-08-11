package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"branchy/internal/project"
	"branchy/internal/sync"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type screen int

const (
	screenProjectPicker screen = iota
	screenTree
	screenSync
	screenLink
	screenMR
	screenDone
)

type projectItem struct {
	id   string
	path string
}

func (i projectItem) Title() string       { return i.id }
func (i projectItem) Description() string { return i.path }
func (i projectItem) FilterValue() string { return i.id + " " + i.path }

// Model is the root Bubble Tea model.
type Model struct {
	screen      screen
	width       int
	height      int
	projects    []*project.Project
	projectList list.Model
	treeView    BranchTreeView
	current     *project.Project
	syncFrom    string
	syncSummary *sync.Summary
	linkParent  string
	linkChild   string
	linkInput   int
	mrFlow      MRFlowModel
	errMsg      string
	quitting    bool
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Sync   key.Binding
	Link   key.Binding
	MR     key.Binding
	Quit   key.Binding
	Yes    key.Binding
	No     key.Binding
	Tab    key.Binding
}

var keys = keyMap{
	Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:  key.NewBinding(key.WithKeys("esc", "b"), key.WithHelp("esc/b", "back")),
	Sync:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sync")),
	Link:  key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "link")),
	MR:    key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mr")),
	Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Yes:   key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
	No:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
	Tab:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
}

// Run starts the TUI for an optional pre-resolved project.
func Run(preselected *project.Project) error {
	projects, err := project.ListAll()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return fmt.Errorf("no projects registered (run branchy init inside a git repo)")
	}

	m := newModel(projects, preselected)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func newModel(projects []*project.Project, preselected *project.Project) Model {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = projectItem{id: p.ID, path: p.Path}
	}

	pl := list.New(items, list.NewDefaultDelegate(), 0, 0)
	pl.Title = "Projects"
	pl.SetShowStatusBar(false)
	pl.SetFilteringEnabled(false)

	bl := newBranchTreeView(nil)

	m := Model{
		screen:      screenProjectPicker,
		projects:    projects,
		projectList: pl,
		treeView:    bl,
	}

	if preselected != nil {
		m.selectProject(preselected)
	} else if len(projects) == 1 {
		m.selectProject(projects[0])
	}

	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.projectList.SetWidth(msg.Width)
		m.projectList.SetHeight(msg.Height - 4)
		m.treeView.width = msg.Width
		if m.screen == screenMR {
			m.mrFlow.width = msg.Width
			m.mrFlow.height = msg.Height
			m.mrFlow.branchList.SetWidth(msg.Width)
			m.mrFlow.branchList.SetHeight(msg.Height - 6)
		}
		return m, nil

	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}

		switch m.screen {
		case screenProjectPicker:
			return m.updateProjectPicker(msg)
		case screenTree:
			return m.updateTree(msg)
		case screenSync:
			return m.updateSync(msg)
		case screenLink:
			return m.updateLink(msg)
		case screenMR:
			var cmd tea.Cmd
			var mrModel tea.Model
			mrModel, cmd = m.mrFlow.Update(msg)
			m.mrFlow = mrModel.(MRFlowModel)
			if m.mrFlow.finished || m.mrFlow.cancelled {
				m.screen = screenTree
				m.mrFlow = MRFlowModel{}
			}
			return m, cmd
		case screenDone:
			if key.Matches(msg, keys.Quit) || key.Matches(msg, keys.Enter) {
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenProjectPicker:
		m.projectList, cmd = m.projectList.Update(msg)
	default:
	}
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.screen == screenMR {
		return m.mrFlow.View()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("branchy"))
	b.WriteString("\n\n")

	switch m.screen {
	case screenProjectPicker:
		b.WriteString(m.projectList.View())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("enter: open  q: quit"))
	case screenTree:
		if m.current != nil {
			b.WriteString(helpStyle.Render(m.current.ID + " — " + m.current.Path))
			b.WriteString("\n\n")
			b.WriteString(m.treeView.View())
			b.WriteString("\n")
		}
		b.WriteString(helpStyle.Render("↑/↓: navigate  s: sync  m: mr  l: link  esc: projects  q: quit"))
	case screenSync:
		b.WriteString(fmt.Sprintf("Sync from %s\n\n", m.syncFrom))
		if m.syncSummary != nil {
			for _, r := range m.syncSummary.Results {
				line := fmt.Sprintf("%s → %s: %s", r.Parent, r.Child, r.Action)
				switch r.Action {
				case "created":
					line = okStyle.Render(line)
				case "failed":
					line = errStyle.Render(line + " — " + r.Message)
				default:
					line = warnStyle.Render(line)
				}
				b.WriteString(line)
				if r.URL != "" {
					b.WriteString("\n  " + r.URL)
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
			b.WriteString(helpStyle.Render("enter/q: back to tree"))
		} else {
			b.WriteString("Press y to confirm sync, n to cancel\n")
			b.WriteString(helpStyle.Render("y: confirm  n: cancel"))
		}
	case screenLink:
		b.WriteString("Link branch\n\n")
		parentMark, childMark := " ", " "
		if m.linkInput == 0 {
			parentMark = "▸"
		} else {
			childMark = "▸"
		}
		b.WriteString(fmt.Sprintf("%s parent: %s\n", parentMark, m.linkParent))
		b.WriteString(fmt.Sprintf("%s child:  %s\n", childMark, m.linkChild))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("type name  tab: switch field  enter: save  esc: cancel"))
	case screenDone:
		if m.errMsg != "" {
			b.WriteString(errStyle.Render(m.errMsg))
		}
	}

	if m.errMsg != "" && m.screen != screenDone {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.errMsg))
	}

	return b.String()
}

func (m *Model) selectProject(p *project.Project) {
	m.current = p
	m.screen = screenTree
	m.errMsg = ""
	m.treeView = newBranchTreeView(p.Tree)
}

func (m Model) updateProjectPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
		m.quitting = true
		return m, tea.Quit
	}
	if key.Matches(msg, keys.Enter) {
		if item, ok := m.projectList.SelectedItem().(projectItem); ok {
			for _, p := range m.projects {
				if p.ID == item.id {
					m.selectProject(p)
					break
				}
			}
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.projectList, cmd = m.projectList.Update(msg)
	return m, cmd
}

func (m Model) updateTree(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
		m.quitting = true
		return m, tea.Quit
	}
	if key.Matches(msg, keys.Back) {
		if len(m.projects) > 1 {
			m.screen = screenProjectPicker
		}
		return m, nil
	}
	if key.Matches(msg, keys.Sync) {
		if name := m.treeView.selectedName(); name != "" {
			m.syncFrom = name
			m.syncSummary = nil
			m.screen = screenSync
			m.errMsg = ""
		}
		return m, nil
	}
	if key.Matches(msg, keys.Link) {
		m.linkParent = m.treeView.selectedName()
		m.linkChild = ""
		m.linkInput = 1
		if m.linkParent == "" {
			m.linkInput = 0
		}
		m.screen = screenLink
		m.errMsg = ""
		return m, nil
	}
	if key.Matches(msg, keys.MR) {
		if m.current == nil {
			return m, nil
		}
		m.mrFlow = newMRFlowModel(m.current, MROptions{
			PrefilledSource: m.treeView.selectedName(),
			Embedded:        true,
		})
		m.mrFlow.width = m.width
		m.mrFlow.height = m.height
		m.mrFlow.branchList.SetWidth(m.width)
		m.mrFlow.branchList.SetHeight(m.height - 6)
		m.screen = screenMR
		m.errMsg = ""
		return m, nil
	}
	if key.Matches(msg, keys.Up) {
		m.treeView.moveUp()
		return m, nil
	}
	if key.Matches(msg, keys.Down) {
		m.treeView.moveDown()
		return m, nil
	}
	return m, nil
}

func (m Model) updateSync(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.syncSummary != nil {
		if key.Matches(msg, keys.Quit) || key.Matches(msg, keys.Enter) || key.Matches(msg, keys.Back) {
			m.screen = screenTree
			m.syncSummary = nil
		}
		return m, nil
	}

	if key.Matches(msg, keys.No) || key.Matches(msg, keys.Back) {
		m.screen = screenTree
		return m, nil
	}
	if key.Matches(msg, keys.Yes) || key.Matches(msg, keys.Enter) {
		summary, err := sync.Run(m.current, sync.Options{
			FromBranch: m.syncFrom,
			Confirm: func(parent, child string) (bool, error) {
				return true, nil
			},
		})
		if err != nil {
			m.errMsg = err.Error()
			m.screen = screenTree
			return m, nil
		}
		m.syncSummary = summary
		return m, nil
	}
	return m, nil
}

func (m Model) updateLink(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Back) {
		m.screen = screenTree
		return m, nil
	}
	if key.Matches(msg, keys.Tab) {
		m.linkInput = 1 - m.linkInput
		return m, nil
	}
	if key.Matches(msg, keys.Enter) {
		if err := m.current.Tree.Link(m.linkParent, m.linkChild); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		if err := m.current.SaveTree(); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.selectProject(m.current)
		return m, nil
	}

	if key.Matches(msg, keys.Quit) {
		m.quitting = true
		return m, tea.Quit
	}

	ch := msg.String()
	if len(ch) != 1 {
		return m, nil
	}
	if ch == "backspace" {
		if m.linkInput == 0 {
			m.linkParent = trimLast(m.linkParent)
		} else {
			m.linkChild = trimLast(m.linkChild)
		}
		return m, nil
	}

	if ch < " " || ch > "~" {
		return m, nil
	}
	if m.linkInput == 0 {
		m.linkParent += ch
	} else {
		m.linkChild += ch
	}
	return m, nil
}

func trimLast(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}
