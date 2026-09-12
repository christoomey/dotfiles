package main

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func mouse(action tea.MouseAction, y int) tea.MouseMsg {
	return tea.MouseMsg{X: 4, Y: y, Action: action, Button: tea.MouseButtonLeft}
}

// The live list on the day this was written: pharmacy is august's worktree
// but sits last, after dotfiles, in the flat order.
func sampleWs() []wsRow {
	aug, moi := "/aug/.git", "/moi/.git"
	return []wsRow{
		{ID: "aug", Label: "august", RepoKey: aug},
		{ID: "sync", Label: "emar-sync-tab-race", RepoKey: aug, Linked: true},
		{ID: "perf", Label: "emar-perf-investigation", RepoKey: aug, Linked: true},
		{ID: "start", Label: "start-end-time-data-model", RepoKey: aug, Linked: true},
		{ID: "moi", Label: "moi", RepoKey: moi},
		{ID: "dot", Label: "dotfiles"},
		{ID: "focus", Label: "focus-mode-website-blocker", RepoKey: moi, Linked: true},
		{ID: "dead", Label: "deadlines", RepoKey: moi, Linked: true},
		{ID: "pharm", Label: "pharmacy-person-mappings", RepoKey: aug, Linked: true},
	}
}

func ids(ws []wsRow) []string {
	out := make([]string, len(ws))
	for i, r := range ws {
		out[i] = r.ID
	}
	return out
}

func nodeIDs(ws []wsRow) []string {
	_, nodes := buildSpaces(ws)
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = ws[n.flat].ID
	}
	return out
}

func TestBuildSpacesGroupsLikeTheSidebar(t *testing.T) {
	want := []string{"aug", "sync", "perf", "start", "pharm", "moi", "focus", "dead", "dot"}
	if got := nodeIDs(sampleWs()); !reflect.DeepEqual(got, want) {
		t.Errorf("nodes %v, want %v", got, want)
	}
}

func TestStepWorktreeAmongSiblings(t *testing.T) {
	ws := sampleWs()
	// pharm (node 4) up past start: insert before start's flat slot 3.
	out, mv, ok := stepNode(ws, 4, -1)
	if !ok || mv.id != "pharm" || mv.insertIndex != 3 {
		t.Fatalf("up: ok=%v mv=%+v", ok, mv)
	}
	if got := ids(out)[3]; got != "pharm" {
		t.Errorf("local order %v", ids(out))
	}
	// start (node 3) down past pharm at flat 8: slot 9 = the end.
	out, mv, ok = stepNode(ws, 3, 1)
	if !ok || mv.id != "start" || mv.insertIndex != 9 {
		t.Fatalf("down: ok=%v mv=%+v", ok, mv)
	}
	if got := ids(out)[8]; got != "start" {
		t.Errorf("local order %v", ids(out))
	}
	// sync (node 1) is the first sibling: no move up.
	if _, _, ok := stepNode(ws, 1, -1); ok {
		t.Error("expected no-op at the top of the space")
	}
	// A worktree never leaves its space: pharm is the last sibling.
	if _, _, ok := stepNode(ws, 4, 1); ok {
		t.Error("expected no-op at the bottom of the space")
	}
}

func TestStepPrimaryMovesWholeSpace(t *testing.T) {
	ws := sampleWs()
	// dot (node 8) up past moi's space: block before moi's first row.
	out, mv, ok := stepNode(ws, 8, -1)
	if !ok || !reflect.DeepEqual(mv.block, []string{"dot"}) || mv.before != "moi" {
		t.Fatalf("up: ok=%v mv=%+v", ok, mv)
	}
	if got := nodeIDs(out); !reflect.DeepEqual(got, []string{"aug", "sync", "perf", "start", "pharm", "dot", "moi", "focus", "dead"}) {
		t.Errorf("after up: %v", got)
	}
	// aug (node 0) down past moi: the block is the whole space in list
	// order, dropped before dotfiles.
	out, mv, ok = stepNode(ws, 0, 1)
	if !ok || !reflect.DeepEqual(mv.block, []string{"aug", "sync", "perf", "start", "pharm"}) || mv.before != "dot" {
		t.Fatalf("down: ok=%v mv=%+v", ok, mv)
	}
	if got := nodeIDs(out); !reflect.DeepEqual(got, []string{"moi", "focus", "dead", "aug", "sync", "perf", "start", "pharm", "dot"}) {
		t.Errorf("after down: %v", got)
	}
	// moi (node 5) down past dot, the last space: no anchor.
	_, mv, ok = stepNode(ws, 5, 1)
	if !ok || mv.before != "" {
		t.Fatalf("to end: ok=%v mv=%+v", ok, mv)
	}
	if _, _, ok := stepNode(ws, 0, -1); ok {
		t.Error("expected no-op at the first space")
	}
}

func TestMoveFlatMatchesServerSemantics(t *testing.T) {
	ws := sampleWs()
	// Verified live: moving flat 2 to slot 1 lands it at 1; moving it back
	// needs slot 3, not 2.
	out := moveFlat(ws, 2, 1)
	if got := ids(out)[:3]; !reflect.DeepEqual(got, []string{"aug", "perf", "sync"}) {
		t.Errorf("slot 1: %v", got)
	}
	back := moveFlat(out, 1, 3)
	if !reflect.DeepEqual(ids(back), ids(ws)) {
		t.Errorf("slot 3 did not restore: %v", ids(back))
	}
	same := moveFlat(out, 1, 2)
	if !reflect.DeepEqual(ids(same), ids(out)) {
		t.Errorf("slot 2 should be a no-op: %v", ids(same))
	}
}

func TestDragStepsToTheRowUnderThePointer(t *testing.T) {
	m := newModel("minimal", viewNew, projects[0])
	m.view = viewSidebar
	m.ws = sampleWs()
	m.wsBranch = map[string]string{}
	m.listRows = 12
	// No branches in the fixture, so rows are nodes: 4 aug, 5 sync, 6 perf, 7 start, 8 pharm.
	if got := m.wsNodeAt(wsListTop + 4); got != 4 {
		t.Fatalf("row → node %d, want 4", got)
	}
	var tm tea.Model = m
	tm, _ = tm.Update(mouse(tea.MouseActionPress, wsListTop+4))  // pick up pharm
	tm, _ = tm.Update(mouse(tea.MouseActionMotion, wsListTop+1)) // drag to sync's row
	got := tm.(model)
	if !got.wsDrag || got.wsCursor != 1 {
		t.Fatalf("drag=%v cursor=%d", got.wsDrag, got.wsCursor)
	}
	if ids := nodeIDs(got.ws); ids[1] != "pharm" || ids[4] != "start" {
		t.Errorf("order after drag: %v", ids)
	}
	tm, _ = tm.Update(mouse(tea.MouseActionRelease, wsListTop+1))
	if tm.(model).wsDrag {
		t.Error("release should end the drag")
	}
}
