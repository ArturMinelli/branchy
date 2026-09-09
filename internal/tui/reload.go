package tui

import "branchy/internal/tree"

type countSnapshot struct {
	inbound  map[string]fileChangeCount
	outbound map[string]fileChangeCount
}

func snapshotCounts(inbound, outbound map[string]fileChangeCount) *countSnapshot {
	return &countSnapshot{inbound: inbound, outbound: outbound}
}

func (s *countSnapshot) Restore(inbound, outbound *map[string]fileChangeCount) {
	if s == nil {
		return
	}
	*inbound = s.inbound
	*outbound = s.outbound
}

func reloadingBadges(doc *tree.Document) map[string]string {
	if doc == nil {
		return nil
	}
	badges := make(map[string]string)
	for _, name := range doc.Names() {
		parent, hasParent := doc.ParentOf(name)
		if !hasParent {
			continue
		}
		_ = parent
		badges[name] = "?"
	}
	return badges
}
