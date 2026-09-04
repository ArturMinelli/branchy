package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"branchy/internal/mr"
	"branchy/internal/project"
	"branchy/internal/sync"
	"branchy/internal/tree"
)

func TestEmbeddedSyncForwardsEdgeResult(t *testing.T) {
	p := &project.Project{
		ID: "test", Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"main": {Children: []string{"develop"}}, "develop": {},
		}},
	}
	m := newModel([]*project.Project{p}, p)
	m.screen = screenSync
	m.syncFlow = testSyncFlowAtConfirm()
	m.syncFlow.step = stepSyncProcessing

	updated, _ := m.Update(edgeResultMsg{result: sync.Result{
		Parent: "main", Child: "develop", Action: mr.ActionCreated, URL: "https://example.com/1",
	}})
	model := updated.(Model)
	if model.syncFlow.edgeIndex != 1 {
		t.Fatalf("expected edgeIndex 1, got %d", model.syncFlow.edgeIndex)
	}
	if model.syncFlow.step != stepSyncEdgeConfirm {
		t.Fatalf("expected stepSyncEdgeConfirm, got %d", model.syncFlow.step)
	}
}

func TestEmbeddedSyncReturnsToTreeWhenFinished(t *testing.T) {
	p := &project.Project{
		ID: "test", Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"main": {Children: []string{"develop"}}, "develop": {},
		}},
	}
	m := newModel([]*project.Project{p}, p)
	m.screen = screenSync
	m.syncFlow = testSyncFlowAtConfirm()
	m.syncFlow.step = stepSyncSummary
	m.syncFlow.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionSkipped, Message: "skipped"}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected remote refresh when returning to tree from sync")
	}
	model := updated.(Model)
	if model.screen != screenTree {
		t.Fatalf("expected screenTree, got %d", model.screen)
	}
}
