package tui

import (
	"strings"
	"testing"

	"branchy/internal/sync"
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
	in := treeHelpFooter(sync.Downward)
	if !strings.Contains(in, "ctrl+↑: outbound") || !strings.Contains(in, "ctrl+↓: inbound") {
		t.Fatalf("inbound chords:\n%s", in)
	}
	if !strings.Contains(in, "counts: inbound (parent→child)") {
		t.Fatalf("inbound mode cue:\n%s", in)
	}
	if strings.Contains(in, "d: show") {
		t.Fatalf("inbound footer must not list d:\n%s", in)
	}
	out := treeHelpFooter(sync.Upward)
	if !strings.Contains(out, "ctrl+↑: outbound") || !strings.Contains(out, "ctrl+↓: inbound") {
		t.Fatalf("outbound chords:\n%s", out)
	}
	if !strings.Contains(out, "counts: outbound (child→parent)") {
		t.Fatalf("outbound mode cue:\n%s", out)
	}
	if strings.Contains(out, "d: show") {
		t.Fatalf("outbound footer must not list d:\n%s", out)
	}
	if !strings.Contains(in, "r: reload") {
		t.Fatalf("inbound footer must list r: reload:\n%s", in)
	}
	if !strings.Contains(out, "r: reload") {
		t.Fatalf("outbound footer must list r: reload:\n%s", out)
	}
}
