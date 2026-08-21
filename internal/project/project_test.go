package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"branchy/internal/config"
	"branchy/internal/tree"
)

func testProject(t *testing.T, doc *tree.Document) *Project {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := &Project{
		ID:   "testproj",
		Path: filepath.Join(home, "repo"),
		Tree: doc,
	}
	if err := p.SaveTree(); err != nil {
		t.Fatalf("SaveTree: %v", err)
	}
	return p
}

func reloadTree(t *testing.T, id string) *tree.Document {
	t.Helper()
	treePath, err := config.TreePath(id)
	if err != nil {
		t.Fatalf("TreePath: %v", err)
	}
	doc, err := tree.Load(treePath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return doc
}

func TestLinkPersistsEdge(t *testing.T) {
	p := testProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"main": {Children: []string{}},
	}})

	if err := p.Link("main", "feature"); err != nil {
		t.Fatalf("Link: %v", err)
	}

	doc := reloadTree(t, p.ID)
	if !doc.HasEdge("main", "feature") {
		t.Fatal("expected main → feature on disk")
	}
}

func TestLinkValidation(t *testing.T) {
	p := testProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"main": {Children: []string{"release"}},
		"release": {},
	}})

	tests := []struct {
		parent, child string
		want          string
	}{
		{"", "child", "parent and child are required"},
		{"main", "", "parent and child are required"},
		{"main", "main", "parent and child must differ"},
		{"main", "release", "edge already exists: main → release"},
	}

	for _, tc := range tests {
		err := p.Link(tc.parent, tc.child)
		if err == nil {
			t.Fatalf("Link(%q, %q): expected error", tc.parent, tc.child)
		}
		if err.Error() != tc.want {
			t.Fatalf("Link(%q, %q): got %q, want %q", tc.parent, tc.child, err.Error(), tc.want)
		}
	}
}

func TestUnlinkEdgeValidation(t *testing.T) {
	p := testProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"main":    {Children: []string{"feature-a"}},
		"feature-a": {Children: []string{"feature-b"}},
		"feature-b": {},
	}})

	_, err := p.Unlink("main", "missing")
	if err == nil || !strings.Contains(err.Error(), `branch "missing" not in tree`) {
		t.Fatalf("missing child: got %v", err)
	}

	_, err = p.Unlink("other", "feature-a")
	if err == nil || !strings.Contains(err.Error(), "edge not found: other → feature-a") {
		t.Fatalf("missing edge: got %v", err)
	}

	doc := reloadTree(t, p.ID)
	if !doc.HasEdge("main", "feature-a") {
		t.Fatal("tree should be unchanged on disk after validation failure")
	}
}

func TestUnlinkSuccess(t *testing.T) {
	p := testProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"main":      {Children: []string{"feature-a"}},
		"feature-a": {Children: []string{"feature-b"}},
		"feature-b": {},
	}})

	res, err := p.Unlink("main", "feature-a")
	if err != nil {
		t.Fatalf("Unlink: %v", err)
	}
	if res.Removed != 2 {
		t.Fatalf("Removed = %d, want 2", res.Removed)
	}
	if res.Parent != "main" || res.Child != "feature-a" {
		t.Fatalf("unexpected result: %+v", res)
	}

	doc := reloadTree(t, p.ID)
	if _, ok := doc.Branches["feature-a"]; ok {
		t.Fatal("feature-a should be gone from disk")
	}
	if _, ok := doc.Branches["feature-b"]; ok {
		t.Fatal("feature-b should be gone from disk")
	}
}

func TestUnlinkRootWithEmptyParent(t *testing.T) {
	p := testProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"main":  {Children: []string{}},
		"orphan": {Children: []string{}},
	}})

	res, err := p.Unlink("", "orphan")
	if err != nil {
		t.Fatalf("Unlink root: %v", err)
	}
	if res.Removed != 1 {
		t.Fatalf("Removed = %d, want 1", res.Removed)
	}

	doc := reloadTree(t, p.ID)
	if _, ok := doc.Branches["orphan"]; ok {
		t.Fatal("orphan should be gone from disk")
	}
}

func TestUnlinkEmptyParentNonRoot(t *testing.T) {
	p := testProject(t, &tree.Document{Branches: map[string]tree.BranchNode{
		"main":      {Children: []string{"feature-a"}},
		"feature-a": {},
	}})

	_, err := p.Unlink("", "feature-a")
	if err == nil || !strings.Contains(err.Error(), `branch "feature-a" is not a tree root`) {
		t.Fatalf("non-root with empty parent: got %v", err)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

func TestInitRegisterAndForce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := initGitRepo(t)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	defer t.Chdir(wd)

	p1, err := Init(InitOptions{Force: false})
	if err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if p1.ID == "" {
		t.Fatal("expected project id")
	}

	_, err = Init(InitOptions{Force: false})
	if err == nil {
		t.Fatal("expected already registered error")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("got %v", err)
	}

	p2, err := Init(InitOptions{Force: true})
	if err != nil {
		t.Fatalf("force Init: %v", err)
	}
	if p2.ID != p1.ID {
		t.Fatalf("force should reuse id: got %q want %q", p2.ID, p1.ID)
	}
}
