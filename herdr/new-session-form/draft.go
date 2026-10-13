package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// A draft is the form's contents, chat fields and feature fields alike,
// written on every change so a prompt outlives the popup: an esc, a crash, or
// a launcher that dies before saving the prompt file. The launcher
// (herdr-new-thread --launch) deletes it once the prompt is on disk; the next
// popup restores whatever is left. The feature, review and debug keys are
// optional, so a draft.json written before they existed still loads.
type draft struct {
	Project string `json:"project"`
	Name    string `json:"name"`
	Prompt  string `json:"prompt"`

	Branch         string `json:"branch,omitempty"`
	FeatureName    string `json:"feature_name,omitempty"`
	FeaturePrompt  string `json:"feature_prompt,omitempty"`
	NameEdited     bool   `json:"feature_name_edited,omitempty"`
	PromptEdited   bool   `json:"feature_prompt_edited,omitempty"`
	FeatureProject string `json:"feature_project,omitempty"`

	ReviewPrompt string `json:"review_prompt,omitempty"`
	DebugTitle   string `json:"debug_title,omitempty"`
	DebugPrompt  string `json:"debug_prompt,omitempty"`

	SavedAt time.Time `json:"saved_at"`
}

// restoreDrafts is off in tests, which would otherwise inherit each other's.
var restoreDrafts = true

func (d draft) empty() bool {
	return strings.TrimSpace(d.Name) == "" && strings.TrimSpace(d.Prompt) == "" && d.featureEmpty() &&
		strings.TrimSpace(d.ReviewPrompt) == "" && strings.TrimSpace(d.DebugPrompt) == "" &&
		strings.TrimSpace(d.DebugTitle) == ""
}

// featureEmpty: nothing typed in the features view. The name and prompt
// follow the branch, so a branch alone is still worth keeping.
func (d draft) featureEmpty() bool {
	return strings.TrimSpace(d.Branch) == "" && strings.TrimSpace(d.FeatureName) == "" &&
		strings.TrimSpace(d.FeaturePrompt) == ""
}

// featureOnly drafts open the popup on the features view, so the restored
// text is in sight.
func (d draft) featureOnly() bool {
	return strings.TrimSpace(d.Name) == "" && strings.TrimSpace(d.Prompt) == "" && !d.featureEmpty()
}

// reviewOnly drafts open on the review view for the same reason.
func (d draft) reviewOnly() bool {
	return strings.TrimSpace(d.Name) == "" && strings.TrimSpace(d.Prompt) == "" && d.featureEmpty() &&
		strings.TrimSpace(d.ReviewPrompt) != ""
}

// debugOnly drafts open on the debug view.
func (d draft) debugOnly() bool {
	return strings.TrimSpace(d.Name) == "" && strings.TrimSpace(d.Prompt) == "" && d.featureEmpty() &&
		strings.TrimSpace(d.ReviewPrompt) == "" &&
		(strings.TrimSpace(d.DebugPrompt) != "" || strings.TrimSpace(d.DebugTitle) != "")
}

func (d draft) same(o draft) bool {
	return d.Project == o.Project && d.Name == o.Name && d.Prompt == o.Prompt &&
		d.Branch == o.Branch && d.FeatureName == o.FeatureName && d.FeaturePrompt == o.FeaturePrompt &&
		d.NameEdited == o.NameEdited && d.PromptEdited == o.PromptEdited && d.FeatureProject == o.FeatureProject &&
		d.ReviewPrompt == o.ReviewPrompt &&
		d.DebugTitle == o.DebugTitle && d.DebugPrompt == o.DebugPrompt
}

// threadsCacheDir holds the draft, the launch log and jobs/. The env
// override keeps tests off the real one.
func threadsCacheDir() string {
	if dir := os.Getenv("HERDR_THREADS_DIR"); dir != "" {
		return dir
	}
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
