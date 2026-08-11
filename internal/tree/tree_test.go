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

func sampleTree() *Document {
	return &Document{Branches: map[string]BranchNode{
		"develop":   {Children: []string{"feature-a"}},
		"feature-a": {Children: []string{"feature-b"}},
		"feature-b": {},
		"orphan":    {},
	}}
}

func TestSubtreeNames(t *testing.T) {
	doc := sampleTree()
	names := doc.SubtreeNames("feature-a")
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %v", names)
	}
	if names[0] != "feature-a" || names[1] != "feature-b" {
		t.Fatalf("unexpected names: %v", names)
	}
	if doc.SubtreeNames("missing") != nil {
		t.Fatal("expected nil for missing branch")
	}
}

func TestParentOf(t *testing.T) {
	doc := sampleTree()
	parent, ok := doc.ParentOf("feature-a")
	if !ok || parent != "develop" {
		t.Fatalf("expected parent develop, got %q ok=%v", parent, ok)
	}
	if _, ok := doc.ParentOf("orphan"); ok {
		t.Fatal("expected no parent for orphan root")
	}
}

func TestHasEdge(t *testing.T) {
	doc := sampleTree()
	if !doc.HasEdge("develop", "feature-a") {
		t.Fatal("expected develop → feature-a edge")
	}
	if doc.HasEdge("develop", "feature-b") {
		t.Fatal("expected no direct edge develop → feature-b")
	}
}

func TestUnlinkSubtreeNested(t *testing.T) {
	doc := sampleTree()
	if err := doc.UnlinkSubtree("feature-a"); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc.Branches["feature-a"]; ok {
		t.Fatal("feature-a should be removed")
	}
	if _, ok := doc.Branches["feature-b"]; ok {
		t.Fatal("feature-b should be removed")
	}
	for _, child := range doc.Branches["develop"].Children {
		if child == "feature-a" {
			t.Fatal("feature-a should be removed from develop children")
		}
	}
	if _, ok := doc.Branches["develop"]; !ok {
		t.Fatal("develop should remain")
	}
	if _, ok := doc.Branches["orphan"]; !ok {
		t.Fatal("orphan root should remain")
	}
}

func TestUnlinkSubtreeLeaf(t *testing.T) {
	doc := sampleTree()
	if err := doc.UnlinkSubtree("feature-b"); err != nil {
		t.Fatal(err)
	}
	if len(doc.Branches) != 3 {
		t.Fatalf("expected 3 branches left, got %d", len(doc.Branches))
	}
	if doc.HasEdge("feature-a", "feature-b") {
		t.Fatal("feature-b edge should be gone")
	}
}

func TestUnlinkSubtreeRoot(t *testing.T) {
	doc := sampleTree()
	if err := doc.UnlinkSubtree("orphan"); err != nil {
		t.Fatal(err)
	}
	if len(doc.Branches) != 3 {
		t.Fatalf("expected 3 branches left, got %d", len(doc.Branches))
	}
}

func TestUnlinkSubtreeNotInTree(t *testing.T) {
	doc := sampleTree()
	if err := doc.UnlinkSubtree("missing"); err == nil {
		t.Fatal("expected error for missing branch")
	}
}
