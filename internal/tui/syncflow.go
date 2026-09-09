package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/browser"
	"branchy/internal/mr"
	"branchy/internal/project"
	"branchy/internal/sync"
	"branchy/internal/tree"
)

type syncStep int

const (
	stepSyncPickRoot syncStep = iota
	stepSyncEdgeConfirm
	stepSyncProcessing
	stepSyncSummary
	stepSyncBrowser
	stepSyncBrowserLoading
	stepSyncEmpty
	stepSyncError
)

// SyncFlowOptions configures the sync TUI flow.
type SyncFlowOptions struct {
	Embedded  bool
	Direction sync.Direction
}

// SyncFlowModel is a multi-step Bubble Tea model for interactive sync.
type SyncFlowModel struct {
	project        *project.Project
	opts           SyncFlowOptions
	fromBranch     string
	edges          []tree.Edge
	edgeIndex      int
	results        []sync.Result
	step           syncStep
	errMsg         string
	browserWarn    string
	branchList     list.Model
	confirm        ConfirmModel
	loading        LoadingModel
	direction      sync.Direction
	inbound        map[string]fileChangeCount
	outbound       map[string]fileChangeCount
	reloadSnapshot *countSnapshot
	cancelled      bool
	finished       bool
	flowWindow
}

// Cancelled reports whether the user cancelled the flow.
func (m SyncFlowModel) Cancelled() bool { return m.cancelled }

// Finished reports whether the flow completed.
func (m SyncFlowModel) Finished() bool { return m.finished }

// RunSync starts a standalone sync TUI for the given project.
func RunSync(p *project.Project, opts SyncFlowOptions) error {
	opts.Embedded = false
	m := newSyncFlowModel(p, "", opts)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	_, err := prog.Run()
	return err
}

func newSyncFlowModel(p *project.Project, fromBranch string, opts SyncFlowOptions) SyncFlowModel {
	m := SyncFlowModel{
		project:    p,
		opts:       opts,
		fromBranch: fromBranch,
		direction:  opts.Direction,
	}
	if p != nil {
		m.inbound = loadInboundCounts(p.Path, p.Tree)
		m.outbound = loadOutboundCounts(p.Path, p.Tree)
	}

	if fromBranch == "" {
		names := p.Tree.Names()
		if len(names) == 0 {
			m.step = stepSyncError
			m.errMsg = "branch tree is empty"
			return m
		}
		m.step = stepSyncPickRoot
		m.branchList = newBranchListWithBadges(names, "Select root branch to sync from", fileChangeBadges(m.activeCounts()))
		return m
	}

	return m.prepareSyncFrom(fromBranch)
}

func (m SyncFlowModel) applyInbound(counts map[string]fileChangeCount) SyncFlowModel {
	return m.applyFileCounts(counts, m.outbound)
}

func (m SyncFlowModel) applyFileCounts(in, out map[string]fileChangeCount) SyncFlowModel {
	m.inbound = in
	m.outbound = out
	return m.refreshCountSurfaces()
}

func (m SyncFlowModel) beginReload() SyncFlowModel {
	if m.project == nil || m.reloadSnapshot != nil {
		return m
	}
	m.reloadSnapshot = snapshotCounts(m.inbound, m.outbound)
	m.inbound = nil
	m.outbound = nil
	return m.refreshCountSurfaces()
}

func (m SyncFlowModel) restoreReloadSnapshot() SyncFlowModel {
	if m.reloadSnapshot == nil {
		return m
	}
	m.reloadSnapshot.Restore(&m.inbound, &m.outbound)
	m.reloadSnapshot = nil
	return m.refreshCountSurfaces()
}

func (m SyncFlowModel) clearReloadSnapshot() SyncFlowModel {
	m.reloadSnapshot = nil
	return m
}

func (m SyncFlowModel) activeCounts() map[string]fileChangeCount {
	if m.direction == sync.Upward {
		return m.outbound
	}
	return m.inbound
}

func (m SyncFlowModel) refreshCountSurfaces() SyncFlowModel {
	if m.project != nil && m.step == stepSyncPickRoot {
		m = m.refreshPicker()
	}
	if m.step == stepSyncEdgeConfirm && len(m.edges) > 0 {
		m = m.resetEdgeConfirm()
	}
	return m
}

func (m SyncFlowModel) refreshPicker() SyncFlowModel {
	idx := m.branchList.Index()
	badges := fileChangeBadges(m.activeCounts())
	if m.reloadSnapshot != nil && m.activeCounts() == nil {
		badges = reloadingBadges(m.project.Tree)
	}
	m.branchList = newBranchListWithBadges(m.project.Tree.Names(), "Select root branch to sync from", badges)
	if m.width > 0 {
		m.branchList.SetWidth(m.width)
		m.branchList.SetHeight(m.height - 6)
	}
	if idx >= 0 {
		m.branchList.Select(idx)
	}
	return m
}

func fileChangeBadges(counts map[string]fileChangeCount) map[string]string {
	if counts == nil {
		return nil
	}
	badges := make(map[string]string, len(counts))
	for name, c := range counts {
		if b := formatFileChangeBadge(c); b != "" {
			badges[name] = b
		}
	}
	return badges
}

func inboundBadges(counts map[string]fileChangeCount) map[string]string {
	return fileChangeBadges(counts)
}

func (m SyncFlowModel) prepareSyncFrom(fromBranch string) SyncFlowModel {
	m.fromBranch = fromBranch
	edges, err := sync.Begin(m.project, fromBranch, m.direction)
	if err != nil {
		m.step = stepSyncError
		m.errMsg = err.Error()
		return m
	}
	if len(edges) == 0 {
		m.step = stepSyncEmpty
		return m
	}

	m.edges = edges
	m.step = stepSyncEdgeConfirm
	return m.resetEdgeConfirm()
}

func (m SyncFlowModel) resetEdgeConfirm() SyncFlowModel {
	edge := m.currentEdge()
	source, target := sync.Ends(edge, m.direction)
	count := lookupInbound(m.activeCounts(), edge.Child)
	m.confirm = NewConfirm(ConfirmOptions{
		Context: []string{
			m.syncContextLine(),
			formatFileChangeConfirm(target, count),
		},
		Question: fmt.Sprintf("Create MR %s → %s?", source, target),
		Progress: fmt.Sprintf("Edge %d of %d", m.edgeIndex+1, len(m.edges)),
		Width:    m.width,
	})
	return m
}

func (m SyncFlowModel) syncContextLine() string {
	if m.direction == sync.Upward {
		return fmt.Sprintf("Sync up to %s", m.fromBranch)
	}
	return fmt.Sprintf("Sync down from %s", m.fromBranch)
}

func syncPickerHelp(dir sync.Direction) string {
	mode := "sync: downward (parent→child)"
	if dir == sync.Upward {
		mode = "sync: upward (child→parent)"
	}
	return "↑/↓: navigate  enter: select  r: reload  " + directionChordHints() + "  esc: cancel  q: quit\n" + mode
}

func (m SyncFlowModel) resetBrowserConfirm() SyncFlowModel {
	m.confirm = NewConfirm(ConfirmOptions{
		Question: "Open MRs in browser?",
		Width:    m.width,
	})
	return m
}

func (m SyncFlowModel) summary() *sync.Summary {
	return &sync.Summary{Results: m.results}
}

func (m SyncFlowModel) currentEdge() tree.Edge {
	return m.edges[m.edgeIndex]
}

func (m SyncFlowModel) Init() tea.Cmd {
	if m.project != nil {
		return remoteUpdateCmd(m.project)
	}
	return nil
}

func (m SyncFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case remoteUpdateMsg:
		if m.project == nil || msg.projectID != m.project.ID {
			return m, nil
		}
		if msg.err != nil {
			if m.reloadSnapshot != nil {
				return m.restoreReloadSnapshot(), nil
			}
			return m, nil
		}
		m = m.clearReloadSnapshot()
		return m.applyFileCounts(
			loadInboundCounts(m.project.Path, m.project.Tree),
			loadOutboundCounts(m.project.Path, m.project.Tree),
		), nil

	case tea.WindowSizeMsg:
		m.onResize(msg)
		m.branchList.SetWidth(msg.Width)
		m.branchList.SetHeight(msg.Height - 6)
		if m.step == stepSyncEdgeConfirm {
			m = m.resetEdgeConfirm()
		} else if m.step == stepSyncBrowser {
			m = m.resetBrowserConfirm()
		}
		return m, nil

	case edgeResultMsg:
		m.loading = m.loading.Clear()
		m.results = append(m.results, msg.result)
		m.edgeIndex++
		if m.edgeIndex < len(m.edges) {
			m.step = stepSyncEdgeConfirm
			return m.resetEdgeConfirm(), nil
		}
		m.step = stepSyncSummary
		return m, nil

	case browserOpenResultMsg:
		m.loading = m.loading.Clear()
		if msg.warn != "" {
			m.browserWarn = msg.warn
			m.step = stepSyncBrowser
			return m.resetBrowserConfirm(), nil
		}
		return m.finish()

	case tea.KeyMsg:
		if m.step == stepSyncProcessing || m.step == stepSyncBrowserLoading {
			return m, nil
		}

		if m.tooSmall {
			if key.Matches(msg, flowKeys.Quit) || key.Matches(msg, flowKeys.Enter) {
				m.finished = true
				if !m.opts.Embedded {
					return m, tea.Quit
				}
			}
			return m, nil
		}

		switch m.step {
		case stepSyncPickRoot:
			return m.updatePickRoot(msg)
		case stepSyncEdgeConfirm:
			return m.updateEdgeConfirm(msg)
		case stepSyncSummary:
			return m.updateSummary(msg)
		case stepSyncBrowser:
			return m.updateBrowser(msg)
		case stepSyncEmpty, stepSyncError:
			if key.Matches(msg, syncKeys.Back) || key.Matches(msg, syncKeys.Enter) || key.Matches(msg, syncKeys.Quit) {
				return m.finish()
			}
		}
	}

	if m.step == stepSyncProcessing || m.step == stepSyncBrowserLoading {
		var cmd tea.Cmd
		m.loading, cmd = m.loading.Update(msg)
		return m, cmd
	}

	if m.step == stepSyncPickRoot {
		var cmd tea.Cmd
		m.branchList, cmd = m.branchList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m SyncFlowModel) updatePickRoot(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, flowKeys.Quit) {
		m.finished = true
		if m.opts.Embedded {
			m.cancelled = true
			return m, nil
		}
		return m, tea.Quit
	}
	if key.Matches(msg, flowKeys.Back) {
		m.cancelled = true
		m.finished = true
		if m.opts.Embedded {
			return m, nil
		}
		return m, tea.Quit
	}
	if key.Matches(msg, syncKeys.DirectionUp) {
		return m.setPickerDirection(sync.Upward), nil
	}
	if key.Matches(msg, syncKeys.DirectionDown) {
		return m.setPickerDirection(sync.Downward), nil
	}
	if key.Matches(msg, syncKeys.Reload) {
		if m.project != nil {
			m = m.beginReload()
			return m, remoteUpdateCmd(m.project)
		}
		return m, nil
	}
	if key.Matches(msg, flowKeys.Enter) {
		item, ok := m.branchList.SelectedItem().(branchItem)
		if !ok {
			return m, nil
		}
		updated := m.prepareSyncFrom(item.name)
		updated.opts = m.opts
		updated.flowWindow = m.flowWindow
		updated.branchList = m.branchList
		return updated, nil
	}
	var cmd tea.Cmd
	m.branchList, cmd = m.branchList.Update(msg)
	return m, cmd
}

func (m SyncFlowModel) View() string {
	if m.tooSmall {
		return m.wrap("")
	}

	var b strings.Builder
	b.WriteString(RenderTitle("branchy sync"))
	b.WriteString("\n\n")

	switch m.step {
	case stepSyncPickRoot:
		b.WriteString(m.branchList.View())
		b.WriteString("\n")
		b.WriteString(RenderHelp(syncPickerHelp(m.direction)))
	case stepSyncEdgeConfirm:
		b.WriteString(m.confirm.View())
		b.WriteString("\n")
		b.WriteString(RenderHelp("esc: cancel remaining  q: quit"))
	case stepSyncProcessing:
		b.WriteString(m.syncContextLine())
		b.WriteString("\n\n")
		b.WriteString(m.loading.View())
	case stepSyncSummary:
		b.WriteString(m.syncContextLine())
		b.WriteString("\n\n")
		b.WriteString(m.renderResults())
		b.WriteString("\n")
		if len(sync.OpenableURLs(m.summary())) > 0 {
			b.WriteString(RenderHelp("enter: continue  q: quit"))
		} else if m.opts.Embedded {
			b.WriteString(RenderHelp("enter/q: back to tree"))
		} else {
			b.WriteString(RenderHelp("enter/q: done"))
		}
	case stepSyncBrowser:
		b.WriteString(m.syncContextLine())
		b.WriteString("\n\n")
		b.WriteString(m.renderResults())
		b.WriteString("\n")
		for _, url := range sync.OpenableURLs(m.summary()) {
			b.WriteString("  " + url)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		if m.browserWarn != "" {
			b.WriteString(warnStyle.Render(m.browserWarn))
			b.WriteString("\n\n")
		}
		b.WriteString(m.confirm.View())
	case stepSyncBrowserLoading:
		b.WriteString(m.syncContextLine())
		b.WriteString("\n\n")
		b.WriteString(m.loading.View())
	case stepSyncEmpty:
		b.WriteString(warnStyle.Render(fmt.Sprintf("No child branches below %q.", m.fromBranch)))
		b.WriteString("\n\n")
		if m.opts.Embedded {
			b.WriteString(RenderHelp("enter/esc: back to tree"))
		} else {
			b.WriteString(RenderHelp("enter/esc: done"))
		}
	case stepSyncError:
		b.WriteString(errStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		if m.opts.Embedded {
			b.WriteString(RenderHelp("enter/esc: back to tree"))
		} else {
			b.WriteString(RenderHelp("enter/esc: done"))
		}
	}

	return b.String()
}

func (m SyncFlowModel) renderResults() string {
	var b strings.Builder
	for _, r := range m.results {
		source, target := r.Arrow()
		line := fmt.Sprintf("%s → %s: %s", source, target, r.Action)
		switch r.Action {
		case mr.ActionCreated:
			line = okStyle.Render(line)
		case mr.ActionFailed:
			line = errStyle.Render(line + " — " + r.Message)
		default:
			line = warnStyle.Render(line)
			if r.Message != "" {
				line = warnStyle.Render(fmt.Sprintf("%s → %s: %s (%s)", source, target, r.Action, r.Message))
			}
		}
		b.WriteString(line)
		b.WriteString("\n")
		if r.URL != "" {
			b.WriteString("  " + r.URL)
			b.WriteString("\n")
		}
	}
	return b.String()
}

type syncKeyMap struct {
	Back          key.Binding
	Enter         key.Binding
	Quit          key.Binding
	Yes           key.Binding
	No            key.Binding
	DirectionUp   key.Binding
	DirectionDown key.Binding
	Reload        key.Binding
}

var syncKeys = syncKeyMap{
	Back:          key.NewBinding(key.WithKeys("esc", "b")),
	Enter:         key.NewBinding(key.WithKeys("enter")),
	Quit:          key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Yes:           key.NewBinding(key.WithKeys("y")),
	No:            key.NewBinding(key.WithKeys("n")),
	DirectionUp:   key.NewBinding(key.WithKeys("ctrl+up")),
	DirectionDown: key.NewBinding(key.WithKeys("ctrl+down")),
	Reload:        key.NewBinding(key.WithKeys("r")),
}

func (m SyncFlowModel) setPickerDirection(dir sync.Direction) SyncFlowModel {
	if m.direction == dir {
		return m
	}
	m.direction = dir
	return m.refreshPicker()
}

type edgeResultMsg struct {
	result sync.Result
}

type browserOpenResultMsg struct {
	warn string
}

func (m SyncFlowModel) updateEdgeConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, syncKeys.Quit) {
		if m.opts.Embedded {
			return m.cancelRemaining()
		}
		m.finished = true
		return m, tea.Quit
	}
	if key.Matches(msg, syncKeys.Back) {
		return m.cancelRemaining()
	}
	if key.Matches(msg, syncKeys.Reload) {
		if m.project != nil {
			m = m.beginReload()
			return m, remoteUpdateCmd(m.project)
		}
		return m, nil
	}

	var choice ConfirmChoice
	m.confirm, choice = m.confirm.Update(msg)
	switch choice {
	case ConfirmYes:
		m.step = stepSyncProcessing
		edge := m.currentEdge()
		source, target := sync.Ends(edge, m.direction)
		m.loading = NewLoading(LoadingOptions{
			Message:  LoadingMessage("Creating MR", fmt.Sprintf("%s → %s", source, target)),
			Progress: fmt.Sprintf("Edge %d of %d", m.edgeIndex+1, len(m.edges)),
		})
		return m, tea.Batch(m.loading.Init(), runEdgeCmd(m.project, edge, m.direction))
	case ConfirmNo:
		m.results = append(m.results, sync.SkippedByUser(m.currentEdge(), m.direction))
		m.edgeIndex++
		if m.edgeIndex < len(m.edges) {
			m.step = stepSyncEdgeConfirm
			return m.resetEdgeConfirm(), nil
		}
		m.step = stepSyncSummary
		return m, nil
	}
	return m, nil
}

func runEdgeCmd(p *project.Project, edge tree.Edge, dir sync.Direction) tea.Cmd {
	return func() tea.Msg {
		return edgeResultMsg{result: sync.RunEdge(p, edge, dir)}
	}
}

func runBrowserOpenCmd(urls []string) tea.Cmd {
	return func() tea.Msg {
		return browserOpenResultMsg{warn: browser.OpenURLs(urls)}
	}
}

func (m SyncFlowModel) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, syncKeys.Quit) {
		if m.opts.Embedded {
			return m.finish()
		}
		m.finished = true
		return m, tea.Quit
	}
	if key.Matches(msg, syncKeys.Enter) || key.Matches(msg, syncKeys.Back) {
		if len(sync.OpenableURLs(m.summary())) > 0 {
			m.step = stepSyncBrowser
			return m.resetBrowserConfirm(), nil
		}
		return m.finish()
	}
	return m, nil
}

func (m SyncFlowModel) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, syncKeys.Quit) {
		if m.opts.Embedded {
			return m.finish()
		}
		m.finished = true
		return m, tea.Quit
	}

	var choice ConfirmChoice
	m.confirm, choice = m.confirm.Update(msg)
	switch choice {
	case ConfirmYes:
		m.browserWarn = ""
		m.step = stepSyncBrowserLoading
		m.loading = NewLoading(LoadingOptions{Message: "Opening MRs in browser…"})
		return m, tea.Batch(m.loading.Init(), runBrowserOpenCmd(sync.OpenableURLs(m.summary())))
	case ConfirmNo:
		return m.finish()
	}
	return m, nil
}

func (m SyncFlowModel) cancelRemaining() (tea.Model, tea.Cmd) {
	m.cancelled = true
	if len(m.results) > 0 {
		m.step = stepSyncSummary
		return m, nil
	}
	return m.finish()
}

func (m SyncFlowModel) finish() (tea.Model, tea.Cmd) {
	m.finished = true
	if m.opts.Embedded {
		return m, nil
	}
	return m, tea.Quit
}
