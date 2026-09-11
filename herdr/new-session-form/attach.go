package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Attachments: ctrl+v in the new-session view dumps the clipboard image (what
// CleanShot leaves there after a capture) to a temp file and lists it under the
// prompt. The launcher moves the files into the session's scratch dir and
// appends their paths to the first prompt; Claude reads images by path.

type attachMsg struct {
	path string
	err  error
}

var errNoImage = errors.New("no image on clipboard")

func (m *model) attachClipboardCmd() tea.Cmd {
	if m.attachDir == "" {
		dir, err := os.MkdirTemp("", "herdr-attach-")
		if err != nil {
			return func() tea.Msg { return attachMsg{err: err} }
		}
		m.attachDir = dir
	}
	path := filepath.Join(m.attachDir, fmt.Sprintf("shot-%d.png", len(m.attachments)+1))
	return func() tea.Msg {
		script := strings.Join([]string{
			`set png to the clipboard as «class PNGf»`,
			`set f to open for access POSIX file "` + path + `" with write permission`,
			`set eof f to 0`,
			`write png to f`,
			`close access f`,
		}, "\n")
		out, err := exec.Command("osascript", "-e", script).CombinedOutput()
		if err != nil {
			if strings.Contains(string(out), "-1700") { // can't coerce clipboard to PNGf
				return attachMsg{err: errNoImage}
			}
			return attachMsg{err: fmt.Errorf("clipboard: %s", strings.TrimSpace(string(out)))}
		}
		return attachMsg{path: path}
	}
}

func (m *model) dropLastAttachment() {
	if n := len(m.attachments); n > 0 {
		os.Remove(m.attachments[n-1])
		m.attachments = m.attachments[:n-1]
	}
}

func (m *model) cleanupAttachments() {
	if m.attachDir != "" {
		os.RemoveAll(m.attachDir)
	}
}

func (m model) attachmentsView() string {
	if len(m.attachments) == 0 {
		return ""
	}
	names := make([]string, len(m.attachments))
	for i, p := range m.attachments {
		names[i] = filepath.Base(p)
	}
	return m.styles.attach.Render("📎 " + strings.Join(names, " · "))
}
