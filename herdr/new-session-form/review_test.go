package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestReviewPromptAndName(t *testing.T) {
	if got := reviewPromptFor(""); got != "/review-queue" {
		t.Errorf("empty: %q", got)
	}
	if got := reviewPromptFor("  just Nick's one  "); got != "/review-queue just Nick's one" {
		t.Errorf("instructions: %q", got)
	}
	if got := reviewName(time.Date(2026, 9, 28, 12, 30, 0, 0, time.Local)); got != "Review queue Sep 28 12:30" {
		t.Errorf("name: %q", got)
	}
}

func TestReviewSubmitsABackgroundReviewQueueJob(t *testing.T) {
	threadsTemp(t)
	get, send := newFormModel(t)
	send(ctrlDigitMsg(4))
	typeText(send, "just nick's")
	send(tea.KeyMsg{Type: tea.KeyCtrlB})

	m := get()
	if m.mode != "bg" {
		t.Fatalf("mode %q, want bg", m.mode)
	}
	res := m.resultFor("bg")
	if res.Project != "august" || res.Prompt != "/review-queue just nick's" || !strings.HasPrefix(res.Name, "Review queue ") {
		t.Errorf("result: %+v", res)
	}
}

// enter on the start field follows the toggle, which starts on background.
func TestReviewStartFieldPicksTheMode(t *testing.T) {
	for _, c := range []struct {
		toggle bool
		want   string
	}{{false, "bg"}, {true, "open"}} {
		threadsTemp(t)
		get, send := newFormModel(t)
		send(ctrlDigitMsg(4))
		send(tea.KeyMsg{Type: tea.KeyTab})
		if c.toggle {
			send(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
		}
		send(tea.KeyMsg{Type: tea.KeyEnter})
		m := get()
		if m.mode != c.want {
			t.Errorf("toggle %v: mode %q, want %q", c.toggle, m.mode, c.want)
		}
		if res := m.resultFor(m.mode); res.Prompt != "/review-queue" || res.Project != "august" {
			t.Errorf("toggle %v: result %+v", c.toggle, res)
		}
	}
}

// Review instructions autosave, and a draft with only them opens on the
// review view.
func TestReviewDraftSurvivesThePopup(t *testing.T) {
	threadsTemp(t)
	get, send := newFormModel(t)
	send(ctrlDigitMsg(4))
	typeText(send, "just nick's")

	next, _ := get().Update(tea.KeyMsg{Type: tea.KeyEscape})
	if next.(model).err == "" {
		t.Error("esc over review instructions should warn first")
	}
	d, ok := loadDraft()
	if !ok || d.ReviewPrompt != "just nick's" || !d.reviewOnly() {
		t.Fatalf("review draft not saved: %+v %v", d, ok)
	}

	withRestore(t)
	get, _ = newFormModel(t)
	m := get()
	if m.view != viewReview || m.reviewPrompt.Value() != "just nick's" {
		t.Errorf("review draft not restored: view %d %q", m.view, m.reviewPrompt.Value())
	}
	if m.notice == "" || !strings.Contains(m.reviewView(), "restored") {
		t.Error("the review view should show the restored-draft notice")
	}
}
