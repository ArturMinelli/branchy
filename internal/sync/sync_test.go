package sync

import (
	"testing"

	"branchy/internal/mr"
	"branchy/internal/project"
	"branchy/internal/tree"
)

func TestCreatedURLsFiltersAndPreservesOrder(t *testing.T) {
	summary := &Summary{
		Results: []Result{
			{Parent: "a", Child: "b", Action: "created", URL: "https://example.com/1"},
			{Parent: "b", Child: "c", Action: "skipped", URL: "https://example.com/2"},
			{Parent: "c", Child: "d", Action: "created", URL: "https://example.com/3"},
			{Parent: "d", Child: "e", Action: "failed", URL: ""},
		},
	}

	got := CreatedURLs(summary)
	want := []string{"https://example.com/1", "https://example.com/3"}
	if len(got) != len(want) {
		t.Fatalf("expected %d URLs, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestOpenableURLsIncludesExistingSkips(t *testing.T) {
	summary := &Summary{
		Results: []Result{
			{Parent: "a", Child: "b", Action: "created", URL: "https://example.com/1"},
			{Parent: "b", Child: "c", Action: "skipped", URL: "https://example.com/2", Message: "open MR already exists"},
			{Parent: "c", Child: "d", Action: "skipped", Message: "skipped by user"},
			{Parent: "d", Child: "e", Action: "failed", URL: ""},
			{Parent: "e", Child: "f", Action: "created", URL: "https://example.com/3"},
		},
	}

	got := OpenableURLs(summary)
	want := []string{"https://example.com/1", "https://example.com/2", "https://example.com/3"}
	if len(got) != len(want) {
		t.Fatalf("expected %d URLs, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestCreatedURLsNilSummary(t *testing.T) {
	if got := CreatedURLs(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := OpenableURLs(nil); got != nil {
		t.Fatalf("expected nil openable, got %v", got)
	}
}

func TestEnds(t *testing.T) {
	edge := tree.Edge{Parent: "develop", Child: "feat-a"}
	src, tgt := Ends(edge, Downward)
	if src != "develop" || tgt != "feat-a" {
		t.Fatalf("downward: got %s → %s", src, tgt)
	}
	src, tgt = Ends(edge, Upward)
	if src != "feat-a" || tgt != "develop" {
		t.Fatalf("upward: got %s → %s", src, tgt)
	}
}

func TestEdgesBelowMatchesTreeWalks(t *testing.T) {
	doc := &tree.Document{Branches: map[string]tree.BranchNode{
		"develop": {Children: []string{"feat-a", "feat-b"}},
		"feat-a":  {Children: []string{"leaf"}},
		"feat-b":  {},
		"leaf":    {},
	}}

	down := EdgesBelow(doc, "develop", Downward)
	wantDown := doc.CollectEdges("develop")
	if len(down) != len(wantDown) {
		t.Fatalf("downward count %d != %d", len(down), len(wantDown))
	}
	for i := range down {
		if down[i] != wantDown[i] {
			t.Fatalf("downward[%d]: got %+v want %+v", i, down[i], wantDown[i])
		}
	}

	up := EdgesBelow(doc, "develop", Upward)
	wantUp := doc.CollectEdgesUpward("develop")
	if len(up) != len(wantUp) {
		t.Fatalf("upward count %d != %d", len(up), len(wantUp))
	}
	for i := range up {
		if up[i] != wantUp[i] {
			t.Fatalf("upward[%d]: got %+v want %+v", i, up[i], wantUp[i])
		}
	}
}

func TestResultArrowPrefersSourceTarget(t *testing.T) {
	r := Result{Parent: "p", Child: "c", Source: "c", Target: "p"}
	src, tgt := r.Arrow()
	if src != "c" || tgt != "p" {
		t.Fatalf("got %s → %s", src, tgt)
	}
	legacy := Result{Parent: "p", Child: "c"}
	src, tgt = legacy.Arrow()
	if src != "p" || tgt != "c" {
		t.Fatalf("legacy fallback: got %s → %s", src, tgt)
	}
}

func testSyncProject() *project.Project {
	return &project.Project{
		ID:   "test",
		Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"main":    {Children: []string{"develop"}},
			"develop": {Children: []string{"feat"}},
			"feat":    {},
			"leaf":    {},
		}},
	}
}

func TestBeginUnknownFromBranch(t *testing.T) {
	p := testSyncProject()
	_, err := Begin(p, "missing", Downward)
	if err == nil || err.Error() != `branch "missing" not in tree` {
		t.Fatalf("got %v", err)
	}
}

func TestBeginEmptyEdgesNoAuth(t *testing.T) {
	p := testSyncProject()
	edges, err := Begin(p, "leaf", Downward)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("expected no edges, got %d", len(edges))
	}
}

func TestSkippedByUser(t *testing.T) {
	edge := tree.Edge{Parent: "main", Child: "develop"}

	down := SkippedByUser(edge, Downward)
	if down.Action != mr.ActionSkipped || down.Message != "skipped by user" {
		t.Fatalf("down: %+v", down)
	}
	if down.Source != "main" || down.Target != "develop" {
		t.Fatalf("down ends: %s → %s", down.Source, down.Target)
	}

	up := SkippedByUser(edge, Upward)
	if up.Source != "develop" || up.Target != "main" {
		t.Fatalf("up ends: %s → %s", up.Source, up.Target)
	}
}
