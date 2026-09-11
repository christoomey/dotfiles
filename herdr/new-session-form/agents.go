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

// Agents view (ctrl+a, or straight in via `--view agents` from prefix+a): a
// fuzzy finder over herdr's live agents. Blocked and done agents are the list
// — blocked first, then most recently finished — with everything else dimmed
// below as a preview; ctrl+a again widens to every agent, most recently
// visited first. Enter hands the pane to the launcher, which focuses it.

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
	Seq       int    // state_change_seq: higher = changed more recently
	Current   bool   // the pane the popup was opened from
	mruRank   int    // 0 = visited most recently; agentMRUMax when never
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

// searchText is the one string the fuzzy matcher sees and the row displays,
// so highlights land where they matched.
func (r agentRow) searchText() string {
	parts := []string{r.Name}
	loc := r.Workspace
	if r.Tab != "" {
		loc += " › " + r.Tab
	}
	if loc != "" && loc != r.Name {
		parts = append(parts, loc)
	}
	if r.Title != "" {
		parts = append(parts, r.Title)
	}
	return strings.Join(parts, " · ")
}

var statusWords = map[string]string{
	"blocked": "needs you",
	"done":    "done",
	"working": "working",
	"idle":    "idle",
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
				ID    string `json:"workspace_id"`
				Label string `json:"label"`
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
	for _, ws := range workspaces.Result.Workspaces {
		wsLabel[ws.ID] = ws.Label
	}
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
	m.err = ""
	m.filter.Blur()
	m.feature.Blur()
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
	actionable, rest := orderAgents(m.agents)
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
	switch k.String() {
	case "esc", "ctrl+c":
		return m, tea.Quit
	case "ctrl+n":
		m.leaveAgents()
		return m, m.leaveResume()
	case "ctrl+r":
		m.leaveAgents()
		return m, m.enterResume()
	case "ctrl+f":
		m.leaveAgents()
		return m, m.enterFeature()
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
	case "down", "ctrl+j":
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
	help := "enter jump · ctrl+j/k move · " + scope + " · ctrl+n new · ctrl+r resume · esc cancel"
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
// hits underlined, plus a status word at the right.
func (m model) agentListView() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	section := lipgloss.NewStyle().Foreground(m.styles.dim).Bold(true).Transform(strings.ToUpper)
	accent := lipgloss.NewStyle().Foreground(m.styles.accent)
	selected := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)
	plain := lipgloss.NewStyle().Foreground(m.styles.text)
	matched := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true).Underline(true)

	// frame padding (4) + cursor/glyph (5) + status column (11)
	textWidth := max(24, m.width-20)
	rows := m.listRows + 2 // no project line on this view
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
		status := statusWords[r.Status]
		return fmt.Sprintf("%s%s  %s %s\n", mark, m.agentGlyph(r.Status), renderTitle(text, textWidth, style, hit, hits), dim.Render(status))
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
	if start > 0 || end < len(m.agentNav) {
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
		{Key: "a", PaneID: "w1:p1", Name: "omnicare-timeout", Status: "blocked", Workspace: "august", Tab: "omnicare-timeout", Seq: 5},
		{Key: "b", PaneID: "w2:p1", Name: "emar-sync", Status: "done", Workspace: "emar-sync-tab-race", Title: "Mutex and bulkUpdate timeout considerations", Seq: 9},
		{Key: "c", PaneID: "w3:p1", Name: "Representative test data generation", Status: "done", Workspace: "emar-perf-investigation", Seq: 3},
		{Key: "d", PaneID: "w4:p1", Name: "dictation-app", Status: "working", Workspace: "dotfiles", mruRank: 1},
		{Key: "e", PaneID: "w5:p1", Name: "start-end-time-data-model", Status: "idle", Workspace: "start-end-time-data-model", mruRank: agentMRUMax},
		{Key: "f", PaneID: "w6:p1", Name: "new-session-thing", Status: "working", Workspace: "dotfiles", Current: true},
	}
}
