// Inbound file-change counts are TUI-only. Scripted CLI paths do not display them.
package tui

import (
	"fmt"
	"strconv"

	"branchy/internal/git"
	"branchy/internal/tree"
)

// inboundCount is the cached comparison of a child against its tree parent.
type inboundCount struct {
	files int
	ok    bool
}

func loadInboundCounts(dir string, doc *tree.Document) map[string]inboundCount {
	out := make(map[string]inboundCount)
	if doc == nil {
		return out
	}
	for _, name := range doc.Names() {
		parent, hasParent := doc.ParentOf(name)
		if !hasParent {
			continue
		}
		n, err := git.InboundFiles(dir, parent, name)
		if err != nil {
			out[name] = inboundCount{ok: false}
			continue
		}
		out[name] = inboundCount{files: n, ok: true}
	}
	return out
}

func lookupInbound(counts map[string]inboundCount, name string) inboundCount {
	if counts == nil {
		return inboundCount{}
	}
	c, ok := counts[name]
	if !ok {
		return inboundCount{}
	}
	return c
}

func formatInboundBadge(c inboundCount) string {
	if !c.ok {
		return "?"
	}
	if c.files == 0 {
		return ""
	}
	return strconv.Itoa(c.files)
}

func formatInboundConfirm(child string, c inboundCount) string {
	if !c.ok {
		return "File count unavailable"
	}
	return fmt.Sprintf("%d files would change on %s", c.files, child)
}
