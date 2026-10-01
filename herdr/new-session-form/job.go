package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// A job is one launch, made durable by the form before the popup closes: the
// prompt exists nowhere else, so it goes to $HERDR_THREADS_DIR/jobs/<id>.json
// (written atomically, never rewritten) with a `queued` line appended to
// <id>.events, and only then is the draft cleared and the result printed.
// The runner claims a job by mkdir'ing <id>.claim and appends its own stages;
// the last line of the events file is the job's state.
type job struct {
	ID        string   `json:"id"`
	CreatedAt string   `json:"created_at"` // RFC3339
	Mode      string   `json:"mode"`
	Kind      string   `json:"kind,omitempty"`
	Project   string   `json:"project"`
	Name      string   `json:"name"`
	Branch    string   `json:"branch"`
	Prompt    string   `json:"prompt"`
	Attach    []string `json:"attachments"`
	SessionID string   `json:"session_id"`
}

// jobModes are the launches worth a spec; focus/agent/workspace only move the
// cursor to something that is running already.
var jobModes = map[string]bool{
	"open": true, "bg": true,
	"feature": true, "feature-bg": true,
	"resume": true, "resume-bg": true,
}

// Keys for the stranded-launch notice, free in every view (and not readline
// keys inside the inputs, unlike ctrl+a/e/k/u/w).
const (
	retryKey = "ctrl+g"
	editKey  = "ctrl+o"
)

const eventLayout = "2006-01-02 15:04:05"

func jobsDir() string { return filepath.Join(threadsCacheDir(), "jobs") }

var nonSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

func jobID(name string, now time.Time) string {
	slug := strings.Trim(nonSlugRe.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	if slug == "" {
		slug = "session"
	}
	var b [2]byte
	_, _ = rand.Read(b[:])
	return now.Format("20060102-150405") + "-" + slug + "-" + hex.EncodeToString(b[:])
}

// writeJob writes the spec through a temp file in the same directory: fsync
// the contents, rename over the final name, then fsync the directory so the
// rename itself survives a crash.
func writeJob(j job) error {
	if err := os.MkdirAll(jobsDir(), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	final := filepath.Join(jobsDir(), j.ID+".json")
	tmp := final + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		return err
	}
	if d, err := os.Open(jobsDir()); err == nil {
		d.Sync()
		d.Close()
	}
	return nil
}

func appendEvent(id, stage, detail string) error {
	if err := os.MkdirAll(jobsDir(), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(jobsDir(), id+".events"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "%s\t%s\t%s\n", time.Now().Format(eventLayout), stage, detail); err != nil {
		return err
	}
	return f.Sync()
}

// queueJob turns a submitted form into a job on disk. It returns the id the
// result JSON carries to the launcher.
func queueJob(res result, now time.Time) (string, error) {
	j := job{
		ID:        jobID(res.Name, now),
		CreatedAt: now.Format(time.RFC3339),
		Mode:      res.Mode,
		Kind:      res.Kind,
		Project:   res.Project,
		Name:      res.Name,
		Branch:    res.Branch,
		Prompt:    res.Prompt,
		Attach:    res.Attach,
		SessionID: res.SessionID,
	}
	if j.Attach == nil {
		j.Attach = []string{}
	}
	if err := writeJob(j); err != nil {
		return "", err
	}
	if err := appendEvent(j.ID, "queued", ""); err != nil {
		return "", err
	}
	return j.ID, nil
}

// lastEvent is the job's current state: the stage of the last line of its
// events file, and when that line was written.
func lastEvent(id string) (string, time.Time, bool) {
	data, err := os.ReadFile(filepath.Join(jobsDir(), id+".events"))
	if err != nil {
		return "", time.Time{}, false
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		parts := strings.Split(lines[i], "\t")
		if len(parts) < 2 {
			continue
		}
		at, err := time.ParseInLocation(eventLayout, parts[0], time.Local)
		if err != nil {
			at = time.Time{}
		}
		return parts[1], at, true
	}
	return "", time.Time{}, false
}

// A strandedJob is a launch whose text reached disk but which never ran:
// still `queued` after the grace period with nobody having claimed it, or
// `failed`. Older than a day and it is no longer worth offering.
type strandedJob struct {
	job
	stage   string
	created time.Time
}

const (
	strandedAfter  = 10 * time.Second
	strandedWindow = 24 * time.Hour
)

func strandedJobs(now time.Time) []strandedJob {
	entries, err := os.ReadDir(jobsDir())
	if err != nil {
		return nil
	}
	var out []strandedJob
	for _, e := range entries {
		// Only the specs; jobs/ also holds events, claims, logs and the
		// runner's own marker files.
		id, ok := strings.CutSuffix(e.Name(), ".json")
		if !ok || e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(jobsDir(), e.Name()))
		if err != nil {
			continue
		}
		var j job
		if json.Unmarshal(data, &j) != nil {
			continue
		}
		created, err := time.Parse(time.RFC3339, j.CreatedAt)
		if err != nil || now.Sub(created) > strandedWindow {
			continue
		}
		stage, at, ok := lastEvent(id)
		if !ok {
			continue
		}
		switch stage {
		case "failed":
		case "queued":
			// The events file's mtime is the finer clock: event lines
			// only carry whole seconds.
			if st, err := os.Stat(filepath.Join(jobsDir(), id+".events")); err == nil {
				at = st.ModTime()
			}
			if now.Sub(at) < strandedAfter {
				continue
			}
			if _, err := os.Stat(filepath.Join(jobsDir(), id+".claim")); err == nil {
				continue
			}
		default:
			continue
		}
		j.ID = id
		out = append(out, strandedJob{job: j, stage: stage, created: created})
	}
	sort.Slice(out, func(i, k int) bool { return out[i].created.After(out[k].created) })
	return out
}

// strandedNotice names the newest one; only that one is actionable.
func strandedNotice(jobs []strandedJob) string {
	if len(jobs) == 0 {
		return ""
	}
	what := "1 launch never started"
	if len(jobs) > 1 {
		what = fmt.Sprintf("%d launches never started", len(jobs))
	}
	return fmt.Sprintf("%s: %q · %s retry · %s edit", what, jobs[0].Name, retryKey, editKey)
}
