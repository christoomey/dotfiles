package main

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Slash completion: typing /word in either prompt opens a menu of the skills
// and slash commands Claude would see there: the project's own .claude dir
// first, then ~/.claude. Tab or enter completes the selection; without a menu
// enter is a newline again.

type skill struct{ name, desc, source string }

// loadSkills reads the two dirs Claude reads, project before user, and keeps
// the first of each name. An empty project means user skills only.
func loadSkills(project string) []skill {
	type source struct{ dir, label string }
	var dirs []source
	if project != "" {
		if root, err := projectRoot(project); err == nil {
			dirs = append(dirs, source{filepath.Join(root, ".claude"), project})
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, source{filepath.Join(home, ".claude"), "user"})
	}

	seen := map[string]bool{}
	var out []skill
	add := func(name, path, label string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, skill{name: name, desc: frontmatterDesc(path), source: label})
	}
	for _, d := range dirs {
		cmds, _ := filepath.Glob(filepath.Join(d.dir, "commands", "*.md"))
		for _, p := range cmds {
			add(strings.TrimSuffix(filepath.Base(p), ".md"), p, d.label)
		}
		skills, _ := filepath.Glob(filepath.Join(d.dir, "skills", "*", "SKILL.md"))
		for _, p := range skills {
			add(filepath.Base(filepath.Dir(p)), p, d.label)
		}
	}
	slices.SortFunc(out, func(a, b skill) int { return strings.Compare(a.name, b.name) })
	return out
}

// frontmatterDesc is the description: line of a skill's YAML frontmatter. A
// block scalar (description: >-) keeps its text on the indented lines below.
func frontmatterDesc(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i > 0 && strings.HasPrefix(line, "---") {
			break
		}
		v, ok := strings.CutPrefix(line, "description:")
		if !ok {
			continue
		}
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		switch v {
		case "", ">", "|", ">-", "|-":
			for _, next := range lines[i+1:] {
				if strings.TrimSpace(next) == "" {
					continue
				}
				if !strings.HasPrefix(next, " ") && !strings.HasPrefix(next, "\t") {
					return ""
				}
				return strings.TrimSpace(next)
			}
			return ""
		}
		return v
	}
	return ""
}

// matchSkills ranks a token (without its slash): prefix matches in name
// order, then the rest of the names it appears anywhere in.
func matchSkills(all []skill, token string) []skill {
	token = strings.ToLower(token)
	var prefix, sub []skill
	for _, s := range all {
		name := strings.ToLower(s.name)
		switch {
		case strings.HasPrefix(name, token):
			prefix = append(prefix, s)
		case token != "" && strings.Contains(name, token):
			sub = append(sub, s)
		}
	}
	return append(prefix, sub...)
}

var slashRestRe = regexp.MustCompile(`^[A-Za-z0-9:_-]*$`)

// slashToken is the /word immediately before cursor (a rune offset), if the
// run of non-whitespace back to the last space is one: "run /ver" yes,
// "http://x" and "a/b" no.
func slashToken(text string, cursor int) (string, bool) {
	runes := []rune(text)
	if cursor < 0 || cursor > len(runes) {
		return "", false
	}
	start := cursor
	for start > 0 && !isSpace(runes[start-1]) {
		start--
	}
	token := string(runes[start:cursor])
	if !strings.HasPrefix(token, "/") || !slashRestRe.MatchString(token[1:]) {
		return "", false
	}
	return token, true
}

func isSpace(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }

// ── the menu on the model ───────────────────────────────────────────────────

// menuHeight is how many matches the menu shows at once, so it can never
// push the popup past its height.
const menuHeight = 3

// skillsFor caches a project's skills on the model. The scan is ~30 small
// files, so the first menu can load them in line.
func (m model) skillsFor(project string) []skill {
	if s, ok := m.skillCache[project]; ok {
		return s
	}
	s := loadSkills(project)
	m.skillCache[project] = s
	return s
}

// menuProject is whose .claude dir the current view completes from; "" is
// user skills only.
func (m model) menuProject() string {
	if m.view == viewNew {
		return projects[m.chatProject]
	}
	project, _, ok := detectProject(m.feature.Value())
	if !ok {
		return ""
	}
	return project
}

// promptToken is the /word under the cursor of whichever prompt has focus.
func (m model) promptToken() (string, bool) {
	var ta textarea.Model
	switch {
	case m.view == viewNew && m.chatFocus == chatPromptField:
		ta = m.chatPrompt
	case m.view == viewFeature && m.featureFocus == featPrompt:
		ta = m.featurePrompt
	default:
		return "", false
	}
	return slashToken(ta.Value(), taCursor(ta))
}

// menuRows are the matches the menu shows right now; none means no menu.
func (m model) menuRows() []skill {
	token, ok := m.promptToken()
	if !ok || m.menuDismissed {
		return nil
	}
	return matchSkills(m.skillsFor(m.menuProject()), strings.TrimPrefix(token, "/"))
}

// syncMenu reopens the menu whenever the token under the cursor changes, so
// an esc only dismisses the one it was typed for.
func (m *model) syncMenu() {
	token, _ := m.promptToken()
	if token != m.menuToken {
		m.menuToken = token
		m.menuDismissed = false
		m.menuIdx = 0
	}
	if rows := m.menuRows(); m.menuIdx >= len(rows) {
		m.menuIdx = max(0, len(rows)-1)
	}
}

// menuKey handles the keys the menu takes over while it is up.
func (m model) menuKey(k tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	rows := m.menuRows()
	if len(rows) == 0 {
		return m, nil, false
	}
	switch k.String() {
	case "esc":
		m.menuDismissed = true
		return m, nil, true
	case "tab", "enter":
		return m.completeSkill(rows[min(m.menuIdx, len(rows)-1)].name)
	case "up", "ctrl+p", "ctrl+k":
		m.menuIdx = max(0, m.menuIdx-1)
		return m, nil, true
	case "down", "ctrl+n", "ctrl+j":
		m.menuIdx = min(m.menuIdx+1, len(rows)-1)
		return m, nil, true
	}
	return m, nil, false
}

// completeSkill swaps the typed token for the full name, by backspacing over
// it so the textarea keeps its own cursor bookkeeping.
func (m model) completeSkill(name string) (tea.Model, tea.Cmd, bool) {
	token, _ := m.promptToken()
	backspace := tea.KeyMsg{Type: tea.KeyBackspace}
	if m.view == viewFeature {
		for range []rune(token) {
			m.featurePrompt, _ = m.featurePrompt.Update(backspace)
		}
		m.featurePrompt.InsertString("/" + name + " ")
		m.promptEdited = true
		m.syncFeature()
	} else {
		for range []rune(token) {
			m.chatPrompt, _ = m.chatPrompt.Update(backspace)
		}
		m.chatPrompt.InsertString("/" + name + " ")
	}
	return m, nil, true
}

func menuHelp(project string) string {
	sources := "user skills"
	if project != "" {
		sources = project + " + user skills"
	}
	return "tab/enter complete · ↑/↓ choose · esc close · " + sources
}

// menuView draws the matches under the prompt box, windowed so the selection
// stays in sight.
func (m model) menuView(rows []skill) string {
	if len(rows) == 0 {
		return ""
	}
	const nameWidth, descWidth = 24, 48
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	selected := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)
	plain := lipgloss.NewStyle().Foreground(m.styles.text)

	idx := min(m.menuIdx, len(rows)-1)
	start := 0
	if idx >= menuHeight {
		start = idx - menuHeight + 1
	}
	var b strings.Builder
	for i := start; i < min(len(rows), start+menuHeight); i++ {
		name := ellipsize("/"+rows[i].name, nameWidth-1)
		cursor, style := "  ", plain
		if i == idx {
			cursor, style = selected.Render("❯ "), selected
		}
		pad := strings.Repeat(" ", nameWidth-len([]rune(name)))
		b.WriteString("  " + cursor + style.Render(name) + pad + dim.Render(ellipsize(rows[i].desc, descWidth)) + "\n")
	}
	return b.String()
}

func ellipsize(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width-1]) + "…"
}

// taCursor is the cursor's rune offset into a textarea's whole value.
// LineInfo is relative to the wrapped row the cursor sits on, and its two
// columns add back up to the column within the logical line.
func taCursor(ta textarea.Model) int {
	li := ta.LineInfo()
	col := li.StartColumn + li.ColumnOffset
	lines := strings.Split(ta.Value(), "\n")
	n := 0
	for i := 0; i < ta.Line() && i < len(lines); i++ {
		n += len([]rune(lines[i])) + 1
	}
	return n + col
}
