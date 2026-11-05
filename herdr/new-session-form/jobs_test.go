package main

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

const jobsFixture = `[
  {"name": "jobs-watchdog", "description": "Flag jobs that missed 2 runs", "enabled": true, "mode": null,
   "schedule": "every 30m", "last_start": "2026-11-04T22:21:40-05:00", "last_end": "2026-11-04T22:21:40-05:00",
   "last_ago": "14m ago", "exit": 0, "result": "ok", "summary": "all jobs healthy", "next_run": "2026-11-04T22:51:40-05:00",
   "next_in": "in 15m", "running": false, "log": "/tmp/w.log", "log_tail": ["# started", "all jobs healthy", "# exit 0 after 0s"],
   "session_id": null, "session_cwd": null},
  {"name": "pr-shepherd", "description": "Fix mechanical CI/merge failures on open PRs", "enabled": true, "mode": "dry-run",
   "schedule": "every 10m", "last_start": "2026-11-04T22:30:23-05:00", "last_end": "2026-11-04T22:30:29-05:00",
   "last_ago": "5m ago", "exit": 1, "result": "pinged", "summary": "dry-run: 2 PRs · 2 would act · 1 waiting on you and a long tail that will not fit",
   "next_run": null, "next_in": "in 4m", "running": false, "log": "/tmp/p.log",
   "log_tail": ["- would fix scalafmt", "# exit 1 after 6s"], "session_id": "sess-123", "session_cwd": "/tmp/shepherd"},
  {"name": "review-queue", "description": "Start /review-queue", "enabled": false, "mode": "live",
   "schedule": "every 15m", "last_start": null, "last_end": null, "last_ago": null, "exit": null, "result": null,
   "summary": null, "next_run": null, "next_in": "-", "running": false, "log": null, "log_tail": [],
   "session_id": null, "session_cwd": null}
]`

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// fakeBgExec answers `job ls --json` from the fixture and herdr tab create
// with a pane, recording every call.
func fakeBgExec(t *testing.T, fixture string) *[]string {
	t.Helper()
	var calls []string
	orig := bgExec
	bgExec = func(name string, args ...string) ([]byte, error) {
		call := strings.Join(append([]string{name}, args...), " ")
		calls = append(calls, call)
		switch {
		case call == "job ls --json":
			return []byte(fixture), nil
		case strings.HasPrefix(call, "herdr tab create"):
			return []byte(`{"id":"x","result":{"tab":{"tab_id":"t9","panes":[{"pane_id":"p42"}]}}}`), nil
		}
		return nil, nil
	}
	t.Cleanup(func() { bgExec = orig })
	return &calls
}

func jobsModel(t *testing.T) (get func() model, send func(tea.Msg), calls *[]string) {
	t.Helper()
	threadsTemp(t)
	calls = fakeBgExec(t, jobsFixture)
	get, send = newFormModel(t)
	send(ctrlDigitMsg(6))
	return get, send, calls
}

func runeKey(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func hasCall(calls []string, want string) bool {
	for _, c := range calls {
		if c == want {
			return true
		}
	}
	return false
}

func TestParseBgJobs(t *testing.T) {
	jobs, err := parseBgJobs([]byte(jobsFixture))
	if err != nil || len(jobs) != 3 {
		t.Fatalf("parse: %v %d", err, len(jobs))
	}
	w, p, r := jobs[0], jobs[1], jobs[2]
	if w.Mode != "" || w.modeLabel() != "-" || w.Exit == nil || *w.Exit != 0 || len(w.LogTail) != 3 {
		t.Errorf("watchdog: %+v", w)
	}
	if p.SessionID != "sess-123" || p.SessionCwd != "/tmp/shepherd" || p.Result != "pinged" {
		t.Errorf("shepherd: %+v", p)
	}
	if r.Enabled || r.modeLabel() != "paused" || r.resultLabel() != "-" || r.Exit != nil || r.lastRunLine() != "never run" {
		t.Errorf("review-queue: %+v", r)
	}
	if !strings.HasSuffix(p.lastRunLine(), "· exit 1 · took 6s") {
		t.Errorf("last run line: %q", p.lastRunLine())
	}
	if _, err := parseBgJobs([]byte("nope")); err == nil {
		t.Error("bad JSON should fail")
	}
}

func TestFirstPaneID(t *testing.T) {
	if got := firstPaneID([]byte(`{"result":{"tab":{"panes":[{"pane_id":"p1"}]}}}`)); got != "p1" {
		t.Errorf("got %q", got)
	}
	if got := firstPaneID([]byte(`{"result":{}}`)); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestJobsViewRendersTableAndDetail(t *testing.T) {
	get, send, _ := jobsModel(t)
	if get().view != viewJobs {
		t.Fatalf("view %d, want jobs", get().view)
	}
	send(runeKey("j"))
	plain := ansiRe.ReplaceAllString(get().View(), "")
	for _, want := range []string{
		"6 JOBS", "ctrl+1..6",
		"JOB", "MODE", "LAST", "RESULT", "NEXT", "SUMMARY",
		"❯ pr-shepherd", "dry-run", "pinged", "in 4m",
		"review-queue   paused",
		"Fix mechanical CI/merge failures on open PRs",
		"exit 1 · took 6s · session sess-123",
		"│ - would fix scalafmt", "│ # exit 1 after 6s",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("view missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "a long tail that will not fit") {
		t.Error("summary should be truncated to the width")
	}
	for _, l := range strings.Split(plain, "\n") {
		if strings.Contains(l, "pr-shepherd") && strings.Contains(l, "SUMMARY") {
			t.Error("header and row on one line")
		}
	}
}

func TestJobsSelectionFollowsTheNameAcrossRefreshes(t *testing.T) {
	get, send, _ := jobsModel(t)
	send(runeKey("j"))
	if got := get().bgJobs[get().bgCursor].Name; got != "pr-shepherd" {
		t.Fatalf("cursor on %q", got)
	}
	jobs, _ := parseBgJobs([]byte(jobsFixture))
	reordered := []bgJob{jobs[2], jobs[0], jobs[1]}
	send(bgJobsMsg{gen: get().bgGen, jobs: reordered})
	if got := get().bgJobs[get().bgCursor].Name; got != "pr-shepherd" {
		t.Errorf("after refresh cursor on %q", got)
	}
	send(bgJobsMsg{gen: get().bgGen - 1, jobs: jobs[:1]})
	if len(get().bgJobs) != 3 {
		t.Error("a stale reply should be dropped")
	}
}

func TestJobsModeToLiveAsksFirst(t *testing.T) {
	get, send, calls := jobsModel(t)
	send(runeKey("j"))
	send(runeKey("m"))
	if get().bgConfirm != "pr-shepherd" || hasCall(*calls, "job mode pr-shepherd live") {
		t.Fatalf("m should ask first: confirm %q calls %v", get().bgConfirm, *calls)
	}
	if !strings.Contains(ansiRe.ReplaceAllString(get().View(), ""), "switch pr-shepherd to live? y/n") {
		t.Error("confirm prompt not shown")
	}
	send(runeKey("n"))
	if get().bgConfirm != "" || hasCall(*calls, "job mode pr-shepherd live") {
		t.Error("n should cancel")
	}
	send(runeKey("m"))
	send(tea.KeyMsg{Type: tea.KeyEscape})
	if m := get(); m.bgConfirm != "" || m.view != viewJobs {
		t.Error("esc should cancel the prompt, not the popup")
	}
	send(runeKey("m"))
	send(runeKey("y"))
	if !hasCall(*calls, "job mode pr-shepherd live") {
		t.Errorf("y should switch to live: %v", *calls)
	}
	if got := get().bgStatus; got != "pr-shepherd -> live" {
		t.Errorf("status %q", got)
	}
}

func TestJobsModeToDryRunNeedsNoConfirm(t *testing.T) {
	get, send, calls := jobsModel(t)
	send(runeKey("j"))
	send(runeKey("j"))
	send(runeKey("m"))
	if get().bgConfirm != "" || !hasCall(*calls, "job mode review-queue dry-run") {
		t.Errorf("live -> dry-run should not ask: %v", *calls)
	}
	send(runeKey("k"))
	send(runeKey("k"))
	n := len(*calls)
	send(runeKey("m"))
	if get().bgConfirm != "" || len(*calls) != n || !strings.Contains(get().bgStatus, "no dry-run/live mode") {
		t.Errorf("a job without a mode: status %q", get().bgStatus)
	}
}

func TestJobsPauseToggles(t *testing.T) {
	get, send, calls := jobsModel(t)
	send(runeKey("p"))
	if !hasCall(*calls, "job pause jobs-watchdog") {
		t.Errorf("p on an enabled job should pause: %v", *calls)
	}
	send(runeKey("j"))
	send(runeKey("j"))
	send(runeKey("p"))
	if !hasCall(*calls, "job resume review-queue") {
		t.Errorf("p on a paused job should resume: %v", *calls)
	}
	if got := get().bgStatus; got != "resumed review-queue" {
		t.Errorf("status %q", got)
	}
}

func TestJobsKickRefreshes(t *testing.T) {
	_, send, calls := jobsModel(t)
	before := strings.Count(strings.Join(*calls, "\n"), "job ls --json")
	send(runeKey("r"))
	if !hasCall(*calls, "job kick jobs-watchdog") {
		t.Errorf("r should kick: %v", *calls)
	}
	if after := strings.Count(strings.Join(*calls, "\n"), "job ls --json"); after != before+1 {
		t.Errorf("kick should refresh: %d -> %d", before, after)
	}
}

func TestJobsOpenSession(t *testing.T) {
	get, send, calls := jobsModel(t)
	send(runeKey("o"))
	if got := get().bgStatus; got != "no Claude session recorded for this run" {
		t.Errorf("status %q", got)
	}
	if hasCall(*calls, "herdr-agent-resume p42 sess-123 pr-shepherd") {
		t.Error("resumed without a session")
	}
	send(runeKey("j"))
	send(runeKey("o"))
	if !hasCall(*calls, "herdr tab create --cwd /tmp/shepherd --label pr-shepherd --env HERDR_THREAD=1 --focus") ||
		!hasCall(*calls, "herdr-agent-resume p42 sess-123 pr-shepherd") {
		t.Errorf("o should resume in a new tab: %v", *calls)
	}
}

func TestJobsEnterFollowsTheLog(t *testing.T) {
	_, send, calls := jobsModel(t)
	send(tea.KeyMsg{Type: tea.KeyEnter})
	if !hasCall(*calls, "herdr pane run p42 job logs jobs-watchdog -f") {
		t.Errorf("enter should follow the log: %v", *calls)
	}
}
