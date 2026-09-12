package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Agents view (ctrl+1, or straight in via `--view agents` from prefix+k): a
// fuzzy finder over herdr's live agents. Blocked and done agents are the list
// — blocked first, then most recently finished — with everything else dimmed
// below as a preview; ctrl+a again widens to every agent, most recently
// visited first. Rows read project · kind · title, the kind being a feature
// worktree or a chat in the project root, and a project picker above the
// filter (all by default) narrows the list. Enter hands the pane to the
// launcher, which focuses it.

type agentRow struct {
	Key       string // MRU identity: Claude session id, else pane id
	PaneID    string
	TabID     string
	Name      string
	Title     string // terminal title when it says more than Name
	Agent     string
	Status    string // blocked / done / working / idle / unknown
	Workspace string
	Tab       string // "" when the tab label is just its number
	Project   string // august / moi / dotfiles, "" when outside all three
	Feature   bool   // lives in a feature worktree, not the project root
	Seq       int    // state_change_seq: higher = changed more recently
	Current   bool   // the pane the popup was opened from
	mruRank   int    // 0 = visited most recently; agentMRUMax when never
}

// Nerd Font glyphs: a git branch for a feature worktree, a comment bubble
// for a chat in the project root.
const (
	glyphFeature = "\ue0a0"
	glyphChat    = "\uf075"
)

func (r agentRow) kindGlyph() string {
	if r.Feature {
		return glyphFeature
	}
	return glyphChat
}

// label is what the row is called: the feature name for a worktree agent
// (its workspace label), the session name for a chat. The tab label carries
// the full session name where herdr's agent name is capped at 32 characters.
func (r agentRow) label() string {
	switch {
	case r.Feature && r.Workspace != "":
		return r.Workspace
	case r.Tab != "":
		return r.Tab
	}
	return r.Name
}

// Blocked agents want you now; finished ones want you next.
func (r agentRow) actionable() bool {
	return r.Status == "blocked" || r.Status == "done"
}

func (r agentRow) statusRank() int {
	switch r.Status {
	case "blocked":
		return 0
	case "done":
		return 1
	case "working":
		return 2
	case "idle":
		return 3
	}
	return 4
}

// projectColumn pads the project to the column width the rows share.
const projectWidth = 9

// searchText is the one string the fuzzy matcher sees and the row displays
// (project column, then the label), so highlights land where they matched.
func (r agentRow) searchText() string {
	project := r.Project
	if project == "" {
		project = "·"
	}
	return fmt.Sprintf("%-*s %s", projectWidth, project, r.label())
}

// projectRoots are the checkout paths each project's workspaces live under.
// moi worktrees sit beside the repo as moi__worktrees/<name>.
func projectRoots() map[string]string {
	roots := map[string]string{}
	for _, p := range projects {
		if root, err := projectRoot(p); err == nil {
			roots[p] = root
		}
	}
	return roots
}

func projectOf(roots map[string]string, path string) string {
	if path == "" {
		return ""
	}
	if real, err := filepath.EvalSymlinks(path); err == nil {
		path = real
	}
	for _, p := range projects {
		root := roots[p]
		if root == "" {
			continue
		}
		if path == root || strings.HasPrefix(path, root+"/") || strings.HasPrefix(path, root+"__worktrees/") {
			return p
		}
	}
	return ""
}

type agentsMsg struct {
	rows []agentRow
	err  error
}

type agentsPollMsg struct{ gen int }

const agentsPollEvery = 2 * time.Second

func herdrJSON(v any, args ...string) error {
	out, err := exec.Command("herdr", args...).Output()
	if err != nil {
		return fmt.Errorf("herdr %s: %w", strings.Join(args, " "), err)
	}
	return json.Unmarshal(out, v)
}

// fetchAgents joins `herdr agent list` with workspace and tab labels. The
// three CLI calls run concurrently; a failed label lookup only degrades the
// display, a failed agent list is the error.
func fetchAgents(mru *agentMRU) ([]agentRow, error) {
	var agents struct {
		Result struct {
			Agents []struct {
				Agent   string `json:"agent"`
				Session *struct {
					Value string `json:"value"`
				} `json:"agent_session"`
				Status      string `json:"agent_status"`
				Cwd         string `json:"cwd"`
				Focused     bool   `json:"focused"`
				Name        string `json:"name"`
				PaneID      string `json:"pane_id"`
				Seq         int    `json:"state_change_seq"`
				TabID       string `json:"tab_id"`
				Title       string `json:"terminal_title_stripped"`
				WorkspaceID string `json:"workspace_id"`
			} `json:"agents"`
		} `json:"result"`
	}
	var workspaces struct {
		Result struct {
			Workspaces []struct {
				ID       string `json:"workspace_id"`
				Label    string `json:"label"`
				Worktree *struct {
					Checkout string `json:"checkout_path"`
					Linked   bool   `json:"is_linked_worktree"`
				} `json:"worktree"`
			} `json:"workspaces"`
		} `json:"result"`
	}
	var tabs struct {
		Result struct {
			Tabs []struct {
				ID     string `json:"tab_id"`
				Label  string `json:"label"`
				Number int    `json:"number"`
			} `json:"tabs"`
		} `json:"result"`
	}

	var wg sync.WaitGroup
	var agentsErr error
	wg.Add(3)
	go func() { defer wg.Done(); agentsErr = herdrJSON(&agents, "agent", "list") }()
	go func() { defer wg.Done(); _ = herdrJSON(&workspaces, "workspace", "list") }()
	go func() { defer wg.Done(); _ = herdrJSON(&tabs, "tab", "list") }()
	wg.Wait()
	if agentsErr != nil {
		return nil, agentsErr
	}

	wsLabel := map[string]string{}
	wsPath := map[string]string{}
	wsLinked := map[string]bool{}
	for _, ws := range workspaces.Result.Workspaces {
		wsLabel[ws.ID] = ws.Label
		if ws.Worktree != nil {
			wsPath[ws.ID] = ws.Worktree.Checkout
			wsLinked[ws.ID] = ws.Worktree.Linked
		}
	}
	roots := projectRoots()
	tabLabel := map[string]string{}
	for _, t := range tabs.Result.Tabs {
		if t.Label != "" && t.Label != fmt.Sprint(t.Number) {
			tabLabel[t.ID] = t.Label
		}
	}

	activePane := os.Getenv("HERDR_ACTIVE_PANE_ID")
	rows := make([]agentRow, 0, len(agents.Result.Agents))
	for _, a := range agents.Result.Agents {
		key := a.PaneID
		if a.Session != nil && a.Session.Value != "" {
			key = a.Session.Value
		}
		// Claude's default title says nothing; fall back through the tab
		// label and the directory before settling on the agent kind.
		title := a.Title
		if title == "Claude Code" {
			title = ""
		}
		name := a.Name
		for _, candidate := range []string{title, tabLabel[a.TabID], filepath.Base(a.Cwd), a.Agent} {
			if name == "" {
				name = candidate
			}
		}
		if title == name {
			title = ""
		}
		// The workspace checkout says which project and whether this is a
		// feature worktree; a plain folder workspace falls back to the cwd.
		where := wsPath[a.WorkspaceID]
		if where == "" {
			where = a.Cwd
		}
		row := agentRow{
			Key:       key,
			PaneID:    a.PaneID,
			TabID:     a.TabID,
			Name:      name,
			Title:     title,
			Agent:     a.Agent,
			Status:    a.Status,
			Workspace: wsLabel[a.WorkspaceID],
			Tab:       tabLabel[a.TabID],
			Project:   projectOf(roots, where),
			Feature:   wsLinked[a.WorkspaceID],
			Seq:       a.Seq,
			Current:   a.PaneID == activePane || (activePane == "" && a.Focused),
		}
		if row.Status == "" {
			row.Status = "unknown"
		}
		if row.Workspace == "" {
			row.Workspace = a.WorkspaceID
		}
		if row.Tab == row.Name {
			row.Tab = ""
		}
		rows = append(rows, row)
	}

	// Where you are now becomes the first "other" row next time, so opening
	// the picker from the target and pressing enter bounces back (alt-tab).
	for _, r := range rows {
		if r.Current {
			mru.touch(r.Key)
			break
		}
	}
	for i := range rows {
		rows[i].mruRank = mru.rank(rows[i].Key)
	}
	return rows, nil
}

func loadAgentsCmd(mru *agentMRU) tea.Cmd {
	return func() tea.Msg {
		rows, err := fetchAgents(mru)
		return agentsMsg{rows: rows, err: err}
	}
}

func agentsPollCmd(gen int) tea.Cmd {
	return tea.Tick(agentsPollEvery, func(time.Time) tea.Msg { return agentsPollMsg{gen: gen} })
}

// orderAgents splits rows into the actionable section and the rest, each in
// resting order. Actionable: blocked before done, most recent change first.
// Rest: most recently visited first, then status; the pane you came from sinks
// to the bottom since jumping there is a no-op.
func orderAgents(rows []agentRow) (actionable, rest []agentRow) {
	for _, r := range rows {
		if r.actionable() && !r.Current {
			actionable = append(actionable, r)
		} else {
			rest = append(rest, r)
		}
	}
	sort.SliceStable(actionable, func(i, j int) bool {
		a, b := actionable[i], actionable[j]
		if a.statusRank() != b.statusRank() {
			return a.statusRank() < b.statusRank()
		}
		return a.Seq > b.Seq
	})
	sort.SliceStable(rest, func(i, j int) bool {
		a, b := rest[i], rest[j]
		if a.Current != b.Current {
			return !a.Current
		}
		if a.mruRank != b.mruRank {
			return a.mruRank < b.mruRank
		}
		if a.statusRank() != b.statusRank() {
			return a.statusRank() < b.statusRank()
		}
		return a.Seq > b.Seq
	})
	return actionable, rest
}

// rankAgents filters rows by the pattern (best score first, resting order on
// ties) and returns the matched rune positions per row for highlighting.
func rankAgents(pattern string, rows []agentRow) ([]agentRow, [][]int) {
	if strings.TrimSpace(pattern) == "" {
		return rows, nil
	}
	texts := make([]string, len(rows))
	for i, r := range rows {
		texts[i] = r.searchText()
	}
	var out []agentRow
	var idx [][]int
	for _, hit := range fuzzyRank(pattern, texts) {
		out = append(out, rows[hit.index])
		idx = append(idx, hit.matched)
	}
	return out, idx
}

func newAgentFilter(s styles) textinput.Model {
	in := textinput.New()
	in.Prompt = "> "
	in.Placeholder = "filter agents"
	in.Width = 60
	in.PromptStyle = lipgloss.NewStyle().Foreground(s.accent)
	in.PlaceholderStyle = lipgloss.NewStyle().Foreground(s.dim)
	return in
}

func (m *model) enterAgents() tea.Cmd {
	m.view = viewAgents
	m.agentGen++
	cmds := []tea.Cmd{m.agentFilter.Focus()}
	if !m.agentsLoading {
		m.agentsLoading = true
		cmds = append(cmds, loadAgentsCmd(m.mru))
	}
	return tea.Batch(cmds...)
}

// leaveAgents stops the poll loop: bumping the generation orphans any tick
// already scheduled.
func (m *model) leaveAgents() {
	m.agentFilter.Blur()
	m.agentGen++
}

// applyAgentFilter recomputes the navigable rows and the dimmed preview. In
// the needs-you scope a pattern that matches nothing actionable falls through
// to the others, so a typed name always lands somewhere.
func (m *model) applyAgentFilter() {
	actionable, rest := orderAgents(m.scopedAgents())
	pattern := m.agentFilter.Value()
	actionable, actIdx := rankAgents(pattern, actionable)
	rest, restIdx := rankAgents(pattern, rest)

	switch {
	case m.agentsAll:
		m.agentNav = append(append([]agentRow{}, actionable...), rest...)
		m.agentNavIdx = nil
		if actIdx != nil || restIdx != nil {
			m.agentNavIdx = append(append([][]int{}, actIdx...), restIdx...)
		}
		m.agentDim = nil
		m.agentNavLabel = ""
	case len(actionable) == 0 && strings.TrimSpace(pattern) != "":
		m.agentNav, m.agentNavIdx, m.agentDim = rest, restIdx, nil
		m.agentNavLabel = "others"
	default:
		m.agentNav, m.agentNavIdx, m.agentDim = actionable, actIdx, rest
		m.agentNavLabel = "needs you"
	}

	m.agentCursor = 0
	for i, r := range m.agentNav {
		if r.Key == m.agentFollow {
			m.agentCursor = i
		}
	}
}

// agentProjects is the picker's choices: every project, or all of them.
var agentProjects = append([]string{"all"}, projects...)

// scopedAgents is the agent list narrowed to the picked project.
func (m model) scopedAgents() []agentRow {
	if m.agentProject == 0 {
		return m.agents
	}
	want := agentProjects[m.agentProject]
	var out []agentRow
	for _, r := range m.agents {
		if r.Project == want {
			out = append(out, r)
		}
	}
	return out
}

func (m *model) setAgentProject(i int) {
	m.agentProject = (i + len(agentProjects)) % len(agentProjects)
	m.followAgent()
	m.applyAgentFilter()
}

func (m *model) followAgent() {
	m.agentFollow = ""
	if m.agentCursor < len(m.agentNav) {
		m.agentFollow = m.agentNav[m.agentCursor].Key
	}
}

func (m model) onAgentsMsg(msg agentsMsg) (tea.Model, tea.Cmd) {
	m.agentsLoading = false
	if msg.err != nil {
		m.agentErr = msg.err.Error()
	} else {
		m.agentErr = ""
		m.agents = msg.rows
		if !m.agentsScoped {
			// Nothing to act on: start wide instead of on an empty list.
			actionable, _ := orderAgents(m.agents)
			m.agentsAll = len(actionable) == 0
			m.agentsScoped = true
		}
		m.applyAgentFilter()
	}
	if m.view != viewAgents {
		return m, nil
	}
	return m, agentsPollCmd(m.agentGen)
}

func (m model) updateAgents(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.agentFilter, cmd = m.agentFilter.Update(msg)
		return m, cmd
	}
	if k.String() == "shift+tab" {
		m.agentProjectFocused = !m.agentProjectFocused
		if m.agentProjectFocused {
			m.agentFilter.Blur()
			return m, nil
		}
		return m, m.agentFilter.Focus()
	}
	if m.agentProjectFocused {
		switch k.String() {
		case "left", "h", "up", "k":
			m.setAgentProject(m.agentProject - 1)
		case "right", "l", "down", "j", " ":
			m.setAgentProject(m.agentProject + 1)
		case "tab", "enter":
			m.agentProjectFocused = false
			return m, m.agentFilter.Focus()
		default:
			if i, ok := projectByInitial(k.String()); ok {
				m.setAgentProject(i + 1)
			}
		}
		return m, nil
	}
	switch k.String() {
	case "ctrl+a", "tab":
		m.agentsAll = !m.agentsAll
		m.agentsScoped = true
		m.followAgent()
		m.applyAgentFilter()
		return m, nil
	case "up", "ctrl+k", "ctrl+p":
		m.agentCursor = max(0, m.agentCursor-1)
		m.followAgent()
		return m, nil
	case "down", "ctrl+j", "ctrl+n":
		m.agentCursor = min(m.agentCursor+1, max(0, len(m.agentNav)-1))
		m.followAgent()
		return m, nil
	case "enter", "ctrl+s":
		if len(m.agentNav) == 0 {
			m.err = "no matching agent"
			return m, nil
		}
		m.agentPicked = m.agentNav[m.agentCursor]
		m.mru.touch(m.agentPicked.Key)
		m.mru.save()
		m.mode = "agent"
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.agentFilter, cmd = m.agentFilter.Update(msg)
	m.err = ""
	m.agentFollow = ""
	m.applyAgentFilter()
	return m, cmd
}

func (m model) agentsView() string {
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	b.WriteString(m.projectPicker(agentProjects[m.agentProject], m.agentProjectFocused) + "\n\n")
	b.WriteString(m.agentFilter.View() + "\n\n")

	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	switch {
	case m.agentErr != "":
		b.WriteString(m.styles.err.Render("  "+m.agentErr) + "\n")
	case m.agentsLoading && m.agents == nil:
		b.WriteString(dim.Render("  listing agents…") + "\n")
	case len(m.agents) == 0:
		b.WriteString(dim.Render("  no agents running") + "\n")
	case len(m.agentNav) == 0 && len(m.agentDim) == 0:
		b.WriteString(dim.Render("  no agents match") + "\n")
	default:
		b.WriteString(m.agentListView())
	}

	scope := "ctrl+a all"
	if m.agentsAll {
		scope = "ctrl+a needs-you only"
	}
	help := "enter jump · ctrl+j/k move · " + scope + " · shift+tab project · " + tabsHelp + " · esc cancel"
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}

func (m model) agentGlyph(status string) string {
	// Same glyphs as herdr's sidebar in status_indicators = "symbols" mode.
	switch status {
	case "blocked":
		return m.styles.err.Render("×")
	case "done":
		return lipgloss.NewStyle().Foreground(m.styles.accent).Render("✓")
	case "working":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render("◐")
	case "idle":
		return lipgloss.NewStyle().Foreground(m.styles.dim).Render("○")
	}
	return lipgloss.NewStyle().Foreground(m.styles.dim).Render("·")
}

// agentListView: the navigable rows in a window around the cursor, then as
// many dimmed preview rows as still fit. Rows are the search text with fuzzy
// hits underlined; the status glyph on the left is the only status shown.
func (m model) agentListView() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	section := lipgloss.NewStyle().Foreground(m.styles.dim).Bold(true).Transform(strings.ToUpper)
	accent := lipgloss.NewStyle().Foreground(m.styles.accent)
	selected := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)
	plain := lipgloss.NewStyle().Foreground(m.styles.text)
	matched := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true).Underline(true)

	// frame padding (4) + cursor/status (5) + kind glyph (2)
	textWidth := max(24, m.width-11)
	rows := m.listRows
	if m.agentNavLabel != "" {
		rows--
	}

	row := func(r agentRow, hits map[int]bool, cursor bool, muted bool) string {
		mark := "  "
		style := plain
		hit := matched
		switch {
		case cursor:
			mark = accent.Render("❯ ")
			style = selected
		case muted:
			style = dim
			hit = dim.Underline(true)
		}
		text := r.searchText()
		if r.Current {
			text += " (here)"
		}
		// The kind glyph sits between the project column and the label, so
		// the two halves of the search text are rendered around it.
		kind := dim.Render(r.kindGlyph())
		if cursor {
			kind = accent.Render(r.kindGlyph())
		}
		split := projectWidth + 1
		runes := []rune(text)
		project := renderTitle(string(runes[:split]), split, style, hit, hits)
		rest := map[int]bool{}
		for i := range hits {
			if i >= split {
				rest[i-split] = true
			}
		}
		label := renderTitle(string(runes[split:]), textWidth-split-2, style, hit, rest)
		return fmt.Sprintf("%s%s  %s%s %s\n", mark, m.agentGlyph(r.Status), project, kind, label)
	}
	hitsFor := func(idx [][]int, i int) map[int]bool {
		if idx == nil || i >= len(idx) {
			return nil
		}
		hits := map[int]bool{}
		for _, j := range idx[i] {
			hits[j] = true
		}
		return hits
	}

	if m.agentNavLabel != "" {
		b.WriteString(section.Render(" "+m.agentNavLabel) + "\n")
	}
	if len(m.agentNav) == 0 {
		b.WriteString(dim.Render("  (none)") + "\n")
		rows--
	}
	navRows := min(len(m.agentNav), rows)
	start := 0
	if m.agentCursor >= navRows {
		start = m.agentCursor - navRows + 1
	}
	end := min(len(m.agentNav), start+navRows)
	for i := start; i < end; i++ {
		b.WriteString(row(m.agentNav[i], hitsFor(m.agentNavIdx, i), i == m.agentCursor, false))
	}
	if navRows > 0 && (start > 0 || end < len(m.agentNav)) {
		b.WriteString(dim.Render(fmt.Sprintf("  %d–%d of %d", start+1, end, len(m.agentNav))) + "\n")
	}

	left := rows - (end - start) - 2 // section header + a possible "n more" line
	if len(m.agentDim) > 0 && left > 1 {
		left--
		b.WriteString("\n" + section.Render(" others") + dim.Render("  ctrl+a to include") + "\n")
		shown := min(len(m.agentDim), left)
		for i := 0; i < shown; i++ {
			b.WriteString(row(m.agentDim[i], nil, false, true))
		}
		if shown < len(m.agentDim) {
			b.WriteString(dim.Render(fmt.Sprintf("  … %d more", len(m.agentDim)-shown)) + "\n")
		}
	}
	return b.String()
}

// ── visited-agent MRU, persisted across popups ──────────────────────────────

const agentMRUMax = 100

type agentMRU struct {
	path string
	keys []string // most recent first
}

func loadAgentMRU() *agentMRU {
	home, _ := os.UserHomeDir()
	m := &agentMRU{path: filepath.Join(home, ".cache", "herdr-threads", "agent-mru.json")}
	if data, err := os.ReadFile(m.path); err == nil {
		_ = json.Unmarshal(data, &m.keys)
	}
	return m
}

func (m *agentMRU) rank(key string) int {
	for i, k := range m.keys {
		if k == key {
			return i
		}
	}
	return agentMRUMax
}

func (m *agentMRU) touch(key string) {
	if key == "" {
		return
	}
	keys := []string{key}
	for _, k := range m.keys {
		if k != key {
			keys = append(keys, k)
		}
	}
	if len(keys) > agentMRUMax {
		keys = keys[:agentMRUMax]
	}
	m.keys = keys
}

// save is best-effort: losing the visit history is not worth failing a jump.
func (m *agentMRU) save() {
	data, err := json.Marshal(m.keys)
	if err != nil || os.MkdirAll(filepath.Dir(m.path), 0o755) != nil {
		return
	}
	tmp := m.path + ".tmp"
	if os.WriteFile(tmp, data, 0o644) == nil {
		_ = os.Rename(tmp, m.path)
	}
}

// sampleAgents is the --demo fixture: one of every status plus the current pane.
func sampleAgents() []agentRow {
	return []agentRow{
		{Key: "a", PaneID: "w1:p1", Name: "omnicare-timeout", Status: "blocked", Workspace: "august", Project: "august", Seq: 5},
		{Key: "b", PaneID: "w2:p1", Name: "emar-sync", Status: "done", Workspace: "emar-sync-tab-race", Project: "august", Feature: true, Seq: 9},
		{Key: "c", PaneID: "w3:p1", Name: "Representative test data generation", Status: "done", Workspace: "emar-perf-investigation", Project: "august", Feature: true, Seq: 3},
		{Key: "d", PaneID: "w4:p1", Name: "dictation-app", Status: "working", Workspace: "dotfiles", Project: "dotfiles", mruRank: 1},
		{Key: "e", PaneID: "w5:p1", Name: "deadlines-agent", Status: "idle", Workspace: "deadlines", Project: "moi", Feature: true, mruRank: agentMRUMax},
		{Key: "f", PaneID: "w6:p1", Name: "new-session-thing", Status: "working", Workspace: "dotfiles", Project: "dotfiles", Current: true},
	}
}
