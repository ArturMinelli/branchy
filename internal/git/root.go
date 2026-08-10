package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Root returns the absolute git repository root for dir (or cwd).
func Root(dir string) (string, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	out, err := exec.Command("git", "-C", abs, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %s", abs)
	}
	return strings.TrimSpace(string(out)), nil
}

// SlugFromPath derives a project ID from the directory name.
func SlugFromPath(root string) string {
	return filepath.Base(root)
}
