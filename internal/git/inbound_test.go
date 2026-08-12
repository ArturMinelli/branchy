package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestInboundFilesRemoteTrackingChild(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	base := strings.TrimSpace(gitOutput(t, dir, "rev-parse", "HEAD"))
	writeCommit(t, dir, "a.txt", "a", "add a")
	writeCommit(t, dir, "b.txt", "b", "add b")
	runGit(t, dir, "update-ref", "refs/remotes/origin/child", base)

	n, err := InboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 inbound files via origin/child, got %d", n)
	}
}

func TestInboundFilesPrefersRemoteChildOverStaleLocal(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "checkout", "-b", "child")
	runGit(t, dir, "checkout", "main")
	writeCommit(t, dir, "a.txt", "a", "add a")
	writeCommit(t, dir, "b.txt", "b", "add b")
	tip := strings.TrimSpace(gitOutput(t, dir, "rev-parse", "HEAD"))
	runGit(t, dir, "update-ref", "refs/remotes/origin/main", tip)
	runGit(t, dir, "update-ref", "refs/remotes/origin/child", tip)

	n, err := InboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("origin/child already has parent tip; stale local child must not count, got %d", n)
	}
}

func TestInboundFilesPrefersRemoteParentOverStaleLocal(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "branch", "child")
	runGit(t, dir, "checkout", "-b", "tmp")
	writeCommit(t, dir, "remote-parent.txt", "x", "on origin/main")
	sha := strings.TrimSpace(gitOutput(t, dir, "rev-parse", "HEAD"))
	runGit(t, dir, "checkout", "main")
	runGit(t, dir, "branch", "-D", "tmp")
	runGit(t, dir, "update-ref", "refs/remotes/origin/main", sha)

	n, err := InboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 inbound file from origin/main, got %d", n)
	}
}

func TestInboundFilesVersionLikeRemoteName(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	base := strings.TrimSpace(gitOutput(t, dir, "rev-parse", "HEAD"))
	writeCommit(t, dir, "patch.txt", "p", "version branch")
	runGit(t, dir, "update-ref", "refs/remotes/origin/develop-1.20.5-gestao-estrategica", base)

	n, err := InboundFiles(dir, "main", "develop-1.20.5-gestao-estrategica")
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 inbound file for version-like remote branch, got %d", n)
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

func TestOutboundFilesChildAhead(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "checkout", "-b", "child")
	writeCommit(t, dir, "a.txt", "a", "add a")
	writeCommit(t, dir, "b.txt", "b", "add b")

	n, err := OutboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 outbound files, got %d", n)
	}
}

func TestOutboundFilesIdentical(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "branch", "child")

	n, err := OutboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected 0 outbound files, got %d", n)
	}
}

func TestOutboundFilesParentOnlyChanges(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "checkout", "-b", "child")
	runGit(t, dir, "checkout", "main")
	writeCommit(t, dir, "only-parent.txt", "x", "parent unique")

	n, err := OutboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("parent-only files must not count as outbound, got %d", n)
	}
}

func TestOutboundFilesMissingRef(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")

	n, err := OutboundFiles(dir, "main", "missing")
	if err == nil {
		t.Fatal("expected error for missing ref")
	}
	if n != 0 {
		t.Fatalf("error path must not return a count, got %d", n)
	}
}

func TestOutboundFilesSymmetryWithInbound(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "base.txt", "base", "initial")
	runGit(t, dir, "checkout", "-b", "child")
	writeCommit(t, dir, "child.txt", "c", "child unique")
	runGit(t, dir, "checkout", "main")
	writeCommit(t, dir, "parent.txt", "p", "parent unique")

	out, err := OutboundFiles(dir, "main", "child")
	if err != nil {
		t.Fatal(err)
	}
	inSwapped, err := InboundFiles(dir, "child", "main")
	if err != nil {
		t.Fatal(err)
	}
	if out != inSwapped {
		t.Fatalf("OutboundFiles(dir,p,c)=%d must equal InboundFiles(dir,c,p)=%d", out, inSwapped)
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
	_ = gitOutput(t, dir, args...)
}

func gitOutput(t *testing.T, dir string, args ...string) string {
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
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}
