//go:build windows

package main

import (
	"unsafe"

	"qalam/engine"
)

const (
	vkBack    = 0x08
	vkTab     = 0x09
	vkReturn  = 0x0D
	vkShift   = 0x10
	vkControl = 0x11
	vkMenu    = 0x12
	vkCapital = 0x14
	vkEscape  = 0x1B
	vkSpace   = 0x20
	vkLWin    = 0x5B
	vkRWin    = 0x5C
	vkLShift  = 0xA0
	vkRShift  = 0xA1
	vkLCtrl   = 0xA2
	vkRCtrl   = 0xA3
	vkLMenu   = 0xA4
	vkRMenu   = 0xA5
)

// keyboardHook sees every key press in every app while Qalam runs.
func keyboardHook(nCode, wParam, lParam uintptr) uintptr {
	if int32(nCode) == 0 {
		kb := (*kbdllHook)(unsafe.Pointer(lParam))
		if kb.Flags&llkhfInjected == 0 && (wParam == wmKeyDown || wParam == wmSysKeyDown) {
			if onKeyDown(kb) {
				return 1 // swallow the key
			}
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, nCode, wParam, lParam)
	return r
}

func isModifier(vk uint32) bool {
	switch vk {
	case vkShift, vkControl, vkMenu, vkCapital, vkLWin, vkRWin, vkLShift, vkRShift, vkLCtrl, vkRCtrl, vkLMenu, vkRMenu:
		return true
	}
	return false
}

var (
	cands []engine.Candidate
	sel   int
)

const (
	vkUp   = 0x26
	vkDown = 0x28
)

// onKeyDown returns true when Qalam handled the key.
func onKeyDown(kb *kbdllHook) bool {
	vk := kb.VkCode
	ctrl := keyDown(vkControl)
	leftAlt := keyDown(vkLMenu)
	alt := keyDown(vkMenu)
	win := keyDown(vkLWin) || keyDown(vkRWin)
	shift := keyDown(vkShift)
	busy := composer.Buffer != ""

	if recording != "" {
		return recordKey(vk, ctrl, alt, leftAlt, shift, win)
	}

	// The user's shortcuts (tray menu → Shortcuts). Defaults: Ctrl+Alt+A Arabic on/off,
	// Ctrl+Alt+O sukūn Off ↔ Smart. Alt means Left Alt only, so AltGr letters on other layouts still work.
	if (ctrl || leftAlt) && !win && !isModifier(vk) {
		hit := func(name string) bool {
			h := settings.hotkey(name)
			return h != nil && h.VK == vk && h.Ctrl == ctrl && h.Alt == leftAlt && h.Alt == alt && h.Shift == shift
		}
		switch {
		case hit("toggle"):
			commitSelected()
			toggle()
			return true
		case hit("sukun"):
			toggleSukun()
			return true
		case hit("harakat"):
			toggleHarakat()
			return true
		}
	}
	if !settings.Enabled || isModifier(vk) {
		return false
	}

	// Shortcuts (Ctrl+C, Alt+Tab…): finish the word, then let the shortcut through.
	if ctrl || alt || win {
		if !busy {
			return false
		}
		commitSelected()
		reinject(vk, kb.Flags)
		return true
	}

	// ↑ ↓ move through the options.
	if busy && (vk == vkUp || vk == vkDown) {
		if len(cands) > 0 {
			step := 1
			if vk == vkUp {
				step = len(cands) - 1
			}
			sel = (sel + step) % len(cands)
			showPanel(cands, sel, composer.Buffer)
		}
		return true
	}

	var key engine.Key
	switch vk {
	case vkBack:
		key = engine.Key{Kind: engine.KeyBackspace}
	case vkEscape:
		key = engine.Key{Kind: engine.KeyEscape}
	case vkReturn:
		key = engine.Key{Kind: engine.KeyEnter}
	case vkSpace:
		key = engine.Key{Kind: engine.KeySpace}
	default:
		if r, ok := usKey(vk, shift); ok {
			key = engine.Key{Kind: engine.KeyChar, Char: r}
		} else {
			key = engine.Key{Kind: engine.KeyOther} // tab, arrows, F-keys…
		}
	}

	out := []string{}
	// A key that ends the word puts in the highlighted option first.
	if busy && finishesWord(key) {
		out = append(out, takeSelected())
		if key.Kind == engine.KeyEnter {
			sendText(out)
			return true
		}
	}

	consumed, actions := composer.Handle(key, settings.Options)
	for _, a := range actions {
		if a.Commit {
			out = append(out, a.Text)
		} else {
			refreshOptions()
		}
	}
	if composer.Buffer == "" {
		cands = nil
		hidePreview()
	}
	if consumed {
		sendText(out)
		return true
	}
	if len(out) > 0 {
		// The word must arrive before the key that finished it, so swallow the key and send both.
		sendText(out)
		reinject(vk, kb.Flags)
		return true
	}
	return false
}

func finishesWord(k engine.Key) bool {
	switch k.Kind {
	case engine.KeySpace, engine.KeyEnter, engine.KeyOther:
		return true
	case engine.KeyChar:
		return !engine.Accepts(k.Char, composer.Buffer)
	}
	return false
}

// refreshOptions recomputes the options for the word being typed.
func refreshOptions() {
	if composer.Buffer == "" {
		cands = nil
		hidePreview()
		return
	}
	cands = engine.Candidates(composer.Buffer, settings.Options)
	sel = 0
	if learned, ok := settings.Learned[composer.Buffer]; ok {
		for i, c := range cands {
			if stripZWJ(c.Text) == learned {
				sel = i
			}
		}
	} else if settings.PrefersPlain {
		for i, c := range cands {
			if c.Kind == "plain" {
				sel = i
			}
		}
	}
	showPanel(cands, sel, composer.Buffer)
}

func stripZWJ(s string) string {
	out := []rune{}
	for _, r := range s {
		if r != 0x200D {
			out = append(out, r)
		}
	}
	return string(out)
}

// takeSelected returns the highlighted option, remembers the choice and clears the word.
func takeSelected() string {
	if composer.Buffer == "" {
		return ""
	}
	if len(cands) == 0 {
		cands = engine.Candidates(composer.Buffer, settings.Options)
		sel = 0
	}
	if sel >= len(cands) {
		sel = 0
	}
	chosen := cands[sel]
	if settings.Learned == nil || len(settings.Learned) > 5000 {
		settings.Learned = map[string]string{}
	}
	settings.Learned[composer.Buffer] = stripZWJ(chosen.Text)
	if chosen.Kind == "plain" {
		settings.PrefersPlain = true
	} else {
		for _, c := range cands {
			if c.Kind == "plain" && c.Text != chosen.Text {
				settings.PrefersPlain = false
			}
		}
	}
	saveSettings(settings)
	composer = engine.Composer{}
	cands = nil
	hidePreview()
	return chosen.Text
}

// commitSelected types the highlighted option now (used before shortcuts).
func commitSelected() {
	if t := takeSelected(); t != "" {
		sendText([]string{t})
	}
}

// pickCandidate: a row in the panel was clicked.
func pickCandidate(i int) {
	if i < len(cands) {
		sel = i
		commitSelected()
	}
}

func toggleSukun() {
	if settings.Sukun == engine.SukunOff {
		settings.Sukun = engine.SukunSmart
	} else {
		settings.Sukun = engine.SukunOff
	}
	saveSettings(settings)
	if composer.Buffer != "" {
		refreshOptions()
	}
}

func toggleHarakat() {
	if settings.Harakat == engine.NoHarakat {
		settings.Harakat = engine.Full
	} else {
		settings.Harakat = engine.NoHarakat
	}
	saveSettings(settings)
	if composer.Buffer != "" {
		refreshOptions()
	}
}

// recording is the shortcut being changed (tray menu → Shortcuts), or "".
var recording string

func startRecording(name, title string) {
	commitSelected()
	recording = name
	showMessage("Press the new shortcut for " + title + ".\nHold Ctrl or Alt with a key.   Backspace: no shortcut.   Esc: cancel.")
}

func stopRecording() {
	recording = ""
	hideMessage()
	updateTray()
}

// recordKey takes the next key press as the new shortcut. Every key is swallowed while recording.
func recordKey(vk uint32, ctrl, anyAlt, alt, shift, win bool) bool {
	if isModifier(vk) {
		return false
	}
	plain := !ctrl && !anyAlt && !win
	switch {
	case anyAlt && !alt:
		showMessage("Use the left Alt key (right Alt / AltGr types letters on many layouts).\nEsc: cancel.")
	case plain && vk == vkEscape:
		stopRecording()
	case plain && vk == vkBack:
		setHotkey(recording, nil)
		stopRecording()
	case win || (!ctrl && !alt):
		showMessage("Hold Ctrl or Alt (not the Windows key) with a key.\nBackspace: no shortcut.   Esc: cancel.")
	default:
		h := &Hotkey{VK: vk, Ctrl: ctrl, Alt: alt, Shift: shift}
		for _, n := range hotkeyNames {
			if n.Name != recording && settings.hotkey(n.Name) != nil && *settings.hotkey(n.Name) == *h {
				showMessage(h.String() + " is already used for " + n.Title + ".\nPress another shortcut, or Esc to cancel.")
				return true
			}
		}
		setHotkey(recording, h)
		stopRecording()
	}
	return true
}

func setHotkey(name string, h *Hotkey) {
	if settings.Keys == nil {
		settings.Keys = map[string]*Hotkey{}
	}
	settings.Keys[name] = h
	saveSettings(settings)
}

func flushAll() []engine.Action {
	return composer.Flush(settings.Options)
}

func sendActions(actions []engine.Action) {
	texts := []string{}
	for _, a := range actions {
		texts = append(texts, a.Text)
	}
	sendText(texts)
}

// sendText types Unicode text into the focused app.
func sendText(texts []string) {
	inputs := []input{}
	for _, t := range texts {
		for _, u := range utf16(t) {
			inputs = append(inputs,
				input{Type: inputKeyboard, WScan: u, DwFlags: keyUnicode},
				input{Type: inputKeyboard, WScan: u, DwFlags: keyUnicode | keyUp})
		}
	}
	if len(inputs) > 0 {
		pSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(input{}))
	}
}

// reinject sends a key press again after Qalam swallowed it.
func reinject(vk uint32, flags uint32) {
	f := uint32(0)
	if flags&llkhfExtended != 0 {
		f = keyExtended
	}
	inputs := []input{
		{Type: inputKeyboard, WVk: uint16(vk), DwFlags: f},
		{Type: inputKeyboard, WVk: uint16(vk), DwFlags: f | keyUp},
	}
	pSendInput.Call(2, uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(input{}))
}

func utf16(s string) []uint16 {
	out := []uint16{}
	for _, r := range s {
		if r >= 0x10000 {
			r -= 0x10000
			out = append(out, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
		} else {
			out = append(out, uint16(r))
		}
	}
	return out
}

// usKey maps a key to its character on the US layout (Qalam's keys are defined on QWERTY).
// Shift decides upper/lower case; Caps Lock is ignored so س never turns into ص by accident.
func usKey(vk uint32, shift bool) (rune, bool) {
	switch {
	case vk >= 'A' && vk <= 'Z':
		if shift {
			return rune(vk), true
		}
		return rune(vk + 32), true
	case vk >= '0' && vk <= '9':
		if shift {
			return rune(")!@#$%^&*("[vk-'0']), true
		}
		return rune(vk), true
	}
	pairs := map[uint32][2]rune{
		0xBA: {';', ':'}, 0xBB: {'=', '+'}, 0xBC: {',', '<'}, 0xBD: {'-', '_'}, 0xBE: {'.', '>'},
		0xBF: {'/', '?'}, 0xC0: {'`', '~'}, 0xDB: {'[', '{'}, 0xDC: {'\\', '|'}, 0xDD: {']', '}'},
		0xDE: {'\'', '"'},
	}
	if p, ok := pairs[vk]; ok {
		if shift {
			return p[1], true
		}
		return p[0], true
	}
	return 0, false
}
