package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"branchy/internal/browser"
	"branchy/internal/mr"
	"branchy/internal/project"
)

type mrStep int

const (
	stepMRSource mrStep = iota
	stepMRTarget
	stepMRTitle
	stepMRConfirm
	stepMRLoading
	stepMRResult
	stepMRBrowser
	stepMRBrowserLoading
	stepMRError
	stepMRFewBranches
)

// MROptions configures the MR TUI flow.
type MROptions struct {
	PrefilledSource string
	Embedded        bool
}

type branchItem struct {
	name  string
	badge string
}

func (i branchItem) Title() string {
	if i.badge == "" {
		return i.name
	}
	return i.name + "  " + i.badge
}
func (i branchItem) Description() string { return "" }
func (i branchItem) FilterValue() string { return i.name }

func newBranchList(names []string, title string) list.Model {
	return newBranchListWithBadges(names, title, nil)
}

func newBranchListWithBadges(names []string, title string, badges map[string]string) list.Model {
	items := make([]list.Item, len(names))
	for i, name := range names {
		badge := ""
		if badges != nil {
			badge = badges[name]
		}
		items[i] = branchItem{name: name, badge: badge}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	return l
}

func targetBranchNames(all []string, source string) []string {
	out := make([]string, 0, len(all)-1)
	for _, name := range all {
		if name != source {
			out = append(out, name)
		}
	}
	return out
}

// MRFlowModel is a multi-step Bubble Tea model for manual MR creation.
type MRFlowModel struct {
	project     *project.Project
	opts        MROptions
	step        mrStep
	source      string
	target      string
	title       string
	result      *mr.CreateResult
	branchList  list.Model
	width       int
	height      int
	errMsg      string
	browserWarn string
	confirm     ConfirmModel
	loading     LoadingModel
	quitting    bool
	cancelled   bool
	finished    bool
	flowWindow
}

// Cancelled reports whether the user cancelled the flow.
func (m MRFlowModel) Cancelled() bool { return m.cancelled }

// Finished reports whether the flow completed (success or error).
func (m MRFlowModel) Finished() bool { return m.finished }

// RunMR starts a standalone MR TUI for the given project.
func RunMR(p *project.Project, opts MROptions) error {
	m := newMRFlowModel(p, opts)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	_, err := prog.Run()
	return err
}

func newMRFlowModel(p *project.Project, opts MROptions) MRFlowModel {
	names := p.Tree.Names()
	m := MRFlowModel{
		project: p,
		opts:    opts,
		source:  opts.PrefilledSource,
	}

	if len(names) < 2 {
		m.step = stepMRFewBranches
		return m
	}

	if opts.PrefilledSource != "" {
		if _, ok := p.Tree.Branches[opts.PrefilledSource]; !ok {
			m.step = stepMRError
			m.errMsg = fmt.Sprintf("branch %q not in tree", opts.PrefilledSource)
			return m
		}
		m.step = stepMRTarget
		m.branchList = newBranchList(targetBranchNames(names, opts.PrefilledSource), "Select target branch")
	} else {
		m.step = stepMRSource
		m.branchList = newBranchList(names, "Select source branch")
	}

	return m
}

func (m MRFlowModel) resetMRConfirm() MRFlowModel {
	m.confirm = NewConfirm(ConfirmOptions{
		Context:  []string{fmt.Sprintf("Title: %s", m.title)},
		Question: fmt.Sprintf("Create MR %s → %s?", m.source, m.target),
		Width:    m.width,
	})
	return m
}

func (m MRFlowModel) resetBrowserConfirm() MRFlowModel {
	m.confirm = NewConfirm(ConfirmOptions{
		Question: "Open in browser?",
		Width:    m.width,
	})
	return m
}

func (m MRFlowModel) Init() tea.Cmd { return nil }

func (m MRFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)
		m.branchList.SetWidth(msg.Width)
		m.branchList.SetHeight(msg.Height - 6)
		if m.step == stepMRConfirm {
			m = m.resetMRConfirm()
		} else if m.step == stepMRBrowser {
			m = m.resetBrowserConfirm()
		}
		return m, nil

	case mrCreateResultMsg:
		m.loading = m.loading.Clear()
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.step = stepMRError
			return m, nil
		}
		m.result = msg.result
		if msg.result.Action == mr.ActionFailed {
			m.errMsg = msg.result.Message
			m.step = stepMRError
			return m, nil
		}
		m.step = stepMRResult
		return m, nil

	case mrBrowserOpenResultMsg:
		m.loading = m.loading.Clear()
		if msg.warn != "" {
			m.browserWarn = msg.warn
		}
		return m.finish()

	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}
		if m.step == stepMRLoading || m.step == stepMRBrowserLoading {
			return m, nil
		}
		if m.tooSmall {
			if key.Matches(msg, mrKeys.Quit) || key.Matches(msg, mrKeys.Enter) {
				return m.cancelOrQuit()
			}
			return m, nil
		}

		switch m.step {
		case stepMRSource, stepMRTarget:
			return m.updateBranchPicker(msg)
		case stepMRTitle:
			return m.updateTitle(msg)
		case stepMRConfirm:
			return m.updateConfirm(msg)
		case stepMRResult:
			return m.updateResult(msg)
		case stepMRBrowser:
			return m.updateBrowser(msg)
		case stepMRError, stepMRFewBranches:
			if key.Matches(msg, mrKeys.Back) || key.Matches(msg, mrKeys.Enter) || key.Matches(msg, mrKeys.Quit) {
				return m.cancelOrQuit()
			}
		}
	}

	if m.step == stepMRLoading || m.step == stepMRBrowserLoading {
		var cmd tea.Cmd
		m.loading, cmd = m.loading.Update(msg)
		return m, cmd
	}

	if m.step == stepMRSource || m.step == stepMRTarget {
		var cmd tea.Cmd
		m.branchList, cmd = m.branchList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m MRFlowModel) View() string {
	if m.quitting {
		return ""
	}
	if m.tooSmall {
		return m.wrap("")
	}

	var b strings.Builder
	b.WriteString(RenderTitle("branchy mr"))
	b.WriteString("\n\n")

	switch m.step {
	case stepMRSource, stepMRTarget:
		if m.step == stepMRTarget && m.source != "" {
			b.WriteString(RenderHelp("Source: " + m.source))
			b.WriteString("\n\n")
		}
		b.WriteString(m.branchList.View())
		b.WriteString("\n")
		b.WriteString(RenderHelp("↑/↓: navigate  enter: select  esc: cancel  q: quit"))
	case stepMRTitle:
		b.WriteString(fmt.Sprintf("Source: %s  →  Target: %s\n\n", m.source, m.target))
		b.WriteString("MR title:\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(m.title))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("type to edit  enter: continue  esc: cancel"))
	case stepMRConfirm:
		b.WriteString(m.confirm.View())
	case stepMRLoading:
		b.WriteString(m.loading.View())
	case stepMRResult:
		if m.result != nil {
			line := fmt.Sprintf("%s → %s: %s", m.result.Source, m.result.Target, m.result.Action)
			switch m.result.Action {
			case mr.ActionCreated:
				line = okStyle.Render(line)
			case mr.ActionFailed:
				line = errStyle.Render(line + " — " + m.result.Message)
			default:
				line = warnStyle.Render(line)
			}
			b.WriteString(line)
			b.WriteString("\n")
			if m.result.URL != "" {
				b.WriteString("  " + m.result.URL)
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
		b.WriteString(RenderHelp("enter: continue"))
	case stepMRBrowser:
		if m.result != nil && m.result.URL != "" {
			b.WriteString("  " + m.result.URL)
			b.WriteString("\n\n")
		}
		if m.browserWarn != "" {
			b.WriteString(warnStyle.Render(m.browserWarn))
			b.WriteString("\n\n")
		}
		b.WriteString(m.confirm.View())
	case stepMRBrowserLoading:
		b.WriteString(m.loading.View())
	case stepMRError:
		b.WriteString(errStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter/esc: exit"))
	case stepMRFewBranches:
		b.WriteString(warnStyle.Render("Need at least 2 branches in tree to create an MR."))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("enter/esc: exit"))
	}

	return b.String()
}

type mrKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
	Yes   key.Binding
	No    key.Binding
}

var mrKeys = mrKeyMap{
	Up:    key.NewBinding(key.WithKeys("up", "k")),
	Down:  key.NewBinding(key.WithKeys("down", "j")),
	Enter: key.NewBinding(key.WithKeys("enter")),
	Back:  key.NewBinding(key.WithKeys("esc", "b")),
	Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Yes:   key.NewBinding(key.WithKeys("y")),
	No:    key.NewBinding(key.WithKeys("n")),
}

type mrCreateResultMsg struct {
	result *mr.CreateResult
	err    error
}

type mrBrowserOpenResultMsg struct {
	warn string
}

func runMRCreateCmd(p *project.Project, req mr.CreateRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := mr.Create(p, req)
		return mrCreateResultMsg{result: result, err: err}
	}
}

func runMRBrowserOpenCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var warn string
		if err := browser.Open(url); err != nil {
			warn = err.Error()
		}
		return mrBrowserOpenResultMsg{warn: warn}
	}
}

func (m MRFlowModel) updateBranchPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, mrKeys.Quit) {
		if m.opts.Embedded {
			return m.cancelOrQuit()
		}
		m.quitting = true
		return m, tea.Quit
	}
	if key.Matches(msg, mrKeys.Back) {
		return m.cancelOrQuit()
	}
	if key.Matches(msg, mrKeys.Enter) {
		item, ok := m.branchList.SelectedItem().(branchItem)
		if !ok {
			return m, nil
		}
		if m.step == stepMRSource {
			m.source = item.name
			m.step = stepMRTarget
			m.branchList = newBranchList(targetBranchNames(m.project.Tree.Names(), m.source), "Select target branch")
			m.branchList.SetWidth(m.width)
			m.branchList.SetHeight(m.height - 6)
		} else {
			m.target = item.name
			m.title = mr.DefaultTitle(m.source, m.target)
			m.step = stepMRTitle
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.branchList, cmd = m.branchList.Update(msg)
	return m, cmd
}

func (m MRFlowModel) updateTitle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, mrKeys.Back) {
		return m.cancelOrQuit()
	}
	if key.Matches(msg, mrKeys.Quit) {
		if m.opts.Embedded {
			return m.cancelOrQuit()
		}
		m.quitting = true
		return m, tea.Quit
	}
	if key.Matches(msg, mrKeys.Enter) {
		if strings.TrimSpace(m.title) == "" {
			m.title = mr.DefaultTitle(m.source, m.target)
		}
		m.step = stepMRConfirm
		return m.resetMRConfirm(), nil
	}

	ch := msg.String()
	if ch == "backspace" {
		m.title = trimLast(m.title)
		return m, nil
	}
	if len(ch) != 1 {
		return m, nil
	}
	if ch < " " || ch > "~" {
		return m, nil
	}
	m.title += ch
	return m, nil
}

func (m MRFlowModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var choice ConfirmChoice
	m.confirm, choice = m.confirm.Update(msg)
	if choice == ConfirmNo {
		return m.cancelOrQuit()
	}
	if choice == ConfirmYes {
		m.step = stepMRLoading
		m.loading = NewLoading(LoadingOptions{
			Message: LoadingMessage("Creating MR", fmt.Sprintf("%s → %s", m.source, m.target)),
		})
		req := mr.CreateRequest{Source: m.source, Target: m.target, Title: m.title}
		return m, tea.Batch(m.loading.Init(), runMRCreateCmd(m.project, req))
	}
	return m, nil
}

func (m MRFlowModel) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, mrKeys.Enter) {
		if m.result != nil && m.result.URL != "" {
			m.step = stepMRBrowser
			return m.resetBrowserConfirm(), nil
		}
		return m.finish()
	}
	return m, nil
}

func (m MRFlowModel) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, mrKeys.Quit) {
		if m.opts.Embedded {
			return m.finish()
		}
		m.quitting = true
		return m, tea.Quit
	}

	var choice ConfirmChoice
	m.confirm, choice = m.confirm.Update(msg)
	if choice == ConfirmNo {
		return m.finish()
	}
	if choice == ConfirmYes && m.result != nil && m.result.URL != "" {
		m.step = stepMRBrowserLoading
		m.loading = NewLoading(LoadingOptions{Message: "Opening in browser…"})
		return m, tea.Batch(m.loading.Init(), runMRBrowserOpenCmd(m.result.URL))
	}
	return m, nil
}

func (m MRFlowModel) cancelOrQuit() (tea.Model, tea.Cmd) {
	m.cancelled = true
	m.finished = true
	if m.opts.Embedded {
		return m, nil
	}
	m.quitting = true
	return m, tea.Quit
}

func (m MRFlowModel) finish() (tea.Model, tea.Cmd) {
	m.finished = true
	if m.opts.Embedded {
		return m, nil
	}
	m.quitting = true
	return m, tea.Quit
}
