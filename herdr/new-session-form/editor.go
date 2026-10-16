package main

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// External editor: ctrl+g (as in Claude Code) or ctrl+x ctrl+e (as in the
// shell) hands the focused prompt to $VISUAL/$EDITOR and loads the file back
// when the editor exits.
const (
	editorKey    = "ctrl+g"
	ctrlXPending = "ctrl+x… (ctrl+e editor · backspace drop image)"
)

type editorDoneMsg struct {
	view view
	text string
	err  error
}

// promptArea is the focused multi-line prompt, if any.
func (m *model) promptArea() *textarea.Model {
	switch {
	case m.view == viewNew && m.chatFocus == chatPromptField:
		return &m.chatPrompt
	case m.view == viewFeature && m.featureFocus == featPrompt:
		return &m.featurePrompt
	case m.view == viewReview && m.reviewFocus == revPrompt:
		return &m.reviewPrompt
	case m.view == viewDebug && m.debugFocus == dbgPrompt:
		return &m.debugPrompt
	}
	return nil
}

func editorCommand() []string {
	for _, v := range []string{os.Getenv("VISUAL"), os.Getenv("EDITOR")} {
		if f := strings.Fields(v); len(f) > 0 {
			return f
		}
	}
	if _, err := exec.LookPath("nvim"); err == nil {
		return []string{"nvim"}
	}
	return []string{"vi"}
}

func (m model) openEditor() (tea.Model, tea.Cmd) {
	ta := m.promptArea()
	if ta == nil {
		m.err = "focus a prompt to edit it in " + editorCommand()[0]
		return m, nil
	}
	f, err := os.CreateTemp("", "herdr-prompt-*.md")
	if err != nil {
		m.err = "editor: " + err.Error()
		return m, nil
	}
	path := f.Name()
	_, err = f.WriteString(ta.Value())
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(path)
		m.err = "editor: " + err.Error()
		return m, nil
	}
	argv := append(editorCommand(), path)
	v := m.view
	return m, tea.Exec(kittyFreeExec{exec.Command(argv[0], argv[1:]...)}, func(err error) tea.Msg {
		defer os.Remove(path)
		if err != nil {
			return editorDoneMsg{view: v, err: err}
		}
		data, rerr := os.ReadFile(path)
		return editorDoneMsg{view: v, text: strings.TrimRight(string(data), "\n"), err: rerr}
	})
}

func (m model) onEditorDone(msg editorDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = "editor: " + msg.err.Error()
		return m, nil
	}
	ta := m.promptArea()
	if ta == nil || m.view != msg.view {
		return m, nil
	}
	ta.SetValue(msg.text)
	if m.view == viewFeature {
		m.promptEdited = msg.text != ""
		m.syncFeature()
	}
	m.err = ""
	return m, nil
}

// kittyFreeExec pops the form's kitty keyboard flags for the editor's run, so
// the editor starts on the terminal's legacy encoding, and pushes them back
// after.
type kittyFreeExec struct{ *exec.Cmd }

func (c kittyFreeExec) Run() error {
	os.Stderr.WriteString(kittyPop)
	defer os.Stderr.WriteString(kittyPush)
	return c.Cmd.Run()
}

func (c kittyFreeExec) SetStdin(r io.Reader) {
	if c.Stdin == nil {
		c.Stdin = r
	}
}

func (c kittyFreeExec) SetStdout(w io.Writer) {
	if c.Stdout == nil {
		c.Stdout = w
	}
}

func (c kittyFreeExec) SetStderr(w io.Writer) {
	if c.Stderr == nil {
		c.Stderr = w
	}
}
