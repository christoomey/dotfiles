package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Jobs view: the background jobs registry (the `job` CLI over
// claude-config/jobs) as a table, with the selected job's last run under it,
// refreshed every few seconds while it shows. r runs a job now, p pauses or
// resumes it, m flips dry-run and live (asking first before going live),
// enter follows its log in a new tab, and o resumes the Claude session its
// last run recorded. The popup stays open throughout.

type bgJob struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	Mode        string   `json:"mode"`
	Schedule    string   `json:"schedule"`
	LastStart   string   `json:"last_start"`
	LastEnd     string   `json:"last_end"`
	LastAgo     string   `json:"last_ago"`
	Exit        *int     `json:"exit"`
	Result      string   `json:"result"`
	Summary     string   `json:"summary"`
	NextIn      string   `json:"next_in"`
	Running     bool     `json:"running"`
	Log         string   `json:"log"`
	LogTail     []string `json:"log_tail"`
	SessionID   string   `json:"session_id"`
	SessionCwd  string   `json:"session_cwd"`
}

type bgJobsMsg struct {
	gen  int
	jobs []bgJob
	err  error
}

type bgJobsPollMsg struct{ gen int }

// bgJobDoneMsg is one action's outcome; reload refreshes the table after it.
type bgJobDoneMsg struct {
	status string
	err    error
	reload bool
}

type bgStatusClearMsg struct{ seq int }

const (
	bgJobsPollEvery = 3 * time.Second
	bgStatusFor     = 6 * time.Second
)

// bgExec runs a command and returns its stdout, with stderr as the error
// text when it fails. Tests swap it out.
var bgExec = func(name string, args ...string) ([]byte, error) {
	out, err := exec.Command(name, args...).Output()
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		err = errors.New(strings.TrimSpace(string(ee.Stderr)))
	}
	return out, err
}

func parseBgJobs(raw []byte) ([]bgJob, error) {
	var jobs []bgJob
	if err := json.Unmarshal(raw, &jobs); err != nil {
		return nil, fmt.Errorf("job ls --json: %w", err)
	}
	return jobs, nil
}

func loadBgJobsCmd(gen int) tea.Cmd {
	return func() tea.Msg {
		out, err := bgExec("job", "ls", "--json")
		if err != nil {
			return bgJobsMsg{gen: gen, err: fmt.Errorf("job ls: %w", err)}
		}
		jobs, err := parseBgJobs(out)
		return bgJobsMsg{gen: gen, jobs: jobs, err: err}
	}
}

func bgJobsPollCmd(gen int) tea.Cmd {
	return tea.Tick(bgJobsPollEvery, func(time.Time) tea.Msg { return bgJobsPollMsg{gen: gen} })
}

func bgJobCLICmd(status string, args ...string) tea.Cmd {
	return func() tea.Msg {
		if _, err := bgExec("job", args...); err != nil {
			return bgJobDoneMsg{err: fmt.Errorf("job %s: %w", strings.Join(args, " "), err), reload: true}
		}
		return bgJobDoneMsg{status: status, reload: true}
	}
}

// openBgTab opens a focused tab and returns its pane.
func openBgTab(args ...string) (string, error) {
	out, err := bgExec("herdr", append([]string{"tab", "create"}, args...)...)
	if err != nil {
		return "", fmt.Errorf("herdr tab create: %w", err)
	}
	pane := firstPaneID(out)
	if pane == "" {
		return "", errors.New("herdr tab create returned no pane")
	}
	return pane, nil
}

func followBgLogCmd(name string) tea.Cmd {
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		pane, err := openBgTab("--cwd", home, "--label", "log: "+name, "--focus")
		if err == nil {
			_, err = bgExec("herdr", "pane", "run", pane, "job logs "+name+" -f")
		}
		if err != nil {
			return bgJobDoneMsg{err: err}
		}
		return tea.QuitMsg{}
	}
}

// resumeBgSessionCmd opens the session the way herdr-open-chat does: a tab
// in its cwd, then herdr-agent-resume so herdr tracks it as an agent.
func resumeBgSessionCmd(j bgJob) tea.Cmd {
	return func() tea.Msg {
		cwd := j.SessionCwd
		if cwd == "" {
			cwd, _ = os.UserHomeDir()
		}
		pane, err := openBgTab("--cwd", cwd, "--label", j.Name, "--env", "HERDR_THREAD=1", "--focus")
		if err == nil {
			_, err = bgExec("herdr-agent-resume", pane, j.SessionID, j.Name)
		}
		if err != nil {
			return bgJobDoneMsg{err: err}
		}
		return tea.QuitMsg{}
	}
}

func firstPaneID(raw []byte) string {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return ""
	}
	return findPaneID(v)
}

func findPaneID(v any) string {
	switch t := v.(type) {
	case map[string]any:
		if s, ok := t["pane_id"].(string); ok && s != "" {
			return s
		}
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if s := findPaneID(t[k]); s != "" {
				return s
			}
		}
	case []any:
		for _, x := range t {
			if s := findPaneID(x); s != "" {
				return s
			}
		}
	}
	return ""
}

// ── model ───────────────────────────────────────────────────────────────────

func (m *model) enterJobs() tea.Cmd {
	m.view = viewJobs
	m.bgGen++
	m.bgLoading = m.bgJobs == nil
	return loadBgJobsCmd(m.bgGen)
}

// leaveJobs orphans any poll tick or reply still in flight.
func (m *model) leaveJobs() {
	m.bgGen++
	m.bgConfirm = ""
}

func (m *model) followBgJob() {
	m.bgCursor = min(m.bgCursor, max(0, len(m.bgJobs)-1))
	for i, j := range m.bgJobs {
		if j.Name == m.bgFollow {
			m.bgCursor = i
		}
	}
}

func (m *model) setBgCursor(i int) {
	m.bgCursor = max(0, min(i, len(m.bgJobs)-1))
	if m.bgCursor < len(m.bgJobs) {
		m.bgFollow = m.bgJobs[m.bgCursor].Name
	}
}

func (m model) selectedBgJob() (bgJob, bool) {
	if m.bgCursor < len(m.bgJobs) {
		return m.bgJobs[m.bgCursor], true
	}
	return bgJob{}, false
}

func (m model) onBgJobsMsg(msg bgJobsMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.bgGen {
		return m, nil
	}
	m.bgLoading = false
	if msg.err != nil {
		m.bgErr = msg.err.Error()
	} else {
		m.bgErr = ""
		m.bgJobs = msg.jobs
		if m.bgFollow == "" && len(m.bgJobs) > 0 {
			m.bgFollow = m.bgJobs[0].Name
		}
		m.followBgJob()
	}
	if m.view != viewJobs {
		return m, nil
	}
	return m, bgJobsPollCmd(m.bgGen)
}

// bgNote sets the transient status line and schedules its clearing.
func (m *model) bgNote(text string, isErr bool) tea.Cmd {
	m.bgStatus, m.bgStatusErr = text, isErr
	m.bgStatusSeq++
	seq := m.bgStatusSeq
	return tea.Tick(bgStatusFor, func(time.Time) tea.Msg { return bgStatusClearMsg{seq: seq} })
}

func (m model) onBgJobDone(msg bgJobDoneMsg) (tea.Model, tea.Cmd) {
	var note tea.Cmd
	if msg.err != nil {
		note = m.bgNote(msg.err.Error(), true)
	} else {
		note = m.bgNote(msg.status, false)
	}
	if !msg.reload || m.view != viewJobs {
		return m, note
	}
	m.bgGen++
	return m, tea.Batch(note, loadBgJobsCmd(m.bgGen))
}

func (m model) updateJobs(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.bgConfirm != "" {
		name := m.bgConfirm
		m.bgConfirm = ""
		if k.String() == "y" || k.String() == "Y" {
			return m, tea.Batch(m.bgNote("switching "+name+" to live…", false), bgJobCLICmd(name+" -> live", "mode", name, "live"))
		}
		return m, m.bgNote(name+" stays on dry-run", false)
	}
	switch k.String() {
	case "up", "k", "ctrl+k", "ctrl+p":
		m.setBgCursor(m.bgCursor - 1)
		return m, nil
	case "down", "j", "ctrl+j", "ctrl+n":
		m.setBgCursor(m.bgCursor + 1)
		return m, nil
	}
	j, ok := m.selectedBgJob()
	if !ok {
		return m, nil
	}
	switch k.String() {
	case "enter":
		return m, followBgLogCmd(j.Name)
	case "r":
		return m, tea.Batch(m.bgNote("kicking "+j.Name+"…", false), bgJobCLICmd("kicked "+j.Name, "kick", j.Name))
	case "p":
		if j.Enabled {
			return m, tea.Batch(m.bgNote("pausing "+j.Name+"…", false), bgJobCLICmd("paused "+j.Name, "pause", j.Name))
		}
		return m, tea.Batch(m.bgNote("resuming "+j.Name+"…", false), bgJobCLICmd("resumed "+j.Name, "resume", j.Name))
	case "m":
		switch j.Mode {
		case "live":
			return m, tea.Batch(m.bgNote("switching "+j.Name+" to dry-run…", false), bgJobCLICmd(j.Name+" -> dry-run", "mode", j.Name, "dry-run"))
		case "dry-run":
			m.bgConfirm = j.Name
			return m, nil
		}
		return m, m.bgNote(j.Name+" has no dry-run/live mode", false)
	case "o":
		if j.SessionID == "" {
			return m, m.bgNote("no Claude session recorded for this run", false)
		}
		return m, resumeBgSessionCmd(j)
	}
	return m, nil
}

// ── rendering ───────────────────────────────────────────────────────────────

func (m model) bgResultStyle(result string) lipgloss.Style {
	switch result {
	case "ok":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1"))
	case "pinged":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387"))
	case "error", "broken":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))
	case "running":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8"))
}

func (j bgJob) modeLabel() string {
	switch {
	case !j.Enabled:
		return "paused"
	case j.Mode != "":
		return j.Mode
	}
	return "-"
}

func (j bgJob) resultLabel() string {
	switch {
	case j.Running:
		return "running"
	case j.Result != "":
		return j.Result
	}
	return "-"
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func padCell(s string, w int) string {
	s = ellipsize(s, max(1, w))
	return s + strings.Repeat(" ", max(0, w-len([]rune(s))))
}

const (
	bgModeW   = 7
	bgLastW   = 8
	bgResultW = 7
	bgNextW   = 7
)

func (m model) jobsWidth() int { return max(60, m.width-4) }

// jobsWindow is the slice of rows on screen, scrolled to keep the cursor in.
func (m model) jobsWindow() (start, end int) {
	rows := max(5, m.listRows-8)
	if m.bgCursor >= rows {
		start = m.bgCursor - rows + 1
	}
	return start, min(len(m.bgJobs), start+rows)
}

func (m model) jobsTableView() string {
	var b strings.Builder
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	accent := lipgloss.NewStyle().Foreground(m.styles.accent)
	selected := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)
	plain := lipgloss.NewStyle().Foreground(m.styles.text)
	section := lipgloss.NewStyle().Foreground(m.styles.dim).Bold(true)

	nameW := 3
	for _, j := range m.bgJobs {
		nameW = max(nameW, len([]rune(j.Name)))
	}
	nameW = min(nameW, 18)
	sumW := max(8, m.jobsWidth()-2-nameW-bgModeW-bgLastW-bgResultW-bgNextW-10)

	cells := func(c ...string) string {
		widths := []int{nameW, bgModeW, bgLastW, bgResultW, bgNextW, sumW}
		out := make([]string, len(c))
		for i := range c {
			out[i] = padCell(c[i], widths[i])
		}
		return strings.TrimRight(strings.Join(out, "  "), " ")
	}
	b.WriteString(section.Render("  "+cells("JOB", "MODE", "LAST", "RESULT", "NEXT", "SUMMARY")) + "\n")

	start, end := m.jobsWindow()
	for i := start; i < end; i++ {
		j := m.bgJobs[i]
		mark, nameStyle, rowStyle := "  ", plain, plain
		if !j.Enabled {
			nameStyle, rowStyle = dim, dim
		}
		if i == m.bgCursor {
			mark, nameStyle = accent.Render("❯ "), selected
		}
		resultStyle := m.bgResultStyle(j.resultLabel())
		if !j.Enabled {
			resultStyle = dim
		}
		row := mark + nameStyle.Render(padCell(j.Name, nameW)) + "  " +
			rowStyle.Render(padCell(j.modeLabel(), bgModeW)) + "  " +
			rowStyle.Render(padCell(orDash(j.LastAgo), bgLastW)) + "  " +
			resultStyle.Render(padCell(j.resultLabel(), bgResultW)) + "  " +
			rowStyle.Render(padCell(orDash(j.NextIn), bgNextW)) + "  " +
			dim.Render(ellipsize(j.Summary, sumW))
		b.WriteString(strings.TrimRight(row, " ") + "\n")
	}
	if start > 0 || end < len(m.bgJobs) {
		b.WriteString(dim.Render(fmt.Sprintf("  %d–%d of %d", start+1, end, len(m.bgJobs))) + "\n")
	}
	return b.String()
}

// lastRunLine is when the last run started, its exit code, and how long it
// took when both ends are known.
func (j bgJob) lastRunLine() string {
	start, err := time.Parse(time.RFC3339, j.LastStart)
	if err != nil {
		return "never run"
	}
	if j.Running {
		return "running since " + start.Local().Format("Mon 15:04")
	}
	parts := []string{"last run " + start.Local().Format("Mon 15:04")}
	if j.Exit != nil {
		parts = append(parts, fmt.Sprintf("exit %d", *j.Exit))
	}
	if end, err := time.Parse(time.RFC3339, j.LastEnd); err == nil && !end.Before(start) {
		parts = append(parts, "took "+end.Sub(start).Round(time.Second).String())
	}
	return strings.Join(parts, " · ")
}

func (m model) jobDetailView(j bgJob) string {
	var b strings.Builder
	w := m.jobsWidth()
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	plain := lipgloss.NewStyle().Foreground(m.styles.text)
	name := lipgloss.NewStyle().Foreground(m.styles.accent).Bold(true)

	b.WriteString(dim.Render(strings.Repeat("─", w)) + "\n")
	head := name.Render(j.Name)
	if j.Schedule != "" {
		head += dim.Render(" · " + j.Schedule)
	}
	b.WriteString(head + "\n")
	if j.Description != "" {
		b.WriteString(plain.Width(w).Render(j.Description) + "\n")
	}
	meta := j.lastRunLine()
	if j.SessionID != "" {
		meta += " · session " + j.SessionID
	}
	b.WriteString(dim.Render(meta) + "\n")
	for _, l := range j.LogTail {
		b.WriteString(dim.Render("  │ "+ellipsize(l, w-4)) + "\n")
	}
	return b.String()
}

func (m model) jobsView() string {
	var b strings.Builder
	b.WriteString(m.tabsView() + "\n\n")
	dim := lipgloss.NewStyle().Foreground(m.styles.dim)
	switch {
	case m.bgErr != "":
		b.WriteString(m.styles.err.Render("  "+m.bgErr) + "\n")
	case m.bgLoading:
		b.WriteString(dim.Render("  listing jobs…") + "\n")
	case len(m.bgJobs) == 0:
		b.WriteString(dim.Render("  no jobs") + "\n")
	default:
		b.WriteString(m.jobsTableView() + "\n")
		if j, ok := m.selectedBgJob(); ok {
			b.WriteString(m.jobDetailView(j))
		}
	}

	switch {
	case m.bgConfirm != "":
		b.WriteString("\n" + m.styles.err.Render("switch "+m.bgConfirm+" to live? y/n"))
	case m.bgStatus != "" && m.bgStatusErr:
		b.WriteString("\n" + m.styles.err.Render(m.bgStatus))
	case m.bgStatus != "":
		b.WriteString("\n" + m.styles.attach.Render(m.bgStatus))
	}

	help := "enter log · r run · p pause · m mode · o session · " + tabsHelp + " · esc cancel"
	b.WriteString("\n" + m.styles.help.Render(help))
	return b.String()
}
