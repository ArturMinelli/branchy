package tui

import (
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
