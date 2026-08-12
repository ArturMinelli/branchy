package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"branchy/internal/tree"
)

var (
	branchNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedStyle   = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("212")).
			Bold(true)
	connectorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
)

// branchRow is one visible line in the flattened tree.
type branchRow struct {
	name      string
	prefix    string
	connector string
}

// BranchTreeView is an always-expanded interactive tree navigator.
type BranchTreeView struct {
	rows      []branchRow
	cursor    int
	width     int
	inbound   map[string]fileChangeCount
	outbound  map[string]fileChangeCount
	direction diffDirection
}

func newBranchTreeView(doc *tree.Document) BranchTreeView {
	if doc == nil {
		return BranchTreeView{}
	}
	return BranchTreeView{
		rows: flattenTree(doc),
	}
}

func flattenTree(doc *tree.Document) []branchRow {
	if doc == nil {
		return nil
	}
	roots := doc.Roots()
	rows := make([]branchRow, 0, len(doc.Branches))
	for i, root := range roots {
		collectRows(doc, root, "", i == len(roots)-1, true, &rows)
	}
	return rows
}

func collectRows(doc *tree.Document, name, prefix string, isLast, isRoot bool, rows *[]branchRow) {
	connector := ""
	if !isRoot {
		if isLast {
			connector = "└── "
		} else {
			connector = "├── "
		}
	}

	*rows = append(*rows, branchRow{
		name:      name,
		prefix:    prefix,
		connector: connector,
	})

	node, ok := doc.Branches[name]
	if !ok {
		return
	}

	children := append([]string(nil), node.Children...)
	sort.Strings(children)

	var childPrefix string
	switch {
	case isRoot:
		childPrefix = ""
	case isLast:
		childPrefix = prefix + "    "
	default:
		childPrefix = prefix + "│   "
	}

	for i, child := range children {
		collectRows(doc, child, childPrefix, i == len(children)-1, false, rows)
	}
}

func (v *BranchTreeView) setFileCounts(inbound, outbound map[string]fileChangeCount) {
	v.inbound = inbound
	v.outbound = outbound
}

func (v *BranchTreeView) setDirection(d diffDirection) {
	v.direction = d
}

func (v BranchTreeView) activeCounts() map[string]fileChangeCount {
	if v.direction == diffOutbound {
		return v.outbound
	}
	return v.inbound
}

func (v *BranchTreeView) rebuild(doc *tree.Document) {
	v.rows = flattenTree(doc)
	if v.cursor >= len(v.rows) {
		v.cursor = max(0, len(v.rows)-1)
	}
}

func (v BranchTreeView) selectedName() string {
	if v.cursor < 0 || v.cursor >= len(v.rows) {
		return ""
	}
	return v.rows[v.cursor].name
}

func (v *BranchTreeView) moveUp() {
	if v.cursor > 0 {
		v.cursor--
	}
}

func (v *BranchTreeView) moveDown() {
	if v.cursor < len(v.rows)-1 {
		v.cursor++
	}
}

func (v BranchTreeView) View() string {
	if len(v.rows) == 0 {
		return connectorStyle.Render("(empty tree)")
	}

	var b strings.Builder
	for i, row := range v.rows {
		line := v.renderRow(row, i == v.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (v BranchTreeView) renderRow(row branchRow, selected bool) string {
	label := row.name
	badge := ""
	if c, ok := v.activeCounts()[row.name]; ok {
		badge = formatFileChangeBadge(c)
	}
	if badge != "" {
		label = row.name + "  " + badge
	}

	var line string
	if selected {
		marker := cursorStyle.Render("▸ ")
		line = marker + selectedStyle.Render(row.prefix+row.connector+label)
	} else {
		prefix := connectorStyle.Render(row.prefix + row.connector)
		name := branchNameStyle.Render(row.name)
		line = prefix + name
		if badge != "" {
			styled := helpStyle.Render(badge)
			if badge == "?" {
				styled = warnStyle.Render(badge)
			}
			line += "  " + styled
		}
	}

	if v.width > 0 {
		line = lipgloss.NewStyle().Width(v.width).Render(line)
	}
	return line
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
