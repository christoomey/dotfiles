package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Sidebar subtab (features › sidebar): herdr's workspace list drawn the way
// the sidebar draws it — grouped into spaces, each a primary checkout with
// its branch and its linked worktrees as a tree — plus the reordering the
// sidebar lacks. J/K (shift+↑/↓) or a mouse drag move the highlighted row:
// a worktree past its sibling (workspace.move), a primary with its whole
// space past the neighbouring space (workspace.move_block). Every step is
// sent as it happens, so the real sidebar beside the popup follows along.
// Enter focuses the workspace.

type wsRow struct {
	ID      string
	Label   string
	Status  string
	Focused bool
	Linked  bool
	RepoKey string
	Path    string // checkout path; "" for a plain folder workspace
	Branch  string // primaries only
	Tabs    int
}

type wsMsg struct {
	gen  int
	rows []wsRow
	err  error
}

type wsPollMsg struct{ gen int }

const wsPollEvery = 2 * time.Second

type wsList struct {
	Workspaces []struct {
		ID       string `json:"workspace_id"`
		Label    string `json:"label"`
		Status   string `json:"agent_status"`
		Focused  bool   `json:"focused"`
		TabCount int    `json:"tab_count"`
		Worktree *struct {
			RepoKey  string `json:"repo_key"`
			Checkout string `json:"checkout_path"`
			Linked   bool   `json:"is_linked_worktree"`
		} `json:"worktree"`
	} `json:"workspaces"`
}

func (l wsList) rows() []wsRow {
	rows := make([]wsRow, 0, len(l.Workspaces))
	for _, w := range l.Workspaces {
		row := wsRow{ID: w.ID, Label: w.Label, Status: w.Status, Focused: w.Focused, Tabs: w.TabCount}
		if w.Worktree != nil {
			row.Linked = w.Worktree.Linked
			row.RepoKey = w.Worktree.RepoKey
			row.Path = w.Worktree.Checkout
		}
		rows = append(rows, row)
	}
	return rows
}

// fillBranches reads each primary's checked-out branch. Plain folder
// workspaces carry no checkout path, so those come from a pane's cwd in the
// session snapshot; branches is a cache keyed by workspace id.
func fillBranches(rows []wsRow, branches map[string]string) {
	var snapshot map[string]string // workspace id → a pane cwd, fetched at most once
	for i := range rows {
		r := &rows[i]
		if r.Linked {
			continue
		}
		if b, ok := branches[r.ID]; ok {
			r.Branch = b
			continue
		}
		path := r.Path
		if path == "" {
			if snapshot == nil {
				snapshot = paneCwds()
			}
			path = snapshot[r.ID]
		}
		r.Branch = gitBranch(path)
		branches[r.ID] = r.Branch
	}
}

func paneCwds() map[string]string {
	var resp struct {
		Snapshot struct {
			Panes []struct {
				WorkspaceID string `json:"workspace_id"`
				Cwd         string `json:"cwd"`
			} `json:"panes"`
		} `json:"snapshot"`
	}
	cwds := map[string]string{}
	if herdrCall("session.snapshot", map[string]any{}, &resp) != nil {
		return cwds
	}
	for _, p := range resp.Snapshot.Panes {
		if _, seen := cwds[p.WorkspaceID]; !seen && p.Cwd != "" {
			cwds[p.WorkspaceID] = p.Cwd
		}
	}
	return cwds
}

func gitBranch(path string) string {
	if path == "" {
		return ""
	}
	head, err := os.ReadFile(filepath.Join(path, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	ref := strings.TrimSpace(string(head))
	if b, ok := strings.CutPrefix(ref, "ref: refs/heads/"); ok {
		return b
	}
	if len(ref) >= 7 {
		return ref[:7]
	}
	return ref
}

func loadWorkspacesCmd(gen int, branches map[string]string) tea.Cmd {
	return func() tea.Msg {
		var list wsList
		err := herdrCall("workspace.list", map[string]any{}, &list)
		rows := list.rows()
		if err == nil {
			fillBranches(rows, branches)
		}
		return wsMsg{gen: gen, rows: rows, err: err}
	}
}

// wsMove is one reorder request, the same shape the local preview applied.
type wsMove struct {
	id          string   // workspace.move: this id …
	insertIndex int      // … before this slot of the pre-move list
	block       []string // workspace.move_block: these ids …
	before      string   // … before this id ("" = the end)
}

func (mv wsMove) cmd(gen int, branches map[string]string) tea.Cmd {
	return func() tea.Msg {
		var list wsList
		var err error
		if mv.block != nil {
			params := map[string]any{"workspace_ids": mv.block, "before_workspace_id": nil}
			if mv.before != "" {
				params["before_workspace_id"] = mv.before
			}
			err = herdrCall("workspace.move_block", params, &list)
		} else {
			params := map[string]any{"workspace_id": mv.id, "insert_index": mv.insertIndex}
			err = herdrCall("workspace.move", params, &list)
		}
		rows := list.rows()
		if err == nil {
			fillBranches(rows, branches)
		}
		return wsMsg{gen: gen, rows: rows, err: err}
	}
}

func wsPollCmd(gen int) tea.Cmd {
	return tea.Tick(wsPollEvery, func(time.Time) tea.Msg { return wsPollMsg{gen: gen} })
}

// ── spaces: the flat list as the sidebar groups it ──────────────────────────

// A space is one repo: its primary checkout (or plain folder) and the linked
// worktrees opened from it, in list order. Spaces sit in the order of their
// first workspace in the list.
type wsSpace struct {
	key     string
	primary int   // flat index, -1 for an orphaned worktree group
	members []int // flat indexes of linked worktrees
	label   string
}

// wsNode is one navigable row of the tree: a primary or a worktree.
type wsNode struct {
	flat  int // index into the flat list
	space int // index into spaces
	last  bool
}

func buildSpaces(ws []wsRow) ([]wsSpace, []wsNode) {
	var spaces []wsSpace
	at := map[string]int{}
	for i, r := range ws {
		key := r.RepoKey
		if key == "" {
			key = "ws:" + r.ID
		}
		s, ok := at[key]
		if !ok {
			s = len(spaces)
			at[key] = s
			spaces = append(spaces, wsSpace{key: key, primary: -1, label: filepath.Base(filepath.Dir(key))})
		}
		if r.Linked {
			spaces[s].members = append(spaces[s].members, i)
		} else {
			spaces[s].primary = i
			spaces[s].label = r.Label
		}
	}
	sort.SliceStable(spaces, func(a, b int) bool { return spaceStart(spaces[a]) < spaceStart(spaces[b]) })

	var nodes []wsNode
	for s, sp := range spaces {
		if sp.primary >= 0 {
			nodes = append(nodes, wsNode{flat: sp.primary, space: s, last: len(sp.members) == 0})
		}
		for j, flat := range sp.members {
			nodes = append(nodes, wsNode{flat: flat, space: s, last: j == len(sp.members)-1})
		}
	}
	return spaces, nodes
}

// spaceStart is the flat position that orders a space: its primary, or its
// first worktree when the primary is not open.
func spaceStart(s wsSpace) int {
	if s.primary >= 0 {
		return s.primary
	}
	return s.members[0]
}

// all is every flat index in the space, in list order.
func (s wsSpace) all() []int {
	idx := append([]int{}, s.members...)
	if s.primary >= 0 {
		idx = append(idx, s.primary)
	}
	sort.Ints(idx)
	return idx
}

// moveFlat applies workspace.move's semantics locally: take the row at from
// and put it before the slot insertBefore of the list as it was.
func moveFlat(ws []wsRow, from, insertBefore int) []wsRow {
	row := ws[from]
	out := append(append([]wsRow{}, ws[:from]...), ws[from+1:]...)
	to := insertBefore
	if insertBefore > from {
		to--
	}
	out = append(out, wsRow{})
	copy(out[to+1:], out[to:])
	out[to] = row
	return out
}

// moveBlock applies workspace.move_block locally: the ids, in list order,
// go before the row with id before, or to the end.
func moveBlock(ws []wsRow, ids []string, before string) []wsRow {
	moving := map[string]bool{}
	for _, id := range ids {
		moving[id] = true
	}
	var block, rest []wsRow
	for _, r := range ws {
		if moving[r.ID] {
			block = append(block, r)
		} else {
			rest = append(rest, r)
		}
	}
	at := len(rest)
	for i, r := range rest {
		if r.ID == before {
			at = i
		}
	}
	out := append([]wsRow{}, rest[:at]...)
	out = append(out, block...)
	return append(out, rest[at:]...)
}

// stepNode moves the node one slot up (dir -1) or down (+1) among what it
// can trade places with: a worktree with its siblings, a primary with the
// neighbouring space. It returns the new list and the request that produces
// it on the server; ok is false at the edge.
func stepNode(ws []wsRow, node int, dir int) (out []wsRow, mv wsMove, ok bool) {
	spaces, nodes := buildSpaces(ws)
	if node < 0 || node >= len(nodes) {
		return ws, mv, false
	}
	n := nodes[node]
	sp := spaces[n.space]
	if ws[n.flat].Linked {
		j := 0
		for k, flat := range sp.members {
			if flat == n.flat {
				j = k
			}
		}
		if j+dir < 0 || j+dir >= len(sp.members) {
			return ws, mv, false
		}
		sibling := sp.members[j+dir]
		mv = wsMove{id: ws[n.flat].ID, insertIndex: sibling}
		if dir > 0 {
			mv.insertIndex = sibling + 1
		}
		return moveFlat(ws, n.flat, mv.insertIndex), mv, true
	}
	target := n.space + dir
	if target < 0 || target >= len(spaces) {
		return ws, mv, false
	}
	for _, flat := range sp.all() {
		mv.block = append(mv.block, ws[flat].ID)
	}
	anchor := target
	if dir > 0 {
		anchor = target + 1
	}
	if anchor < len(spaces) {
		mv.before = ws[spaces[anchor].all()[0]].ID
	}
	return moveBlock(ws, mv.block, mv.before), mv, true
}

// ── model ───────────────────────────────────────────────────────────────────

func (m *model) enterSidebar() tea.Cmd {
	m.view = viewSidebar
	m.wsGen++
	m.wsLoading = m.ws == nil
	if m.wsBranch == nil {
		m.wsBranch = map[string]string{}
	}
	return loadWorkspacesCmd(m.wsGen, m.branchCache())
}

// leaveSidebar orphans any poll tick or reply still in flight.
func (m *model) leaveSidebar() {
	m.wsGen++
	m.wsDrag = false
}

// branchCache is a copy for a command goroutine to read and grow; the
// model's own map is only touched in Update.
func (m model) branchCache() map[string]string {
	c := make(map[string]string, len(m.wsBranch))
	for k, v := range m.wsBranch {
		c[k] = v
	}
	return c
}

func (m *model) followWs() {
	_, nodes := buildSpaces(m.ws)
	m.wsCursor = min(m.wsCursor, max(0, len(nodes)-1))
	for i, n := range nodes {
		if m.ws[n.flat].ID == m.wsFollow {
			m.wsCursor = i
		}
	}
}

func (m model) onWsMsg(msg wsMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.wsGen {
		return m, nil
	}
	m.wsLoading = false
	if msg.err != nil {
		m.wsErr = msg.err.Error()
	} else {
		m.wsErr = ""
		m.ws = msg.rows
		for _, r := range msg.rows {
			if !r.Linked {
				m.wsBranch[r.ID] = r.Branch
			}
		}
		m.followWs()
	}
	if m.view != viewSidebar {
		return m, nil
	}
	return m, wsPollCmd(m.wsGen)
}

// step applies one reorder locally and sends it; the reply (or the next
// poll) brings the server's order back.
func (m *model) step(dir int) tea.Cmd {
	out, mv, ok := stepNode(m.ws, m.wsCursor, dir)
	if !ok {
		return nil
	}
	m.ws = out
	m.followWs()
	m.wsGen++
	return mv.cmd(m.wsGen, m.branchCache())
}

func (m *model) setWsCursor(i int) {
	_, nodes := buildSpaces(m.ws)
	m.wsCursor = max(0, min(i, len(nodes)-1))
	if m.wsCursor < len(nodes) {
		m.wsFollow = m.ws[nodes[m.wsCursor].flat].ID
	}
}

func (m model) updateSidebar(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return m.sidebarMouse(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "ctrl+k", "ctrl+p":
			m.setWsCursor(m.wsCursor - 1)
		case "down", "j", "ctrl+j", "ctrl+n":
			m.setWsCursor(m.wsCursor + 1)
		case "K", "shift+up":
			return m, m.step(-1)
		case "J", "shift+down":
			return m, m.step(1)
		case "enter", "ctrl+s":
			return m.pickWs()
		}
	}
	return m, nil
}

func (m model) pickWs() (tea.Model, tea.Cmd) {
	_, nodes := buildSpaces(m.ws)
	if len(nodes) == 0 {
		m.err = "no workspaces"
		return m, nil
	}
	m.wsPicked = m.ws[nodes[m.wsCursor].flat]
	m.mode = "workspace"
	return m, tea.Quit
}

// sidebarMouse: press picks a row up, dragging over another row steps it
// there one trade at a time, release drops it. A wheel moves the cursor.
func (m model) sidebarMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Button == tea.MouseButtonWheelUp:
		m.setWsCursor(m.wsCursor - 1)
		return m, nil
	case msg.Button == tea.MouseButtonWheelDown:
		m.setWsCursor(m.wsCursor + 1)
		return m, nil
	case msg.Action == tea.MouseActionRelease:
		m.wsDrag = false
		return m, nil
	case msg.Button != tea.MouseButtonLeft:
		return m, nil
	}
	target := m.wsNodeAt(msg.Y)
	if target < 0 {
		return m, nil
	}
	if msg.Action == tea.MouseActionPress {
		m.setWsCursor(target)
		m.wsDrag = true
		return m, nil
	}
	if msg.Action != tea.MouseActionMotion || !m.wsDrag {
		return m, nil
	}
	var cmds []tea.Cmd
	for guard := 0; m.wsCursor != target && guard < 32; guard++ {
		dir := 1
		if target < m.wsCursor {
			dir = -1
		}
		cmd := m.step(dir)
		if cmd == nil {
			break
		}
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// ── rendering ───────────────────────────────────────────────────────────────

type wsLineKind int

const (
	lineHeader wsLineKind = iota // an orphaned space's repo name
	linePrimary
	lineBranch
	lineMember
)

type wsLine struct {
	kind wsLineKind
	node int // linePrimary/lineMember: index into nodes; else the space
}

// wsLines lays the tree out one line per row, in the same shape as the
// sidebar: primary, its branch beneath, then ├─/└─ worktrees.
func wsLines(ws []wsRow) []wsLine {
	spaces, nodes := buildSpaces(ws)
	var lines []wsLine
	node := 0
	for s, sp := range spaces {
		if sp.primary >= 0 {
			lines = append(lines, wsLine{kind: linePrimary, node: node})
			node++
			if ws[sp.primary].Branch != "" {
				lines = append(lines, wsLine{kind: lineBranch, node: s})
			}
		} else {
			lines = append(lines, wsLine{kind: lineHeader, node: s})
		}
		for range sp.members {
			lines = append(lines, wsLine{kind: lineMember, node: node})
			node++
		}
	}
	_ = nodes
	return lines
}

// Rows above the first tree line inside the popup: frame padding, the
// section and subtab lines, a blank, and the SPACES header.
const wsListTop = 5

// wsWindow is the slice of lines on screen, scrolled so the cursor's line
// stays visible.
func (m model) wsWindow() (lines []wsLine, start, end int) {
	lines = wsLines(m.ws)
	rows := m.listRows + 3 // no project line or filter on this view
	cursorLine := 0
	for i, l := range lines {
		if (l.kind == linePrimary || l.kind == lineMember) && l.node == m.wsCursor {
			cursorLine = i
		}
	}
	if cursorLine >= rows {
		start = cursorLine - rows + 1
	}
	end = min(len(lines), start+rows)
	return lines, start, end
}

// wsNodeAt maps a screen row to the node drawn there, or -1.
func (m model) wsNodeAt(y int) int {
	lines, start, end := m.wsWindow()
	i := start + y - wsListTop
	if i < start || i >= end {
		return -1
	}
	if l := lines[i]; l.kind == linePrimary || l.kind == lineMember {
		return l.node
	}
	return -1
}

func (m model) sidebarView() string {
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")

	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	section := lipgloss.NewStyle().Foreground(m.styles.dim).Bold(true).Transform(strings.ToUpper)
	b.WriteString(section.Render(" spaces") + "\n")
	switch {
	case m.wsErr != "":
		b.WriteString(m.styles.err.Render("  "+m.wsErr) + "\n")
	case m.wsLoading:
		b.WriteString(dim.Render("  listing workspaces…") + "\n")
	case len(m.ws) == 0:
		b.WriteString(dim.Render("  no workspaces") + "\n")
	default:
		b.WriteString(m.wsTreeView())
	}

	help := "enter focus · j/k move · J/K or drag to reorder · " + tabsHelp + " · esc cancel"
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}

func (m model) wsTreeView() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	accent := lipgloss.NewStyle().Foreground(m.styles.accent)
	selected := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)
	plain := lipgloss.NewStyle().Foreground(m.styles.text)
	here := lipgloss.NewStyle().Foreground(m.styles.text).Bold(true)

	spaces, nodes := buildSpaces(m.ws)
	lines, start, end := m.wsWindow()
	for _, l := range lines[start:end] {
		switch l.kind {
		case lineHeader:
			b.WriteString(dim.Render("    "+spaces[l.node].label) + "\n")
		case lineBranch:
			r := m.ws[spaces[l.node].primary]
			style := dim
			if r.Focused {
				style = accent
			}
			b.WriteString(style.Render("    "+r.Branch) + "\n")
		default:
			n := nodes[l.node]
			r := m.ws[n.flat]
			mark, style := "  ", plain
			switch {
			case l.node == m.wsCursor:
				mark, style = accent.Render("❯ "), selected
			case r.Focused:
				style = here
			}
			var tree string
			if l.kind == lineMember {
				tree = "  ├─ "
				if n.last {
					tree = "  └─ "
				}
			}
			b.WriteString(fmt.Sprintf("%s%s%s %s\n", mark, dim.Render(tree), m.agentGlyph(r.Status), style.Render(r.Label)))
		}
	}
	if start > 0 || end < len(lines) {
		b.WriteString(dim.Render(fmt.Sprintf("  %d–%d of %d", start+1, end, len(lines))) + "\n")
	}
	return b.String()
}
