package tui

import (
	"strings"
	"testing"

	"branchy/internal/project"
)

func TestRenderBackendTreeIntegration(t *testing.T) {
	p, err := project.Load("comissionamento-backend")
	if err != nil {
		t.Skip("comissionamento-backend not registered:", err)
	}

	v := newBranchTreeView(p.Tree)
	out := stripANSI(v.View())

	for _, want := range []string{"main", "release", "develop", "└──", "├──"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q:\n%s", want, out)
		}
	}

	t.Logf("tree view:\n%s", out)
}

func stripANSI(s string) string {
	var b strings.Builder
	skip := false
	for _, r := range s {
		if r == '\x1b' {
			skip = true
			continue
		}
		if skip {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
				skip = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
