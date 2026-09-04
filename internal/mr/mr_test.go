package mr

import (
	"strings"
	"testing"

	"branchy/internal/project"
	"branchy/internal/tree"
)

func testTree() *tree.Document {
	return &tree.Document{Branches: map[string]tree.BranchNode{
		"develop":   {},
		"feature-x": {},
	}}
}

func TestValidateBranches(t *testing.T) {
	doc := testTree()

	if err := ValidateBranches(doc, "feature-x", "develop"); err != nil {
		t.Fatalf("expected valid branches, got %v", err)
	}
	if err := ValidateBranches(doc, "develop", "develop"); err == nil {
		t.Fatal("expected same-branch error")
	}
	if err := ValidateBranches(doc, "missing", "develop"); err == nil {
		t.Fatal("expected missing source error")
	}
	if err := ValidateBranches(doc, "develop", "missing"); err == nil {
		t.Fatal("expected missing target error")
	}
}

func TestCreateValidationReturnsError(t *testing.T) {
	p := &project.Project{Tree: testTree()}
	cases := []CreateRequest{
		{Source: "missing", Target: "develop"},
		{Source: "develop", Target: "missing"},
		{Source: "develop", Target: "develop"},
	}
	for _, req := range cases {
		res, err := Create(p, req)
		if err == nil {
			t.Fatalf("%+v: expected error", req)
		}
		if res != nil {
			t.Fatalf("%+v: expected nil result, got %+v", req, res)
		}
	}
}

func TestDefaultTitle(t *testing.T) {
	got := DefaultTitle("a", "b")
	want := "MR: a → b"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDefaultDescription(t *testing.T) {
	desc := DefaultDescription()
	if !strings.Contains(desc, "Manual merge request created by branchy on") {
		t.Fatalf("unexpected description: %q", desc)
	}
}
