package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// session is one past Claude Code conversation from a project's root dir
// (~/.claude/projects/<encoded-root>/<id>.jsonl).
type session struct {
	ID       string
	Title    string
	FirstAt  time.Time
	LastAt   time.Time
	Turns    int
	LiveTab  string // herdr tab id when a claude process currently holds it
	LiveName string // herdr agent name for that tab
}

// augustRoot mirrors herdr-new-thread: paths.features from the augie config,
// symlink-resolved, since Claude encodes the real cwd into the project dir.
func augustRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "augie", "config.yaml"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if key, value, ok := strings.Cut(strings.TrimSpace(line), ":"); ok && key == "features" {
			return filepath.EvalSymlinks(strings.TrimSpace(value))
		}
	}
	return "", os.ErrNotExist
}

// projectRoot mirrors project_root in herdr-new-thread.
func projectRoot(project string) (string, error) {
	if project == "august" {
		return augustRoot()
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "code", project), nil
}

func projectDir(root string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	encoded := strings.NewReplacer("/", "-", ".", "-", "_", "-").Replace(root)
	return filepath.Join(home, ".claude", "projects", encoded), nil
}

// loadSessions scans every top-level session file in the project dir (subdirs
// hold subagent transcripts). Files are append-only and multi-megabyte, so
// only the few record kinds we need are JSON-decoded; everything else is a
// substring test per line.
func loadSessions(project string) ([]session, error) {
	root, err := projectRoot(project)
	if err != nil {
		return nil, err
	}
	dir, err := projectDir(root)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var sessions []session
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		s := scanSession(filepath.Join(dir, entry.Name()))
		if s.Turns == 0 {
			continue
		}
		s.ID = strings.TrimSuffix(entry.Name(), ".jsonl")
		s.LastAt = info.ModTime()
		sessions = append(sessions, s)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].LastAt.After(sessions[j].LastAt) })
	return sessions, nil
}

var (
	markAITitle     = []byte(`"type":"ai-title"`)
	markCustomTitle = []byte(`"type":"custom-title"`)
	markLastPrompt  = []byte(`"type":"last-prompt"`)
	markUser        = []byte(`"type":"user"`)
)

func scanSession(path string) session {
	var s session
	f, err := os.Open(path)
	if err != nil {
		return s
	}
	defer f.Close()

	var aiTitle, customTitle, firstPrompt string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		switch {
		case bytes.Contains(line, markLastPrompt):
			s.Turns++
		case bytes.Contains(line, markCustomTitle):
			var rec struct {
				CustomTitle string `json:"customTitle"`
			}
			if json.Unmarshal(line, &rec) == nil && rec.CustomTitle != "" {
				customTitle = rec.CustomTitle
			}
		case bytes.Contains(line, markAITitle):
			var rec struct {
				AITitle string `json:"aiTitle"`
			}
			if json.Unmarshal(line, &rec) == nil && rec.AITitle != "" {
				aiTitle = rec.AITitle
			}
		case firstPrompt == "" && bytes.Contains(line, markUser):
			var rec struct {
				Timestamp time.Time `json:"timestamp"`
				Message   struct {
					Content json.RawMessage `json:"content"`
				} `json:"message"`
			}
			if json.Unmarshal(line, &rec) != nil {
				continue
			}
			var text string
			// Human prompts are a plain string; tool results are arrays.
			if json.Unmarshal(rec.Message.Content, &text) == nil && text != "" && !strings.HasPrefix(text, "<") {
				firstPrompt = text
				s.FirstAt = rec.Timestamp
			}
		}
	}

	switch {
	case customTitle != "":
		s.Title = customTitle
	case aiTitle != "":
		s.Title = aiTitle
	default:
		s.Title = firstLine(firstPrompt)
	}
	return s
}

func firstLine(text string) string {
	text = strings.TrimSpace(text)
	if idx := strings.IndexByte(text, '\n'); idx != -1 {
		text = text[:idx]
	}
	return text
}

// liveSessions maps Claude session id → (tab id, agent name) for every claude
// agent herdr is currently running, so a live session is focused rather than
// resumed a second time.
func liveSessions() map[string]session {
	live := map[string]session{}
	out, err := exec.Command("herdr", "agent", "list").Output()
	if err != nil {
		return live
	}
	var resp struct {
		Result struct {
			Agents []struct {
				Agent   string `json:"agent"`
				Name    string `json:"name"`
				TabID   string `json:"tab_id"`
				Session struct {
					Value string `json:"value"`
				} `json:"agent_session"`
			} `json:"agents"`
		} `json:"result"`
	}
	if json.Unmarshal(out, &resp) != nil {
		return live
	}
	for _, a := range resp.Result.Agents {
		if a.Agent == "claude" && a.Session.Value != "" {
			live[a.Session.Value] = session{LiveTab: a.TabID, LiveName: a.Name}
		}
	}
	return live
}

func humanAge(t time.Time) string {
	if t.IsZero() {
		return "?"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return strconv.Itoa(int(d.Minutes())) + "m"
	case d < 24*time.Hour:
		return strconv.Itoa(int(d.Hours())) + "h"
	case d < 14*24*time.Hour:
		return strconv.Itoa(int(d.Hours()/24)) + "d"
	default:
		return strconv.Itoa(int(d.Hours()/(24*7))) + "w"
	}
}
