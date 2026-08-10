package config

import (
	"os"
	"path/filepath"
)

// Dir returns the branchy config root (~/.config/branchy).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "branchy"), nil
}

// IndexPath returns ~/.config/branchy/index.yaml.
func IndexPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "index.yaml"), nil
}

// ProjectDir returns ~/.config/branchy/projects/<id>.
func ProjectDir(id string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects", id), nil
}

// TreePath returns ~/.config/branchy/projects/<id>/branch-tree.yaml.
func TreePath(id string) (string, error) {
	dir, err := ProjectDir(id)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "branch-tree.yaml"), nil
}
