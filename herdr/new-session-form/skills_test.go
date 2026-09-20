package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSlashToken(t *testing.T) {
	for _, c := range []struct {
		text  string
		at    int
		token string
		ok    bool
	}{
		{"/ver", 4, "/ver", true},
		{"run /ver", 8, "/ver", true},
		{"/", 1, "/", true},
		{"/linear-review", 14, "/linear-review", true},
		{"/august:docs", 12, "/august:docs", true},
		{"a/b", 3, "", false},
		{"http://x", 8, "", false},
		{"/a/b", 4, "", false},
		{"/ver ", 5, "", false},
		{"", 0, "", false},
		{"go /ver now", 7, "/ver", true},  // cursor mid-text
		{"go /ver now", 6, "/ve", true},   // mid-token
		{"look\n/ver", 9, "/ver", true},   // a later line
		{"look\n/ver x", 9, "/ver", true}, // cursor at the end of the token
	} {
		token, ok := slashToken(c.text, c.at)
		if token != c.token || ok != c.ok {
			t.Errorf("%q@%d: %q %v, want %q %v", c.text, c.at, token, ok, c.token, c.ok)
		}
	}
}

func TestLoadSkills(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	write := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(home, ".claude", "commands", "wip.md"), "---\ndescription: Summarize in-flight work\n---\n")
	write(filepath.Join(home, ".claude", "skills", "verify", "SKILL.md"),
		"---\nname: verify\ndescription: >-\n  Mint a dev token and drive the API\n  over several lines.\n---\n")
	if err := os.MkdirAll(filepath.Join(home, ".claude", "skills", "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(home, "code", "dotfiles", ".claude", "skills", "wip", "SKILL.md"),
		"---\ndescription: \"the dotfiles one\"\n---\n")

	user := loadSkills("")
	if len(user) != 2 || user[0].name != "verify" || user[1].name != "wip" {
		t.Fatalf("user skills: %+v", user)
	}
	if user[0].desc != "Mint a dev token and drive the API" || user[0].source != "user" {
		t.Errorf("block-scalar description: %+v", user[0])
	}

	got := loadSkills("dotfiles")
	if len(got) != 2 || got[1].name != "wip" || got[1].source != "dotfiles" || got[1].desc != "the dotfiles one" {
		t.Errorf("the project skill should win the name clash: %+v", got)
	}
}

// fixtureSkills plants a user skill the completion tests can complete to.
func fixtureSkills(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".claude", "skills", "verify")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\ndescription: Mint a token\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSlashCompletionInTheChatPrompt(t *testing.T) {
	fixtureSkills(t)
	threadsTemp(t)
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "/ve")
	if rows := get().menuRows(); len(rows) != 1 || rows[0].name != "verify" {
		t.Fatalf("menu rows: %+v", rows)
	}

	send(tea.KeyMsg{Type: tea.KeyTab})
	m := get()
	if m.chatPrompt.Value() != "/verify " {
		t.Errorf("tab did not complete: %q", m.chatPrompt.Value())
	}
	if m.chatFocus != chatPromptField {
		t.Errorf("focus left the prompt: %d", m.chatFocus)
	}
	// With the menu gone, tab is next-field again.
	send(tea.KeyMsg{Type: tea.KeyTab})
	if got := get().chatFocus; got != chatStartField {
		t.Errorf("tab without a menu: focus %d, want start", got)
	}
}

// esc closes the menu and nothing else; the next one is the ordinary
// unsent-prompt warning.
func TestEscClosesTheMenuOnly(t *testing.T) {
	fixtureSkills(t)
	threadsTemp(t)
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "/ve")

	next, cmd := get().Update(tea.KeyMsg{Type: tea.KeyEscape})
	m := next.(model)
	if cmd != nil || m.mode != "" {
		t.Fatal("esc over the menu quit the popup")
	}
	if len(m.menuRows()) != 0 {
		t.Error("esc did not close the menu")
	}
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil || next.(model).err == "" {
		t.Error("the second esc should warn about the unsent prompt")
	}
}

func TestEnterCompletesAndCtrlJKMoveTheMenu(t *testing.T) {
	fixtureSkills(t)
	cmds := filepath.Join(os.Getenv("HOME"), ".claude", "commands")
	if err := os.MkdirAll(cmds, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmds, "grill-me.md"), []byte("---\ndescription: Grill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	threadsTemp(t)
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "/")
	rows := get().menuRows()
	if len(rows) < 2 {
		t.Fatalf("want at least two rows, got %+v", rows)
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if got := get().menuIdx; got != 1 {
		t.Fatalf("ctrl+j: idx %d, want 1", got)
	}
	send(tea.KeyMsg{Type: tea.KeyCtrlK})
	if got := get().menuIdx; got != 0 {
		t.Fatalf("ctrl+k: idx %d, want 0", got)
	}
	send(tea.KeyMsg{Type: tea.KeyCtrlJ})
	send(tea.KeyMsg{Type: tea.KeyEnter})
	if got, want := get().chatPrompt.Value(), "/"+rows[1].name+" "; got != want {
		t.Errorf("enter over the menu: %q, want %q", got, want)
	}
	send(tea.KeyMsg{Type: tea.KeyEnter})
	if got := get().chatPrompt.Value(); !strings.HasSuffix(got, "\n") {
		t.Errorf("enter without a menu should be a newline: %q", got)
	}
}

func TestASlashMidWordOpensNoMenu(t *testing.T) {
	fixtureSkills(t)
	threadsTemp(t)
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "a/b")
	if rows := get().menuRows(); len(rows) != 0 {
		t.Errorf("a/b opened a menu: %+v", rows)
	}
}

// Completing in the feature prompt counts as taking the prompt over from
// the branch, the same as typing does.
func TestSlashCompletionInTheFeaturePrompt(t *testing.T) {
	fixtureSkills(t)
	threadsTemp(t)
	get, send := newFormModel(t)
	send(tea.KeyMsg{Type: tea.KeyCtrlF})
	typeText(send, "christoomey/toom-9-fix-thing")
	send(tea.KeyMsg{Type: tea.KeyTab})
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, " /ve")
	send(tea.KeyMsg{Type: tea.KeyTab})

	m := get()
	if m.featurePrompt.Value() != "/linear-review TOOM-9 /verify " || !m.promptEdited {
		t.Errorf("feature completion: %q edited %v", m.featurePrompt.Value(), m.promptEdited)
	}
}

func TestStartFieldSubmits(t *testing.T) {
	threadsTemp(t)
	for _, c := range []struct {
		toggle bool
		want   string
	}{{false, "open"}, {true, "bg"}} {
		get, send := newFormModel(t)
		typeText(send, "dose bug")
		send(tea.KeyMsg{Type: tea.KeyTab})
		send(tea.KeyMsg{Type: tea.KeyTab})
		if c.toggle {
			send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
		}
		send(tea.KeyMsg{Type: tea.KeyEnter})
		if got := get().mode; got != c.want {
			t.Errorf("enter on start: mode %q, want %q", got, c.want)
		}
	}
}

func TestSubmitWithoutANameAsksForOne(t *testing.T) {
	threadsTemp(t)
	get, send := newFormModel(t)
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "go look into it")
	send(tea.KeyMsg{Type: tea.KeyCtrlS})

	m := get()
	if m.mode != "" || m.err != "name is required" {
		t.Fatalf("submit: mode %q err %q", m.mode, m.err)
	}
	if m.chatFocus != chatNameField {
		t.Errorf("focus %d, want the name field", m.chatFocus)
	}
}

// The chat fields are boxed like the feature ones, not barred.
func TestChatViewDrawsBoxedFields(t *testing.T) {
	get, _ := newFormModel(t)
	view := get().View()
	if !strings.Contains(view, "╭") {
		t.Error("no rounded box in the chat view")
	}
	if strings.Contains(view, "┃") {
		t.Error("the chat view still draws a focus bar")
	}
}
