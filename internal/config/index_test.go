package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueID(t *testing.T) {
	idx := &Index{Projects: map[string]ProjectEntry{
		"backend": {Path: "/a"},
	}}
	if got := idx.UniqueID("backend"); got != "backend-2" {
		t.Fatalf("got %q", got)
	}
	if got := idx.UniqueID("frontend"); got != "frontend" {
		t.Fatalf("got %q", got)
	}
}

func TestRegisterAndFind(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	idx := &Index{Projects: map[string]ProjectEntry{}}
	path := filepath.Join(dir, "repo")
	if err := idx.Register("repo", path); err != nil {
		t.Fatal(err)
	}
	id, _, ok := idx.FindByPath(path)
	if !ok || id != "repo" {
		t.Fatalf("find failed: %s %v", id, ok)
	}
}
