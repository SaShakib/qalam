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
		real := kb.Flags&llkhfInjected == 0 || (typeTesting && kb.DwExtraInfo != qalamMark)
		if real && (wParam == wmKeyDown || wParam == wmSysKeyDown) {
			if onKeyDown(kb) {
				return 1 // swallow the key
			}
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, nCode, wParam, lParam)
	return r
}

// mouseHook: a click moves the cursor, so the word being typed stays where it is and ends there.
// (A click on the options panel is the panel's own business.)
func mouseHook(nCode, wParam, lParam uintptr) uintptr {
	if int32(nCode) == 0 && composer.Buffer != "" && (wParam == wmLButtonDown || wParam == wmRButtonDown || wParam == wmMButtonDown) {
		m := (*msllHook)(unsafe.Pointer(lParam))
		h, _, _ := pWindowFromPoint.Call(uintptr(uint32(m.Pt.X)) | uintptr(uint32(m.Pt.Y))<<32)
		if h != previewHwnd {
			endWordInPlace()
		}
	}
	r, _, _ := pCallNextHookEx.Call(0, nCode, wParam, lParam)
	return r
}

// endWordInPlace finishes the word where it is: the cursor or the app has changed, so nothing
// already in the app may be touched again.
func endWordInPlace() {
	if composer.Buffer == "" {
		return
	}
	if shownInline != "" {
		takeSelected()
		shownInline = ""
	} else {
		commitSelected()
	}
}

func isModifier(vk uint32) bool {
	switch vk {
	case vkShift, vkControl, vkMenu, vkCapital, vkLWin, vkRWin, vkLShift, vkRShift, vkLCtrl, vkRCtrl, vkLMenu, vkRMenu:
		return true
	}
	return false
}

var (
	cands       []engine.Candidate
	sel         int
	shownInline string  // the word as it stands in the app right now (see inline.go)
	composeFg   uintptr // the window the word is being typed in
	typingFg    uintptr // the last app typed in, outside Qalam (the tray's Qur'an item goes back to it)
)

const (
	vkLeft  = 0x25
	vkUp    = 0x26
	vkRight = 0x27
	vkDown  = 0x28
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
	fg, _, _ := pGetForegroundWindow.Call()
	if !isOwnWindow(fg) {
		typingFg = fg
	}
	if busy && fg != composeFg { // another window came forward: the word stays in the old one
		endWordInPlace()
		busy = false
	}

	if recording != "" {
		return recordKey(vk, ctrl, alt, leftAlt, shift, win)
	}
	// The Qur'an box (Ctrl+Alt+Q) gets every key while it's open.
	if quranOpen {
		return quranKey(vk, ctrl, leftAlt, alt, shift, win)
	}

	// The user's shortcuts (tray menu → Shortcuts). Defaults: Ctrl+Alt+A Arabic on/off,
	// Ctrl+Alt+O sukūn Off ↔ Smart, Ctrl+Alt+Q insert from the Qur'an. Alt means Left Alt only, so AltGr letters on other layouts still work.
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
		case hit("quran"):
			commitSelected()
			openQuran()
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
			showInline()
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

	sent := false
	// A key that ends the word puts in the highlighted option first.
	if busy && finishesWord(key) {
		sent = sendCommit(takeSelected())
		if key.Kind == engine.KeyEnter {
			return true
		}
	}
	if !busy {
		composeFg = fg
	}

	out := ""
	consumed, actions := composer.Handle(key, typingOptions())
	for _, a := range actions {
		if a.Commit {
			out += a.Text
		} else {
			refreshOptions()
		}
	}
	if composer.Buffer == "" {
		cands = nil
		hidePreview()
	}
	if out != "" && sendCommit(out) {
		sent = true
	}
	if consumed {
		return true
	}
	if sent {
		// The word must arrive before the key that finished it, so swallow the key and send both.
		reinject(vk, kb.Flags)
		return true
	}
	return false
}

// showInline puts the highlighted option in the app, replacing what is showing there.
func showInline() {
	if settings.NoInline {
		return
	}
	target := ""
	if composer.Buffer != "" && sel < len(cands) {
		target = cands[sel].Text
	}
	replaceInline(target)
}

// replaceInline changes the word showing in the app into target; false if nothing had to be sent.
func replaceInline(target string) bool {
	back, add := inlineEdit(shownInline, target)
	shownInline = target
	if back == 0 && add == "" {
		return false
	}
	inputs := make([]input, 0, 2*back+2*len(add))
	for i := 0; i < back; i++ {
		inputs = append(inputs,
			input{Type: inputKeyboard, WVk: vkBack, DwExtraInfo: qalamMark},
			input{Type: inputKeyboard, WVk: vkBack, DwFlags: keyUp, DwExtraInfo: qalamMark})
	}
	inputs = append(inputs, textInputs(add)...)
	sendInputs(inputs)
	return true
}

// sendCommit makes text final: what is showing becomes text, and the next word starts fresh.
func sendCommit(text string) bool {
	sent := replaceInline(text)
	shownInline = ""
	return sent
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
		showInline()
		return
	}
	cands = engine.Candidates(composer.Buffer, typingOptions())
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
	showInline()
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
		cands = engine.Candidates(composer.Buffer, typingOptions())
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
		sendCommit(t)
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
	return composer.Flush(typingOptions())
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
		inputs = append(inputs, textInputs(t)...)
	}
	sendInputs(inputs)
}

func textInputs(t string) []input {
	inputs := []input{}
	for _, u := range utf16(t) {
		inputs = append(inputs,
			input{Type: inputKeyboard, WScan: u, DwFlags: keyUnicode, DwExtraInfo: qalamMark},
			input{Type: inputKeyboard, WScan: u, DwFlags: keyUnicode | keyUp, DwExtraInfo: qalamMark})
	}
	return inputs
}

func sendInputs(inputs []input) {
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
	sendInputs([]input{
		{Type: inputKeyboard, WVk: uint16(vk), DwFlags: f, DwExtraInfo: qalamMark},
		{Type: inputKeyboard, WVk: uint16(vk), DwFlags: f | keyUp, DwExtraInfo: qalamMark},
	})
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
