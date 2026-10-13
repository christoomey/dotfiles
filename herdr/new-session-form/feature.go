package main

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Feature view (features › new): paste a branch. One with a Linear id picks
// the project from its team key; any other branch, or a bare feature name
// (which becomes $USER/<name>), needs the project picked by hand. The
// feature name and the first prompt follow the branch as you type (the name
// the way `af up` derives it, the prompt `/linear-review <ID>` when there is
// an id) until you edit them yourself; clearing one hands it back to the
// branch. ctrl+s runs the project's worktree bootstrap (`af up` for august,
// `bin/up` for moi) in a new tab with claude started on the prompt; ctrl+b
// does the same without taking focus. Dotfiles has no feature flow.

var projects = []string{"august", "moi", "dotfiles"}

// featureProjects are the projects with a worktree flow, in picker order.
var featureProjects = projects[:2]

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

// featureBranch is the branch a submit checks out: the input as typed, or a
// bare feature name kebab-cased under the user's prefix.
func featureBranch(input string) (string, error) {
	input = strings.TrimSpace(input)
	switch {
	case input == "":
		return "", errors.New("branch or feature name is required")
	case strings.Contains(input, "/"):
		if strings.ContainsAny(input, " \t") {
			return "", errors.New("a branch cannot contain spaces")
		}
		return input, nil
	}
	name := featureNameFor(input)
	if name == "" {
		return "", errors.New("no usable feature name in " + strconv.Quote(input))
	}
	user := os.Getenv("USER")
	if user == "" {
		user = "christoomey"
	}
	return user + "/" + name, nil
}

func defaultPrompt(branch string) string {
	if id := linearID(branch); id != "" {
		return "/linear-review " + id
	}
	return ""
}

const (
	featBranch = iota
	featProject
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

// featureTarget is the project a submit bootstraps: the branch's team key
// when it has a Linear id, else the one picked by hand ("" until then).
func (m model) featureTarget() (project string, inferred bool) {
	if p, _, ok := detectProject(m.feature.Value()); ok {
		return p, true
	}
	return m.featurePicked, false
}

// focusFeature moves to field i. The project picker only takes focus when
// the branch leaves the project open; otherwise focus steps over it.
func (m *model) focusFeature(i int) tea.Cmd {
	i = (i + featFields) % featFields
	if _, inferred := m.featureTarget(); i == featProject && inferred {
		if m.featureFocus == featName {
			i = featBranch
		} else {
			i = featName
		}
	}
	m.blurFeature()
	m.featureFocus = i
	switch m.featureFocus {
	case featBranch:
		return m.feature.Focus()
	case featProject:
		return nil
	case featName:
		return m.featureName.Focus()
	default:
		return m.featurePrompt.Focus()
	}
}

// pickFeatureProject sets the hand-picked project; i indexes projects and
// wraps over the ones with a feature flow.
func (m model) pickFeatureProject(i int) (tea.Model, tea.Cmd) {
	n := len(featureProjects)
	m.featurePicked = featureProjects[(i%n+n)%n]
	m.err = ""
	return m, nil
}

func (m model) updateFeatureProject(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if i, ok := projectByInitial(k.String()); ok && i < len(featureProjects) {
		return m.pickFeatureProject(i)
	}
	cur := slices.Index(featureProjects, m.featurePicked)
	switch k.String() {
	case "left", "h", "up", "k":
		if cur < 0 {
			cur = 0
		}
		return m.pickFeatureProject(cur - 1)
	case "right", "l", "down", "j", " ":
		return m.pickFeatureProject(cur + 1)
	}
	return m, nil
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
	if strings.TrimSpace(branch) != "" && linearID(branch) == "" {
		m.featurePrompt.Placeholder = "what to work on…"
	} else {
		m.featurePrompt.Placeholder = "/linear-review MEDS-1234"
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
		if m.featureFocus == featProject {
			return m.updateFeatureProject(k)
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
	branch, err := featureBranch(m.feature.Value())
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	project, _ := m.featureTarget()
	if project == "" {
		m.err = "no Linear id in the branch: pick a project (a august · m moi)"
		return m, m.focusFeature(featProject)
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

func featureRunner(project string) string {
	if project == "moi" {
		return "runs bin/up in ~/code/moi"
	}
	return "runs af up"
}

func (m model) featureView() string {
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	accent := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)

	input := strings.TrimSpace(m.feature.Value())
	var hint string
	switch project, key, ok := detectProject(input); {
	case input == "":
		hint = dim.Render("a Linear branch picks the project by team key (toom → moi, else august)")
	case ok:
		hint = dim.Render(key+" → ") + accent.Render(project) + dim.Render("   "+featureRunner(project))
	default:
		hint = dim.Render("no Linear id · pick the project below")
		if branch, err := featureBranch(input); err == nil && branch != input {
			hint += dim.Render(" · branch ") + accent.Render(branch)
		}
	}

	project, inferred := m.featureTarget()
	var picker string
	switch {
	case inferred:
		picker = project + dim.Render("  from the branch")
	case project == "":
		picker = dim.Render("none · a august · m moi")
	default:
		picker = project + dim.Render("  "+featureRunner(project))
	}

	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(m.inputField(m.featureFocus == featBranch, "Branch", dim.Render("Linear branch, any branch, or a feature name"), m.feature.View()) + "\n  " + hint + "\n\n")
	b.WriteString(m.projectPicker(picker, m.featureFocus == featProject) + "\n\n")
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
