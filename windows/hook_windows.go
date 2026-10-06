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

// onKeyDown returns true when Qalam handled the key.
func onKeyDown(kb *kbdllHook) bool {
	vk := kb.VkCode
	ctrl := keyDown(vkControl)
	leftAlt := keyDown(vkLMenu)
	alt := keyDown(vkMenu)
	win := keyDown(vkLWin) || keyDown(vkRWin)
	shift := keyDown(vkShift)

	// Ctrl + Left Alt + A: switch Arabic on/off (Left Alt only, so AltGr+A still works).
	if vk == 'A' && ctrl && leftAlt && !win {
		if composer.Buffer != "" {
			sendActions(flushAll())
		}
		toggle()
		return true
	}
	if !settings.Enabled || isModifier(vk) {
		return false
	}

	// Shortcuts (Ctrl+C, Alt+Tab…): finish the word, then let the shortcut through.
	if ctrl || alt || win {
		if composer.Buffer == "" {
			return false
		}
		sendActions(flushAll())
		hidePreview()
		reinject(vk, kb.Flags)
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

	consumed, actions := composer.Handle(key, settings.Options)
	commits := []string{}
	for _, a := range actions {
		if a.Commit {
			commits = append(commits, a.Text)
		} else {
			showPreview(a.Text, composer.Buffer)
		}
	}
	if composer.Buffer == "" {
		hidePreview()
	}
	if consumed {
		sendText(commits)
		return true
	}
	if len(commits) > 0 {
		// The word must arrive before the key that finished it, so swallow the key and send both.
		sendText(commits)
		reinject(vk, kb.Flags)
		return true
	}
	return false
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
