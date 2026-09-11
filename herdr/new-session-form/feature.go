package main

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Feature view (ctrl+f): paste a Linear branch, the project is inferred from
// the team key, and the launcher runs that project's worktree bootstrap
// (`af up` for august, `bin/up` for moi) in a new tab. Dotfiles has no
// feature flow, so only branches with a Linear id are accepted.

var projects = []string{"august", "moi", "dotfiles"}

var linearBranchRe = regexp.MustCompile(`^(?:[^/]+/)?([A-Za-z]+)-[0-9]+(?:-|$)`)

// detectProject maps a Linear branch to a project by team key: toom is moi,
// every other team is august.
func detectProject(branch string) (project, key string, ok bool) {
	m := linearBranchRe.FindStringSubmatch(strings.TrimSpace(branch))
	if m == nil {
		return "", "", false
	}
	key = strings.ToLower(m[1])
	if key == "toom" {
		return "moi", key, true
	}
	return "august", key, true
}

func newFeatureInput(s styles) textinput.Model {
	in := textinput.New()
	in.Prompt = "> "
	in.Placeholder = "christoomey/meds-1234-short-description"
	in.Width = 64
	in.PromptStyle = lipgloss.NewStyle().Foreground(s.accent)
	in.PlaceholderStyle = lipgloss.NewStyle().Foreground(s.dim)
	return in
}

func (m *model) enterFeature() tea.Cmd {
	m.view = viewFeature
	m.err = ""
	m.filter.Blur()
	return m.feature.Focus()
}

func (m model) updateFeature(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.feature, cmd = m.feature.Update(msg)
		return m, cmd
	}
	switch k.String() {
	case "esc", "ctrl+c":
		return m, tea.Quit
	case "ctrl+n":
		m.feature.Blur()
		return m, m.leaveResume()
	case "ctrl+r":
		m.feature.Blur()
		return m, m.enterResume()
	case "ctrl+a":
		return m, m.enterAgents()
	case "enter", "ctrl+s":
		branch := strings.TrimSpace(m.feature.Value())
		project, _, ok := detectProject(branch)
		if !ok {
			m.err = "need a Linear branch like user/team-123-description"
			return m, nil
		}
		m.branch = branch
		m.featureProject = project
		m.mode = "feature"
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.feature, cmd = m.feature.Update(msg)
	m.err = ""
	return m, cmd
}

func (m model) featureView() string {
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(m.styles.theme.Focused.Title.Render("Linear branch") + "\n")
	b.WriteString(m.feature.View() + "\n\n")

	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	accent := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)
	branch := strings.TrimSpace(m.feature.Value())
	switch project, key, ok := detectProject(branch); {
	case branch == "":
		b.WriteString(dim.Render("  paste a branch; the team key picks the project (toom → moi, else august)") + "\n")
	case !ok:
		b.WriteString(dim.Render("  no Linear id found yet") + "\n")
	case project == "moi":
		b.WriteString(dim.Render("  "+key+" → ") + accent.Render("moi") + dim.Render("   runs bin/up in ~/code/moi") + "\n")
	default:
		b.WriteString(dim.Render("  "+key+" → ") + accent.Render("august") + dim.Render("   runs af up") + "\n")
	}

	help := "enter start feature · ctrl+n new session · ctrl+r resume · esc cancel"
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}
