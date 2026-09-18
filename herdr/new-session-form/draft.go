package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// A draft is the new-session form's contents, written on every change so a
// prompt outlives the popup: an esc, a crash, or a launcher that dies before
// saving the prompt file. The launcher (herdr-new-thread --launch) deletes it
// once the prompt is on disk; the next popup restores whatever is left.
type draft struct {
	Project string    `json:"project"`
	Name    string    `json:"name"`
	Prompt  string    `json:"prompt"`
	SavedAt time.Time `json:"saved_at"`
}

// restoreDrafts is off in tests, which would otherwise inherit each other's.
var restoreDrafts = true

func (d draft) empty() bool {
	return strings.TrimSpace(d.Name) == "" && strings.TrimSpace(d.Prompt) == ""
}

func (d draft) same(o draft) bool {
	return d.Project == o.Project && d.Name == o.Name && d.Prompt == o.Prompt
}

func threadsCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "herdr-threads")
}

func draftPath() string { return filepath.Join(threadsCacheDir(), "draft.json") }

func loadDraft() (draft, bool) {
	var d draft
	data, err := os.ReadFile(draftPath())
	if err != nil || json.Unmarshal(data, &d) != nil || d.empty() {
		return draft{}, false
	}
	return d, true
}

func saveDraft(d draft) error {
	if d.empty() {
		return clearDraft()
	}
	d.SavedAt = time.Now()
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(threadsCacheDir(), 0o755); err != nil {
		return err
	}
	tmp := draftPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, draftPath())
}

func clearDraft() error {
	if err := os.Remove(draftPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// formLog records how each popup ended (and every esc it saw), so a launch
// that never happened can be told apart from one that failed later on.
func formLog(format string, args ...any) {
	f, err := os.OpenFile(filepath.Join(threadsCacheDir(), "launch.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s [form:%d] %s\n", time.Now().Format("2006-01-02 15:04:05"), os.Getpid(), fmt.Sprintf(format, args...))
}
