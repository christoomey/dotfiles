package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Same name as Bubble Tea's unexported message, which is all translateKitty
// keys on.
type unknownCSISequenceMsg []byte

func TestTranslateKitty(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"\x1b[49;5u", "ctrl+1"},
		{"\x1b[52;5u", "ctrl+4"},
		{"\x1b[27u", "esc"},
		{"\x1b[115;5u", "ctrl+s"},
		{"\x1b[106;5u", "ctrl+j"},
		{"\x1b[9;2u", "shift+tab"},
		{"\x1b[97;3u", "alt+a"},
		{"\x1b[32;5u", "ctrl+@"},
		{"\x1b[49u", "1"},
		{"\x1b[49:33;129u", "1"},
	}
	for _, c := range cases {
		got := translateKitty(unknownCSISequenceMsg(c.in))
		var s string
		switch g := got.(type) {
		case ctrlDigitMsg:
			s = "ctrl+" + string(rune('0'+g))
		case tea.KeyMsg:
			s = g.String()
		default:
			s = "untranslated"
		}
		if s != c.want {
			t.Errorf("%q: got %s, want %s", c.in, s, c.want)
		}
	}
	if _, ok := translateKitty(unknownCSISequenceMsg("\x1b[?1u")).(unknownCSISequenceMsg); !ok {
		t.Error("non-key CSI should pass through")
	}
	if _, ok := translateKitty(tea.KeyMsg{Type: tea.KeyEnter}).(tea.KeyMsg); !ok {
		t.Error("KeyMsg should pass through")
	}
}

func TestCtrlDigitSwitchesTabs(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew)
	for _, c := range []struct {
		digit int
		want  view
	}{{2, viewResume}, {4, viewAgents}, {3, viewFeature}, {1, viewNew}, {9, viewNew}} {
		m, _ = m.Update(ctrlDigitMsg(c.digit))
		if got := m.(model).view; got != c.want {
			t.Errorf("ctrl+%d: view %d, want %d", c.digit, got, c.want)
		}
	}
}

func TestKittySequencesReachModel(t *testing.T) {
	in := strings.NewReader("\x1b[52;5u\x1b[27u")
	p := tea.NewProgram(newModel("minimal", viewNew), tea.WithInput(in), tea.WithoutRenderer(), tea.WithoutSignalHandler())
	done := make(chan tea.Model, 1)
	go func() {
		final, err := p.Run()
		if err != nil {
			t.Error(err)
		}
		done <- final
	}()
	select {
	case final := <-done:
		if got := final.(model).view; got != viewAgents {
			t.Errorf("view %d, want agents", got)
		}
	case <-time.After(5 * time.Second):
		p.Kill()
		t.Fatal("program did not quit on CSI 27 u")
	}
}
