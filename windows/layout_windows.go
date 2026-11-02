//go:build windows

package main

import "unsafe"

// Qalam reads QWERTY physical keys and sends Arabic back as Unicode. That is safe with ordinary
// Windows keyboard layouts, but legacy Bengali tools such as Bijoy can also transform those same
// keys before an app receives them. Keep Bijoy installed, but temporarily ask the focused app for
// the normal English layout while Qalam is on. The prior layout comes back when Qalam is off.

const (
	englishUSLayout = "00000409"
	klfNotTellShell = 0x0080
)

var (
	qalamLayoutTarget uintptr
	qalamPreviousHKL  uintptr
)

// useQalamLayout runs on the UI thread or in the low-level hook. It only posts a Windows message;
// it never waits for the other application, so it is safe before the current key is delivered.
func useQalamLayout(hwnd uintptr) {
	if !settings.Enabled || hwnd == 0 || isOwnWindow(hwnd) || hwnd == qalamLayoutTarget {
		return
	}
	restoreQalamLayout()
	tid, _, _ := pGetWindowThreadProcessId.Call(hwnd, 0)
	current, _, _ := pGetKeyboardLayout.Call(tid)
	english, _, _ := pLoadKeyboardLayoutW.Call(uintptr(unsafe.Pointer(utf16Ptr(englishUSLayout))), klfNotTellShell)
	if english == 0 {
		return
	}
	if current == english {
		// Remember the app anyway, so we do not call the layout APIs for every key.
		qalamLayoutTarget = hwnd
		return
	}
	qalamLayoutTarget, qalamPreviousHKL = hwnd, current
	pPostMessageW.Call(hwnd, wmInputLangChangeRequest, 0, english)
}

// restoreQalamLayout gives the focused app back the layout it had before Qalam took over.
func restoreQalamLayout() {
	if qalamLayoutTarget != 0 && qalamPreviousHKL != 0 {
		pPostMessageW.Call(qalamLayoutTarget, wmInputLangChangeRequest, 0, qalamPreviousHKL)
	}
	qalamLayoutTarget, qalamPreviousHKL = 0, 0
}
