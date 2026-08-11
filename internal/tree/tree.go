package tree

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// BranchNode is one node in the branch hierarchy.
type BranchNode struct {
	Children []string `yaml:"children"`
}

// Document is the on-disk branch-tree.yaml shape.
type Document struct {
	Branches map[string]BranchNode `yaml:"branches"`
}

// Edge is a parent→child sync pair.
type Edge struct {
	Parent string
	Child  string
}

// Load reads a branch-tree.yaml file.
func Load(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse branch tree: %w", err)
	}
	if doc.Branches == nil {
		doc.Branches = map[string]BranchNode{}
	}
	doc.Normalize()
	return &doc, nil
}

// Normalize ensures every branch referenced in a parent's children list
// also exists as a key in Branches.
func (d *Document) Normalize() {
	for _, node := range d.Branches {
		for _, child := range node.Children {
			d.EnsureNode(child)
		}
	}
}

// Save writes a branch-tree.yaml file.
func (d *Document) Save(path string) error {
	data, err := yaml.Marshal(d)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Names returns all branch names sorted.
func (d *Document) Names() []string {
	names := make([]string, 0, len(d.Branches))
	for name := range d.Branches {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Roots returns branches that are never listed as a child.
func (d *Document) Roots() []string {
	childOf := map[string]bool{}
	for _, node := range d.Branches {
		for _, child := range node.Children {
			childOf[child] = true
		}
	}
	var roots []string
	for name := range d.Branches {
		if !childOf[name] {
			roots = append(roots, name)
		}
	}
	sort.Strings(roots)
	return roots
}

// CollectEdges returns all parent→child edges below root (DFS).
func (d *Document) CollectEdges(root string) []Edge {
	var edges []Edge
	d.collectEdges(root, &edges)
	return edges
}

func (d *Document) collectEdges(parent string, edges *[]Edge) {
	node, ok := d.Branches[parent]
	if !ok {
		return
	}
	children := append([]string(nil), node.Children...)
	sort.Strings(children)
	for _, child := range children {
		*edges = append(*edges, Edge{Parent: parent, Child: child})
		d.collectEdges(child, edges)
	}
}

// Link adds parent→child edge; creates nodes if missing.
func (d *Document) Link(parent, child string) error {
	if parent == "" || child == "" {
		return fmt.Errorf("parent and child are required")
	}
	if parent == child {
		return fmt.Errorf("parent and child must differ")
	}

	if _, ok := d.Branches[parent]; !ok {
		d.Branches[parent] = BranchNode{}
	}
	if _, ok := d.Branches[child]; !ok {
		d.Branches[child] = BranchNode{}
	}

	node := d.Branches[parent]
	for _, existing := range node.Children {
		if existing == child {
			return fmt.Errorf("edge already exists: %s → %s", parent, child)
		}
	}
	node.Children = append(node.Children, child)
	sort.Strings(node.Children)
	d.Branches[parent] = node
	return nil
}

// EnsureNode creates an empty branch entry if missing.
func (d *Document) EnsureNode(name string) {
	if _, ok := d.Branches[name]; !ok {
		d.Branches[name] = BranchNode{}
	}
}

// SubtreeNames returns root and all descendants via DFS.
func (d *Document) SubtreeNames(root string) []string {
	if _, ok := d.Branches[root]; !ok {
		return nil
	}
	var names []string
	d.collectSubtreeNames(root, &names)
	return names
}

func (d *Document) collectSubtreeNames(name string, names *[]string) {
	*names = append(*names, name)
	node, ok := d.Branches[name]
	if !ok {
		return
	}
	children := append([]string(nil), node.Children...)
	sort.Strings(children)
	for _, child := range children {
		d.collectSubtreeNames(child, names)
	}
}

// ParentOf returns the parent branch name if child appears in any children list.
func (d *Document) ParentOf(child string) (string, bool) {
	for parent, node := range d.Branches {
		for _, c := range node.Children {
			if c == child {
				return parent, true
			}
		}
	}
	return "", false
}

// HasEdge reports whether parent lists child in its children.
func (d *Document) HasEdge(parent, child string) bool {
	node, ok := d.Branches[parent]
	if !ok {
		return false
	}
	for _, c := range node.Children {
		if c == child {
			return true
		}
	}
	return false
}

// UnlinkSubtree removes root and all descendants from the document.
func (d *Document) UnlinkSubtree(root string) error {
	if root == "" {
		return fmt.Errorf("branch name is required")
	}
	if _, ok := d.Branches[root]; !ok {
		return fmt.Errorf("branch %q not in tree", root)
	}

	parent, hasParent := d.ParentOf(root)
	members := d.SubtreeNames(root)
	for _, name := range members {
		delete(d.Branches, name)
	}

	if hasParent {
		node := d.Branches[parent]
		filtered := node.Children[:0]
		for _, c := range node.Children {
			if c != root {
				filtered = append(filtered, c)
			}
		}
		node.Children = filtered
		d.Branches[parent] = node
	}

	return nil
}
