package main

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Every test gets a throwaway HOME: the form autosaves its draft and appends
// to the launch log under ~/.cache/herdr-threads.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "new-session-form-test")
	if err != nil {
		panic(err)
	}
	os.Setenv("HOME", home)
	restoreDrafts = false
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}

func newFormModel(t *testing.T) (get func() model, send func(tea.Msg)) {
	t.Helper()
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	m = drive(m, m.Init())
	get = func() model { return m.(model) }
	send = func(msg tea.Msg) {
		next, cmd := m.Update(msg)
		m = drive(next, cmd)
	}
	return get, send
}

func typeText(send func(tea.Msg), s string) {
	send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
}

func TestDraftSurvivesThePopup(t *testing.T) {
	t.Cleanup(func() { clearDraft() })
	_, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "go look into it")

	d, ok := loadDraft()
	if !ok || d.Name != "dose bug" || d.Prompt != "go look into it" || d.Project != projects[0] {
		t.Fatalf("draft not saved as typed: %+v %v", d, ok)
	}

	restoreDrafts = true
	defer func() { restoreDrafts = false }()
	get, _ := newFormModel(t)
	if m := get(); *m.name != "dose bug" || *m.prompt != "go look into it" || m.notice == "" {
		t.Errorf("draft not restored: name %q prompt %q notice %q", *m.name, *m.prompt, m.notice)
	}
}

func TestEscOverAnUnsentPromptNeedsASecondPress(t *testing.T) {
	t.Cleanup(func() { clearDraft() })
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "go look into it")

	next, cmd := get().Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil || next.(model).err == "" {
		t.Fatal("first esc quit instead of warning")
	}
	_, cmd = next.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("second esc did not quit")
	}
	if _, ok := loadDraft(); !ok {
		t.Error("draft should stay on disk until main sees the discard")
	}
}

func TestEscWithNothingTypedQuitsAtOnce(t *testing.T) {
	get, _ := newFormModel(t)
	if _, cmd := get().Update(tea.KeyMsg{Type: tea.KeyEscape}); cmd == nil {
		t.Error("esc on an empty form should quit")
	}
}

func TestCtrlAInThePromptEditsInsteadOfSwitching(t *testing.T) {
	t.Cleanup(func() { clearDraft() })
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "look")
	send(tea.KeyMsg{Type: tea.KeyCtrlA})
	typeText(send, "go ")
	if m := get(); m.view != viewNew || *m.prompt != "go look" {
		t.Errorf("view %d prompt %q", m.view, *m.prompt)
	}
}
