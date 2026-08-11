package tree

import (
	"os"
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

func TestNormalizeEnsuresChildNodes(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{
		"1.20": {Children: []string{"develop-1.20.7-gestao-estrategica"}},
	}}
	doc.Normalize()
	if _, ok := doc.Branches["develop-1.20.7-gestao-estrategica"]; !ok {
		t.Fatal("expected child branch to be created")
	}
	names := doc.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 branch names, got %v", names)
	}
}

func TestLoadNormalizesChildren(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/branch-tree.yaml"
	data := []byte(`branches:
  1.20:
    children:
      - develop-1.20.7-gestao-estrategica
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := doc.Branches["develop-1.20.7-gestao-estrategica"]; !ok {
		t.Fatal("expected child branch to be created on load")
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
