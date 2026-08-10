package tree

import (
	"testing"
)

func TestCollectEdges(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{
		"main":     {Children: []string{"release"}},
		"release":  {Children: []string{"develop"}},
		"develop":  {Children: []string{"feat-a", "feat-b"}},
		"feat-a":   {},
		"feat-b":   {},
	}}

	edges := doc.CollectEdges("main")
	if len(edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(edges))
	}
	if edges[0].Parent != "main" || edges[0].Child != "release" {
		t.Fatalf("unexpected first edge: %+v", edges[0])
	}
}

func TestLink(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{
		"main": {},
	}}
	if err := doc.Link("main", "release"); err != nil {
		t.Fatal(err)
	}
	if err := doc.Link("main", "release"); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestRoots(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{
		"main":    {Children: []string{"a"}},
		"orphan":  {},
		"a":       {},
	}}
	roots := doc.Roots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %v", roots)
	}
}
