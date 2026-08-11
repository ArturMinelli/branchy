package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/mr"
	"branchy/internal/project"
	"branchy/internal/sync"
	"branchy/internal/tree"
)

func testSyncProject() *project.Project {
	return &project.Project{
		ID:   "test",
		Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"main":      {Children: []string{"develop"}},
			"develop":   {Children: []string{"feature-a", "release"}},
			"feature-a": {},
			"release":   {},
		}},
	}
}

func testSyncFlowAtConfirm() SyncFlowModel {
	p := testSyncProject()
	edges := p.Tree.CollectEdges("main")
	return SyncFlowModel{
		project:    p,
		fromBranch: "main",
		edges:      edges,
		edgeIndex:  0,
		step:       stepSyncEdgeConfirm,
		opts:       SyncFlowOptions{Embedded: true},
	}
}

func TestSyncFlowEmptyEdges(t *testing.T) {
	p := &project.Project{
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"leaf": {},
		}},
	}
	m := newSyncFlowModel(p, "leaf", SyncFlowOptions{})
	if m.step != stepSyncEmpty {
		t.Fatalf("expected stepSyncEmpty, got %d", m.step)
	}
}

func TestSyncFlowPickRootWhenFromEmpty(t *testing.T) {
	m := newSyncFlowModel(testSyncProject(), "", SyncFlowOptions{})
	if m.step != stepSyncPickRoot {
		t.Fatalf("expected stepSyncPickRoot, got %d", m.step)
	}
}

func TestSyncFlowPickRootSelectsBranch(t *testing.T) {
	m := newSyncFlowModel(testSyncProject(), "", SyncFlowOptions{})
	m.branchList.Select(0)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(SyncFlowModel)
	if flow.fromBranch == "" {
		t.Fatal("expected fromBranch set after pick")
	}
	if flow.step != stepSyncEdgeConfirm && flow.step != stepSyncEmpty && flow.step != stepSyncError {
		t.Fatalf("unexpected step after pick: %d", flow.step)
	}
}

func TestSyncFlowStandaloneFinishQuits(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.opts = SyncFlowOptions{Embedded: false}
	m.step = stepSyncSummary
	m.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionSkipped, Message: "skipped"}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(SyncFlowModel)
	if !flow.finished {
		t.Fatal("expected finished")
	}
	if cmd == nil {
		t.Fatal("expected quit command for standalone flow")
	}
}

func TestSyncFlowDeclineSkipsEdge(t *testing.T) {
	m := testSyncFlowAtConfirm()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd != nil {
		t.Fatal("decline should not dispatch async cmd")
	}
	flow := updated.(SyncFlowModel)
	if len(flow.results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(flow.results))
	}
	if flow.results[0].Action != mr.ActionSkipped {
		t.Fatalf("expected skipped, got %q", flow.results[0].Action)
	}
	if flow.edgeIndex != 1 {
		t.Fatalf("expected edgeIndex 1, got %d", flow.edgeIndex)
	}
	if flow.step != stepSyncEdgeConfirm {
		t.Fatalf("expected next edge confirm, got %d", flow.step)
	}
}

func TestSyncFlowCancelGoesToSummary(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionCreated, URL: "https://example.com/1"}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	flow := updated.(SyncFlowModel)
	if !flow.cancelled {
		t.Fatal("expected cancelled")
	}
	if flow.step != stepSyncSummary {
		t.Fatalf("expected stepSyncSummary, got %d", flow.step)
	}
}

func TestSyncFlowEdgeResultAdvances(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncProcessing
	updated, _ := m.Update(edgeResultMsg{result: sync.Result{
		Parent: "main", Child: "develop", Action: mr.ActionCreated, URL: "https://example.com/1",
	}})
	flow := updated.(SyncFlowModel)
	if flow.edgeIndex != 1 {
		t.Fatalf("expected edgeIndex 1, got %d", flow.edgeIndex)
	}
	if flow.step != stepSyncEdgeConfirm {
		t.Fatalf("expected stepSyncEdgeConfirm, got %d", flow.step)
	}
}

func TestSyncFlowLastEdgeGoesToSummary(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.edgeIndex = len(m.edges) - 1
	m.step = stepSyncProcessing
	updated, _ := m.Update(edgeResultMsg{result: sync.Result{
		Parent: "develop", Child: "release", Action: mr.ActionCreated, URL: "https://example.com/2",
	}})
	flow := updated.(SyncFlowModel)
	if flow.step != stepSyncSummary {
		t.Fatalf("expected stepSyncSummary, got %d", flow.step)
	}
}

func TestSyncFlowSummaryToBrowserWhenCreated(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncSummary
	m.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionCreated, URL: "https://example.com/1"}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(SyncFlowModel)
	if flow.step != stepSyncBrowser {
		t.Fatalf("expected stepSyncBrowser, got %d", flow.step)
	}
}

func TestSyncFlowSummarySkipsBrowserWhenNoCreates(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncSummary
	m.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionSkipped, Message: "skipped by user"}}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(SyncFlowModel)
	if !flow.finished {
		t.Fatal("expected finished when no creates")
	}
	if cmd != nil {
		t.Fatal("embedded finish should not quit program")
	}
}

func TestSyncFlowBrowserNoFinishes(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncBrowser
	m.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionCreated, URL: "https://example.com/1"}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	flow := updated.(SyncFlowModel)
	if !flow.finished {
		t.Fatal("expected finished after declining browser")
	}
}
