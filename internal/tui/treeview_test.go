package tui

import (
	"strings"
	"testing"

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
	v.setInbound(map[string]inboundCount{
		"release": {files: 12, ok: true},
		"develop": {files: 0, ok: true},
	})

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
	v.setInbound(map[string]inboundCount{
		"ok":      {files: 4, ok: true},
		"missing": {ok: false},
	})

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
