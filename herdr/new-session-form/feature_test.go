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
	if got := labels(); len(got) != 8 {
		t.Fatalf("all: %v", got)
	}
	send(tea.KeyMsg{Type: tea.KeyCtrlD})
	if got := labels(); len(got) != 2 || got[0] != "dictation app" {
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
	if got := labels(); len(got) != 8 || m.(model).agentProject != 0 {
		t.Errorf("wrapped back to all: %v (project %d)", got, m.(model).agentProject)
	}
}

func TestFeatureBranch(t *testing.T) {
	t.Setenv("USER", "someone")
	for _, c := range []struct{ in, branch string }{
		{"christoomey/meds-3241-expose-data", "christoomey/meds-3241-expose-data"},
		{"christoomey/just-words", "christoomey/just-words"},
		{"medpass-hyphen-filter", "someone/medpass-hyphen-filter"},
		{"  Medpass Hyphen_Filter ", "someone/medpass-hyphen-filter"},
		{"", ""},
		{"christoomey/has space", ""},
		{"---", ""},
	} {
		got, err := featureBranch(c.in)
		if c.branch == "" {
			if err == nil {
				t.Errorf("%q: want an error, got %q", c.in, got)
			}
		} else if err != nil || got != c.branch {
			t.Errorf("%q: %q %v, want %q", c.in, got, err, c.branch)
		}
	}
}

// Without a Linear id the project has to be picked: submit refuses and moves
// to the picker, a/m or ctrl+m pick, and the bare name gets the user prefix.
func TestFeatureWithoutLinearIDNeedsAProject(t *testing.T) {
	t.Setenv("USER", "someone")
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

	send(tea.KeyMsg{Type: tea.KeyCtrlF})
	typeText("hyphen-filter")
	if mm := m.(model); mm.featureName.Value() != "hyphen-filter" || mm.featurePrompt.Value() != "" {
		t.Fatalf("derived %q %q", mm.featureName.Value(), mm.featurePrompt.Value())
	}
	send(tea.KeyMsg{Type: tea.KeyCtrlS})
	mm := m.(model)
	if mm.mode != "" || mm.err == "" || mm.featureFocus != featProject {
		t.Fatalf("submit without a project: mode %q err %q focus %d", mm.mode, mm.err, mm.featureFocus)
	}
	typeText("m")
	if p, _ := m.(model).featureTarget(); p != "moi" {
		t.Errorf("m should pick moi: %q", p)
	}
	send(tea.KeyMsg{Type: tea.KeyRight})
	if p, _ := m.(model).featureTarget(); p != "august" {
		t.Errorf("right should wrap to august: %q", p)
	}
	send(tea.KeyMsg{Type: tea.KeyShiftTab})
	send(ctrlMMsg{})
	if p, _ := m.(model).featureTarget(); p != "moi" {
		t.Errorf("ctrl+m should pick moi from the branch field: %q", p)
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlB})
	mm = m.(model)
	if mm.mode != "feature-bg" || mm.featureProject != "moi" || mm.branch != "someone/hyphen-filter" || mm.picked.Title != "hyphen-filter" {
		t.Errorf("submit: mode %q project %q branch %q name %q", mm.mode, mm.featureProject, mm.branch, mm.picked.Title)
	}
}

// A Linear branch decides the project, so tab steps over the picker and a
// hand pick cannot override it.
func TestFeatureLinearBranchSkipsThePicker(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	m = drive(m, m.Init())
	send := func(msg tea.Msg) {
		next, cmd := m.Update(msg)
		m = drive(next, cmd)
	}
	send(tea.KeyMsg{Type: tea.KeyCtrlF})
	for _, r := range "christoomey/meds-1-thing" {
		send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	send(ctrlMMsg{})
	send(tea.KeyMsg{Type: tea.KeyTab})
	mm := m.(model)
	if mm.featureFocus != featName {
		t.Errorf("tab should skip the picker: focus %d", mm.featureFocus)
	}
	if p, inferred := mm.featureTarget(); p != "august" || !inferred {
		t.Errorf("target %q inferred %v", p, inferred)
	}
	send(tea.KeyMsg{Type: tea.KeyShiftTab})
	if f := m.(model).featureFocus; f != featBranch {
		t.Errorf("shift+tab should skip the picker back: focus %d", f)
	}
}
