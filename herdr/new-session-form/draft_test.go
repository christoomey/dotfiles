package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Every test gets a throwaway HOME and cache dir: the form autosaves its
// draft and writes job specs under $HERDR_THREADS_DIR. Tests that look at
// either call threadsTemp for one of their own.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "new-session-form-test")
	if err != nil {
		panic(err)
	}
	os.Setenv("HOME", home)
	os.Setenv("HERDR_THREADS_DIR", filepath.Join(home, "threads"))
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

// The feature fields autosave too, and a draft with nothing but feature text
// opens the popup on that view.
func TestFeatureDraftSurvivesThePopup(t *testing.T) {
	threadsTemp(t)
	_, send := newFormModel(t)
	send(tea.KeyMsg{Type: tea.KeyCtrlF})
	typeText(send, "christoomey/toom-9-fix-thing")
	send(tea.KeyMsg{Type: tea.KeyTab})
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "dig into the retry path")

	d, ok := loadDraft()
	if !ok || d.Branch != "christoomey/toom-9-fix-thing" || d.FeatureName != "fix-thing" ||
		!strings.HasSuffix(d.FeaturePrompt, "dig into the retry path") || !d.PromptEdited {
		t.Fatalf("feature draft not saved: %+v %v", d, ok)
	}

	withRestore(t)
	get, _ := newFormModel(t)
	m := get()
	if m.view != viewFeature {
		t.Errorf("a feature-only draft should open on the features view, got %d", m.view)
	}
	if m.feature.Value() != d.Branch || m.featureName.Value() != d.FeatureName ||
		m.featurePrompt.Value() != d.FeaturePrompt || !m.promptEdited {
		t.Errorf("feature draft not restored: %q %q %q", m.feature.Value(), m.featureName.Value(), m.featurePrompt.Value())
	}
	if m.notice == "" || !strings.Contains(m.featureView(), "restored") {
		t.Error("the features view should show the restored-draft notice")
	}
}

// A branch alone is worth keeping; the prompt the form filled in itself is
// not unsent text.
func TestADefaultFeaturePromptIsNotUnsent(t *testing.T) {
	threadsTemp(t)
	get, send := newFormModel(t)
	send(tea.KeyMsg{Type: tea.KeyCtrlF})
	typeText(send, "christoomey/toom-9-fix-thing")

	m := get()
	if m.unsent() {
		t.Error("a derived /linear-review prompt should not block esc")
	}
	if d, ok := loadDraft(); !ok || d.Branch == "" {
		t.Errorf("a typed branch belongs in the draft: %+v %v", d, ok)
	}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape}); cmd == nil {
		t.Error("esc should quit at once")
	}
}

// draft.json written before the feature keys existed still loads.
func TestOldDraftsStillLoad(t *testing.T) {
	threadsTemp(t)
	if err := os.MkdirAll(threadsCacheDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `{"project":"moi","name":"dose bug","prompt":"go look","saved_at":"2026-09-18T10:00:00-04:00"}`
	if err := os.WriteFile(draftPath(), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	d, ok := loadDraft()
	if !ok || d.Name != "dose bug" || d.Prompt != "go look" || d.Project != "moi" || !d.featureEmpty() {
		t.Fatalf("legacy draft: %+v %v", d, ok)
	}

	withRestore(t)
	get, _ := newFormModel(t)
	m := get()
	if m.view != viewNew || *m.name != "dose bug" || *m.prompt != "go look" {
		t.Errorf("legacy draft not restored: view %d %q %q", m.view, *m.name, *m.prompt)
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
