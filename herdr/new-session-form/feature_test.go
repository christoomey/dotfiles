package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFeatureNameAndPromptDerivation(t *testing.T) {
	for _, c := range []struct{ branch, name, id string }{
		{"christoomey/meds-3241-expose-data-needed", "expose-data-needed", "MEDS-3241"},
		{"christoomey/TOOM-12-Weekly_Review", "weekly-review", "TOOM-12"},
		{"meds-77-no-user", "no-user", "MEDS-77"},
		{"christoomey/just-words", "just-words", ""},
		{"", "", ""},
	} {
		if got := featureNameFor(c.branch); got != c.name {
			t.Errorf("%q: name %q, want %q", c.branch, got, c.name)
		}
		if got := linearID(c.branch); got != c.id {
			t.Errorf("%q: id %q, want %q", c.branch, got, c.id)
		}
	}
}

// Name and prompt follow the branch until edited; clearing one hands it back.
func TestFeatureFieldsFollowBranchUntilEdited(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	m = drive(m, m.Init())
	send := func(msg tea.Msg) {
		next, cmd := m.Update(msg)
		m = drive(next, cmd)
	}
	typeText := func(s string) {
		for _, r := range s {
			send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
	}
	state := func() (string, string) {
		mm := m.(model)
		return mm.featureName.Value(), mm.featurePrompt.Value()
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlF})
	typeText("christoomey/toom-9-fix-thing")
	if n, p := state(); n != "fix-thing" || p != "/linear-review TOOM-9" {
		t.Fatalf("after branch: %q %q", n, p)
	}
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText("s")
	typeText("")
	send(tea.KeyMsg{Type: tea.KeyShiftTab})
	typeText("-more")
	if n, _ := state(); n != "fix-things" {
		t.Errorf("edited name should not follow the branch: %q", n)
	}
	send(tea.KeyMsg{Type: tea.KeyTab})
	for range "fix-things" {
		send(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	if n, _ := state(); n != "fix-thing-more" {
		t.Errorf("cleared name should follow the branch again: %q", n)
	}
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText("x")
	send(tea.KeyMsg{Type: tea.KeyBackspace})
	if _, p := state(); p != "/linear-review TOOM-9" {
		t.Errorf("cleared prompt should follow the branch again: %q", p)
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlB})
	mm := m.(model)
	if mm.mode != "feature-bg" || mm.featureProject != "moi" || mm.picked.Title != "fix-thing-more" {
		t.Errorf("submit: mode %q project %q name %q", mm.mode, mm.featureProject, mm.picked.Title)
	}
}

// The agents picker scopes the list by project: all by default, a/m/d in
// the picker, ctrl+m/d from the filter.
func TestAgentsProjectScope(t *testing.T) {
	var m tea.Model = newModel("minimal", viewAgents, projects[0])
	m = drive(m, m.Init())
	send := func(msg tea.Msg) {
		next, cmd := m.Update(msg)
		m = drive(next, cmd)
	}
	mm := m.(model)
	mm.agentsScoped = false
	m = mm
	send(agentsMsg{rows: sampleAgents()})
	send(tea.KeyMsg{Type: tea.KeyCtrlA}) // widen to every agent
	labels := func() []string {
		var out []string
		for _, r := range m.(model).agentNav {
			out = append(out, r.label())
		}
		return out
	}
	if got := labels(); len(got) != 6 {
		t.Fatalf("all: %v", got)
	}
	send(tea.KeyMsg{Type: tea.KeyCtrlD})
	if got := labels(); len(got) != 2 || got[0] != "dictation-app" {
		t.Errorf("ctrl+d: %v", got)
	}
	send(tea.KeyMsg{Type: tea.KeyShiftTab})
	send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if got := labels(); len(got) != 1 || got[0] != "deadlines" {
		t.Errorf("picker m: %v", got)
	}
	if !m.(model).agentProjectFocused {
		t.Error("picker should keep focus")
	}
	send(tea.KeyMsg{Type: tea.KeyLeft})
	send(tea.KeyMsg{Type: tea.KeyLeft})
	if got := labels(); len(got) != 6 || m.(model).agentProject != 0 {
		t.Errorf("wrapped back to all: %v (project %d)", got, m.(model).agentProject)
	}
}
