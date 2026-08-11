package tui

import (
	"os/exec"
	"testing"

	"branchy/internal/project"
	"branchy/internal/tree"
)

func TestRemoteUpdateCmdNoRemotes(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	p := &project.Project{
		ID:   "probe",
		Path: dir,
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{"main": {}}},
	}
	teaCmd := remoteUpdateCmd(p)
	if teaCmd == nil {
		t.Fatal("expected cmd")
	}
	msg := teaCmd()
	ru, ok := msg.(remoteUpdateMsg)
	if !ok {
		t.Fatalf("expected remoteUpdateMsg, got %T", msg)
	}
	if ru.projectID != "probe" {
		t.Fatalf("projectID %q", ru.projectID)
	}
	if ru.err != nil {
		t.Fatalf("no remotes must be nil error, got %v", ru.err)
	}
}
