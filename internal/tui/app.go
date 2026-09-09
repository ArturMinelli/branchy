package tui

import (
	"fmt"
	"strings"

	"branchy/internal/project"
	"branchy/internal/sync"
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
	screenUnlink
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
	direction   sync.Direction
	syncFlow    SyncFlowModel
	linkFlow    LinkFlowModel
	unlinkFlow  UnlinkFlowModel
	mrFlow      MRFlowModel
	errMsg         string
	quitting       bool
	reloadSnapshot *countSnapshot
}

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Enter         key.Binding
	Back          key.Binding
	Sync          key.Binding
	Link          key.Binding
	Unlink        key.Binding
	MR            key.Binding
	DirectionUp   key.Binding
	DirectionDown key.Binding
	Reload        key.Binding
	Quit          key.Binding
}

var keys = keyMap{
	Up:            key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:          key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:         key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:          key.NewBinding(key.WithKeys("esc", "b"), key.WithHelp("esc/b", "back")),
	Sync:          key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sync")),
	Link:          key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "link")),
	Unlink:        key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "unlink")),
	MR:            key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mr")),
	DirectionUp:   key.NewBinding(key.WithKeys("ctrl+up"), key.WithHelp("ctrl+↑", "outbound")),
	DirectionDown: key.NewBinding(key.WithKeys("ctrl+down"), key.WithHelp("ctrl+↓", "inbound")),
	Reload:        key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reload")),
	Quit:          key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
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

func (m Model) beginReload() Model {
	if m.current == nil || m.reloadSnapshot != nil {
		return m
	}
	m.reloadSnapshot = snapshotCounts(m.treeView.inbound, m.treeView.outbound)
	m.treeView.setFileCounts(nil, nil)
	return m
}

func (m Model) applyRemoteUpdate(msg remoteUpdateMsg) Model {
	if m.current == nil || msg.projectID != m.current.ID {
		return m
	}
	if msg.err != nil {
		if m.reloadSnapshot != nil {
			var in, out map[string]fileChangeCount
			m.reloadSnapshot.Restore(&in, &out)
			m.treeView.setFileCounts(in, out)
			m.reloadSnapshot = nil
			if m.screen == screenSync {
				m.syncFlow = m.syncFlow.restoreReloadSnapshot()
			}
			return m
		}
		if m.screen == screenSync && m.syncFlow.reloadSnapshot != nil {
			m.syncFlow = m.syncFlow.restoreReloadSnapshot()
			return m
		}
		if m.treeView.inbound == nil && m.treeView.outbound == nil {
			return m.withFreshFileCounts()
		}
		return m
	}
	m.reloadSnapshot = nil
	m = m.withFreshFileCounts()
	if m.screen == screenSync {
		m.syncFlow = m.syncFlow.applyFileCounts(m.treeView.inbound, m.treeView.outbound)
		m.syncFlow.clearReloadSnapshot()
	}
	return m
}

func (m Model) withFreshFileCounts() Model {
	in := loadInboundCounts(m.current.Path, m.current.Tree)
	out := loadOutboundCounts(m.current.Path, m.current.Tree)
	m.treeView.setFileCounts(in, out)
	m.treeView.setDirection(m.direction)
	return m
}

func (m Model) returnToTreeWithRefresh() (Model, tea.Cmd) {
	m.screen = screenTree
	if m.current == nil {
		return m, nil
	}
	return m, remoteUpdateCmd(m.current)
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
			m.syncFlow = SyncFlowModel{}
			return m.returnToTreeWithRefresh()
		}
		return m, cmd
	}
	if m.screen == screenMR {
		var cmd tea.Cmd
		var mrModel tea.Model
		mrModel, cmd = m.mrFlow.Update(msg)
		m.mrFlow = mrModel.(MRFlowModel)
		if m.mrFlow.finished || m.mrFlow.cancelled {
			m.mrFlow = MRFlowModel{}
			return m.returnToTreeWithRefresh()
		}
		return m, cmd
	}
	if m.screen == screenLink {
		var cmd tea.Cmd
		var linkModel tea.Model
		linkModel, cmd = m.linkFlow.Update(msg)
		m.linkFlow = linkModel.(LinkFlowModel)
		if m.linkFlow.finished {
			return m.finishEmbeddedLink()
		}
		return m, cmd
	}
	if m.screen == screenUnlink {
		var cmd tea.Cmd
		var unlinkModel tea.Model
		unlinkModel, cmd = m.unlinkFlow.Update(msg)
		m.unlinkFlow = unlinkModel.(UnlinkFlowModel)
		if m.unlinkFlow.finished {
			return m.finishEmbeddedUnlink()
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
	if m.screen == screenLink {
		return m.linkFlow.View()
	}
	if m.screen == screenUnlink {
		return m.unlinkFlow.View()
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
		for i, line := range strings.Split(treeHelpFooter(m.direction), "\n") {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(RenderHelp(line))
		}
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
	m.treeView.setDirection(m.direction)
	// Counts load after the background fetch in applyRemoteUpdate so
	// remote-tracking refs are fresh; badges show ? until then.
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
			m.syncFlow = newSyncFlowModel(m.current, name, SyncFlowOptions{
				Embedded:  true,
				Direction: m.direction,
			})
			m.syncFlow.width = m.width
			m.syncFlow.height = m.height
			m.screen = screenSync
			m.errMsg = ""
			return m, m.syncFlow.Init()
		}
		return m, nil
	}
	if key.Matches(msg, keys.Link) {
		m.linkFlow = newLinkFlowModel(m.current, LinkFlowOptions{
			Embedded:      true,
			PrefillParent: m.treeView.selectedName(),
		})
		m.sizeLinkFlow()
		m.screen = screenLink
		m.errMsg = ""
		return m, m.linkFlow.Init()
	}
	if key.Matches(msg, keys.Unlink) {
		name := m.treeView.selectedName()
		if name == "" || m.current == nil {
			return m, nil
		}
		m.unlinkFlow = newUnlinkFlowModel(m.current, UnlinkFlowOptions{
			Embedded:      true,
			PrefillTarget: name,
		})
		m.sizeUnlinkFlow()
		m.screen = screenUnlink
		m.errMsg = ""
		return m, m.unlinkFlow.Init()
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
	if key.Matches(msg, keys.Reload) {
		if m.current != nil {
			m = m.beginReload()
			return m, remoteUpdateCmd(m.current)
		}
		return m, nil
	}
	if key.Matches(msg, keys.DirectionUp) {
		m.direction = sync.Upward
		m.treeView.setDirection(m.direction)
		return m, nil
	}
	if key.Matches(msg, keys.DirectionDown) {
		m.direction = sync.Downward
		m.treeView.setDirection(m.direction)
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

func (m *Model) sizeLinkFlow() {
	m.linkFlow.width = m.width
	m.linkFlow.height = m.height
	if m.linkFlow.step == stepLinkParent {
		m.linkFlow.branchList.SetWidth(m.width)
		m.linkFlow.branchList.SetHeight(m.height - 6)
	}
}

func (m *Model) sizeUnlinkFlow() {
	m.unlinkFlow.width = m.width
	m.unlinkFlow.height = m.height
	if m.unlinkFlow.step == stepUnlinkPick {
		m.unlinkFlow.branchList.SetWidth(m.width)
		m.unlinkFlow.branchList.SetHeight(m.height - 6)
	}
	if m.unlinkFlow.step == stepUnlinkConfirm {
		m.unlinkFlow = m.unlinkFlow.withConfirm(m.unlinkFlow.target)
	}
}

func (m Model) finishEmbeddedLink() (tea.Model, tea.Cmd) {
	success := !m.linkFlow.cancelled && m.linkFlow.step == stepLinkSuccess
	m.linkFlow = LinkFlowModel{}
	if success {
		return m, m.selectProject(m.current)
	}
	m.screen = screenTree
	return m, nil
}

func (m Model) finishEmbeddedUnlink() (tea.Model, tea.Cmd) {
	success := !m.unlinkFlow.cancelled && m.unlinkFlow.step == stepUnlinkSuccess
	m.unlinkFlow = UnlinkFlowModel{}
	if success {
		return m, m.selectProject(m.current)
	}
	m.screen = screenTree
	return m, nil
}

func trimLast(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}
