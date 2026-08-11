package tui

import (
	"fmt"
	"strings"

	"branchy/internal/project"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenProjectPicker screen = iota
	screenTree
	screenSync
	screenLink
	screenUnlinkConfirm
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
	screen        screen
	width         int
	height        int
	projects      []*project.Project
	projectList   list.Model
	treeView      BranchTreeView
	current       *project.Project
	syncFlow      SyncFlowModel
	linkParent    string
	linkChild     string
	linkInput     int
	unlinkTarget  string
	unlinkCount   int
	unlinkConfirm ConfirmModel
	mrFlow        MRFlowModel
	errMsg        string
	quitting      bool
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Sync   key.Binding
	Link   key.Binding
	Unlink key.Binding
	MR     key.Binding
	Quit   key.Binding
	Yes    key.Binding
	No     key.Binding
	Tab    key.Binding
}

var keys = keyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:   key.NewBinding(key.WithKeys("esc", "b"), key.WithHelp("esc/b", "back")),
	Sync:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sync")),
	Link:   key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "link")),
	Unlink: key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unlink")),
	MR:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mr")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Yes:    key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
	No:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
	Tab:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
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

func (m Model) Init() tea.Cmd {
	if m.current != nil {
		return remoteUpdateCmd(m.current)
	}
	return nil
}

func (m Model) applyRemoteUpdate(msg remoteUpdateMsg) Model {
	if m.current == nil || msg.projectID != m.current.ID {
		return m
	}
	if msg.err != nil {
		return m
	}
	counts := loadInboundCounts(m.current.Path, m.current.Tree)
	m.treeView.setInbound(counts)
	if m.screen == screenSync {
		m.syncFlow = m.syncFlow.applyInbound(counts)
	}
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ru, ok := msg.(remoteUpdateMsg); ok {
		return m.applyRemoteUpdate(ru), nil
	}
	if m.screen == screenSync {
		var cmd tea.Cmd
		var syncModel tea.Model
		syncModel, cmd = m.syncFlow.Update(msg)
		m.syncFlow = syncModel.(SyncFlowModel)
		if m.syncFlow.finished {
			m.screen = screenTree
			m.syncFlow = SyncFlowModel{}
		}
		return m, cmd
	}
	if m.screen == screenMR {
		var cmd tea.Cmd
		var mrModel tea.Model
		mrModel, cmd = m.mrFlow.Update(msg)
		m.mrFlow = mrModel.(MRFlowModel)
		if m.mrFlow.finished || m.mrFlow.cancelled {
			m.screen = screenTree
			m.mrFlow = MRFlowModel{}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.projectList.SetWidth(msg.Width)
		m.projectList.SetHeight(msg.Height - 4)
		m.treeView.width = msg.Width
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
		case screenLink:
			return m.updateLink(msg)
		case screenUnlinkConfirm:
			return m.updateUnlink(msg)
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
	if m.screen == screenSync {
		return m.syncFlow.View()
	}

	var b strings.Builder
	b.WriteString(RenderTitle("branchy"))
	b.WriteString("\n\n")

	switch m.screen {
	case screenProjectPicker:
		b.WriteString(m.projectList.View())
		b.WriteString("\n")
		b.WriteString(RenderHelp("enter: open  q: quit"))
	case screenTree:
		if m.current != nil {
			b.WriteString(RenderHelp(m.current.ID + " — " + m.current.Path))
			b.WriteString("\n\n")
			b.WriteString(m.treeView.View())
			b.WriteString("\n")
		}
		b.WriteString(RenderHelp("↑/↓: navigate  s: sync  m: mr  l: link  u: unlink  esc: projects  q: quit"))
	case screenUnlinkConfirm:
		b.WriteString(m.unlinkConfirm.View())
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
		b.WriteString(RenderHelp("type name  tab: switch field  enter: save  esc: cancel"))
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

func (m *Model) selectProject(p *project.Project) tea.Cmd {
	m.current = p
	m.screen = screenTree
	m.errMsg = ""
	m.treeView = newBranchTreeView(p.Tree)
	if p == nil {
		return nil
	}
	m.treeView.setInbound(loadInboundCounts(p.Path, p.Tree))
	return remoteUpdateCmd(p)
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
					return m, m.selectProject(p)
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
			m.syncFlow = newSyncFlowModel(m.current, name, SyncFlowOptions{Embedded: true})
			m.syncFlow.width = m.width
			m.syncFlow.height = m.height
			m.screen = screenSync
			m.errMsg = ""
			return m, m.syncFlow.Init()
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
	if key.Matches(msg, keys.Unlink) {
		name := m.treeView.selectedName()
		if name == "" || m.current == nil {
			return m, nil
		}
		if _, ok := m.current.Tree.Branches[name]; !ok {
			m.errMsg = fmt.Sprintf("branch %q not in tree", name)
			return m, nil
		}
		m.unlinkTarget = name
		m.unlinkCount = len(m.current.Tree.SubtreeNames(name))
		label := "branch"
		if m.unlinkCount != 1 {
			label = "branches"
		}
		m.unlinkConfirm = NewConfirm(ConfirmOptions{
			Question: fmt.Sprintf(`Remove "%s" and %d %s from tree?`, name, m.unlinkCount, label),
			Width:    m.width,
		})
		m.screen = screenUnlinkConfirm
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
		return m, m.selectProject(m.current)
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

func (m Model) updateUnlink(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
		m.quitting = true
		return m, tea.Quit
	}

	var choice ConfirmChoice
	m.unlinkConfirm, choice = m.unlinkConfirm.Update(msg)
	if choice == ConfirmNo {
		m.screen = screenTree
		m.unlinkTarget = ""
		m.unlinkCount = 0
		return m, nil
	}
	if choice == ConfirmYes {
		if err := m.current.Tree.UnlinkSubtree(m.unlinkTarget); err != nil {
			m.errMsg = err.Error()
			m.screen = screenTree
			m.unlinkTarget = ""
			m.unlinkCount = 0
			return m, nil
		}
		if err := m.current.SaveTree(); err != nil {
			m.errMsg = err.Error()
			m.screen = screenTree
			m.unlinkTarget = ""
			m.unlinkCount = 0
			return m, nil
		}
		m.unlinkTarget = ""
		m.unlinkCount = 0
		return m, m.selectProject(m.current)
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
