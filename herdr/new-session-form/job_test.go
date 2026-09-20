package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// threadsTemp points the cache dir (draft, jobs, launch log) at a temp dir
// for one test, so nothing touches the real one.
func threadsTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HERDR_THREADS_DIR", dir)
	return dir
}

// withRestore turns draft restoring back on for one test.
func withRestore(t *testing.T) {
	t.Helper()
	restoreDrafts = true
	t.Cleanup(func() { restoreDrafts = false })
}

// writeJobAt plants a job on disk as a previous popup would have left it.
func writeJobAt(t *testing.T, j job, created time.Time, events ...[2]string) string {
	t.Helper()
	j.ID = jobID(j.Name, created)
	j.CreatedAt = created.Format(time.RFC3339)
	if err := writeJob(j); err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, e := range events {
		lines = append(lines, e[0]+"\t"+e[1]+"\t")
	}
	if len(lines) > 0 {
		path := filepath.Join(jobsDir(), j.ID+".events")
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		// The scan reads the file's mtime, not just the last line.
		last, err := time.ParseInLocation(eventLayout, events[len(events)-1][0], time.Local)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, last, last); err != nil {
			t.Fatal(err)
		}
	}
	return j.ID
}

func stamp(at time.Time) string { return at.Format(eventLayout) }

func TestQueueJobWritesTheSpecAndTheQueuedEvent(t *testing.T) {
	dir := threadsTemp(t)
	now := time.Date(2026, 9, 19, 21, 4, 5, 0, time.Local)
	id, err := queueJob(result{Mode: "open", Project: "moi", Name: "Dose Unit Case Bug!", Prompt: "go look"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^20260919-210405-dose-unit-case-bug-[0-9a-f]{4}$`).MatchString(id) {
		t.Errorf("job id %q", id)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "jobs", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"attachments":[]`) {
		t.Errorf("attachments should be an empty list: %s", raw)
	}
	var j job
	if err := json.Unmarshal(raw, &j); err != nil {
		t.Fatal(err)
	}
	if j.ID != id || j.Mode != "open" || j.Project != "moi" || j.Prompt != "go look" || j.Name != "Dose Unit Case Bug!" {
		t.Errorf("spec %+v", j)
	}
	if _, err := time.Parse(time.RFC3339, j.CreatedAt); err != nil {
		t.Errorf("created_at %q: %v", j.CreatedAt, err)
	}

	stage, _, ok := lastEvent(id)
	if !ok || stage != "queued" {
		t.Errorf("last event %q (%v)", stage, ok)
	}
	// The atomic write leaves nothing behind.
	entries, _ := os.ReadDir(filepath.Join(dir, "jobs"))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestJobIDSlugIsCapped(t *testing.T) {
	threadsTemp(t)
	now := time.Now()
	id := jobID(strings.Repeat("long name ", 12), now)
	slug := strings.TrimSuffix(strings.TrimPrefix(id, now.Format("20060102-150405")+"-"), id[len(id)-5:])
	if len(slug) > 40 {
		t.Errorf("slug %q is %d chars", slug, len(slug))
	}
	if got := jobID("!!!", now); !strings.Contains(got, "-session-") {
		t.Errorf("nameless job: %q", got)
	}
}

func TestSubmitWritesTheJobBeforeClearingTheDraft(t *testing.T) {
	dir := threadsTemp(t)
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "go look into it")
	if _, ok := loadDraft(); !ok {
		t.Fatal("nothing autosaved")
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlS})
	m := get()
	if m.mode != "open" || m.jobID == "" {
		t.Fatalf("submit: mode %q job %q", m.mode, m.jobID)
	}
	if _, ok := loadDraft(); ok {
		t.Error("draft should be gone once the spec is on disk")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "jobs", m.jobID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var j job
	json.Unmarshal(raw, &j)
	if j.Name != "dose bug" || j.Prompt != "go look into it" || j.Mode != "open" {
		t.Errorf("spec %+v", j)
	}
	if res := m.resultFor(m.mode); res.Name != "dose bug" || res.Prompt != "go look into it" {
		t.Errorf("result %+v", res)
	}
}

func TestSubmitFailureKeepsTheFormAndTheText(t *testing.T) {
	dir := threadsTemp(t)
	// A file where jobs/ belongs: the spec cannot be written.
	if err := os.WriteFile(filepath.Join(dir, "jobs"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	get, send := newFormModel(t)
	typeText(send, "dose bug")
	send(tea.KeyMsg{Type: tea.KeyTab})
	typeText(send, "go look into it")
	send(tea.KeyMsg{Type: tea.KeyCtrlS})

	m := get()
	if m.mode != "" {
		t.Fatalf("form quit on a failed write: mode %q", m.mode)
	}
	if m.err == "" {
		t.Error("no error shown")
	}
	if m.chatName.Value() != "dose bug" || m.chatPrompt.Value() != "go look into it" {
		t.Errorf("text lost: %q %q", m.chatName.Value(), m.chatPrompt.Value())
	}
	if d, ok := loadDraft(); !ok || d.Prompt != "go look into it" {
		t.Errorf("draft not kept: %+v %v", d, ok)
	}
}

func TestStrandedJobsAreTheLaunchesThatNeverRan(t *testing.T) {
	threadsTemp(t)
	now := time.Now()
	old := now.Add(-30 * time.Second)

	queued := writeJobAt(t, job{Mode: "open", Name: "never started"}, old, [2]string{stamp(old), "queued"})
	failed := writeJobAt(t, job{Mode: "feature", Name: "blew up"}, now.Add(-time.Minute),
		[2]string{stamp(old), "queued"}, [2]string{stamp(now), "failed"})
	writeJobAt(t, job{Mode: "open", Name: "just queued"}, now, [2]string{stamp(now), "queued"})
	writeJobAt(t, job{Mode: "open", Name: "finished"}, old, [2]string{stamp(old), "queued"}, [2]string{stamp(now), "done"})
	writeJobAt(t, job{Mode: "open", Name: "yesterday"}, now.Add(-25*time.Hour), [2]string{stamp(now.Add(-25 * time.Hour)), "queued"})
	claimed := writeJobAt(t, job{Mode: "open", Name: "claimed"}, old, [2]string{stamp(old), "queued"})
	if err := os.Mkdir(filepath.Join(jobsDir(), claimed+".claim"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := strandedJobs(now)
	var ids []string
	for _, s := range got {
		ids = append(ids, s.ID)
	}
	if len(ids) != 2 || ids[0] != queued || ids[1] != failed {
		t.Fatalf("stranded %v, want [%s %s]", ids, queued, failed)
	}
	if n := strandedNotice(got); !strings.Contains(n, "2 launches never started") ||
		!strings.Contains(n, "never started") || !strings.Contains(n, retryKey) || !strings.Contains(n, editKey) {
		t.Errorf("notice %q", n)
	}
}

func TestRetryReportsTheStrandedJob(t *testing.T) {
	threadsTemp(t)
	old := time.Now().Add(-time.Minute)
	id := writeJobAt(t, job{Mode: "open", Name: "never started"}, old, [2]string{stamp(old), "queued"})

	get, send := newFormModel(t)
	if n := get().jobNotice; !strings.Contains(n, "never started") {
		t.Fatalf("no notice: %q", n)
	}
	send(tea.KeyMsg{Type: tea.KeyCtrlG})
	m := get()
	if m.mode != "retry" || m.jobID != id {
		t.Fatalf("retry: mode %q job %q", m.mode, m.jobID)
	}
	res := m.resultFor(m.mode)
	res.JobID = m.jobID
	out, _ := json.Marshal(res)
	if string(out) != `{"mode":"retry","job_id":"`+id+`"}` {
		t.Errorf("result %s", out)
	}
}

func TestEditingAStrandedFeatureJobSupersedesIt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	threadsTemp(t)
	old := time.Now().Add(-time.Minute)
	id := writeJobAt(t, job{Mode: "feature", Project: "moi", Branch: "christoomey/toom-9-fix-thing",
		Name: "fix-thing", Prompt: "/linear-review TOOM-9"}, old, [2]string{stamp(old), "queued"})

	get, send := newFormModel(t)
	send(tea.KeyMsg{Type: tea.KeyCtrlO})
	m := get()
	if m.view != viewFeature || m.feature.Value() != "christoomey/toom-9-fix-thing" ||
		m.featureName.Value() != "fix-thing" || m.featurePrompt.Value() != "/linear-review TOOM-9" {
		t.Fatalf("job not loaded: view %d %q %q %q", m.view, m.feature.Value(), m.featureName.Value(), m.featurePrompt.Value())
	}
	if m.jobNotice != "" {
		t.Error("notice should go once the job is being edited")
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlB})
	m = get()
	if m.mode != "feature-bg" || m.jobID == "" || m.jobID == id {
		t.Fatalf("resubmit: mode %q job %q", m.mode, m.jobID)
	}
	stage, _, _ := lastEvent(id)
	if stage != "superseded" {
		t.Errorf("old job stage %q", stage)
	}
	events, _ := os.ReadFile(filepath.Join(jobsDir(), id+".events"))
	if !strings.Contains(string(events), m.jobID) {
		t.Errorf("superseded event should name the new job: %s", events)
	}
}

func TestFeatureRefusesAnExistingWorktree(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	threadsTemp(t)
	if err := os.MkdirAll(filepath.Join(home, "code", "moi__worktrees", "fix-thing"), 0o755); err != nil {
		t.Fatal(err)
	}

	get, send := newFormModel(t)
	send(tea.KeyMsg{Type: tea.KeyCtrlF})
	typeText(send, "christoomey/toom-9-fix-thing")
	send(tea.KeyMsg{Type: tea.KeyCtrlS})

	m := get()
	if m.mode != "" {
		t.Fatalf("launched over an existing worktree: mode %q", m.mode)
	}
	if !strings.Contains(m.err, "exists already") {
		t.Errorf("err %q", m.err)
	}
}
