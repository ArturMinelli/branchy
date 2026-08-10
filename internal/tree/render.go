package tree

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	branchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	edgeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// RenderASCII returns an indented tree string for display.
func (d *Document) RenderASCII() string {
	roots := d.Roots()
	if len(roots) == 0 {
		return edgeStyle.Render("(empty tree)")
	}
	var b strings.Builder
	for i, root := range roots {
		d.renderNode(&b, root, "", i == len(roots)-1)
	}
	return b.String()
}

func (d *Document) renderNode(b *strings.Builder, name, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		b.WriteString(branchStyle.Render(name))
	} else {
		b.WriteString(prefix + connector + branchStyle.Render(name))
	}
	b.WriteString("\n")

	node, ok := d.Branches[name]
	if !ok {
		return
	}
	children := append([]string(nil), node.Children...)
	sort.Strings(children)

	childPrefix := prefix
	if prefix == "" {
		childPrefix = ""
	} else if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + "│   "
	}

	for i, child := range children {
		d.renderNode(b, child, childPrefix, i == len(children)-1)
	}
}
