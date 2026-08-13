// File-change counts are TUI-only. Scripted CLI paths do not display them.
package tui

import (
	"fmt"
	"strconv"

	"branchy/internal/git"
	"branchy/internal/tree"
)

// diffDirection is the session-scoped main-tree count mode.
type diffDirection int

const (
	diffInbound diffDirection = iota
	diffOutbound
)

// fileChangeCount is the cached comparison of a child against its tree parent.
type fileChangeCount struct {
	files int
	ok    bool
}

func loadInboundCounts(dir string, doc *tree.Document) map[string]fileChangeCount {
	return loadFileChangeCounts(dir, doc, git.InboundFiles)
}

func loadOutboundCounts(dir string, doc *tree.Document) map[string]fileChangeCount {
	return loadFileChangeCounts(dir, doc, git.OutboundFiles)
}

func loadFileChangeCounts(dir string, doc *tree.Document, compare func(dir, parent, child string) (int, error)) map[string]fileChangeCount {
	out := make(map[string]fileChangeCount)
	if doc == nil {
		return out
	}
	for _, name := range doc.Names() {
		parent, hasParent := doc.ParentOf(name)
		if !hasParent {
			continue
		}
		n, err := compare(dir, parent, name)
		if err != nil {
			out[name] = fileChangeCount{ok: false}
			continue
		}
		out[name] = fileChangeCount{files: n, ok: true}
	}
	return out
}

func lookupInbound(counts map[string]fileChangeCount, name string) fileChangeCount {
	if counts == nil {
		return fileChangeCount{}
	}
	c, ok := counts[name]
	if !ok {
		return fileChangeCount{}
	}
	return c
}

func formatFileChangeBadge(c fileChangeCount) string {
	if !c.ok {
		return "?"
	}
	if c.files == 0 {
		return ""
	}
	return strconv.Itoa(c.files)
}

func formatFileChangeConfirm(branch string, c fileChangeCount) string {
	if !c.ok {
		return "File count unavailable"
	}
	return fmt.Sprintf("%d files would change on %s", c.files, branch)
}

func formatInboundConfirm(child string, c fileChangeCount) string {
	return formatFileChangeConfirm(child, c)
}

func treeHelpFooter(d diffDirection) string {
	switch d {
	case diffOutbound:
		return "↑/↓: navigate  d: show inbound  s: sync  m: mr  l: link  u: unlink  esc: projects  q: quit\ncounts: outbound (child→parent)"
	default:
		return "↑/↓: navigate  d: show outbound  s: sync  m: mr  l: link  u: unlink  esc: projects  q: quit\ncounts: inbound (parent→child)"
	}
}
