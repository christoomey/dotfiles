package main

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// projectLine is what the form's project field draws, since the help line
// mentions project names too.
func projectLine(m tea.Model) string {
	plain := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(m.View(), "")
	lines := strings.Split(plain, "\n")
	for i, l := range lines {
		if strings.Contains(l, "PROJECT") && i+1 < len(lines) {
			return strings.TrimSpace(lines[i+1])
		}
	}
	return ""
}

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
		{"\x1b[97;5u", "ctrl+a"},
		{"\x1b[109;5u", "ctrl+m"},
		{"\x1b[13u", "enter"},
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
		case ctrlMMsg:
			s = "ctrl+m"
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

// ctrl+N enters a section's last subtab; pressed again it cycles the
// subtabs; digits past the last section are ignored.
func TestCtrlDigitSwitchesTabs(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	for _, c := range []struct {
		digit int
		want  view
	}{
		{1, viewAgents}, {3, viewFeature}, {3, viewSidebar}, {3, viewFeature},
		{2, viewNew}, {2, viewResume}, {9, viewResume}, {3, viewFeature}, {2, viewResume},
	} {
		m, _ = m.Update(ctrlDigitMsg(c.digit))
		if got := m.(model).view; got != c.want {
			t.Errorf("ctrl+%d: view %d, want %d", c.digit, got, c.want)
		}
	}
}

func TestCtrlTCyclesSubtabs(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	for _, want := range []view{viewResume, viewNew, viewResume} {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
		if got := m.(model).view; got != want {
			t.Errorf("ctrl+t: view %d, want %d", got, want)
		}
	}
	m, _ = m.Update(ctrlDigitMsg(1))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if got := m.(model).view; got != viewAgents {
		t.Errorf("ctrl+t in a one-view section: view %d, want agents", got)
	}
}

// A left click on a section label or a subtab label switches the view;
// clicks elsewhere on those rows do nothing.
func TestClickingTabsSwitchesViews(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	click := func(x, y int) {
		m, _ = m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	}
	_, secs, subs := m.(model).tabsLayout()
	if len(secs) != 3 || len(subs) != 2 {
		t.Fatalf("layout: %d sections, %d subtabs", len(secs), len(subs))
	}
	click(subs[1].x0, subtabsRow)
	if got := m.(model).view; got != viewResume {
		t.Errorf("subtab click: view %d, want resume", got)
	}
	click(secs[2].x1-1, tabsRow)
	if got := m.(model).view; got != viewFeature {
		t.Errorf("section click: view %d, want feature", got)
	}
	_, _, subs = m.(model).tabsLayout()
	click(subs[1].x0+1, subtabsRow)
	if got := m.(model).view; got != viewSidebar {
		t.Errorf("features subtab click: view %d, want sidebar", got)
	}
	click(secs[1].x0, tabsRow)
	if got := m.(model).view; got != viewResume {
		t.Errorf("section click should return to its last subtab: view %d", got)
	}
	click(secs[0].x1, tabsRow) // the gap after the label
	if got := m.(model).view; got != viewResume {
		t.Errorf("click in the gap switched to %d", got)
	}
}

// a/m/d jump both project pickers straight to a project, but only while the
// picker itself has focus — in the name field they are just letters.
func TestProjectInitialsJumpPickers(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	m = drive(m, m.Init())
	// huh moves focus through commands, so run each key's command tree.
	key := func(k tea.KeyMsg) {
		next, cmd := m.Update(k)
		m = drive(next, cmd)
	}
	rune_ := func(r rune) { key(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}) }

	rune_('m')
	if got := *m.(model).project; got != "august" {
		t.Errorf("m in the name field changed the project to %q", got)
	}
	key(tea.KeyMsg{Type: tea.KeyShiftTab})
	rune_('m')
	if got := *m.(model).project; got != "moi" {
		t.Errorf("form picker: project %q, want moi", got)
	}
	rune_('d')
	if got, sel := *m.(model).project, m.(model).projSel.GetValue(); got != "dotfiles" || sel != "dotfiles" {
		t.Errorf("form picker: project %q / select %v, want dotfiles", got, sel)
	}
	// The select draws its own cursor, not the bound value: check the frame.
	if got := projectLine(m); !strings.Contains(got, "dotfiles") {
		t.Errorf("form picker: the field draws %q", got)
	}

	key(tea.KeyMsg{Type: tea.KeyCtrlR})
	rune_('d')
	if got := m.(model).resumeProject; got != 0 {
		t.Errorf("d in the filter changed the resume project to %d", got)
	}
	key(tea.KeyMsg{Type: tea.KeyShiftTab})
	rune_('d')
	if got := m.(model).resumeProject; got != 2 {
		t.Errorf("resume picker: project %d, want dotfiles (2)", got)
	}
	rune_('a')
	if got := m.(model).resumeProject; got != 0 {
		t.Errorf("resume picker: project %d, want august (0)", got)
	}
}

// ctrl+m / ctrl+d pick moi / dotfiles from anywhere in the new and resume
// views, leaving focus and typed text where they are.
func TestProjectHotkeysWithoutLeavingTheField(t *testing.T) {
	var m tea.Model = newModel("minimal", viewNew, projects[0])
	m = drive(m, m.Init())
	send := func(msg tea.Msg) {
		next, cmd := m.Update(msg)
		m = drive(next, cmd)
	}
	project := func() string { return *m.(model).project }

	send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dose")})
	send(tea.KeyMsg{Type: tea.KeyCtrlD})
	if project() != "dotfiles" || projectLine(m) != "┃ dotfiles" {
		t.Errorf("ctrl+d: %q draws %q", project(), projectLine(m))
	}
	send(ctrlMMsg{})
	if project() != "moi" || projectLine(m) != "┃ moi" {
		t.Errorf("ctrl+m: %q draws %q", project(), projectLine(m))
	}
	if got := *m.(model).name; got != "dose" {
		t.Errorf("name field lost or gained text: %q", got)
	}
	if m.(model).form.GetFocusedField() == m.(model).projSel {
		t.Error("focus moved to the project field")
	}

	send(tea.KeyMsg{Type: tea.KeyCtrlR})
	send(tea.KeyMsg{Type: tea.KeyCtrlD})
	if got := m.(model).resumeProject; got != projectDotfiles {
		t.Errorf("resume ctrl+d: %d", got)
	}
	send(ctrlMMsg{})
	if got := m.(model).resumeProject; got != projectMoi {
		t.Errorf("resume ctrl+m: %d", got)
	}
	if m.(model).projectFocused {
		t.Error("resume picker took focus")
	}
}

func TestKittySequencesReachModel(t *testing.T) {
	in := strings.NewReader("\x1b[49;5u\x1b[27u")
	p := tea.NewProgram(newModel("minimal", viewNew, projects[0]), tea.WithInput(in), tea.WithoutRenderer(), tea.WithoutSignalHandler())
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
