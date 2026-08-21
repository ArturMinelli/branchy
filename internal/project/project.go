package project

import (
	"fmt"
	"os"
	"path/filepath"

	"branchy/internal/config"
	"branchy/internal/git"
	"branchy/internal/tree"
)

// Project is a registered application with its branch tree.
type Project struct {
	ID   string
	Path string
	Tree *tree.Document
}

// Load resolves a project by ID from the global index.
func Load(id string) (*Project, error) {
	idx, err := config.LoadIndex()
	if err != nil {
		return nil, err
	}
	entry, ok := idx.Projects[id]
	if !ok {
		return nil, fmt.Errorf("project %q not found", id)
	}
	return loadFromEntry(id, entry.Path)
}

// ResolveFromCWD finds a project for the current working directory.
func ResolveFromCWD() (*Project, error) {
	root, err := git.Root("")
	if err != nil {
		return nil, err
	}
	idx, err := config.LoadIndex()
	if err != nil {
		return nil, err
	}
	id, entry, ok := idx.FindByPath(root)
	if !ok {
		return nil, fmt.Errorf("no branchy project registered for %s (run branchy init)", root)
	}
	return loadFromEntry(id, entry.Path)
}

func loadFromEntry(id, path string) (*Project, error) {
	treePath, err := config.TreePath(id)
	if err != nil {
		return nil, err
	}
	doc, err := tree.Load(treePath)
	if err != nil {
		return nil, err
	}
	return &Project{ID: id, Path: path, Tree: doc}, nil
}

// Link validates, adds parent→child, and persists the branch tree.
func (p *Project) Link(parent, child string) error {
	if err := p.Tree.Link(parent, child); err != nil {
		return err
	}
	return p.SaveTree()
}

// UnlinkResult summarizes a successful unlink.
type UnlinkResult struct {
	Parent  string
	Child   string
	Removed int
}

// Unlink validates, removes the child subtree, and persists the branch tree.
// When parent is empty, child must be a tree root (interactive root unlink).
func (p *Project) Unlink(parent, child string) (*UnlinkResult, error) {
	if child == "" {
		return nil, fmt.Errorf("parent and child are required")
	}
	if parent != "" {
		if parent == child {
			return nil, fmt.Errorf("parent and child must differ")
		}
		if _, ok := p.Tree.Branches[child]; !ok {
			return nil, fmt.Errorf("branch %q not in tree", child)
		}
		if !p.Tree.HasEdge(parent, child) {
			return nil, fmt.Errorf("edge not found: %s → %s", parent, child)
		}
	} else {
		if _, ok := p.Tree.Branches[child]; !ok {
			return nil, fmt.Errorf("branch %q not in tree", child)
		}
		if _, hasParent := p.Tree.ParentOf(child); hasParent {
			return nil, fmt.Errorf("branch %q is not a tree root", child)
		}
	}

	removed := len(p.Tree.SubtreeNames(child))
	if err := p.Tree.UnlinkSubtree(child); err != nil {
		return nil, err
	}
	if err := p.SaveTree(); err != nil {
		return nil, err
	}
	return &UnlinkResult{Parent: parent, Child: child, Removed: removed}, nil
}

// SaveTree persists the branch tree to the global store.
func (p *Project) SaveTree() error {
	treePath, err := config.TreePath(p.ID)
	if err != nil {
		return err
	}
	projectDir, err := config.ProjectDir(p.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return err
	}
	return p.Tree.Save(treePath)
}

// InitOptions controls project registration.
type InitOptions struct {
	Force bool
}

// Init registers the git repo at cwd and imports an existing branch-tree if present.
func Init(opts InitOptions) (*Project, error) {
	root, err := git.Root("")
	if err != nil {
		return nil, err
	}

	idx, err := config.LoadIndex()
	if err != nil {
		return nil, err
	}

	baseID := config.Slugify(git.SlugFromPath(root))
	id := idx.UniqueID(baseID)

	if existingID, _, ok := idx.FindByPath(root); ok && !opts.Force {
		return nil, fmt.Errorf("already registered as %q (use --force to re-import)", existingID)
	}
	if existingID, _, ok := idx.FindByPath(root); ok && opts.Force {
		id = existingID
	}

	projectDir, err := config.ProjectDir(id)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return nil, err
	}

	treePath, err := config.TreePath(id)
	if err != nil {
		return nil, err
	}

	var doc *tree.Document
	importPath := filepath.Join(root, "repo", "branch-tree.yaml")
	if _, err := os.Stat(importPath); err == nil {
		doc, err = tree.Load(importPath)
		if err != nil {
			return nil, fmt.Errorf("import %s: %w", importPath, err)
		}
	} else if _, err := os.Stat(treePath); err == nil {
		doc, err = tree.Load(treePath)
		if err != nil {
			return nil, err
		}
	} else {
		doc = &tree.Document{Branches: map[string]tree.BranchNode{
			"main": {Children: []string{}},
		}}
	}

	if err := doc.Save(treePath); err != nil {
		return nil, err
	}
	if err := idx.Register(id, root); err != nil {
		return nil, err
	}

	return &Project{ID: id, Path: root, Tree: doc}, nil
}

// ListAll returns all registered projects.
func ListAll() ([]*Project, error) {
	idx, err := config.LoadIndex()
	if err != nil {
		return nil, err
	}
	var projects []*Project
	for _, id := range idx.All() {
		p, err := Load(id)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}
