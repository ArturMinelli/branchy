package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResolveRef maps a branch tree name to a git revision.
// Preference: remote-tracking branch (origin first), then a local branch,
// then whatever git already resolves (tag, etc.).
//
// Remote-tracking wins so inbound counts match GitLab after fetch, even when
// a stale local checkout of the same name still exists.
func ResolveRef(dir, name string) (string, error) {
	if dir == "" || name == "" {
		return "", fmt.Errorf("dir and name are required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for _, ref := range remoteTrackingRefs(abs, name) {
		if revExists(abs, ref) {
			return ref, nil
		}
	}

	local := "refs/heads/" + name
	if revExists(abs, local) {
		return local, nil
	}

	if revExists(abs, name) {
		return name, nil
	}
	return "", fmt.Errorf("unknown ref %q", name)
}

func revExists(dir, rev string) bool {
	err := exec.Command("git", "-C", dir, "rev-parse", "--verify", "--quiet", rev+"^{commit}").Run()
	return err == nil
}

func remoteTrackingRefs(dir, name string) []string {
	out, err := exec.Command("git", "-C", dir, "for-each-ref", "--format=%(refname)", "refs/remotes/").Output()
	if err != nil {
		return nil
	}

	var origin []string
	var others []string
	for _, line := range strings.Split(string(out), "\n") {
		ref := strings.TrimSpace(line)
		rest, ok := strings.CutPrefix(ref, "refs/remotes/")
		if !ok {
			continue
		}
		remote, branch, ok := strings.Cut(rest, "/")
		if !ok || branch != name {
			continue
		}
		if remote == "origin" {
			origin = append(origin, ref)
			continue
		}
		others = append(others, ref)
	}
	return append(origin, others...)
}
