package main

import (
	"reflect"
	"regexp"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

// Legacy key encoding has no bytes for ctrl+1..4, so the form asks the
// terminal for the kitty keyboard protocol's disambiguate flag. Bubble Tea v1
// surfaces the resulting CSI u sequences as an unexported "unknown CSI"
// message; translateKitty turns those back into the KeyMsgs the rest of the
// program expects, plus ctrlDigitMsg for the digits.
const (
	kittyPush = "\x1b[>1u"
	kittyPop  = "\x1b[<u"
)

type ctrlDigitMsg int

var csiURe = regexp.MustCompile(`^\x1b\[(\d+)(?::[\d:]*)?(?:;(\d+)(?::\d+)?)?u$`)

func translateKitty(msg tea.Msg) tea.Msg {
	v := reflect.ValueOf(msg)
	if !v.IsValid() || v.Type().Name() != "unknownCSISequenceMsg" || v.Kind() != reflect.Slice {
		return msg
	}
	parts := csiURe.FindSubmatch(v.Bytes())
	if parts == nil {
		return msg
	}
	code, _ := strconv.Atoi(string(parts[1]))
	mods := 1
	if len(parts[2]) > 0 {
		mods, _ = strconv.Atoi(string(parts[2]))
	}
	mods--
	shift, alt, ctrl := mods&1 != 0, mods&2 != 0, mods&4 != 0

	switch {
	case ctrl && code >= '0' && code <= '9':
		return ctrlDigitMsg(code - '0')
	case ctrl && code >= 'a' && code <= 'z':
		return tea.KeyMsg{Type: tea.KeyCtrlA + tea.KeyType(code-'a'), Alt: alt}
	case ctrl && code >= 'A' && code <= 'Z':
		return tea.KeyMsg{Type: tea.KeyCtrlA + tea.KeyType(code-'A'), Alt: alt}
	case ctrl && code == ' ':
		return tea.KeyMsg{Type: tea.KeyCtrlAt, Alt: alt}
	case ctrl && code == '\\':
		return tea.KeyMsg{Type: tea.KeyCtrlBackslash, Alt: alt}
	case ctrl && code == ']':
		return tea.KeyMsg{Type: tea.KeyCtrlCloseBracket, Alt: alt}
	case ctrl:
		return msg
	}
	switch code {
	case 27:
		return tea.KeyMsg{Type: tea.KeyEscape, Alt: alt}
	case 9:
		if shift {
			return tea.KeyMsg{Type: tea.KeyShiftTab, Alt: alt}
		}
		return tea.KeyMsg{Type: tea.KeyTab, Alt: alt}
	case 13:
		return tea.KeyMsg{Type: tea.KeyEnter, Alt: alt}
	case 127:
		return tea.KeyMsg{Type: tea.KeyBackspace, Alt: alt}
	case ' ':
		return tea.KeyMsg{Type: tea.KeySpace, Alt: alt}
	}
	if code < ' ' {
		return msg
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(code)}, Alt: alt}
}
