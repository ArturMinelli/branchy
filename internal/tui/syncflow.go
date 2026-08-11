package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/gitlab"
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
	stepSyncEmpty
	stepSyncError
)

// SyncFlowOptions configures the sync TUI flow.
type SyncFlowOptions struct {
	Embedded bool
}

// SyncFlowModel is a multi-step Bubble Tea model for interactive sync.
type SyncFlowModel struct {
	project    *project.Project
	opts       SyncFlowOptions
	fromBranch string
	edges      []tree.Edge
	edgeIndex  int
	results    []sync.Result
	step       syncStep
	errMsg     string
	browserWarn string
	branchList list.Model
	cancelled  bool
	finished   bool
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
	}

	if fromBranch == "" {
		names := p.Tree.Names()
		if len(names) == 0 {
			m.step = stepSyncError
			m.errMsg = "branch tree is empty"
			return m
		}
		m.step = stepSyncPickRoot
		m.branchList = newBranchList(names, "Select root branch to sync from")
		return m
	}

	return m.prepareSyncFrom(fromBranch)
}

func (m SyncFlowModel) prepareSyncFrom(fromBranch string) SyncFlowModel {
	m.fromBranch = fromBranch
	edges := m.project.Tree.CollectEdges(fromBranch)
	if len(edges) == 0 {
		m.step = stepSyncEmpty
		return m
	}

	client := &gitlab.Client{Dir: m.project.Path}
	if err := client.AuthOK(); err != nil {
		m.step = stepSyncError
		m.errMsg = fmt.Sprintf("glab auth: %v (run: glab auth login)", err)
		return m
	}

	m.edges = edges
	m.step = stepSyncEdgeConfirm
	return m
}

func (m SyncFlowModel) summary() *sync.Summary {
	return &sync.Summary{Results: m.results}
}

func (m SyncFlowModel) currentEdge() tree.Edge {
	return m.edges[m.edgeIndex]
}

func (m SyncFlowModel) Init() tea.Cmd { return nil }

func (m SyncFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)
		m.branchList.SetWidth(msg.Width)
		m.branchList.SetHeight(msg.Height - 6)
		return m, nil

	case edgeResultMsg:
		m.results = append(m.results, msg.result)
		m.edgeIndex++
		if m.edgeIndex < len(m.edges) {
			m.step = stepSyncEdgeConfirm
			return m, nil
		}
		m.step = stepSyncSummary
		return m, nil

	case tea.KeyMsg:
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
		b.WriteString(RenderHelp("↑/↓: navigate  enter: select  esc: cancel  q: quit"))
	case stepSyncEdgeConfirm:
		edge := m.currentEdge()
		b.WriteString(fmt.Sprintf("Sync from %s\n\n", m.fromBranch))
		b.WriteString(fmt.Sprintf("Create MR %s → %s? [y/N]\n", edge.Parent, edge.Child))
		b.WriteString(fmt.Sprintf("(edge %d of %d)\n\n", m.edgeIndex+1, len(m.edges)))
		b.WriteString(RenderHelp("y: create  n: skip  esc: cancel remaining  q: quit"))
	case stepSyncProcessing:
		edge := m.currentEdge()
		b.WriteString(fmt.Sprintf("Sync from %s\n\n", m.fromBranch))
		b.WriteString(fmt.Sprintf("Creating MR %s → %s...\n", edge.Parent, edge.Child))
	case stepSyncSummary:
		b.WriteString(fmt.Sprintf("Sync from %s\n\n", m.fromBranch))
		b.WriteString(m.renderResults())
		b.WriteString("\n")
		if len(sync.CreatedURLs(m.summary())) > 0 {
			b.WriteString(RenderHelp("enter: continue  q: quit"))
		} else if m.opts.Embedded {
			b.WriteString(RenderHelp("enter/q: back to tree"))
		} else {
			b.WriteString(RenderHelp("enter/q: done"))
		}
	case stepSyncBrowser:
		b.WriteString(fmt.Sprintf("Sync from %s\n\n", m.fromBranch))
		b.WriteString(m.renderResults())
		b.WriteString("\n")
		b.WriteString("Open created MRs in browser? [y/N]\n\n")
		for _, url := range sync.CreatedURLs(m.summary()) {
			b.WriteString("  " + url)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		if m.browserWarn != "" {
			b.WriteString(warnStyle.Render(m.browserWarn))
			b.WriteString("\n\n")
		}
		b.WriteString(RenderHelp("y: open  n/enter: done  q: quit"))
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
		line := fmt.Sprintf("%s → %s: %s", r.Parent, r.Child, r.Action)
		switch r.Action {
		case mr.ActionCreated:
			line = okStyle.Render(line)
		case mr.ActionFailed:
			line = errStyle.Render(line + " — " + r.Message)
		default:
			line = warnStyle.Render(line)
			if r.Message != "" {
				line = warnStyle.Render(fmt.Sprintf("%s → %s: %s (%s)", r.Parent, r.Child, r.Action, r.Message))
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
	Back  key.Binding
	Enter key.Binding
	Quit  key.Binding
	Yes   key.Binding
	No    key.Binding
}

var syncKeys = syncKeyMap{
	Back:  key.NewBinding(key.WithKeys("esc", "b")),
	Enter: key.NewBinding(key.WithKeys("enter")),
	Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Yes:   key.NewBinding(key.WithKeys("y")),
	No:    key.NewBinding(key.WithKeys("n")),
}

type edgeResultMsg struct {
	result sync.Result
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
	if key.Matches(msg, syncKeys.No) {
		m.results = append(m.results, sync.Result{
			Parent:  m.currentEdge().Parent,
			Child:   m.currentEdge().Child,
			Action:  mr.ActionSkipped,
			Message: "skipped by user",
		})
		m.edgeIndex++
		if m.edgeIndex < len(m.edges) {
			return m, nil
		}
		m.step = stepSyncSummary
		return m, nil
	}
	if key.Matches(msg, syncKeys.Yes) || key.Matches(msg, syncKeys.Enter) {
		m.step = stepSyncProcessing
		edge := m.currentEdge()
		return m, runEdgeCmd(m.project, edge)
	}
	return m, nil
}

func runEdgeCmd(p *project.Project, edge tree.Edge) tea.Cmd {
	return func() tea.Msg {
		return edgeResultMsg{result: sync.RunEdge(p, edge)}
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
		if len(sync.CreatedURLs(m.summary())) > 0 {
			m.step = stepSyncBrowser
			return m, nil
		}
		return m.finish()
	}
	return m, nil
}

func (m SyncFlowModel) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, syncKeys.Yes) {
		if warn := sync.OpenURLs(sync.CreatedURLs(m.summary())); warn != "" {
			m.browserWarn = warn
		}
		return m.finish()
	}
	if key.Matches(msg, syncKeys.No) || key.Matches(msg, syncKeys.Enter) || key.Matches(msg, syncKeys.Back) {
		return m.finish()
	}
	if key.Matches(msg, syncKeys.Quit) {
		if m.opts.Embedded {
			return m.finish()
		}
		m.finished = true
		return m, tea.Quit
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
