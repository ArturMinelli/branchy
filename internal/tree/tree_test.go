package tree

import (
	"os"
	"testing"
)

func TestCollectEdges(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{
		"main":    {Children: []string{"release"}},
		"release": {Children: []string{"develop"}},
		"develop": {Children: []string{"feat-a", "feat-b"}},
		"feat-a":  {},
		"feat-b":  {},
	}}

	edges := doc.CollectEdges("main")
	if len(edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(edges))
	}
	if edges[0].Parent != "main" || edges[0].Child != "release" {
		t.Fatalf("unexpected first edge: %+v", edges[0])
	}
}

func TestCollectEdgesUpward(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{
		"develop": {Children: []string{"feat-a", "feat-b"}},
		"feat-a":  {Children: []string{"leaf"}},
		"feat-b":  {},
		"leaf":    {},
	}}

	down := doc.CollectEdges("develop")
	up := doc.CollectEdgesUpward("develop")
	if len(up) != len(down) {
		t.Fatalf("membership size: down %d up %d", len(down), len(up))
	}

	type pair struct{ p, c string }
	set := map[pair]int{}
	for _, e := range down {
		set[pair{e.Parent, e.Child}]++
	}
	for _, e := range up {
		key := pair{e.Parent, e.Child}
		if set[key] == 0 {
			t.Fatalf("upward extra edge %s → %s", e.Parent, e.Child)
		}
		set[key]--
	}
	for k, n := range set {
		if n != 0 {
			t.Fatalf("missing or extra %s → %s (%d)", k.p, k.c, n)
		}
	}

	if len(up) != 3 {
		t.Fatalf("expected 3 edges below develop, got %d", len(up))
	}
	if up[0].Parent != "feat-a" || up[0].Child != "leaf" {
		t.Fatalf("first upward edge should be deepest, got %+v", up[0])
	}
	if up[1].Parent != "develop" || up[1].Child != "feat-a" {
		t.Fatalf("second upward edge should be develop→feat-a, got %+v", up[1])
	}
	if up[2].Parent != "develop" || up[2].Child != "feat-b" {
		t.Fatalf("sibling order should stay feat-a then feat-b, got %+v", up[2])
	}

	// Reversed pre-order would start at develop→feat-b, not feat-a→leaf.
	if down[len(down)-1].Child == "feat-b" && up[0].Child != "leaf" {
		t.Fatal("CollectEdgesUpward must not be a reversed CollectEdges slice")
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
		"main":   {Children: []string{"a"}},
		"orphan": {},
		"a":      {},
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

func TestWalkDisplayEmpty(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{}}
	if got := doc.WalkDisplay(); len(got) != 0 {
		t.Fatalf("expected empty walk, got %+v", got)
	}
}

func TestWalkDisplayMembershipOrderAndSpine(t *testing.T) {
	doc := &Document{Branches: map[string]BranchNode{
		"main":    {Children: []string{"release"}},
		"release": {Children: []string{"zeta", "alpha"}},
		"alpha":   {},
		"zeta":    {},
		"orphan":  {},
	}}

	got := doc.WalkDisplay()
	// Roots() sorts: main, orphan. Children sort: alpha before zeta.
	want := []struct {
		name   string
		depth  int
		isRoot bool
		isLast bool
	}{
		{"main", 0, true, false},
		{"release", 1, false, true},
		{"alpha", 2, false, false},
		{"zeta", 2, false, true},
		{"orphan", 0, true, true},
	}
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d (%+v)", len(got), len(want), got)
	}
	names := map[string]bool{}
	for i, n := range got {
		w := want[i]
		names[n.Name] = true
		if n.Name != w.name || n.Depth != w.depth || n.IsRoot != w.isRoot || n.IsLast != w.isLast {
			t.Fatalf("node %d: got %+v want %+v", i, n, w)
		}
		if len(n.LastAtDepth) != n.Depth {
			t.Fatalf("%s: LastAtDepth len %d want %d", n.Name, len(n.LastAtDepth), n.Depth)
		}
	}
	for name := range doc.Branches {
		if !names[name] {
			t.Fatalf("missing %s in walk", name)
		}
	}

	// release is the only child of main; main is not the last root.
	if !got[1].IsLast || got[1].LastAtDepth[0] {
		t.Fatalf("release spine: %+v", got[1])
	}
	// alpha's parent (release) is last among main's children; skip root flag at [0].
	if got[2].LastAtDepth[1] != true || got[2].IsLast {
		t.Fatalf("alpha spine: %+v", got[2])
	}
}
