package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Index maps registered project IDs to their git roots.
type Index struct {
	Projects map[string]ProjectEntry `yaml:"projects"`
}

// ProjectEntry holds a registered application directory.
type ProjectEntry struct {
	Path string `yaml:"path"`
}

// LoadIndex reads ~/.config/branchy/index.yaml.
func LoadIndex() (*Index, error) {
	path, err := IndexPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Index{Projects: map[string]ProjectEntry{}}, nil
		}
		return nil, err
	}

	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	if idx.Projects == nil {
		idx.Projects = map[string]ProjectEntry{}
	}
	return &idx, nil
}

// SaveIndex writes ~/.config/branchy/index.yaml.
func (idx *Index) SaveIndex() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path, err := IndexPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(idx)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// FindByPath returns the project ID for an absolute git root path.
func (idx *Index) FindByPath(absPath string) (string, *ProjectEntry, bool) {
	absPath = filepath.Clean(absPath)
	for id, entry := range idx.Projects {
		if filepath.Clean(entry.Path) == absPath {
			e := entry
			return id, &e, true
		}
	}
	return "", nil, false
}

// Register adds or updates a project in the index.
func (idx *Index) Register(id, absPath string) error {
	absPath = filepath.Clean(absPath)
	for existingID, entry := range idx.Projects {
		if existingID != id && filepath.Clean(entry.Path) == absPath {
			return fmt.Errorf("path already registered as %q", existingID)
		}
	}
	idx.Projects[id] = ProjectEntry{Path: absPath}
	return idx.SaveIndex()
}

// All returns sorted project IDs.
func (idx *Index) All() []string {
	ids := make([]string, 0, len(idx.Projects))
	for id := range idx.Projects {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// UniqueID returns id or id-2, id-3, … if id is taken.
func (idx *Index) UniqueID(base string) string {
	if _, ok := idx.Projects[base]; !ok {
		return base
	}
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if _, ok := idx.Projects[candidate]; !ok {
			return candidate
		}
	}
}

// Slugify normalizes a string for use as project ID.
func Slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteRune('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "project"
	}
	return out
}
