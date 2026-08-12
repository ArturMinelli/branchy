package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// InboundFiles returns how many files would change on child if parent were
// merged into it — the files-changed count GitLab shows on MR parent → child.
//
// Comparison is a local three-dot diff: git diff --name-only child...parent.
// Branch names resolve via ResolveRef (remote-tracking, then local head).
// Missing or unresolvable refs return an error; callers must not treat that as 0.
func InboundFiles(dir, parent, child string) (int, error) {
	if dir == "" || parent == "" || child == "" {
		return 0, fmt.Errorf("dir, parent, and child are required")
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return 0, err
	}

	parentRef, err := ResolveRef(abs, parent)
	if err != nil {
		return 0, fmt.Errorf("compare %s → %s: %w", parent, child, err)
	}
	childRef, err := ResolveRef(abs, child)
	if err != nil {
		return 0, fmt.Errorf("compare %s → %s: %w", parent, child, err)
	}

	spec := childRef + "..." + parentRef
	out, err := exec.Command("git", "-C", abs, "diff", "--name-only", spec).Output()
	if err != nil {
		return 0, fmt.Errorf("compare %s → %s: %w", parent, child, err)
	}

	n := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n, nil
}
