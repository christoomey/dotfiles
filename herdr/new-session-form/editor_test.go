package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEditorKeys(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	m = drive(m, m.Init())
	var last tea.Cmd
	send := func(msg tea.Msg) {
		var next tea.Model
		next, last = m.Update(msg)
		m = next
	}
	typeText := func(s string) {
		for _, r := range s {
			send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlG})
	if err := m.(model).err; !strings.Contains(err, "focus a prompt") {
		t.Errorf("ctrl+g on the name field: err %q", err)
	}

	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText("draft")
	send(tea.KeyMsg{Type: tea.KeyCtrlX})
	if !m.(model).ctrlX || m.(model).err != ctrlXPending {
		t.Fatalf("ctrl+x should arm the chord")
	}
	typeText("!")
	if got := m.(model).chatPrompt.Value(); got != "draft!" {
		t.Errorf("a key after ctrl+x should pass through: %q", got)
	}
	if m.(model).err != "" {
		t.Errorf("chord hint should clear: %q", m.(model).err)
	}

	mm := m.(model)
	mm.attachments = []string{"/tmp/a.png", "/tmp/b.png"}
	m = mm
	send(tea.KeyMsg{Type: tea.KeyCtrlX})
	send(tea.KeyMsg{Type: tea.KeyBackspace})
	if got := m.(model).attachments; len(got) != 1 || got[0] != "/tmp/a.png" {
		t.Errorf("ctrl+x backspace should drop the last image: %v", got)
	}
	if got := m.(model).chatPrompt.Value(); got != "draft!" {
		t.Errorf("ctrl+x backspace should not edit the prompt: %q", got)
	}

	for _, keys := range [][]tea.KeyType{{tea.KeyCtrlG}, {tea.KeyCtrlX, tea.KeyCtrlE}} {
		for _, k := range keys {
			send(tea.KeyMsg{Type: k})
		}
		if last == nil {
			t.Errorf("%v on the prompt should start the editor", keys)
		}
	}

	send(editorDoneMsg{view: viewNew, text: "from vim"})
	if got := m.(model).chatPrompt.Value(); got != "from vim" {
		t.Errorf("editor result not loaded: %q", got)
	}
	send(editorDoneMsg{view: viewReview, text: "stale"})
	if got := m.(model).chatPrompt.Value(); got != "from vim" {
		t.Errorf("a result for another view should be ignored: %q", got)
	}
}

func TestKittyFreeExecRunsEditor(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "ed")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf ' edited' >> \"$1\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", script)
	if got := editorCommand(); len(got) != 1 || got[0] != script {
		t.Fatalf("editorCommand: %v", got)
	}
	file := filepath.Join(dir, "prompt.md")
	os.WriteFile(file, []byte("draft"), 0o644)
	if err := (kittyFreeExec{exec.Command(script, file)}).Run(); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(file); string(data) != "draft edited" {
		t.Errorf("file after editor: %q", data)
	}
}
