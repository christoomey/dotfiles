// herdr-new-session-form: the popup behind prefix+ctrl+t. Three sections on
// ctrl+1..3, two of them with subtabs (ctrl+t, or the section's own key
// again, cycles):
//
//	1 agents            (also prefix+k via --view agents) a fuzzy finder over
//	                    live herdr agents, needs-you first
//	2 chats › new       the new-session form
//	        › resume    a fuzzy list of past sessions in a project root
//	3 features › new    a Linear-branch input that bootstraps a feature worktree
//	           › sidebar herdr's spaces as the sidebar shows them, reordered
//	                    with J/K or a mouse drag
//
// Ctrl+digits need the kitty keyboard protocol (see kitty.go); ctrl+r/f/a
// jump straight to resume/feature/agents as legacy fallbacks. Both project
// pickers default to august and sit above the main flow: shift+tab reaches
// them, a/m/d pick a project while one has focus, and from anywhere in the
// view ctrl+m / ctrl+d pick moi / dotfiles (ctrl+m needs kitty: in legacy
// encoding it is enter). Prints one JSON result on stdout; exits 1 on cancel.
//
//	{"mode":"open"|"bg", "project","name","prompt","attachments"}   new session (attachments: png paths)
//	{"mode":"resume"|"resume-bg", "project","session_id","name"}    resume a past session
//	{"mode":"focus", "tab_id","name"}                               session is live already
//	{"mode":"feature"|"feature-bg", "project","branch","name","prompt"}  bootstrap a feature worktree
//	{"mode":"agent", "tab_id","pane_id","name"}                     jump to a live agent's pane
//	{"mode":"workspace", "workspace_id","name"}                     focus a workspace from the sidebar tab
//
//	herdr-new-session-form [--theme NAME] [--view agents]
//	herdr-new-session-form --demo NAME fresh|filled|resume|agents|sidebar   # print one frame (ANSI)
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
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
	PaneID    string   `json:"pane_id,omitempty"`
	Workspace string   `json:"workspace_id,omitempty"`
}

type view int

const (
	viewAgents view = iota
	viewNew
	viewResume
	viewFeature
	viewSidebar
)

// startViews are the --view names: a section opens on its first subtab.
var startViews = map[string]view{
	"agents":   viewAgents,
	"chats":    viewNew,
	"new":      viewNew,
	"resume":   viewResume,
	"features": viewFeature,
	"sidebar":  viewSidebar,
}

// sections are the ctrl+1..3 tabs; a section with several views shows them
// as subtabs.
var sections = []struct {
	label string
	views []view
	subs  []string
}{
	{"agents", []view{viewAgents}, nil},
	{"chats", []view{viewNew, viewResume}, []string{"new", "resume"}},
	{"features", []view{viewFeature, viewSidebar}, []string{"new", "sidebar"}},
}

const tabsHelp = "ctrl+1..3 section · ctrl+t subtab"

func sectionOf(v view) int {
	for i, s := range sections {
		for _, sv := range s.views {
			if sv == v {
				return i
			}
		}
	}
	return 0
}

// nextSub is the subtab after the current one within its section.
func (m model) nextSub() view {
	views := sections[sectionOf(m.view)].views
	for i, v := range views {
		if v == m.view {
			return views[(i+1)%len(views)]
		}
	}
	return m.view
}

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
	projSel *huh.Select[string] // the form's project field, for a/m/d jumps
	project *string
	name    *string
	prompt  *string
	openNow *bool
	mode    string
	err     string
	styles  styles

	attachDir   string
	attachments []string

	feature        textinput.Model // Linear branch
	featureName    textinput.Model
	featurePrompt  textarea.Model
	featureFocus   int
	nameEdited     bool // the user took the name over from the branch
	promptEdited   bool
	branch         string
	featureProject string

	view     view
	sub      []view // last-used view per section, so ctrl+N returns where you were
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
	initCmd        tea.Cmd              // enters a non-default start view once
	projectFocused bool                 // resume view: picker has focus instead of the filter
	loaded         map[string][]session // per-project scan cache
	width          int

	// Agents view; see agents.go.
	agentFilter         textinput.Model
	agents              []agentRow
	agentNav            []agentRow // navigable rows after scope + filter
	agentNavIdx         [][]int    // fuzzy hit positions per nav row; nil without a pattern
	agentNavLabel       string     // section header over the nav rows; "" in the all scope
	agentDim            []agentRow // dimmed preview under the nav rows (needs-you scope)
	agentCursor         int
	agentFollow         string // key of the row the cursor should stick to across polls
	agentsAll           bool   // every agent is navigable, not just blocked/done
	agentsScoped        bool   // scope has been decided (by the user, or auto on first load)
	agentsLoading       bool
	agentErr            string
	agentGen            int // poll generation; a tick from an older one is dropped
	agentPicked         agentRow
	mru                 *agentMRU
	agentProject        int // index into agentProjects; 0 = all
	agentProjectFocused bool

	// Sidebar view; see sidebar.go.
	ws        []wsRow
	wsCursor  int
	wsFollow  string // id of the row the cursor sticks to across reloads
	wsLoading bool
	wsErr     string
	wsGen     int // reply/poll generation; older ones are dropped
	wsPicked  wsRow
	wsBranch  map[string]string // primary branch by workspace id, read once
	wsDrag    bool              // mouse button held on a row
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

func newModel(themeName string, start view, project string) model {
	m := model{project: new(string), name: new(string), prompt: new(string), openNow: new(bool), view: start}
	*m.openNow = true
	*m.project = project
	m.resumeProject = slices.Index(projects, project)
	m.styles = themes[themeName]
	m.sub = make([]view, len(sections))
	for i, s := range sections {
		m.sub[i] = s.views[0]
	}
	m.sub[sectionOf(start)] = start

	keys := huh.NewDefaultKeyMap()
	keys.Input.Next = key.NewBinding(key.WithKeys("tab", "enter"))
	keys.Text.Next = key.NewBinding(key.WithKeys("tab"))
	keys.Text.NewLine = key.NewBinding(key.WithKeys("enter", "ctrl+j"))
	keys.Text.Submit = key.NewBinding(key.WithKeys("ctrl+s"))
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
	m.projSel = projectField
	m.form = huh.NewForm(huh.NewGroup(projectField, nameField, promptField, confirmField)).
		WithKeyMap(keys).WithShowHelp(false).WithTheme(m.styles.theme).WithWidth(77)
	// huh sizes a focused text input one cell past the form width, which
	// clips a bordered field's right edge; size the fields themselves narrower.
	for _, f := range []huh.Field{nameField, projectField, promptField, confirmField} {
		f.WithWidth(73)
	}

	m.filter = textinput.New()
	m.filter.Prompt = "> "
	m.filter.Placeholder = "filter sessions"
	m.filter.Width = 60
	m.filter.PromptStyle = lipgloss.NewStyle().Foreground(m.styles.accent)
	m.filter.PlaceholderStyle = lipgloss.NewStyle().Foreground(m.styles.dim)
	m.initFeature(m.styles)
	m.agentFilter = newAgentFilter(m.styles)
	m.mru = loadAgentMRU()
	m.loading = true
	m.listRows = 12
	m.width = 86
	m.loaded = map[string][]session{}
	switch start {
	case viewAgents:
		m.agentFilter.Focus()
		m.agentsLoading = true
	case viewNew:
	default:
		m.view = viewNew
		m.initCmd = m.switchView(start)
	}
	return m
}

// Init skips past the project picker so typing starts in the name field.
func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.form.Init(), huh.NextField, loadSessionsCmd(projects[m.resumeProject]), loadLiveCmd}
	if m.view == viewAgents {
		cmds = append(cmds, loadAgentsCmd(m.mru), textinput.Blink)
	}
	if m.initCmd != nil {
		cmds = append(cmds, m.initCmd, textinput.Blink)
	}
	return tea.Batch(cmds...)
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

// switchView is every tab change: tidies whichever view is being left, then
// enters the target and remembers it as its section's subtab.
func (m *model) switchView(v view) tea.Cmd {
	if v == m.view {
		return nil
	}
	switch m.view {
	case viewAgents:
		m.leaveAgents()
	case viewSidebar:
		m.leaveSidebar()
	case viewResume:
		m.filter.Blur()
	case viewFeature:
		m.blurFeature()
	}
	m.err = ""
	m.view = v
	m.sub[sectionOf(v)] = v
	switch v {
	case viewResume:
		m.projectFocused = false
		return m.filter.Focus()
	case viewFeature:
		return m.focusFeature(featBranch)
	case viewAgents:
		return m.enterAgents()
	case viewSidebar:
		return m.enterSidebar()
	default:
		return m.form.Init()
	}
}

// switchSection is ctrl+N: the section's last subtab, or the next subtab
// when its section is already showing.
func (m *model) switchSection(n int) tea.Cmd {
	if n < 0 || n >= len(sections) {
		return nil
	}
	if n == sectionOf(m.view) {
		return m.switchView(m.nextSub())
	}
	return m.switchView(m.sub[n])
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	msg = translateKitty(msg)
	switch msg := msg.(type) {
	case ctrlDigitMsg:
		return m, m.switchSection(int(msg) - 1)
	case ctrlMMsg:
		return m.pickProject(projectMoi)
	case tea.WindowSizeMsg:
		// sections(1) + subtabs(1) + blank + project(2) + blank + filter + blank + list + blank + help, inside the frame.
		m.listRows = max(3, msg.Height-2-11)
		m.width = msg.Width
		return m, nil
	case agentsMsg:
		return m.onAgentsMsg(msg)
	case wsMsg:
		return m.onWsMsg(msg)
	case wsPollMsg:
		if msg.gen != m.wsGen || m.view != viewSidebar {
			return m, nil
		}
		return m, loadWorkspacesCmd(m.wsGen, m.branchCache())
	case agentsPollMsg:
		if msg.gen != m.agentGen || m.view != viewAgents {
			return m, nil
		}
		return m, loadAgentsCmd(m.mru)
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
	if mm, ok := msg.(tea.MouseMsg); ok && mm.Action == tea.MouseActionPress && mm.Button == tea.MouseButtonLeft {
		if v, ok := m.tabAt(mm.X, mm.Y); ok {
			return m, m.switchView(v)
		}
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "ctrl+d":
			return m.pickProject(projectDotfiles)
		case "ctrl+t":
			return m, m.switchView(m.nextSub())
		case "ctrl+r":
			return m, m.switchView(viewResume)
		case "ctrl+f":
			return m, m.switchView(viewFeature)
		case "ctrl+a":
			// In the agents view ctrl+a is the scope toggle instead.
			if m.view != viewAgents {
				return m, m.switchView(viewAgents)
			}
		}
	}
	switch m.view {
	case viewResume:
		return m.updateResume(msg)
	case viewFeature:
		return m.updateFeature(msg)
	case viewAgents:
		return m.updateAgents(msg)
	case viewSidebar:
		return m.updateSidebar(msg)
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		if i, ok := projectByInitial(k.String()); ok && m.form.GetFocusedField() == m.projSel {
			return m.jumpProject(i)
		}
		switch k.String() {
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
		}
	}
	return m.forwardToForm(msg)
}

// pickProject sets the active view's project without moving focus. Views
// without a picker ignore it.
func (m model) pickProject(i int) (tea.Model, tea.Cmd) {
	switch m.view {
	case viewNew:
		return m.jumpProject(i)
	case viewResume:
		return m, m.setResumeProject(i)
	case viewAgents:
		m.setAgentProject(i + 1)
	}
	return m, nil
}

// jumpProject moves the form's inline project select to projects[i] by
// stepping the field itself with arrow keys, whichever field has focus.
// huh's select only re-reads its bound value on binding, so writing the
// pointer would leave its cursor (and what it draws) behind.
func (m model) jumpProject(i int) (tea.Model, tea.Cmd) {
	cur := 0
	for j, p := range projects {
		if p == *m.project {
			cur = j
		}
	}
	step := tea.KeyMsg{Type: tea.KeyRight}
	if i < cur {
		step = tea.KeyMsg{Type: tea.KeyLeft}
	}
	var cmds []tea.Cmd
	for n := 0; n < max(i-cur, cur-i); n++ {
		_, cmd := m.projSel.Update(step)
		cmds = append(cmds, cmd)
	}
	// The form draws a cached render of its fields that only its own Update
	// rebuilds; any message will do.
	out, cmd := m.forwardToForm(formRefreshMsg{})
	return out, tea.Batch(append(cmds, cmd)...)
}

type formRefreshMsg struct{}

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
		if i, ok := projectByInitial(k.String()); ok {
			return m, m.setResumeProject(i)
		}
		return m, nil
	}
	switch k.String() {
	case "up", "ctrl+k", "ctrl+p":
		m.cursor = max(0, m.cursor-1)
		return m, nil
	case "down", "ctrl+j", "ctrl+n":
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
	case viewAgents:
		body = m.agentsView()
	case viewSidebar:
		body = m.sidebarView()
	default:
		help := "ctrl+s open now · ctrl+b in background · ctrl+v attach image · ctrl+m/d moi/dotfiles · " + tabsHelp + " · esc cancel"
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

// tabHit is the column span of one clickable label on the tabs lines.
type tabHit struct {
	x0, x1 int
	v      view
}

// Rows and left edge of the tabs lines inside the popup: the frame pads one
// row above and two columns left.
const (
	tabsRow    = 1
	subtabsRow = 2
	tabsLeft   = 2
)

// tabsView is two lines: the sections, numbered by ctrl+N key, with the
// active one in the theme's focused label style; then that section's
// subtabs indented to sit under it. Always two lines so the views below
// don't jump when switching to a section without subtabs.
func (m model) tabsView() string {
	text, _, _ := m.tabsLayout()
	return text
}

// tabsLayout renders the tabs lines and reports where each label sits, so
// a click can be mapped back to a view.
func (m model) tabsLayout() (text string, secHits, subHits []tabHit) {
	t := m.styles.theme
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	sep := dim.Render("  ·  ")
	sepW := lipgloss.Width(sep)
	active := sectionOf(m.view)

	parts := make([]string, len(sections))
	x := tabsLeft
	indent := 0
	for i, s := range sections {
		label := fmt.Sprintf("%d %s", i+1, s.label)
		if i == active {
			parts[i] = t.Focused.Title.Render(label)
			indent = x - tabsLeft
		} else {
			parts[i] = t.Blurred.Title.Render(label)
		}
		w := lipgloss.Width(parts[i])
		secHits = append(secHits, tabHit{x0: x, x1: x + w, v: m.sub[i]})
		x += w + sepW
	}
	text = strings.Join(parts, sep) + "\n"

	s := sections[active]
	if len(s.views) < 2 {
		return text, secHits, nil
	}
	subs := make([]string, len(s.views))
	x = tabsLeft + indent + 2
	for j, v := range s.views {
		if v == m.view {
			subs[j] = t.Focused.Title.Render(s.subs[j])
		} else {
			subs[j] = dim.Render(s.subs[j])
		}
		w := lipgloss.Width(subs[j])
		subHits = append(subHits, tabHit{x0: x, x1: x + w, v: v})
		x += w + sepW
	}
	return text + strings.Repeat(" ", indent+2) + strings.Join(subs, sep), secHits, subHits
}

// tabAt is the view a click at (x, y) selects, if it landed on a label.
func (m model) tabAt(x, y int) (view, bool) {
	_, secHits, subHits := m.tabsLayout()
	var hits []tabHit
	switch y {
	case tabsRow:
		hits = secHits
	case subtabsRow:
		hits = subHits
	}
	for _, h := range hits {
		if x >= h.x0 && x < h.x1 {
			return h.v, true
		}
	}
	return 0, false
}

// projectPicker renders a view's project line in the same shape as the
// form's inline select: a label, then ← project → when focused.
func (m model) projectPicker(name string, focused bool) string {
	t := m.styles.theme
	if focused {
		return t.Focused.Base.Render(t.Focused.Title.Render("Project") + "\n" +
			t.Focused.PrevIndicator.String() + name + t.Focused.NextIndicator.String())
	}
	return t.Blurred.Base.Render(t.Blurred.Title.Render("Project") + "\n" + name)
}

func (m model) resumeView() string {
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(m.projectPicker(projects[m.resumeProject], m.projectFocused) + "\n\n")
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

	help := "enter resume · ctrl+b in background · ↑/↓ move · ctrl+m/d moi/dotfiles · " + tabsHelp + " · esc cancel"
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
	mm := newModel(themeName, viewNew, projects[0])
	mm.form = mm.form.WithHeight(40)
	var m tea.Model = mm
	m = drive(m, m.Init())
	// agents[-sample][-all][:pattern] — live rows, or a synthetic mix of
	// every status so the needs-you layout can be seen without one.
	if rest, ok := strings.CutPrefix(state, "agents"); ok {
		rest, pattern, _ := strings.Cut(rest, ":")
		rows, _ := fetchAgents(m.(model).mru)
		if strings.Contains(rest, "-sample") {
			rows = sampleAgents()
		}
		next, c := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
		m = drive(next, c)
		// drive already delivered a live agentsMsg; let the fixture decide the scope.
		mm2 := m.(model)
		mm2.agentsScoped = false
		next, c = mm2.Update(agentsMsg{rows: rows})
		m = drive(next, c)
		if strings.Contains(rest, "-all") {
			next, c = m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
			m = drive(next, c)
		}
		for _, r := range pattern {
			next, c := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			m = drive(next, c)
		}
	}
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
	if state == "sidebar" {
		next, c := m.Update(ctrlDigitMsg(3))
		m = drive(next, c)
		next, c = m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
		m = drive(next, c)
		// The socket round trip is slower than drive's budget; fetch inline.
		mm2 := m.(model)
		var list wsList
		err := herdrCall("workspace.list", map[string]any{}, &list)
		rows := list.rows()
		fillBranches(rows, map[string]string{})
		next, c = mm2.Update(wsMsg{gen: mm2.wsGen, rows: rows, err: err})
		m = drive(next, c)
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
	start := viewNew
	project := projects[0]
	for i := 0; i+1 < len(args); i += 2 {
		switch args[i] {
		case "--theme":
			themeName = args[i+1]
		case "--view":
			v, ok := startViews[args[i+1]]
			if !ok {
				fmt.Fprintln(os.Stderr, "unknown view:", args[i+1])
				os.Exit(2)
			}
			start = v
		case "--project":
			if !slices.Contains(projects, args[i+1]) {
				fmt.Fprintln(os.Stderr, "unknown project:", args[i+1])
				os.Exit(2)
			}
			project = args[i+1]
		}
	}
	if _, ok := themes[themeName]; !ok {
		fmt.Fprintln(os.Stderr, "unknown theme:", themeName)
		os.Exit(2)
	}
	p := tea.NewProgram(newModel(themeName, start, project), tea.WithOutput(os.Stderr), tea.WithMouseCellMotion())
	os.Stderr.WriteString(kittyPush)
	final, err := p.Run()
	os.Stderr.WriteString(kittyPop)
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
	case "feature", "feature-bg":
		m.cleanupAttachments()
		res = result{Mode: m.mode, Project: m.featureProject, Branch: m.branch, Name: m.picked.Title, Prompt: strings.TrimSpace(m.featurePrompt.Value())}
	case "focus":
		m.cleanupAttachments()
		// herdr only reports a name for agents it started itself.
		name := m.picked.LiveName
		if name == "" {
			name = m.picked.Title
		}
		res = result{Mode: m.mode, Name: name, TabID: m.picked.LiveTab}
	case "agent":
		m.cleanupAttachments()
		res = result{Mode: m.mode, Name: m.agentPicked.Name, TabID: m.agentPicked.TabID, PaneID: m.agentPicked.PaneID}
	case "workspace":
		m.cleanupAttachments()
		res = result{Mode: m.mode, Name: m.wsPicked.Label, Workspace: m.wsPicked.ID}
	default:
		m.cleanupAttachments()
		res = result{Mode: m.mode, Project: projects[m.resumeProject], Name: m.picked.Title, SessionID: m.picked.ID}
	}
	out, _ := json.Marshal(res)
	fmt.Println(string(out))
}
