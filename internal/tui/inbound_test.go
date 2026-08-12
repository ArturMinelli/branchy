package tui

import (
	"strings"
	"testing"

	"branchy/internal/tree"
)

func TestFormatFileChangeBadge(t *testing.T) {
	if got := formatFileChangeBadge(fileChangeCount{files: 12, ok: true}); got != "12" {
		t.Fatalf("positive: got %q", got)
	}
	if got := formatFileChangeBadge(fileChangeCount{files: 0, ok: true}); got != "" {
		t.Fatalf("zero must be hidden, got %q", got)
	}
	if got := formatFileChangeBadge(fileChangeCount{ok: false}); got != "?" {
		t.Fatalf("unknown: got %q", got)
	}
}

func TestFormatInboundConfirm(t *testing.T) {
	if got := formatInboundConfirm("feature", fileChangeCount{files: 4, ok: true}); got != "4 files would change on feature" {
		t.Fatalf("positive: got %q", got)
	}
	if got := formatInboundConfirm("feature", fileChangeCount{files: 0, ok: true}); got != "0 files would change on feature" {
		t.Fatalf("known zero: got %q", got)
	}
	if got := formatInboundConfirm("feature", fileChangeCount{ok: false}); got != "File count unavailable" {
		t.Fatalf("unknown: got %q", got)
	}
}

func TestLoadOutboundCountsOmitsRoots(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":    {Children: []string{"release"}},
		"release": {},
	}}
	// Path is fake; all comparisons fail → unknown entries for non-roots only.
	out := loadOutboundCounts("/nonexistent-repo-for-counts", doc)
	if _, ok := out["main"]; ok {
		t.Fatal("root must be omitted from outbound map")
	}
	if c, ok := out["release"]; !ok || c.ok {
		t.Fatalf("child with missing repo must be unknown, got %#v ok=%v", c, ok)
	}
}

func TestTreeHelpFooter(t *testing.T) {
	in := treeHelpFooter(diffInbound)
	if !strings.Contains(in, "d: show outbound") || !strings.Contains(in, "counts: inbound (parent→child)") {
		t.Fatalf("inbound footer:\n%s", in)
	}
	out := treeHelpFooter(diffOutbound)
	if !strings.Contains(out, "d: show inbound") || !strings.Contains(out, "counts: outbound (child→parent)") {
		t.Fatalf("outbound footer:\n%s", out)
	}
}
