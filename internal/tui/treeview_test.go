package tui

import (
	"strings"
	"testing"

	"branchy/internal/sync"
	"branchy/internal/tree"
)

func TestFlattenTree(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":    {Children: []string{"release"}},
		"release": {Children: []string{"develop"}},
		"develop": {},
		"orphan":  {},
	}}

	v := newBranchTreeView(doc)
	if len(v.rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(v.rows))
	}
	if v.rows[0].name != "main" {
		t.Fatalf("expected main first, got %s", v.rows[0].name)
	}
	if v.rows[1].name != "release" {
		t.Fatalf("expected release second, got %s", v.rows[1].name)
	}
	if v.rows[1].connector != "└── " {
		t.Fatalf("expected └── connector, got %q", v.rows[1].connector)
	}
}

func TestTreeViewNavigation(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"a": {Children: []string{"b"}},
		"b": {},
	}}
	v := newBranchTreeView(doc)
	v.moveDown()
	if v.selectedName() != "b" {
		t.Fatalf("expected b, got %s", v.selectedName())
	}
	v.moveUp()
	if v.selectedName() != "a" {
		t.Fatalf("expected a, got %s", v.selectedName())
	}
}

func TestTreeViewInboundBadges(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":    {Children: []string{"release"}},
		"release": {Children: []string{"develop"}},
		"develop": {},
	}}
	v := newBranchTreeView(doc)
	v.setFileCounts(map[string]fileChangeCount{
		"release": {files: 12, ok: true},
		"develop": {files: 0, ok: true},
	}, nil)

	out := stripANSI(v.View())
	if !strings.Contains(out, "release  12") {
		t.Fatalf("expected release badge 12:\n%s", out)
	}
	if strings.Contains(out, "develop  0") {
		t.Fatalf("zero inbound must be hidden:\n%s", out)
	}
	if strings.Contains(out, "main  ") {
		t.Fatalf("root must not show inbound badge:\n%s", out)
	}
}

func TestTreeViewUnknownBadge(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":    {Children: []string{"ok", "missing"}},
		"ok":      {},
		"missing": {},
	}}
	v := newBranchTreeView(doc)
	v.setFileCounts(map[string]fileChangeCount{
		"ok":      {files: 4, ok: true},
		"missing": {ok: false},
	}, nil)

	out := stripANSI(v.View())
	if !strings.Contains(out, "ok  4") {
		t.Fatalf("expected sibling count:\n%s", out)
	}
	if !strings.Contains(out, "missing  ?") {
		t.Fatalf("expected unknown placeholder:\n%s", out)
	}
	if strings.Contains(out, "missing  0") {
		t.Fatalf("unknown must not look like zero:\n%s", out)
	}
}

func TestTreeViewDirectionToggleBadges(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":  {Children: []string{"child"}},
		"child": {},
	}}
	v := newBranchTreeView(doc)
	v.setFileCounts(
		map[string]fileChangeCount{"child": {files: 5, ok: true}},
		map[string]fileChangeCount{"child": {files: 2, ok: true}},
	)

	in := stripANSI(v.View())
	if !strings.Contains(in, "child  5") {
		t.Fatalf("default inbound badge:\n%s", in)
	}

	v.setDirection(sync.Upward)
	out := stripANSI(v.View())
	if !strings.Contains(out, "child  2") {
		t.Fatalf("outbound badge:\n%s", out)
	}
	if strings.Contains(out, "child  5") {
		t.Fatalf("inbound badge must not remain in outbound mode:\n%s", out)
	}

	v.setDirection(sync.Downward)
	back := stripANSI(v.View())
	if !strings.Contains(back, "child  5") {
		t.Fatalf("restored inbound badge:\n%s", back)
	}
}

func TestTreeViewZeroHidePerDirection(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":  {Children: []string{"child"}},
		"child": {},
	}}
	v := newBranchTreeView(doc)
	v.setFileCounts(
		map[string]fileChangeCount{"child": {files: 3, ok: true}},
		map[string]fileChangeCount{"child": {files: 0, ok: true}},
	)

	in := stripANSI(v.View())
	if !strings.Contains(in, "child  3") {
		t.Fatalf("inbound non-zero:\n%s", in)
	}

	v.setDirection(sync.Upward)
	out := stripANSI(v.View())
	if strings.Contains(out, "child  0") || strings.Contains(out, "child  3") {
		t.Fatalf("outbound zero must hide badge:\n%s", out)
	}
}

func TestTreeViewUnknownBothDirections(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":    {Children: []string{"missing"}},
		"missing": {},
	}}
	v := newBranchTreeView(doc)
	v.setFileCounts(
		map[string]fileChangeCount{"missing": {ok: false}},
		map[string]fileChangeCount{"missing": {ok: false}},
	)

	in := stripANSI(v.View())
	if !strings.Contains(in, "missing  ?") {
		t.Fatalf("inbound unknown:\n%s", in)
	}
	v.setDirection(sync.Upward)
	out := stripANSI(v.View())
	if !strings.Contains(out, "missing  ?") {
		t.Fatalf("outbound unknown:\n%s", out)
	}
}

func TestTreeViewDirectionArrows(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"main":    {Children: []string{"release"}},
		"release": {Children: []string{"develop"}},
		"develop": {},
		"orphan":  {},
	}}
	v := newBranchTreeView(doc)
	v.setFileCounts(map[string]fileChangeCount{
		"release": {files: 12, ok: true},
		"develop": {files: 0, ok: true},
	}, map[string]fileChangeCount{
		"release": {files: 0, ok: true},
		"develop": {ok: false},
	})

	in := stripANSI(v.View())
	if !strings.Contains(in, "▸") {
		t.Fatalf("selected marker missing:\n%s", in)
	}
	if !strings.Contains(in, "release  12 ↓") {
		t.Fatalf("inbound arrow after count:\n%s", in)
	}
	if strings.Contains(in, "↓ main") || strings.Contains(in, "↓ develop") || strings.Contains(in, "↓ orphan") {
		t.Fatalf("rows without a visible count must not show an arrow:\n%s", in)
	}
	if strings.Contains(in, "develop  0") {
		t.Fatalf("hidden zero must stay hidden:\n%s", in)
	}

	v.setDirection(sync.Upward)
	out := stripANSI(v.View())
	if strings.Contains(out, "12 ↓") || strings.Contains(out, "release  12") {
		t.Fatalf("outbound zero must hide count and arrow:\n%s", out)
	}
	if !strings.Contains(out, "develop  ? ↑") {
		t.Fatalf("unknown outbound must keep arrow after ?:\n%s", out)
	}
	if strings.Contains(out, "↑ main") || strings.Contains(out, "↑ orphan") {
		t.Fatalf("rows without a count must not show an arrow:\n%s", out)
	}
}
