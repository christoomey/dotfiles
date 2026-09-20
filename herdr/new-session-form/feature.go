package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Feature view (features › new): paste a Linear branch and the project is
// inferred from the team key. The feature name and the first prompt follow
// the branch as you type (the name the way `af up` derives it, the prompt
// `/linear-review <ID>`) until you edit them yourself; clearing one hands
// it back to the branch. ctrl+s runs the project's worktree bootstrap
// (`af up` for august, `bin/up` for moi) in a new tab with claude started
// on the prompt; ctrl+b does the same without taking focus. Dotfiles has
// no feature flow, so only branches with a Linear id are accepted.

var projects = []string{"august", "moi", "dotfiles"}

const (
	projectMoi      = 1
	projectDotfiles = 2
)

// projectByInitial maps a typed letter to a project while a project picker
// has focus: a → august, m → moi, d → dotfiles.
func projectByInitial(key string) (int, bool) {
	for i, p := range projects {
		if key == p[:1] {
			return i, true
		}
	}
	return 0, false
}

var (
	linearBranchRe = regexp.MustCompile(`^(?:[^/]+/)?([A-Za-z]+)-([0-9]+)(?:-|$)`)
	userPrefixRe   = regexp.MustCompile(`^[^/]+/`)
	linearPrefixRe = regexp.MustCompile(`^[A-Za-z]+-[0-9]+-?`)
	nonKebabRe     = regexp.MustCompile(`[^a-z0-9]+`)
)

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

// linearID is the issue id in a branch, upper-cased: MEDS-3241, TOOM-12.
func linearID(branch string) string {
	m := linearBranchRe.FindStringSubmatch(strings.TrimSpace(branch))
	if m == nil {
		return ""
	}
	return strings.ToUpper(m[1]) + "-" + m[2]
}

// featureNameFor derives the worktree name the way `af up` and moi's
// `bin/up` do: drop the user prefix and the team-id prefix, kebab-case the
// rest. christoomey/meds-3241-expose-data → expose-data.
func featureNameFor(branch string) string {
	name := strings.TrimSpace(branch)
	name = userPrefixRe.ReplaceAllString(name, "")
	name = linearPrefixRe.ReplaceAllString(name, "")
	name = nonKebabRe.ReplaceAllString(strings.ToLower(name), "-")
	return strings.Trim(name, "-")
}

func defaultPrompt(branch string) string {
	if id := linearID(branch); id != "" {
		return "/linear-review " + id
	}
	return ""
}

const (
	featBranch = iota
	featName
	featPrompt
	featFields
)

// newInput and newPromptArea are the one-line and multi-line fields both the
// feature and the chat views are built from.
func newInput(s styles, placeholder string) textinput.Model {
	in := textinput.New()
	in.Prompt = "> "
	in.Placeholder = placeholder
	in.Width = 70
	in.PromptStyle = lipgloss.NewStyle().Foreground(s.accent)
	in.PlaceholderStyle = lipgloss.NewStyle().Foreground(s.dim)
	in.TextStyle = lipgloss.NewStyle().Foreground(s.text)
	return in
}

func newPromptArea(s styles, placeholder string, height int) textarea.Model {
	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.Placeholder = placeholder
	ta.CharLimit = 0
	ta.SetWidth(72)
	ta.SetHeight(height)
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("enter", "ctrl+j"))
	for _, st := range []*textarea.Style{&ta.FocusedStyle, &ta.BlurredStyle} {
		st.Base = lipgloss.NewStyle()
		st.CursorLine = lipgloss.NewStyle()
		st.Placeholder = lipgloss.NewStyle().Foreground(s.dim)
		st.Text = lipgloss.NewStyle().Foreground(s.text)
		st.EndOfBuffer = lipgloss.NewStyle()
	}
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(s.accent)
	return ta
}

func (m *model) initFeature(s styles) {
	m.feature = newInput(s, "christoomey/meds-1234-short-description")
	m.featureName = newInput(s, "short-description")
	m.featurePrompt = newPromptArea(s, "/linear-review MEDS-1234", 3)
}

// setFeature fills the three inputs from a draft or a job. SetValue keeps the
// old cursor position, so the inputs are moved to the end as well.
func (m *model) setFeature(branch, name, prompt string) {
	m.feature.SetValue(branch)
	m.feature.CursorEnd()
	m.featureName.SetValue(name)
	m.featureName.CursorEnd()
	m.featurePrompt.SetValue(prompt)
}

func (m *model) focusFeature(i int) tea.Cmd {
	m.blurFeature()
	m.featureFocus = (i + featFields) % featFields
	switch m.featureFocus {
	case featBranch:
		return m.feature.Focus()
	case featName:
		return m.featureName.Focus()
	default:
		return m.featurePrompt.Focus()
	}
}

func (m *model) blurFeature() {
	m.feature.Blur()
	m.featureName.Blur()
	m.featurePrompt.Blur()
}

// syncFeature re-derives the name and prompt from the branch for whichever
// of them the user has not taken over.
func (m *model) syncFeature() {
	branch := m.feature.Value()
	if !m.nameEdited {
		// SetValue keeps the old cursor position; typing would land mid-word.
		m.featureName.SetValue(featureNameFor(branch))
		m.featureName.CursorEnd()
	}
	if !m.promptEdited {
		m.featurePrompt.SetValue(defaultPrompt(branch))
	}
}

func (m model) updateFeature(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "tab":
			return m, m.focusFeature(m.featureFocus + 1)
		case "shift+tab":
			return m, m.focusFeature(m.featureFocus - 1)
		case "enter":
			if m.featureFocus != featPrompt {
				return m, m.focusFeature(m.featureFocus + 1)
			}
		case "ctrl+s", "ctrl+b":
			return m.submitFeature(k.String() == "ctrl+b")
		}
	}
	var cmd tea.Cmd
	switch m.featureFocus {
	case featBranch:
		m.feature, cmd = m.feature.Update(msg)
		m.syncFeature()
	case featName:
		before := m.featureName.Value()
		m.featureName, cmd = m.featureName.Update(msg)
		if after := m.featureName.Value(); after != before {
			m.nameEdited = after != ""
			m.syncFeature()
		}
	default:
		before := m.featurePrompt.Value()
		m.featurePrompt, cmd = m.featurePrompt.Update(msg)
		if after := m.featurePrompt.Value(); after != before {
			m.promptEdited = after != ""
			m.syncFeature()
		}
	}
	m.err = ""
	return m, cmd
}

func (m model) submitFeature(background bool) (tea.Model, tea.Cmd) {
	branch := strings.TrimSpace(m.feature.Value())
	project, _, ok := detectProject(branch)
	if !ok {
		m.err = "need a Linear branch like user/team-123-description"
		return m, nil
	}
	name := strings.TrimSpace(m.featureName.Value())
	if name == "" {
		m.err = "feature name is required"
		return m, nil
	}
	// The bootstrap would fail later in a tab nobody is looking at.
	if dir, err := projectWorktreeRoot(project); err == nil {
		wt := filepath.Join(dir, name)
		if _, err := os.Stat(wt); err == nil {
			m.err = "worktree exists already: " + wt
			return m, nil
		}
	}
	m.branch = branch
	m.featureProject = project
	m.picked.Title = name
	mode := "feature"
	if background {
		mode = "feature-bg"
	}
	return m.submit(mode)
}

func (m model) featureView() string {
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	accent := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)

	branch := strings.TrimSpace(m.feature.Value())
	var hint string
	switch project, key, ok := detectProject(branch); {
	case branch == "":
		hint = dim.Render("paste a branch; the team key picks the project (toom → moi, else august)")
	case !ok:
		hint = dim.Render("no Linear id found yet")
	case project == "moi":
		hint = dim.Render(key+" → ") + accent.Render("moi") + dim.Render("   runs bin/up in ~/code/moi")
	default:
		hint = dim.Render(key+" → ") + accent.Render("august") + dim.Render("   runs af up")
	}

	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(m.inputField(m.featureFocus == featBranch, "Linear branch", "", m.feature.View()) + "\n  " + hint + "\n\n")
	b.WriteString(m.inputField(m.featureFocus == featName, "Feature name", "", m.featureName.View()) + "\n\n")
	b.WriteString(m.inputField(m.featureFocus == featPrompt, "Initial prompt", dim.Render("optional · enter for a new line"), m.featurePrompt.View()) + "\n")
	rows := m.menuRows()
	b.WriteString(m.menuView(rows))
	b.WriteString(m.noticesView())

	help := "ctrl+s start feature · ctrl+b in background · tab next field · " + tabsHelp + m.strandedHelp() + " · esc cancel"
	if len(rows) > 0 {
		help = menuHelp(m.menuProject())
	}
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}
