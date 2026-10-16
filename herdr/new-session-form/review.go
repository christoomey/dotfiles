package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Review view: one optional instructions box over /review-queue, which
// reviews every unreviewed PR on the Perspective board from a session in
// august. It runs in the background by default; ctrl+s opens it instead.
// The result is a plain chats › new launch, so the launcher needs nothing
// of its own for it.

const (
	revPrompt = iota
	revStart
	revFields
)

func (m *model) initReview(s styles) {
	m.reviewPrompt = newPromptArea(s, "just Nick's one in the Now list", 6)
}

func reviewName(now time.Time) string {
	return "Review queue " + now.Format("Jan 2 15:04")
}

func reviewPromptFor(instructions string) string {
	if s := strings.TrimSpace(instructions); s != "" {
		return "/review-queue " + s
	}
	return "/review-queue"
}

func (m *model) focusReview(i int) tea.Cmd {
	m.blurReview()
	m.reviewFocus = (i + revFields) % revFields
	if m.reviewFocus == revPrompt {
		return m.reviewPrompt.Focus()
	}
	return nil
}

func (m *model) blurReview() {
	m.reviewPrompt.Blur()
}

func (m model) updateReview(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m.forwardReview(msg)
	}
	switch k.String() {
	case "ctrl+s":
		return m.submitReview("open")
	}
	if m.reviewFocus == revStart {
		switch k.String() {
		case "tab", "left", "right", "h", "l", " ":
			m.reviewOpen = !m.reviewOpen
		case "shift+tab":
			return m, m.focusReview(revPrompt)
		case "enter":
			if m.reviewOpen {
				return m.submitReview("open")
			}
			return m.submitReview("bg")
		}
		return m, nil
	}
	switch k.String() {
	case "tab", "shift+tab":
		return m, m.focusReview(revStart)
	}
	m.err = ""
	return m.forwardReview(msg)
}

// submitReview fixes the session name now, so the job spec and the printed
// result cannot straddle a minute.
func (m model) submitReview(mode string) (tea.Model, tea.Cmd) {
	m.picked.Title = reviewName(time.Now())
	return m.submit(mode)
}

func (m model) forwardReview(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.reviewFocus != revPrompt {
		return m, nil
	}
	var cmd tea.Cmd
	m.reviewPrompt, cmd = m.reviewPrompt.Update(msg)
	return m, cmd
}

func (m model) reviewView() string {
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(dim.Width(76).Render("Runs /review-queue in august: reviews every unreviewed PR on the Perspective board, links each session and report to its action, then closes itself.") + "\n\n")
	b.WriteString(m.inputField(m.reviewFocus == revPrompt, "Instructions", dim.Render("optional · which PRs, or how · enter for a new line"), m.reviewPrompt.View()) + "\n\n")
	b.WriteString(m.startField(m.reviewFocus == revStart, m.reviewOpen) + "\n")
	b.WriteString(m.noticesView())

	help := "ctrl+enter in background · ctrl+g editor · ctrl+s open now · tab next field · " + tabsHelp + m.strandedHelp() + " · esc cancel"
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}
