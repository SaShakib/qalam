//go:build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"qalam/engine"
)

// The ق icon in the system tray: left-click switches Arabic on/off, right-click opens the menu.

var (
	hIconOn         uintptr
	hIconOff        uintptr
	taskbarCreated  uintptr
	trayCallbackPtr = syscall.NewCallback(trayProc)
)

const (
	cmdToggle = iota + 1
	cmdEveryday
	cmdQuran
	cmdFull
	cmdAsTyped
	cmdNone
	cmdDigits
	cmdSpelling
	cmdGuide
	cmdStartup
	cmdUninstall
	cmdCheckUpdate
	cmdAutoUpdate
	cmdSukunSmart
	cmdSukunFull
	cmdSukunOff
	cmdOptions
	cmdQuit
	cmdKeysReset
	cmdKeysFirst // + index into hotkeyNames
)

func loadIcon(data []byte, name string) uintptr {
	path := filepath.Join(appDataDir(), name)
	os.WriteFile(path, data, 0o644)
	cx, _, _ := pGetSystemMetrics.Call(smCxSmIcon)
	cy, _, _ := pGetSystemMetrics.Call(smCySmIcon)
	h, _, _ := pLoadImageW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(path))), imageIcon, cx, cy, lrLoadFromFile)
	return h
}

func createTray() {
	className := utf16Ptr("QalamTray")
	wc := wndClassEx{LpfnWndProc: trayCallbackPtr, HInstance: hInstance, LpszClassName: className}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	trayHwnd, _, _ = pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("Qalam"))), 0, 0, 0, 0, 0, 0, 0, hInstance, 0)
	taskbarCreated, _, _ = pRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(utf16Ptr("TaskbarCreated"))))
	hIconOn = loadIcon(iconOn, "qalam-on.ico")
	hIconOff = loadIcon(iconOff, "qalam-off.ico")
	trayCall(nimAdd)
}

func trayData() notifyIconData {
	var d notifyIconData
	d.CbSize = uint32(unsafe.Sizeof(d))
	d.HWnd = trayHwnd
	d.UID = 1
	d.UFlags = nifMessage | nifIcon | nifTip
	d.UCallbackMessage = wmTray
	if settings.Enabled {
		d.HIcon = hIconOn
		copyUTF16(d.SzTip[:], "Qalam: Arabic ON"+switchHint())
	} else {
		d.HIcon = hIconOff
		copyUTF16(d.SzTip[:], "Qalam: Arabic OFF"+switchHint())
	}
	return d
}

func switchHint() string {
	if h := settings.hotkey("toggle"); h != nil {
		return " (" + h.String() + " to switch)"
	}
	return " (click to switch)"
}

// keyTab puts the shortcut at the right edge of a menu item.
func keyTab(name string) string {
	if h := settings.hotkey(name); h != nil {
		return "\t" + h.String()
	}
	return ""
}

func trayCall(op uintptr) {
	d := trayData()
	pShellNotifyIconW.Call(op, uintptr(unsafe.Pointer(&d)))
}

func updateTray() { trayCall(nimModify) }
func removeTray() { trayCall(nimDelete) }

func balloon(title, text string) {
	d := trayData()
	d.UFlags |= nifInfo
	d.DwInfoFlags = niifInfo
	copyUTF16(d.SzInfoTitle[:], title)
	copyUTF16(d.SzInfo[:], text)
	pShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&d)))
}

func trayProc(hwnd, message, wParam, lParam uintptr) uintptr {
	switch {
	case message == wmTray:
		switch lParam & 0xFFFF {
		case wmLButtonUp:
			toggle()
		case wmRButtonUp:
			showMenu()
		}
		return 0
	case message == taskbarCreated && taskbarCreated != 0:
		trayCall(nimAdd) // Explorer restarted
		return 0
	case message == wmClose:
		pDestroyWindow.Call(hwnd)
		return 0
	case message == wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return r
}

func showMenu() {
	menu, _, _ := pCreatePopupMenu.Call()
	defer pDestroyMenu.Call(menu)
	item := func(id int, text string, checked bool) {
		flags := uintptr(mfString)
		if checked {
			flags |= mfChecked
		}
		pAppendMenuW.Call(menu, flags, uintptr(id), uintptr(unsafe.Pointer(utf16Ptr(text))))
	}
	sep := func() { pAppendMenuW.Call(menu, mfSeparator, 0, 0) }

	o := settings.Options
	item(cmdToggle, "Arabic typing"+keyTab("toggle"), settings.Enabled)
	sep()
	item(cmdEveryday, "Everyday style", o.Style == engine.Everyday)
	item(cmdQuran, "Qur'an style (ٱ ـٰ ـٓ)", o.Style == engine.Quran)
	sep()
	item(cmdFull, "Full harakat", o.Harakat == engine.Full)
	item(cmdAsTyped, "Only the harakat I type", o.Harakat == engine.AsTyped)
	item(cmdNone, "No harakat"+keyTab("harakat"), o.Harakat == engine.NoHarakat)
	sep()
	item(cmdSukunSmart, "Sukūn: Smart (only where needed)", o.Sukun == engine.SukunSmart)
	item(cmdSukunFull, "Sukūn: Full (every stop)", o.Sukun == engine.SukunFull)
	item(cmdSukunOff, "Sukūn: Off"+keyTab("sukun"), o.Sukun == engine.SukunOff)
	item(cmdOptions, "Show options while typing", !settings.NoOptions)

	// Shortcuts submenu: click one, then press the new keys.
	keys, _, _ := pCreatePopupMenu.Call()
	for i, n := range hotkeyNames {
		pAppendMenuW.Call(keys, mfString, uintptr(cmdKeysFirst+i),
			uintptr(unsafe.Pointer(utf16Ptr(n.Title+": "+settings.hotkeyText(n.Name)+"   (change…)"))))
	}
	pAppendMenuW.Call(keys, mfSeparator, 0, 0)
	pAppendMenuW.Call(keys, mfString, cmdKeysReset, uintptr(unsafe.Pointer(utf16Ptr("Reset to Ctrl+Alt+A / Ctrl+Alt+O"))))
	pAppendMenuW.Call(menu, mfPopup, keys, uintptr(unsafe.Pointer(utf16Ptr("Shortcuts"))))
	sep()
	item(cmdDigits, "Arabic digits ١٢٣", o.ArabicDigits)
	item(cmdSpelling, "Smart spelling (اللَّه، هَٰذَا)", o.SpellingWords)
	sep()
	item(cmdGuide, "Letters, words && practice…", false)
	item(cmdStartup, "Start with Windows", startupEnabled())
	item(cmdCheckUpdate, "Check for updates… (version "+version+")", false)
	item(cmdAutoUpdate, "Install updates automatically", !settings.NoAutoUpdate)
	if runningFromInstallDir() {
		item(cmdUninstall, "Uninstall Qalam…", false)
	}
	item(cmdQuit, "Quit Qalam", false)

	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(trayHwnd)
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmReturnCmd|tpmRightButton, uintptr(pt.X), uintptr(pt.Y), 0, trayHwnd, 0)
	pPostMessageW.Call(trayHwnd, wmNull, 0, 0)

	switch int(cmd) {
	case cmdToggle:
		toggle()
		return
	case cmdEveryday:
		settings.Style = engine.Everyday
	case cmdQuran:
		settings.Style = engine.Quran
	case cmdFull:
		settings.Harakat = engine.Full
	case cmdAsTyped:
		settings.Harakat = engine.AsTyped
	case cmdNone:
		settings.Harakat = engine.NoHarakat
	case cmdDigits:
		settings.ArabicDigits = !settings.ArabicDigits
	case cmdSpelling:
		settings.SpellingWords = !settings.SpellingWords
	case cmdGuide:
		openGuide()
		return
	case cmdStartup:
		setStartup(!startupEnabled())
		return
	case cmdSukunSmart:
		settings.Sukun = engine.SukunSmart
	case cmdSukunFull:
		settings.Sukun = engine.SukunFull
	case cmdSukunOff:
		settings.Sukun = engine.SukunOff
	case cmdOptions:
		settings.NoOptions = !settings.NoOptions
	case cmdCheckUpdate:
		go checkForUpdate(true)
		return
	case cmdAutoUpdate:
		settings.NoAutoUpdate = !settings.NoAutoUpdate
	case cmdUninstall:
		startUninstall()
		return
	case cmdQuit:
		pDestroyWindow.Call(trayHwnd)
		return
	case cmdKeysReset:
		settings.Keys = nil
		updateTray()
	default:
		if i := int(cmd) - cmdKeysFirst; i >= 0 && i < len(hotkeyNames) {
			startRecording(hotkeyNames[i].Name, hotkeyNames[i].Title)
		}
		return
	}
	saveSettings(settings)
}
