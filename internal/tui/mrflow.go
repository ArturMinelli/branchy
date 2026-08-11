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
	stepMRResult
	stepMRBrowser
	stepMRError
	stepMRFewBranches
)

// MROptions configures the MR TUI flow.
type MROptions struct {
	PrefilledSource string
	Embedded        bool
}

type branchItem struct {
	name string
}

func (i branchItem) Title() string       { return i.name }
func (i branchItem) Description() string { return "" }
func (i branchItem) FilterValue() string { return i.name }

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
	quitting    bool
	cancelled   bool
	finished    bool
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

func newBranchList(names []string, title string) list.Model {
	items := make([]list.Item, len(names))
	for i, name := range names {
		items[i] = branchItem{name: name}
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

func (m MRFlowModel) Init() tea.Cmd { return nil }

func (m MRFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.branchList.SetWidth(msg.Width)
		m.branchList.SetHeight(msg.Height - 6)
		return m, nil

	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
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

	var b strings.Builder
	b.WriteString(titleStyle.Render("branchy mr"))
	b.WriteString("\n\n")

	switch m.step {
	case stepMRSource, stepMRTarget:
		if m.step == stepMRTarget && m.source != "" {
			b.WriteString(helpStyle.Render("Source: " + m.source))
			b.WriteString("\n\n")
		}
		b.WriteString(m.branchList.View())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓: navigate  enter: select  esc: cancel  q: quit"))
	case stepMRTitle:
		b.WriteString(fmt.Sprintf("Source: %s  →  Target: %s\n\n", m.source, m.target))
		b.WriteString("MR title:\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(m.title))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("type to edit  enter: continue  esc: cancel"))
	case stepMRConfirm:
		b.WriteString(fmt.Sprintf("Create MR %s → %s?\n\n", m.source, m.target))
		b.WriteString(fmt.Sprintf("Title: %s\n\n", m.title))
		b.WriteString(helpStyle.Render("y: create  n/esc: cancel"))
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
		b.WriteString(helpStyle.Render("enter: continue"))
	case stepMRBrowser:
		b.WriteString("Open in browser? [y/N]\n\n")
		if m.result != nil && m.result.URL != "" {
			b.WriteString("  " + m.result.URL)
			b.WriteString("\n\n")
		}
		if m.browserWarn != "" {
			b.WriteString(warnStyle.Render(m.browserWarn))
			b.WriteString("\n\n")
		}
		b.WriteString(helpStyle.Render("y: open  n/enter: done  q: quit"))
	case stepMRError:
		b.WriteString(errStyle.Render(m.errMsg))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter/esc: exit"))
	case stepMRFewBranches:
		b.WriteString(warnStyle.Render("Need at least 2 branches in tree to create an MR."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter/esc: exit"))
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
		return m, nil
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
	if key.Matches(msg, mrKeys.No) || key.Matches(msg, mrKeys.Back) {
		return m.cancelOrQuit()
	}
	if key.Matches(msg, mrKeys.Yes) || key.Matches(msg, mrKeys.Enter) {
		result, err := mr.Create(m.project, mr.CreateRequest{
			Source: m.source,
			Target: m.target,
			Title:  m.title,
		})
		if err != nil {
			m.errMsg = err.Error()
			m.step = stepMRError
			return m, nil
		}
		m.result = result
		if result.Action == mr.ActionFailed {
			m.errMsg = result.Message
			m.step = stepMRError
			return m, nil
		}
		m.step = stepMRResult
		return m, nil
	}
	return m, nil
}

func (m MRFlowModel) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, mrKeys.Enter) {
		if m.result != nil && m.result.URL != "" {
			m.step = stepMRBrowser
			return m, nil
		}
		return m.finish()
	}
	return m, nil
}

func (m MRFlowModel) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, mrKeys.Yes) {
		if m.result != nil && m.result.URL != "" {
			if err := browser.Open(m.result.URL); err != nil {
				m.browserWarn = err.Error()
			}
		}
		return m.finish()
	}
	if key.Matches(msg, mrKeys.No) || key.Matches(msg, mrKeys.Enter) {
		return m.finish()
	}
	if key.Matches(msg, mrKeys.Quit) {
		if m.opts.Embedded {
			return m.finish()
		}
		m.quitting = true
		return m, tea.Quit
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
