// herdr-new-session-form: the popup behind prefix+ctrl+t. Two views: the
// new-session form, (ctrl+r) a fuzzy list of past sessions in a project root
// to resume, and (ctrl+f) a Linear-branch input that bootstraps a feature
// worktree. Both project pickers default to august and sit above the main
// flow: shift+tab reaches them. Prints one JSON result on stdout; exits 1 on
// cancel.
//
//	{"mode":"open"|"bg", "project","name","prompt","attachments"}   new session (attachments: png paths)
//	{"mode":"resume"|"resume-bg", "project","session_id","name"}    resume a past session
//	{"mode":"focus", "tab_id","name"}                               session is live already
//	{"mode":"feature", "project","branch"}                          bootstrap a feature worktree
//
//	herdr-new-session-form [--theme NAME]
//	herdr-new-session-form --demo NAME fresh|filled|resume   # print one frame (ANSI)
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type result struct {
	Mode      string   `json:"mode"`
	Project   string   `json:"project,omitempty"`
	Name      string   `json:"name,omitempty"`
	Branch    string   `json:"branch,omitempty"`
	Prompt    string   `json:"prompt,omitempty"`
	Attach    []string `json:"attachments,omitempty"`
	SessionID string   `json:"session_id,omitempty"`
	TabID     string   `json:"tab_id,omitempty"`
}

type view int

const (
	viewNew view = iota
	viewResume
	viewFeature
)

type sessionsMsg struct {
	project  string
	sessions []session
	err      error
}

type liveMsg struct{ live map[string]session }

// Field values are heap-allocated: huh keeps pointers to them, and Bubble Tea
// passes the model by value, so plain struct fields would be pointed at a
// stale copy.
type model struct {
	form    *huh.Form
	project *string
	name    *string
	prompt  *string
	openNow *bool
	mode    string
	err     string
	styles  styles

	attachDir   string
	attachments []string

	feature        textinput.Model
	branch         string
	featureProject string

	view     view
	filter   textinput.Model
	sessions []session // current project, newest first
	matches  []session // after fuzzy filter
	matchIdx [][]int
	cursor   int
	loading  bool
	loadErr  string
	live     map[string]session
	picked   session
	listRows int

	resumeProject  int                  // index into projects
	projectFocused bool                 // resume view: picker has focus instead of the filter
	loaded         map[string][]session // per-project scan cache
}

type styles struct {
	frame  lipgloss.Style
	help   lipgloss.Style
	err    lipgloss.Style
	attach lipgloss.Style
	accent lipgloss.Color
	dim    lipgloss.Color
	text   lipgloss.Color
	theme  *huh.Theme
}

func newModel(themeName string) model {
	m := model{project: new(string), name: new(string), prompt: new(string), openNow: new(bool)}
	*m.openNow = true
	*m.project = projects[0]
	m.styles = themes[themeName]

	keys := huh.NewDefaultKeyMap()
	keys.Input.Next = key.NewBinding(key.WithKeys("tab", "enter"))
	keys.Text.Next = key.NewBinding(key.WithKeys("tab"))
	keys.Text.NewLine = key.NewBinding(key.WithKeys("enter", "ctrl+j"))
	keys.Text.Submit = key.NewBinding(key.WithKeys("ctrl+d"))
	keys.Confirm.Next = key.NewBinding(key.WithKeys("enter"))
	keys.Confirm.Toggle = key.NewBinding(key.WithKeys("tab", "left", "right", "h", "l", " "))

	projectField := huh.NewSelect[string]().
		Title("Project").
		Options(huh.NewOptions(projects...)...).
		Inline(true).
		Filtering(false).
		Value(m.project)
	nameField := huh.NewInput().
		Title("Session name").
		Placeholder("dose unit case bug").
		Value(m.name).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("name is required")
			}
			return nil
		})
	promptField := huh.NewText().
		Title("First prompt").
		Description("optional · enter for a new line").
		Placeholder("go look into…").
		Lines(4).
		Value(m.prompt)
	confirmField := huh.NewConfirm().
		Title("Start").
		Affirmative("Open now").
		Negative("Run in background").
		Value(m.openNow)
	m.form = huh.NewForm(huh.NewGroup(projectField, nameField, promptField, confirmField)).
		WithKeyMap(keys).WithShowHelp(false).WithTheme(m.styles.theme).WithWidth(70)
	// huh sizes a focused text input one cell past the form width, which
	// clips a bordered field's right edge; size the fields themselves narrower.
	for _, f := range []huh.Field{nameField, projectField, promptField, confirmField} {
		f.WithWidth(66)
	}

	m.filter = textinput.New()
	m.filter.Prompt = "> "
	m.filter.Placeholder = "filter sessions"
	m.filter.Width = 60
	m.filter.PromptStyle = lipgloss.NewStyle().Foreground(m.styles.accent)
	m.filter.PlaceholderStyle = lipgloss.NewStyle().Foreground(m.styles.dim)
	m.feature = newFeatureInput(m.styles)
	m.loading = true
	m.listRows = 12
	m.loaded = map[string][]session{}
	return m
}

// Init skips past the project picker so typing starts in the name field.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.form.Init(), huh.NextField, loadSessionsCmd(projects[m.resumeProject]), loadLiveCmd)
}

func loadSessionsCmd(project string) tea.Cmd {
	return func() tea.Msg {
		sessions, err := loadSessions(project)
		return sessionsMsg{project: project, sessions: sessions, err: err}
	}
}

// setResumeProject switches the resume list to another project, scanning it
// on first use.
func (m *model) setResumeProject(idx int) tea.Cmd {
	m.resumeProject = (idx + len(projects)) % len(projects)
	m.cursor = 0
	m.loadErr = ""
	project := projects[m.resumeProject]
	if cached, ok := m.loaded[project]; ok {
		m.sessions = cached
		m.loading = false
		m.decorateLive()
		return nil
	}
	m.sessions = nil
	m.matches = nil
	m.loading = true
	return loadSessionsCmd(project)
}

func loadLiveCmd() tea.Msg { return liveMsg{live: liveSessions()} }

// applyFilter recomputes the visible rows from the filter text. Empty filter
// keeps recency order; otherwise best fuzzy score first, recency on ties.
func (m *model) applyFilter() {
	pattern := strings.TrimSpace(m.filter.Value())
	m.matches = m.matches[:0]
	m.matchIdx = m.matchIdx[:0]
	if pattern == "" {
		m.matches = append(m.matches, m.sessions...)
		m.matchIdx = nil
	} else {
		titles := make([]string, len(m.sessions))
		for i, s := range m.sessions {
			titles[i] = s.Title
		}
		for _, hit := range fuzzyRank(pattern, titles) {
			m.matches = append(m.matches, m.sessions[hit.index])
			m.matchIdx = append(m.matchIdx, hit.matched)
		}
	}
	if m.cursor >= len(m.matches) {
		m.cursor = max(0, len(m.matches)-1)
	}
}

func (m *model) decorateLive() {
	for i := range m.sessions {
		if l, ok := m.live[m.sessions[i].ID]; ok {
			m.sessions[i].LiveTab = l.LiveTab
			m.sessions[i].LiveName = l.LiveName
		}
	}
	m.applyFilter()
}

func (m *model) enterResume() tea.Cmd {
	m.view = viewResume
	m.err = ""
	m.projectFocused = false
	m.form.WithShowHelp(false)
	return m.filter.Focus()
}

func (m *model) leaveResume() tea.Cmd {
	m.view = viewNew
	m.err = ""
	m.filter.Blur()
	return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// header(1) + blank + project(2) + blank + filter + blank + list + blank + help, inside the frame.
		m.listRows = max(3, msg.Height-2-10)
		return m, nil
	case sessionsMsg:
		if msg.err == nil {
			m.loaded[msg.project] = msg.sessions
		}
		if msg.project != projects[m.resumeProject] {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
			return m, nil
		}
		m.sessions = msg.sessions
		m.decorateLive()
		return m, nil
	case liveMsg:
		m.live = msg.live
		m.decorateLive()
		return m, nil
	case attachMsg:
		switch {
		case msg.err == errNoImage:
			// Plain text on the clipboard: let the field paste it.
			return m.forwardToForm(tea.KeyMsg{Type: tea.KeyCtrlV})
		case msg.err != nil:
			m.err = msg.err.Error()
		default:
			m.attachments = append(m.attachments, msg.path)
		}
		return m, nil
	}
	if m.view == viewResume {
		return m.updateResume(msg)
	}
	if m.view == viewFeature {
		return m.updateFeature(msg)
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+r":
			return m, m.enterResume()
		case "ctrl+f":
			return m, m.enterFeature()
		case "ctrl+v":
			m.err = ""
			return m, m.attachClipboardCmd()
		case "ctrl+x":
			m.dropLastAttachment()
			return m, nil
		case "ctrl+s", "ctrl+b":
			if strings.TrimSpace(*m.name) == "" {
				m.err = "name is required"
				return m, nil
			}
			m.mode = "open"
			if k.String() == "ctrl+b" {
				m.mode = "bg"
			}
			return m, tea.Quit
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m.forwardToForm(msg)
}

func (m model) forwardToForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	f, cmd := m.form.Update(msg)
	if form, ok := f.(*huh.Form); ok {
		m.form = form
	}
	if m.form.State == huh.StateCompleted {
		if strings.TrimSpace(*m.name) == "" {
			return m, tea.Quit
		}
		m.mode = "open"
		if !*m.openNow {
			m.mode = "bg"
		}
		return m, tea.Quit
	}
	if m.form.State == huh.StateAborted {
		return m, tea.Quit
	}
	return m, cmd
}

func (m model) updateResume(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}
	switch k.String() {
	case "esc", "ctrl+c":
		return m, tea.Quit
	case "ctrl+n":
		return m, m.leaveResume()
	case "ctrl+f":
		return m, m.enterFeature()
	case "shift+tab":
		m.projectFocused = !m.projectFocused
		if m.projectFocused {
			m.filter.Blur()
			return m, nil
		}
		return m, m.filter.Focus()
	}
	if m.projectFocused {
		switch k.String() {
		case "left", "h", "up", "k":
			return m, m.setResumeProject(m.resumeProject - 1)
		case "right", "l", "down", "j", " ":
			return m, m.setResumeProject(m.resumeProject + 1)
		case "tab", "enter":
			m.projectFocused = false
			return m, m.filter.Focus()
		}
		return m, nil
	}
	switch k.String() {
	case "up", "ctrl+k", "ctrl+p":
		m.cursor = max(0, m.cursor-1)
		return m, nil
	case "down", "ctrl+j":
		m.cursor = min(m.cursor+1, max(0, len(m.matches)-1))
		return m, nil
	case "enter", "ctrl+s", "ctrl+b":
		if len(m.matches) == 0 {
			m.err = "no matching session"
			return m, nil
		}
		m.picked = m.matches[m.cursor]
		switch {
		case m.picked.LiveTab != "":
			m.mode = "focus"
		case k.String() == "ctrl+b":
			m.mode = "resume-bg"
		default:
			m.mode = "resume"
		}
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.err = ""
	m.cursor = 0
	m.applyFilter()
	return m, cmd
}

func (m model) View() string {
	if m.mode != "" {
		return ""
	}
	var body string
	switch m.view {
	case viewResume:
		body = m.resumeView()
	case viewFeature:
		body = m.featureView()
	default:
		help := "ctrl+s open now · ctrl+b run in background · ctrl+v attach clipboard image · ctrl+r resume · ctrl+f feature · esc cancel"
		body = m.tabsView() + "\n\n" + m.form.View()
		if a := m.attachmentsView(); a != "" {
			body += "\n" + a + "\n"
		}
		body += "\n" + m.styles.help.Render(help)
	}
	if m.err != "" {
		body += "\n" + m.styles.err.Render(m.err)
	}
	return m.styles.frame.Render(body)
}

// tabsView is the view switcher line: the active view in the theme's focused
// label style, the others blurred.
func (m model) tabsView() string {
	t := m.styles.theme
	sep := lipgloss.NewStyle().Foreground(m.styles.dim).Render("  ·  ")
	tabs := []struct {
		v     view
		label string
	}{{viewNew, "new session"}, {viewResume, "resume"}, {viewFeature, "feature"}}
	parts := make([]string, len(tabs))
	for i, tab := range tabs {
		if tab.v == m.view {
			parts[i] = t.Focused.Title.Render(tab.label)
		} else {
			parts[i] = t.Blurred.Title.Render(tab.label)
		}
	}
	return strings.Join(parts, sep)
}

// projectPicker renders the resume view's project line in the same shape as
// the form's inline select: a label, then ← project → when focused.
func (m model) projectPicker() string {
	t := m.styles.theme
	name := projects[m.resumeProject]
	if m.projectFocused {
		return t.Focused.Base.Render(t.Focused.Title.Render("Project") + "\n" +
			t.Focused.PrevIndicator.String() + name + t.Focused.NextIndicator.String())
	}
	return t.Blurred.Base.Render(t.Blurred.Title.Render("Project") + "\n" + name)
}

func (m model) resumeView() string {
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(m.projectPicker() + "\n\n")
	b.WriteString(m.filter.View() + "\n\n")

	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	switch {
	case m.loadErr != "":
		b.WriteString(m.styles.err.Render("  "+m.loadErr) + "\n")
	case m.loading:
		b.WriteString(dim.Render("  scanning sessions…") + "\n")
	case len(m.matches) == 0:
		b.WriteString(dim.Render("  no sessions match") + "\n")
	default:
		b.WriteString(m.listView())
	}

	help := "enter resume · ctrl+b in background · ↑/↓ move · shift+tab project · ctrl+n new · ctrl+f feature · esc cancel"
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}

const (
	titleWidth = 42
	dateLayout = "Jan _2 15:04"
)

func (m model) listView() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	live := lipgloss.NewStyle().Foreground(m.styles.accent)
	selected := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)
	plain := lipgloss.NewStyle().Foreground(m.styles.text)
	matched := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true).Underline(true)

	// Keep the cursor inside the window.
	start := 0
	if m.cursor >= m.listRows {
		start = m.cursor - m.listRows + 1
	}
	end := min(len(m.matches), start+m.listRows)

	for i := start; i < end; i++ {
		s := m.matches[i]
		cursor := "  "
		rowStyle := plain
		if i == m.cursor {
			cursor = live.Render("❯ ")
			rowStyle = selected
		}
		dot := " "
		if s.LiveTab != "" {
			dot = live.Render("●")
		}
		var hits map[int]bool
		if m.matchIdx != nil {
			hits = map[int]bool{}
			for _, idx := range m.matchIdx[i] {
				hits[idx] = true
			}
		}
		title := renderTitle(s.Title, titleWidth, rowStyle, matched, hits)
		meta := fmt.Sprintf("%s %4s %3d⟲", s.LastAt.Format(dateLayout), humanAge(s.LastAt), s.Turns)
		b.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, dot, title, dim.Render(meta)))
	}
	if start > 0 || end < len(m.matches) {
		b.WriteString(dim.Render(fmt.Sprintf("  %d–%d of %d", start+1, end, len(m.matches))) + "\n")
	}
	return b.String()
}

// renderTitle pads/truncates to width and underlines fuzzy-matched runes.
func renderTitle(title string, width int, base, hit lipgloss.Style, hits map[int]bool) string {
	runes := []rune(title)
	if len(runes) > width {
		runes = append(runes[:width-1], '…')
	}
	var b strings.Builder
	for i, r := range runes {
		if hits[i] {
			b.WriteString(hit.Render(string(r)))
		} else {
			b.WriteString(base.Render(string(r)))
		}
	}
	for i := len(runes); i < width; i++ {
		b.WriteByte(' ')
	}
	return b.String()
}

// ── themes ──────────────────────────────────────────────────────────────────

var themes = map[string]styles{}

func init() {
	dim := lipgloss.Color("#6c7086")
	frame := lipgloss.NewStyle().Padding(1, 2)
	help := lipgloss.NewStyle().Foreground(dim)
	errS := lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))

	// A: huh's stock Charm theme, just with colour actually on.
	themes["charm"] = styles{frame: frame, help: help, err: errS, theme: huh.ThemeCharm(),
		accent: lipgloss.Color("#f5c2e7"), dim: dim, text: lipgloss.Color("#cdd6f4"),
		attach: lipgloss.NewStyle().Foreground(lipgloss.Color("#f5c2e7")).PaddingLeft(2)}

	// B: boxed — every field in its own rounded box; the focused box lights up.
	{
		accent := lipgloss.Color("#89b4fa")
		text := lipgloss.Color("#cdd6f4")
		t := huh.ThemeBase()
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
		t.Focused.Base = box.BorderForeground(accent)
		t.Blurred.Base = box.BorderForeground(dim)
		t.Focused.Title = lipgloss.NewStyle().Bold(true).Foreground(accent)
		t.Blurred.Title = lipgloss.NewStyle().Bold(true).Foreground(text)
		t.Focused.Description = lipgloss.NewStyle().Foreground(dim)
		t.Blurred.Description = t.Focused.Description
		for _, f := range []*huh.FieldStyles{&t.Focused, &t.Blurred} {
			f.TextInput.Placeholder = lipgloss.NewStyle().Foreground(dim).Italic(true)
			f.TextInput.Prompt = lipgloss.NewStyle().Foreground(accent)
			f.TextInput.Text = lipgloss.NewStyle().Foreground(text)
			f.TextInput.Cursor = lipgloss.NewStyle().Foreground(accent)
		}
		t.Focused.FocusedButton = lipgloss.NewStyle().Bold(true).Padding(0, 2).Foreground(lipgloss.Color("#1e1e2e")).Background(accent)
		t.Focused.BlurredButton = lipgloss.NewStyle().Padding(0, 2).Foreground(text).Background(lipgloss.Color("#313244"))
		t.Blurred.FocusedButton = t.Focused.BlurredButton
		t.Blurred.BlurredButton = t.Focused.BlurredButton
		t.Focused.ErrorMessage = errS
		t.Blurred.ErrorMessage = errS
		themes["boxed"] = styles{frame: frame, help: help, err: errS, theme: t, accent: accent, dim: dim, text: text,
			attach: lipgloss.NewStyle().Foreground(accent).PaddingLeft(2)}
	}

	// C: minimal — small-caps-style dim labels, a thick accent bar marks the
	// focused field, buttons as bracketed text with the active one inverted.
	{
		accent := lipgloss.Color("#a6e3a1")
		text := lipgloss.Color("#cdd6f4")
		t := huh.ThemeBase()
		bar := lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, false, false, true).PaddingLeft(1)
		t.Focused.Base = bar.BorderForeground(accent)
		t.Blurred.Base = bar.BorderForeground(lipgloss.Color("#313244"))
		label := lipgloss.NewStyle().Foreground(dim).Transform(strings.ToUpper)
		t.Focused.Title = label.Foreground(accent)
		t.Blurred.Title = label
		t.Focused.Description = lipgloss.NewStyle().Foreground(dim)
		t.Blurred.Description = t.Focused.Description
		for _, f := range []*huh.FieldStyles{&t.Focused, &t.Blurred} {
			f.TextInput.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a"))
			f.TextInput.Prompt = lipgloss.NewStyle().Foreground(accent)
			f.TextInput.Text = lipgloss.NewStyle().Foreground(text)
			f.TextInput.Cursor = lipgloss.NewStyle().Foreground(accent)
		}
		t.Focused.FocusedButton = lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(lipgloss.Color("#1e1e2e")).Background(accent)
		t.Focused.BlurredButton = lipgloss.NewStyle().Padding(0, 1).Foreground(dim)
		t.Blurred.FocusedButton = lipgloss.NewStyle().Padding(0, 1).Foreground(text)
		t.Blurred.BlurredButton = t.Focused.BlurredButton
		t.Focused.ErrorMessage = errS
		t.Blurred.ErrorMessage = errS
		themes["minimal"] = styles{frame: frame, help: help, err: errS, theme: t, accent: accent, dim: dim, text: text,
			attach: lipgloss.NewStyle().Foreground(accent).PaddingLeft(2)}
	}
}

// ── demo rendering (no terminal needed) ─────────────────────────────────────

// drive runs a command tree to completion, feeding every message back into
// the model, so a frame can be rendered without a tea.Program. Commands that
// don't answer within 50ms (cursor-blink ticks) are dropped, or they'd loop
// forever.
func drive(m tea.Model, cmd tea.Cmd) tea.Model {
	if cmd == nil {
		return m
	}
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-ch:
	case <-time.After(50 * time.Millisecond):
		return m
	}
	switch v := msg.(type) {
	case nil:
		return m
	case tea.BatchMsg:
		for _, c := range v {
			m = drive(m, c)
		}
		return m
	}
	next, c := m.Update(msg)
	return drive(next, c)
}

func demo(themeName, state string) {
	mm := newModel(themeName)
	mm.form = mm.form.WithHeight(40)
	var m tea.Model = mm
	m = drive(m, m.Init())
	if pattern, ok := strings.CutPrefix(state, "resume:"); ok || state == "resume" {
		mm2 := m.(model)
		mm2.sessions, _ = loadSessions("august")
		mm2.live = liveSessions()
		mm2.loading = false
		mm2.decorateLive()
		mm2.listRows = 12
		m = mm2
		next, c := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
		m = drive(next, c)
		for _, r := range pattern {
			next, c := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			m = drive(next, c)
		}
	}
	if pattern, ok := strings.CutPrefix(state, "feature:"); ok {
		next, c := m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
		m = drive(next, c)
		for _, r := range pattern {
			next, c := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			m = drive(next, c)
		}
	}
	if state == "project" {
		next, c := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		m = drive(next, c)
	}
	if state == "resume-project" {
		mm2 := m.(model)
		mm2.loaded["august"], _ = loadSessions("august")
		mm2.sessions = mm2.loaded["august"]
		mm2.loading = false
		mm2.decorateLive()
		m = mm2
		for _, k := range []tea.KeyType{tea.KeyCtrlR, tea.KeyShiftTab, tea.KeyRight} {
			next, c := m.Update(tea.KeyMsg{Type: k})
			m = drive(next, c)
		}
	}
	if state == "filled" {
		typeText := func(t string) {
			for _, r := range t {
				next, c := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
				m = drive(next, c)
			}
		}
		tab := func() {
			next, c := m.Update(tea.KeyMsg{Type: tea.KeyTab})
			m = drive(next, c)
		}
		typeText("dose unit case bug")
		tab()
		typeText("Look into why HL7 dose units like \"Ml\" blank the review unit select.")
		tab()
	}
	fmt.Print(m.View())
}

func main() {
	lipgloss.SetColorProfile(termenv.TrueColor)
	themeName := "minimal"
	args := os.Args[1:]
	if len(args) >= 3 && args[0] == "--demo" {
		demo(args[1], args[2])
		return
	}
	if len(args) >= 2 && args[0] == "--theme" {
		themeName = args[1]
	}
	if _, ok := themes[themeName]; !ok {
		fmt.Fprintln(os.Stderr, "unknown theme:", themeName)
		os.Exit(2)
	}
	p := tea.NewProgram(newModel(themeName), tea.WithOutput(os.Stderr))
	final, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	m := final.(model)
	if m.mode == "" {
		m.cleanupAttachments()
		os.Exit(1)
	}
	var res result
	switch m.mode {
	case "open", "bg":
		res = result{Mode: m.mode, Project: *m.project, Name: strings.TrimSpace(*m.name), Prompt: strings.TrimSpace(*m.prompt), Attach: m.attachments}
	case "feature":
		m.cleanupAttachments()
		res = result{Mode: m.mode, Project: m.featureProject, Branch: m.branch}
	case "focus":
		m.cleanupAttachments()
		// herdr only reports a name for agents it started itself.
		name := m.picked.LiveName
		if name == "" {
			name = m.picked.Title
		}
		res = result{Mode: m.mode, Name: name, TabID: m.picked.LiveTab}
	default:
		m.cleanupAttachments()
		res = result{Mode: m.mode, Project: projects[m.resumeProject], Name: m.picked.Title, SessionID: m.picked.ID}
	}
	out, _ := json.Marshal(res)
	fmt.Println(string(out))
}
