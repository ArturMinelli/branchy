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
	return filesChanged(dir, child, parent, parent, child)
}

// OutboundFiles returns how many files would change on parent if child were
// merged into it — the files-changed count GitLab shows on MR child → parent.
//
// Comparison is a local three-dot diff: git diff --name-only parent...child.
// Argument order matches InboundFiles (parent, then child). Missing or
// unresolvable refs return an error; callers must not treat that as 0.
func OutboundFiles(dir, parent, child string) (int, error) {
	return filesChanged(dir, parent, child, parent, child)
}

// filesChanged counts non-empty lines from git diff --name-only left...right
// after resolving branch names. labelParent/labelChild are used only in errors.
func filesChanged(dir, left, right, labelParent, labelChild string) (int, error) {
	if dir == "" || left == "" || right == "" {
		return 0, fmt.Errorf("dir, parent, and child are required")
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return 0, err
	}

	leftRef, err := ResolveRef(abs, left)
	if err != nil {
		return 0, fmt.Errorf("compare %s → %s: %w", labelParent, labelChild, err)
	}
	rightRef, err := ResolveRef(abs, right)
	if err != nil {
		return 0, fmt.Errorf("compare %s → %s: %w", labelParent, labelChild, err)
	}

	spec := leftRef + "..." + rightRef
	out, err := exec.Command("git", "-C", abs, "diff", "--name-only", spec).Output()
	if err != nil {
		return 0, fmt.Errorf("compare %s → %s: %w", labelParent, labelChild, err)
	}

	n := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n, nil
}
