package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDebugName(t *testing.T) {
	now := time.Date(2026, 9, 30, 9, 5, 0, 0, time.Local)
	for _, c := range []struct{ prompt, want string }{
		{"Why did the nightly export job fail last night?", "why did the nightly export"},
		{"\n  \n...\nExport: timeout!", "export timeout"},
		{"supercalifragilisticexpialidocious antidisestablishmentarianism", "supercalifragilisticexpialidocious antid"},
		{"?!\n", "debug Sep 30 09:05"},
		{"", "debug Sep 30 09:05"},
	} {
		if got := debugName(c.prompt, now); got != c.want {
			t.Errorf("debugName(%q) = %q, want %q", c.prompt, got, c.want)
		}
	}
}

func TestDebugRequiresAPrompt(t *testing.T) {
	threadsTemp(t)
	get, send := newFormModel(t)
	send(ctrlDigitMsg(5))
	send(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m := get(); m.mode != "" || m.err == "" {
		t.Errorf("empty prompt submitted: mode %q err %q", m.mode, m.err)
	}
}

func TestDebugSubmitsAnOpenDebugJob(t *testing.T) {
	dir := threadsTemp(t)
	get, send := newFormModel(t)
	send(ctrlDigitMsg(5))
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "why did the export fail")
	send(tea.KeyMsg{Type: tea.KeyEnter})
	typeText(send, "last night")
	send(tea.KeyMsg{Type: tea.KeyCtrlS})

	m := get()
	if m.mode != "open" {
		t.Fatalf("mode %q, want open", m.mode)
	}
	res := m.resultFor("open")
	want := result{Mode: "open", Kind: "debug", Project: "august", Name: "why did the export fail", Prompt: "why did the export fail\nlast night"}
	if res.Mode != want.Mode || res.Kind != want.Kind || res.Project != want.Project || res.Name != want.Name || res.Prompt != want.Prompt {
		t.Errorf("result %+v, want %+v", res, want)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "jobs", m.jobID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var j job
	json.Unmarshal(raw, &j)
	if j.Kind != "debug" || j.Mode != "open" || j.Prompt != want.Prompt {
		t.Errorf("spec %+v", j)
	}
}

func TestDebugTitleNamesTheSession(t *testing.T) {
	threadsTemp(t)
	get, send := newFormModel(t)
	send(ctrlDigitMsg(5))
	typeText(send, "export failures")
	send(tea.KeyMsg{Type: tea.KeyEnter})
	typeText(send, "take a look at this")
	send(ctrlEnterMsg{})

	m := get()
	res := m.resultFor(m.mode)
	if m.mode != "bg" || res.Name != "export failures" || res.Prompt != "take a look at this" {
		t.Errorf("mode %q result %+v", m.mode, res)
	}
}

func TestDebugCtrlBSubmitsABackgroundDebugJob(t *testing.T) {
	dir := threadsTemp(t)
	get, send := newFormModel(t)
	send(ctrlDigitMsg(5))
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "why did the export fail")
	send(ctrlEnterMsg{})

	m := get()
	if m.mode != "bg" {
		t.Fatalf("mode %q, want bg", m.mode)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "jobs", m.jobID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var j job
	json.Unmarshal(raw, &j)
	if j.Kind != "debug" || j.Mode != "bg" {
		t.Errorf("spec %+v", j)
	}
}

func TestDebugDraftSurvivesThePopup(t *testing.T) {
	threadsTemp(t)
	get, send := newFormModel(t)
	send(ctrlDigitMsg(5))
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "why did the export fail")

	d, ok := loadDraft()
	if !ok || d.DebugPrompt != "why did the export fail" || !d.debugOnly() {
		t.Fatalf("debug draft not saved: %+v %v", d, ok)
	}

	withRestore(t)
	get, _ = newFormModel(t)
	if m := get(); m.view != viewDebug || m.debugPrompt.Value() != "why did the export fail" {
		t.Errorf("debug draft not restored: view %d %q", m.view, m.debugPrompt.Value())
	}
}
