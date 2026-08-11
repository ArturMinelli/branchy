package git

import (
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

func TestDefaultRemoteNone(t *testing.T) {
	dir := initTestRepo(t)
	if _, ok := DefaultRemote(dir); ok {
		t.Fatal("expected no default remote")
	}
}

func TestDefaultRemotePrefersOrigin(t *testing.T) {
	dir := initTestRepo(t)
	runGit(t, dir, "remote", "add", "upstream", "file:///tmp/upstream.git")
	runGit(t, dir, "remote", "add", "origin", "file:///tmp/origin.git")
	name, ok := DefaultRemote(dir)
	if !ok || name != "origin" {
		t.Fatalf("expected origin, got %q ok=%v", name, ok)
	}
}

func TestDefaultRemoteFirstWhenNoOrigin(t *testing.T) {
	dir := initTestRepo(t)
	runGit(t, dir, "remote", "add", "upstream", "file:///tmp/upstream.git")
	name, ok := DefaultRemote(dir)
	if !ok || name != "upstream" {
		t.Fatalf("expected upstream, got %q ok=%v", name, ok)
	}
}

func TestFetchDefaultRemoteNoRemotes(t *testing.T) {
	dir := initTestRepo(t)
	if err := FetchDefaultRemote(dir); err != nil {
		t.Fatalf("no remotes must be a no-op, got %v", err)
	}
}

func TestFetchDefaultRemoteUpdatesTracking(t *testing.T) {
	bare := t.TempDir()
	runGit(t, bare, "init", "--bare", "-b", "main")

	src := initTestRepo(t)
	writeCommit(t, src, "a.txt", "a", "first")
	runGit(t, src, "remote", "add", "origin", bare)
	runGit(t, src, "push", "-u", "origin", "main")

	dst := t.TempDir()
	clone := exec.Command("git", "clone", bare, dst)
	if out, err := clone.CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
	runGit(t, dst, "config", "commit.gpgsign", "false")

	writeCommit(t, src, "b.txt", "b", "second")
	runGit(t, src, "push", "origin", "main")

	before := gitOutput(t, dst, "rev-parse", "refs/remotes/origin/main")
	if err := FetchDefaultRemote(dst); err != nil {
		t.Fatal(err)
	}
	after := gitOutput(t, dst, "rev-parse", "refs/remotes/origin/main")
	if before == after {
		t.Fatal("expected origin/main to advance after fetch")
	}
	want := gitOutput(t, src, "rev-parse", "HEAD")
	if after != want {
		t.Fatalf("dst origin/main %q != src HEAD %q", after, want)
	}
}

func TestFetchDefaultRemoteUnreachable(t *testing.T) {
	dir := initTestRepo(t)
	writeCommit(t, dir, "a.txt", "a", "first")
	missing := filepath.Join(t.TempDir(), "missing.git")
	runGit(t, dir, "remote", "add", "origin", "file://"+missing)
	if err := FetchDefaultRemote(dir); err == nil {
		t.Fatal("expected error for unreachable remote")
	}
}

func TestFetchDefaultRemoteConcurrentShare(t *testing.T) {
	bare := t.TempDir()
	runGit(t, bare, "init", "--bare", "-b", "main")
	src := initTestRepo(t)
	writeCommit(t, src, "a.txt", "a", "first")
	runGit(t, src, "remote", "add", "origin", bare)
	runGit(t, src, "push", "-u", "origin", "main")

	dst := t.TempDir()
	if out, err := exec.Command("git", "clone", bare, dst).CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = FetchDefaultRemote(dst)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
}
