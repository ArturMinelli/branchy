package tui

import (
	"testing"

	"branchy/internal/tree"
)

func TestSnapshotCountsRoundTrip(t *testing.T) {
	in := map[string]fileChangeCount{"a": {files: 3, ok: true}}
	out := map[string]fileChangeCount{"b": {files: 1, ok: true}}
	snap := snapshotCounts(in, out)

	var restoredIn, restoredOut map[string]fileChangeCount
	snap.Restore(&restoredIn, &restoredOut)
	if restoredIn["a"].files != 3 || restoredOut["b"].files != 1 {
		t.Fatal("restore must return original maps")
	}
}

func TestReloadingBadgesMarksChildrenOnly(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":  {Children: []string{"child"}},
		"child": {},
	}}
	badges := reloadingBadges(doc)
	if badges["child"] != "?" {
		t.Fatalf("expected ? for child, got %q", badges["child"])
	}
	if badges["main"] != "" {
		t.Fatal("root must not get a reload badge")
	}
}
