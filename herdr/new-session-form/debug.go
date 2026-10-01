package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Debug view: an optional title over a required prompt that starts an
// Abulafia session (pi on
// Bedrock, read-only AWS) in august, in front (ctrl+s) or in the background
// (ctrl+b). The result is a chats › new launch with kind "debug", which the
// launcher routes to launch_debug; that notifies when a background session
// is waiting on an MFA code.

const (
	dbgTitle = iota
	dbgPrompt
	dbgFields
)

func (m *model) initDebug(s styles) {
	m.debugTitle = newInput(s, "nightly export failures")
	m.debugPrompt = newPromptArea(s, "why did the nightly export job fail last night?", 4)
}

// debugName is the first few words of the prompt's first line, else a
// dated fallback.
func debugName(prompt string, now time.Time) string {
	for _, line := range strings.Split(prompt, "\n") {
		words := strings.Fields(nonSlugRe.ReplaceAllString(strings.ToLower(line), " "))
		if len(words) == 0 {
			continue
		}
		name := strings.Join(words[:min(5, len(words))], " ")
		if len(name) > 40 {
			name = strings.TrimSpace(name[:40])
		}
		return name
	}
	return "debug " + now.Format("Jan 2 15:04")
}

func (m *model) focusDebug() tea.Cmd {
	m.blurDebug()
	if m.debugFocus == dbgTitle {
		return m.debugTitle.Focus()
	}
	return m.debugPrompt.Focus()
}

func (m *model) blurDebug() {
	m.debugTitle.Blur()
	m.debugPrompt.Blur()
}

func (m model) updateDebug(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+s":
			return m.submitDebug("open")
		case "ctrl+b":
			return m.submitDebug("bg")
		case "tab", "shift+tab":
			m.debugFocus = (m.debugFocus + 1) % dbgFields
			return m, m.focusDebug()
		case "enter":
			if m.debugFocus == dbgTitle {
				m.debugFocus = dbgPrompt
				return m, m.focusDebug()
			}
		}
		m.err = ""
	}
	var cmd tea.Cmd
	if m.debugFocus == dbgTitle {
		m.debugTitle, cmd = m.debugTitle.Update(msg)
	} else {
		m.debugPrompt, cmd = m.debugPrompt.Update(msg)
	}
	return m, cmd
}

// submitDebug pins the name now, like submitReview: the title, else the
// prompt's first words.
func (m model) submitDebug(mode string) (tea.Model, tea.Cmd) {
	prompt := strings.TrimSpace(m.debugPrompt.Value())
	if prompt == "" {
		m.err = "prompt is required"
		return m, nil
	}
	m.picked.Title = strings.TrimSpace(m.debugTitle.Value())
	if m.picked.Title == "" {
		m.picked.Title = debugName(prompt, time.Now())
	}
	return m.submit(mode)
}

func (m model) debugView() string {
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(dim.Width(76).Render("Starts an Abulafia read-only debugging session (pi on Bedrock) in august. If it needs an MFA code, a background launch notifies you.") + "\n\n")
	b.WriteString(m.inputField(m.debugFocus == dbgTitle, "Title", dim.Render("optional · defaults to the prompt's first words"), m.debugTitle.View()) + "\n\n")
	b.WriteString(m.inputField(m.debugFocus == dbgPrompt, "First prompt", dim.Render("required · enter for a new line"), m.debugPrompt.View()) + "\n")
	b.WriteString(m.noticesView())

	help := "ctrl+b in background · ctrl+s open now · tab next field · " + tabsHelp + m.strandedHelp() + " · esc cancel"
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}
