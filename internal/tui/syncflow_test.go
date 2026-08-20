package tui

import (
	"errors"
	"strings"
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
	m := SyncFlowModel{
		project:    p,
		fromBranch: "main",
		edges:      edges,
		edgeIndex:  0,
		step:       stepSyncEdgeConfirm,
		opts:       SyncFlowOptions{Embedded: true},
	}
	return m.resetEdgeConfirm()
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

func TestSyncFlowSummaryToBrowserWhenExistingSkip(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncSummary
	m.results = []sync.Result{{
		Parent: "main", Child: "develop", Action: mr.ActionSkipped,
		URL: "https://example.com/existing", Message: "open MR already exists",
	}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	flow := updated.(SyncFlowModel)
	if flow.step != stepSyncBrowser {
		t.Fatalf("expected stepSyncBrowser for existing MR URL, got %d", flow.step)
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

func TestSyncFlowBrowserWarnStaysOnStep(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncBrowserLoading
	m.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionCreated, URL: "https://example.com/1"}}
	updated, _ := m.Update(browserOpenResultMsg{warn: "could not open"})
	flow := updated.(SyncFlowModel)
	if flow.step != stepSyncBrowser {
		t.Fatalf("expected return to browser step on warn, got %d", flow.step)
	}
	if flow.browserWarn == "" {
		t.Fatal("expected browserWarn set")
	}
	if flow.finished {
		t.Fatal("must not finish while warning is shown")
	}
}

func TestSyncFlowBrowserNoFinishes(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncBrowser
	m.results = []sync.Result{{Parent: "main", Child: "develop", Action: mr.ActionCreated, URL: "https://example.com/1"}}
	m = m.resetBrowserConfirm()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	flow := updated.(SyncFlowModel)
	if !flow.finished {
		t.Fatal("expected finished after declining browser")
	}
}

func TestSyncFlowConfirmViewUsesPanel(t *testing.T) {
	m := testSyncFlowAtConfirm()
	view := m.View()
	if strings.Contains(view, "[y/N]") {
		t.Fatal("sync confirm view must not contain [y/N]")
	}
	if !strings.Contains(view, " No ") || !strings.Contains(view, " Yes ") {
		t.Fatal("expected confirm buttons in view")
	}
}

func TestSyncFlowPickerInboundBadges(t *testing.T) {
	m := newSyncFlowModel(testSyncProject(), "", SyncFlowOptions{})
	m.inbound = map[string]fileChangeCount{
		"develop":   {files: 7, ok: true},
		"feature-a": {files: 0, ok: true},
		"release":   {files: 2, ok: true},
	}
	m.branchList = newBranchListWithBadges(m.project.Tree.Names(), "Select root branch to sync from", inboundBadges(m.inbound))

	byName := map[string]branchItem{}
	for _, item := range m.branchList.Items() {
		bi, ok := item.(branchItem)
		if !ok {
			t.Fatal("expected branchItem")
		}
		byName[bi.name] = bi
	}

	if byName["develop"].Title() != "develop  7" {
		t.Fatalf("develop title: %q", byName["develop"].Title())
	}
	if byName["develop"].FilterValue() != "develop" {
		t.Fatalf("filter must stay bare name, got %q", byName["develop"].FilterValue())
	}
	if byName["feature-a"].Title() != "feature-a" {
		t.Fatalf("zero badge must be hidden, got %q", byName["feature-a"].Title())
	}
	if byName["main"].Title() != "main" {
		t.Fatalf("root must have no badge, got %q", byName["main"].Title())
	}
}

func TestSyncFlowConfirmShowsInboundCount(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.inbound = map[string]fileChangeCount{
		"develop": {files: 7, ok: true},
	}
	m = m.resetEdgeConfirm()
	view := m.View()
	if !strings.Contains(view, "7 files would change on develop") {
		t.Fatalf("expected confirm count, got:\n%s", view)
	}
	if !strings.Contains(view, "Create MR main → develop?") {
		t.Fatal("question text must stay unchanged")
	}
	if !strings.Contains(view, "Sync down from main") {
		t.Fatalf("expected downward context, got:\n%s", view)
	}
}

func TestSyncFlowConfirmShowsKnownZero(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.inbound = map[string]fileChangeCount{
		"develop": {files: 0, ok: true},
	}
	m = m.resetEdgeConfirm()
	view := m.View()
	if !strings.Contains(view, "0 files would change on develop") {
		t.Fatalf("expected known zero on confirm, got:\n%s", view)
	}
}

func TestSyncFlowUnknownInbound(t *testing.T) {
	m := newSyncFlowModel(testSyncProject(), "", SyncFlowOptions{})
	m.inbound = map[string]fileChangeCount{
		"develop":   {files: 3, ok: true},
		"feature-a": {ok: false},
	}
	m.branchList = newBranchListWithBadges(m.project.Tree.Names(), "Select root branch to sync from", inboundBadges(m.inbound))

	byName := map[string]branchItem{}
	for _, item := range m.branchList.Items() {
		bi := item.(branchItem)
		byName[bi.name] = bi
	}
	if byName["develop"].Title() != "develop  3" {
		t.Fatalf("sibling must still show N, got %q", byName["develop"].Title())
	}
	if byName["feature-a"].Title() != "feature-a  ?" {
		t.Fatalf("unknown picker badge: %q", byName["feature-a"].Title())
	}

	m.fromBranch = "develop"
	m.edges = []tree.Edge{{Parent: "develop", Child: "feature-a"}}
	m.edgeIndex = 0
	m.step = stepSyncEdgeConfirm
	m = m.resetEdgeConfirm()
	view := m.View()
	if !strings.Contains(view, "File count unavailable") {
		t.Fatalf("expected unavailable confirm line, got:\n%s", view)
	}
}

func TestSyncFlowApplyInboundRefreshesPickerAndConfirm(t *testing.T) {
	m := newSyncFlowModel(testSyncProject(), "", SyncFlowOptions{})
	m = m.applyInbound(map[string]fileChangeCount{"develop": {files: 7, ok: true}})
	found := false
	for _, item := range m.branchList.Items() {
		bi := item.(branchItem)
		if bi.name == "develop" && bi.Title() == "develop  7" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected picker badge to refresh to 7")
	}

	m = testSyncFlowAtConfirm()
	m = m.applyInbound(map[string]fileChangeCount{"develop": {files: 11, ok: true}})
	if !strings.Contains(m.View(), "11 files would change on develop") {
		t.Fatalf("expected confirm refresh, got:\n%s", m.View())
	}
}

func TestSyncFlowRemoteUpdateFailureKeepsSnapshot(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.inbound = map[string]fileChangeCount{"develop": {files: 7, ok: true}}
	m = m.resetEdgeConfirm()
	updated, _ := m.Update(remoteUpdateMsg{projectID: "test", path: "/tmp/test", err: errors.New("offline")})
	flow := updated.(SyncFlowModel)
	if !strings.Contains(flow.View(), "7 files would change on develop") {
		t.Fatalf("failed update must keep confirm count:\n%s", flow.View())
	}
	if strings.Contains(strings.ToLower(flow.View()), "fetch") {
		t.Fatalf("view must not mention fetch:\n%s", flow.View())
	}
}

func TestSyncFlowKeysWorkDuringRemoteUpdate(t *testing.T) {
	m := testSyncFlowAtConfirm()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd != nil {
		t.Fatal("decline should not dispatch async cmd")
	}
	flow := updated.(SyncFlowModel)
	if len(flow.results) != 1 {
		t.Fatalf("expected skip while fetch may be in flight, got %d results", len(flow.results))
	}
}

func TestSyncFlowLoadingIgnoresKeys(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.step = stepSyncProcessing
	m.loading = NewLoading(LoadingOptions{Message: "Creating MR"})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	flow := updated.(SyncFlowModel)
	if flow.step != stepSyncProcessing {
		t.Fatalf("expected to stay on processing, got %d", flow.step)
	}
	if cmd != nil {
		t.Fatal("expected no command while loading")
	}
}

func TestSyncFlowYesEntersLoading(t *testing.T) {
	m := testSyncFlowAtConfirm()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	flow := updated.(SyncFlowModel)
	if flow.step != stepSyncProcessing {
		t.Fatalf("expected stepSyncProcessing, got %d", flow.step)
	}
	if cmd == nil {
		t.Fatal("expected async edge command")
	}
	if !flow.loading.Active() {
		t.Fatal("expected loading active")
	}
}

func testDeepSyncProject() *project.Project {
	return &project.Project{
		ID:   "test",
		Path: "/tmp/test",
		Tree: &tree.Document{Branches: map[string]tree.BranchNode{
			"develop":   {Children: []string{"feature-a", "feature-b"}},
			"feature-a": {Children: []string{"leaf"}},
			"feature-b": {},
			"leaf":      {},
		}},
	}
}

func testSyncFlowUpward() SyncFlowModel {
	p := testDeepSyncProject()
	m := SyncFlowModel{
		project:    p,
		fromBranch: "develop",
		direction:  sync.Upward,
		edges:      sync.EdgesBelow(p.Tree, "develop", sync.Upward),
		edgeIndex:  0,
		step:       stepSyncEdgeConfirm,
		opts:       SyncFlowOptions{Embedded: true, Direction: sync.Upward},
	}
	return m.resetEdgeConfirm()
}

func TestSyncFlowUpwardOfferOrder(t *testing.T) {
	m := testSyncFlowUpward()
	if len(m.edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(m.edges))
	}
	if m.edges[0].Parent != "feature-a" || m.edges[0].Child != "leaf" {
		t.Fatalf("first edge should be deepest, got %+v", m.edges[0])
	}
	for _, e := range m.edges {
		if e.Parent == "main" {
			t.Fatal("must not include edges above the ceiling")
		}
	}
	view := m.View()
	if !strings.Contains(view, "Create MR leaf → feature-a?") {
		t.Fatalf("expected child→parent question, got:\n%s", view)
	}
	if !strings.Contains(view, "Sync up to develop") {
		t.Fatalf("expected upward context, got:\n%s", view)
	}
	if strings.Contains(view, "d: show") {
		t.Fatalf("confirm must not offer direction toggle:\n%s", view)
	}
}

func TestSyncFlowUpwardDeclineContinues(t *testing.T) {
	m := testSyncFlowUpward()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd != nil {
		t.Fatal("decline should not dispatch async cmd")
	}
	flow := updated.(SyncFlowModel)
	if len(flow.results) != 1 || flow.results[0].Action != mr.ActionSkipped {
		t.Fatalf("expected skipped result, got %+v", flow.results)
	}
	if flow.results[0].Source != "leaf" || flow.results[0].Target != "feature-a" {
		t.Fatalf("skip should record MR ends, got %s → %s", flow.results[0].Source, flow.results[0].Target)
	}
	if flow.step != stepSyncEdgeConfirm {
		t.Fatalf("expected next confirm, got %d", flow.step)
	}
	if !strings.Contains(flow.View(), "Create MR feature-a → develop?") {
		t.Fatalf("second edge should be feature-a → develop:\n%s", flow.View())
	}
}

func TestSyncFlowUpwardCountsOnParent(t *testing.T) {
	m := testSyncFlowUpward()
	m.inbound = map[string]fileChangeCount{"leaf": {files: 1, ok: true}}
	m.outbound = map[string]fileChangeCount{"leaf": {files: 9, ok: true}}
	m = m.resetEdgeConfirm()
	view := m.View()
	if !strings.Contains(view, "9 files would change on feature-a") {
		t.Fatalf("upward confirm should show outbound on parent:\n%s", view)
	}
	if strings.Contains(view, "1 files would change") {
		t.Fatalf("must not show inbound count:\n%s", view)
	}
}

func TestSyncFlowDownwardKeepsInboundCount(t *testing.T) {
	m := testSyncFlowAtConfirm()
	m.inbound = map[string]fileChangeCount{"develop": {files: 4, ok: true}}
	m.outbound = map[string]fileChangeCount{"develop": {files: 9, ok: true}}
	m = m.resetEdgeConfirm()
	view := m.View()
	if !strings.Contains(view, "4 files would change on develop") {
		t.Fatalf("downward confirm should show inbound on child:\n%s", view)
	}
	if strings.Contains(view, "9 files would change") {
		t.Fatalf("must not show outbound count:\n%s", view)
	}
	if !strings.Contains(view, "Sync down from main") {
		t.Fatalf("expected downward context:\n%s", view)
	}
}

func TestSyncFlowUpwardKnownZeroOnParent(t *testing.T) {
	m := testSyncFlowUpward()
	m.outbound = map[string]fileChangeCount{"leaf": {files: 0, ok: true}}
	m = m.resetEdgeConfirm()
	if !strings.Contains(m.View(), "0 files would change on feature-a") {
		t.Fatalf("expected known zero on parent:\n%s", m.View())
	}
}

func TestSyncFlowStandalonePickerStartsDownward(t *testing.T) {
	m := newSyncFlowModel(testSyncProject(), "", SyncFlowOptions{})
	if m.direction != sync.Downward {
		t.Fatalf("standalone must start downward, got %d", m.direction)
	}
	view := m.View()
	if !strings.Contains(view, "sync: downward (parent→child)") || !strings.Contains(view, "ctrl+↑: outbound") || !strings.Contains(view, "ctrl+↓: inbound") {
		t.Fatalf("picker help:\n%s", view)
	}
	if strings.Contains(view, "d: show") {
		t.Fatalf("picker must not list d:\n%s", view)
	}
}

func TestSyncFlowStandalonePickerChords(t *testing.T) {
	m := newSyncFlowModel(testDeepSyncProject(), "", SyncFlowOptions{})
	m.inbound = map[string]fileChangeCount{
		"feature-a": {files: 1, ok: true},
		"leaf":      {files: 2, ok: true},
	}
	m.outbound = map[string]fileChangeCount{
		"feature-a": {files: 8, ok: true},
		"leaf":      {files: 9, ok: true},
	}
	m = m.refreshPicker()

	byName := pickerByName(t, m)
	if byName["leaf"].Title() != "leaf  2" {
		t.Fatalf("downward badge: %q", byName["leaf"].Title())
	}
	if strings.HasPrefix(byName["leaf"].Title(), "↓ ") || strings.HasPrefix(byName["leaf"].Title(), "↑ ") {
		t.Fatalf("picker row must not use tree arrows: %q", byName["leaf"].Title())
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlUp})
	flow := updated.(SyncFlowModel)
	if flow.direction != sync.Upward {
		t.Fatal("ctrl+up should set upward")
	}
	view := flow.View()
	if !strings.Contains(view, "sync: upward (child→parent)") || !strings.Contains(view, "ctrl+↑: outbound") {
		t.Fatalf("upward picker help:\n%s", view)
	}
	if strings.Contains(view, "d: show") {
		t.Fatalf("picker must not list d:\n%s", view)
	}
	byName = pickerByName(t, flow)
	if byName["leaf"].Title() != "leaf  9" {
		t.Fatalf("upward badge: %q", byName["leaf"].Title())
	}

	updated, _ = flow.Update(tea.KeyMsg{Type: tea.KeyCtrlUp})
	flow = updated.(SyncFlowModel)
	if flow.direction != sync.Upward {
		t.Fatal("second ctrl+up must stay upward")
	}

	updated, _ = flow.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	flow = updated.(SyncFlowModel)
	if flow.direction != sync.Upward {
		t.Fatal("d must not change picker direction")
	}

	flow.fromBranch = "develop"
	flow.edges = sync.EdgesBelow(flow.project.Tree, "develop", flow.direction)
	flow.step = stepSyncEdgeConfirm
	flow = flow.resetEdgeConfirm()
	if flow.direction != sync.Upward {
		t.Fatal("direction must stay upward after root is chosen")
	}
	cv := flow.View()
	if !strings.Contains(cv, "Create MR leaf → feature-a?") {
		t.Fatalf("expected upward first confirm:\n%s", cv)
	}
	if strings.Contains(cv, "d: show") || strings.Contains(cv, "ctrl+↑") {
		t.Fatalf("confirm must not list direction chords:\n%s", cv)
	}

	updated, _ = flow.Update(tea.KeyMsg{Type: tea.KeyCtrlDown})
	flow = updated.(SyncFlowModel)
	if flow.direction != sync.Upward {
		t.Fatal("confirm ctrl+down must not change direction")
	}

	fresh := newSyncFlowModel(testDeepSyncProject(), "", SyncFlowOptions{})
	if fresh.direction != sync.Downward {
		t.Fatal("new standalone model must start downward")
	}
}

func pickerByName(t *testing.T, m SyncFlowModel) map[string]branchItem {
	t.Helper()
	byName := map[string]branchItem{}
	for _, item := range m.branchList.Items() {
		bi, ok := item.(branchItem)
		if !ok {
			t.Fatal("expected branchItem")
		}
		byName[bi.name] = bi
	}
	return byName
}
