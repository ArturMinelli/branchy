package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestComissionamentoFrontendCountsAfterFetch(t *testing.T) {
	dir := "/home/arturpeixoto/apps/5labs/comissionamento-frontend"
	if _, err := os.Stat(dir); err != nil {
		t.Skip("comissionamento-frontend not available")
	}

	if err := FetchDefaultRemote(dir); err != nil {
		t.Fatalf("fetch: %v", err)
	}

	edges := [][2]string{
		{"main", "release"},
		{"release", "homolog"},
		{"homolog", "develop"},
		{"develop", "develop-1.20-gestao-estrategica"},
	}
	for _, e := range edges {
		parent, child := e[0], e[1]
		wantIn, err := gitDiffCount(dir, child, parent)
		if err != nil {
			t.Fatalf("%s->%s inbound baseline: %v", parent, child, err)
		}
		wantOut, err := gitDiffCount(dir, parent, child)
		if err != nil {
			t.Fatalf("%s->%s outbound baseline: %v", parent, child, err)
		}

		in, err := InboundFiles(dir, parent, child)
		if err != nil {
			t.Errorf("%s->%s inbound: %v", parent, child, err)
			continue
		}
		out, err := OutboundFiles(dir, parent, child)
		if err != nil {
			t.Errorf("%s->%s outbound: %v", parent, child, err)
			continue
		}
		if in != wantIn {
			t.Errorf("%s->%s inbound=%d want %d", parent, child, in, wantIn)
		}
		if out != wantOut {
			t.Errorf("%s->%s outbound=%d want %d", parent, child, out, wantOut)
		}
	}
}

func gitDiffCount(dir, left, right string) (int, error) {
	childRef, err := ResolveRef(dir, left)
	if err != nil {
		return 0, err
	}
	parentRef, err := ResolveRef(dir, right)
	if err != nil {
		return 0, err
	}
	out, err := exec.Command("git", "-C", dir, "diff", "--name-only", childRef+"..."+parentRef).Output()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n, nil
}

func TestResolveRefUsesOriginForFrontendDevelop(t *testing.T) {
	dir := "/home/arturpeixoto/apps/5labs/comissionamento-frontend"
	if _, err := os.Stat(dir); err != nil {
		t.Skip("comissionamento-frontend not available")
	}
	ref, err := ResolveRef(dir, "develop")
	if err != nil {
		t.Fatal(err)
	}
	want := "refs/remotes/origin/develop"
	if ref != want {
		t.Fatalf("ResolveRef(develop)=%q want %q", ref, want)
	}
	local, _ := exec.Command("git", "-C", dir, "rev-parse", "refs/heads/develop").Output()
	remote, _ := exec.Command("git", "-C", dir, "rev-parse", "refs/remotes/origin/develop").Output()
	if string(local) != string(remote) {
		t.Logf("local develop differs from origin (expected); ResolveRef must use origin")
	}
	abs, _ := filepath.Abs(dir)
	if !revExists(abs, ref) {
		t.Fatal("origin/develop must exist")
	}
}
