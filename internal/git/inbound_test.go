package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInboundFilesParentAhead(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "checkout", "-b", "child")
	runGit(t, dir, "checkout", "main")
	writeCommit(t, dir, "a.txt", "a", "add a")
	writeCommit(t, dir, "b.txt", "b", "add b")
	writeCommit(t, dir, "c.txt", "c", "add c")

	n, err := InboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("expected 3 inbound files, got %d", n)
	}
}

func TestInboundFilesIdentical(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "branch", "child")

	n, err := InboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected 0 inbound files, got %d", n)
	}
}

func TestInboundFilesMissingRef(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")

	n, err := InboundFiles(dir, "main", "missing")
	if err == nil {
		t.Fatal("expected error for missing ref")
	}
	if n != 0 {
		t.Fatalf("error path must not return a count, got %d", n)
	}
}

func TestInboundFilesChildOnlyChanges(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "checkout", "-b", "child")
	writeCommit(t, dir, "only-child.txt", "x", "child unique")

	n, err := InboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("child-only files must not count as inbound, got %d", n)
	}
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.name", "test")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

func writeCommit(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", name)
	runGit(t, dir, "commit", "-m", msg)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_NOSYSTEM=1",
		"EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
