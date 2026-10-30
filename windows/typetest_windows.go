//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"qalam/engine"
)

// `Qalam.exe --typetest result.txt` is CI's proof on real Windows that typing works: it presses
// real keys (SendInput) into a real edit box, a rich edit box and an Edge window, with the same
// hook as the running app, and reads back what arrived. Word by word it checks that the Arabic
// shows as it is typed, that Shift counts, and that the Qur'an and on/off shortcuts fire.

var typeTesting bool

const (
	wmGetText       = 0x000D
	wmSetText       = 0x000C
	esMultiline     = 0x0004
	wsChildVisible  = 0x50000000
	wsOverlappedWin = 0x00CF0000
	wsVisible       = 0x10000000
)

var (
	pGetWindowTextW     = user32.NewProc("GetWindowTextW")
	pLoadLibraryW       = kernel32.NewProc("LoadLibraryW")
	pBringWindowToTop   = user32.NewProc("BringWindowToTop")
	pSetFocus           = user32.NewProc("SetFocus")
	pGetWindowRect      = user32.NewProc("GetWindowRect")
	pAttachThreadInput  = user32.NewProc("AttachThreadInput")
	pGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
	pSetCursorPos       = user32.NewProc("SetCursorPos")
	pMouseEvent         = user32.NewProc("mouse_event")
)

func tkey(vk uint16, up bool) input {
	f := uint32(0)
	if up {
		f = keyUp
	}
	return input{Type: inputKeyboard, WVk: vk, DwFlags: f}
}

// press types keys as a person would: "A" holds Shift, words go letter by letter.
func press(s string) {
	for _, r := range s {
		var seq []input
		switch {
		case r >= 'a' && r <= 'z':
			seq = []input{tkey(uint16(r-32), false), tkey(uint16(r-32), true)}
		case r >= 'A' && r <= 'Z':
			seq = []input{tkey(vkShift, false), tkey(uint16(r), false), tkey(uint16(r), true), tkey(vkShift, true)}
		case r >= '0' && r <= '9':
			seq = []input{tkey(uint16(r), false), tkey(uint16(r), true)}
		case r == ' ':
			seq = []input{tkey(vkSpace, false), tkey(vkSpace, true)}
		case r == ':':
			seq = []input{tkey(vkShift, false), tkey(0xBA, false), tkey(0xBA, true), tkey(vkShift, true)}
		}
		sendInputs(seq)
		time.Sleep(45 * time.Millisecond)
	}
}

func tap(vk uint16) {
	sendInputs([]input{tkey(vk, false), tkey(vk, true)})
	time.Sleep(90 * time.Millisecond)
}

func chord(mods []uint16, vk uint16) {
	var seq []input
	for _, m := range mods {
		seq = append(seq, tkey(m, false))
	}
	seq = append(seq, tkey(vk, false), tkey(vk, true))
	for i := len(mods) - 1; i >= 0; i-- {
		seq = append(seq, tkey(mods[i], true))
	}
	sendInputs(seq)
	time.Sleep(200 * time.Millisecond)
}

func windowText(h uintptr) string {
	buf := make([]uint16, 8192)
	n, _, _ := pSendMessageW.Call(h, wmGetText, uintptr(len(buf)), uintptr(unsafe.Pointer(&buf[0])))
	return strings.ReplaceAll(syscall.UTF16ToString(buf[:n]), "\r", "")
}

func windowTitle(h uintptr) string {
	buf := make([]uint16, 4096)
	n, _, _ := pGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:n])
}

func editorWindow(class string, top uintptr) uintptr {
	h, _, _ := pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(class))), 0, wsChildVisible|esMultiline,
		0, 0, 780, 360, top, 0, hInstance, 0)
	return h
}

func typeTest(out string) int {
	typeTesting = true
	pSetProcessDpiAwarenessContext.Call(^uintptr(3))
	settings = Settings{Options: engine.DefaultOptions(), Enabled: true, SetupDone: true, FirstRunDone: true, AskedInstall: true, NoAutoUpdate: true}
	hInstance, _, _ = pGetModuleHandleW.Call(0)
	loadPreviewFont()
	createPreviewWindow()
	createQuranWindow()
	createTray()
	hook, _, _ := pSetWindowsHookExW.Call(whKeyboardLL, hookProc, hInstance, 0)
	pSetWindowsHookExW.Call(whMouseLL, mouseProc, hInstance, 0)

	top, _, _ := pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(utf16Ptr("STATIC"))), uintptr(unsafe.Pointer(utf16Ptr("Qalam type test"))),
		wsOverlappedWin|wsVisible, 40, 40, 820, 420, 0, 0, hInstance, 0)
	pLoadLibraryW.Call(uintptr(unsafe.Pointer(utf16Ptr("Msftedit.dll"))))
	edit := editorWindow("EDIT", top)
	rich := editorWindow("RICHEDIT50W", top)

	var log []string
	fails := 0
	check := func(ok bool, format string, args ...any) {
		line := "ok   "
		if !ok {
			line = "FAIL "
			fails++
		}
		log = append(log, line+fmt.Sprintf(format, args...))
	}
	if hook == 0 {
		check(false, "keyboard hook")
	}
	want := func(s string) string { return engine.Text(s, engine.DefaultOptions()) }
	onUI := func(h uintptr) {
		tap(0x87) // F24, nothing uses it: a key press lets this process take the foreground
		runOnUI(func() {
			fg0, _, _ := pGetForegroundWindow.Call()
			ft, _, _ := pGetWindowThreadProcessId.Call(fg0, 0)
			me, _, _ := pGetCurrentThreadId.Call()
			pAttachThreadInput.Call(me, ft, 1) // sharing input with the window in front lets this one take over
			ok, _, _ := pSetForegroundWindow.Call(top)
			defer pAttachThreadInput.Call(me, ft, 0)
			pBringWindowToTop.Call(top)
			f, _, _ := pSetFocus.Call(h)
			log = append(log, fmt.Sprintf("info focus: SetForegroundWindow=%d SetFocus=%d", ok, f))
		})
		time.Sleep(300 * time.Millisecond)
		var g guiThreadInfo
		g.CbSize = uint32(unsafe.Sizeof(g))
		fg, _, _ := pGetForegroundWindow.Call()
		pGetGUIThreadInfo.Call(0, uintptr(unsafe.Pointer(&g)))
		log = append(log, fmt.Sprintf("info foreground=%x top=%x focus=%x editor=%x", fg, top, g.HwndFocus, h))
	}

	result := make(chan int, 1)
	scenario := func() {
		box := func(name string, h uintptr, read func() string) {
			reset := func() {
				pSendMessageW.Call(h, wmSetText, 0, uintptr(unsafe.Pointer(utf16Ptr(""))))
				time.Sleep(60 * time.Millisecond)
			}
			// Live: the word shows before Space.
			press("kataba")
			time.Sleep(150 * time.Millisecond)
			check(read() == want("kataba"), "%s: live %q, want %q", name, read(), want("kataba"))
			tap(vkSpace)
			time.Sleep(150 * time.Millisecond)
			check(read() == want("kataba")+" ", "%s: after Space %q", name, read())
			// Each key rewrites only the end: the earlier word must survive.
			press("kitaAbuN")
			tap(vkSpace)
			time.Sleep(150 * time.Millisecond)
			check(read() == want("kataba")+" "+want("kitaAbuN")+" ", "%s: Shift + second word %q", name, read())
			reset()
			// Backspace inside a word.
			press("kat")
			tap(vkBack)
			time.Sleep(120 * time.Millisecond)
			check(read() == want("ka"), "%s: Backspace %q, want %q", name, read(), want("ka"))
			tap(vkBack)
			tap(vkBack)
			time.Sleep(120 * time.Millisecond)
			check(read() == "", "%s: Backspace to nothing %q", name, read())
			// Esc gives the Latin back.
			press("qul")
			tap(vkEscape)
			time.Sleep(120 * time.Millisecond)
			check(read() == "qul", "%s: Esc %q", name, read())
			reset()
			// An option from the list replaces the word in place.
			press("saAala")
			before := read()
			tap(vkDown)
			time.Sleep(120 * time.Millisecond)
			n := len(cands)
			sels := ""
			if sel < n {
				sels = cands[sel].Text
			}
			check(n < 2 || (read() == sels && read() != before), "%s: ↓ option %q, want %q", name, read(), sels)
			tap(vkEscape)
			reset()
		}

		onUI(edit)
		box("edit", edit, func() string { return windowText(edit) })
		pShowWindow.Call(edit, swHide)
		onUI(rich)
		box("richedit", rich, func() string { return windowText(rich) })

		// Shortcuts: real Ctrl + left Alt.
		pSendMessageW.Call(rich, wmSetText, 0, uintptr(unsafe.Pointer(utf16Ptr(""))))
		chord([]uint16{vkLCtrl, vkLMenu}, 'A')
		check(!settings.Enabled, "Ctrl+Alt+A switches Arabic off")
		press("hello")
		time.Sleep(150 * time.Millisecond)
		check(windowText(rich) == "hello", "typing stays English when off: %q", windowText(rich))
		chord([]uint16{vkLCtrl, vkLMenu}, 'A')
		check(settings.Enabled, "Ctrl+Alt+A switches Arabic on")
		pSendMessageW.Call(rich, wmSetText, 0, uintptr(unsafe.Pointer(utf16Ptr(""))))
		time.Sleep(100 * time.Millisecond)

		chord([]uint16{vkLCtrl, vkLMenu}, 'Q')
		time.Sleep(300 * time.Millisecond)
		check(quranOpen, "Ctrl+Alt+Q opens the Qur'an box")
		press("112:1")
		time.Sleep(900 * time.Millisecond)
		tap(vkReturn)
		time.Sleep(900 * time.Millisecond)
		check(!quranOpen, "Enter closes the box")
		got := windowText(rich)
		check(strings.Contains(got, "(112:1)"), "Enter inserts the āyah: %q", got)

		// Edge: Chromium's own text box. The page copies the box into its title.
		page := "data:text/html;charset=utf-8,<title>ready</title><textarea autofocus id=b rows=9 cols=60 oninput='document.title=%22v:%22%2Bb.value'></textarea>"
		cmd := exec.Command(`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`, "--app="+page, "--no-first-run", "--user-data-dir="+os.TempDir()+`\qalam-edge-test`)
		if err := cmd.Start(); err != nil {
			log = append(log, "skip Edge: "+err.Error())
		} else {
			var eh uintptr
			for i := 0; i < 80 && eh == 0; i++ {
				time.Sleep(250 * time.Millisecond)
				fg, _, _ := pGetForegroundWindow.Call()
				if strings.Contains(windowTitle(fg), "ready") {
					eh = fg
				}
			}
			if eh == 0 {
				log = append(log, "skip Edge: its window did not come forward")
			} else {
				time.Sleep(4 * time.Second) // until the page is ready
				var rc rect
				pGetWindowRect.Call(eh, uintptr(unsafe.Pointer(&rc)))
				pSetCursorPos.Call(uintptr(rc.Left+120), uintptr(rc.Top+110))
				pMouseEvent.Call(0x0002, 0, 0, 0, 0) // left button down
				pMouseEvent.Call(0x0004, 0, 0, 0, 0) // and up
				time.Sleep(500 * time.Millisecond)
				read := func() string { time.Sleep(160 * time.Millisecond); return strings.TrimPrefix(windowTitle(eh), "v:") }
				noZWJ := func(s string) string { return strings.ReplaceAll(s, "\u200D", "") }
				for _, k := range "kataba" { // key by key: every step must arrive
					press(string(k))
					log = append(log, fmt.Sprintf("info edge key %c: %q shown %q", k, read(), shownInline))
				}
				check(read() == want("kataba"), "edge: live %q, want %q", read(), want("kataba"))
				tap(vkSpace)
				press("lillaAhi")
				tap(vkSpace)
				got := noZWJ(read())
				check(strings.TrimSpace(got) == want("kataba")+" "+want("lillaAhi"), "edge: two words %q", got) // a title drops the last space
				press("kat")
				tap(vkBack)
				tap(vkBack)
				got = noZWJ(read())
				check(got == want("kataba")+" "+want("lillaAhi")+" "+want("k"), "edge: Backspace by one mark %q", got)
				tap(vkSpace)
				press("kitaAbuN")
				tap(vkBack)
				tap(vkBack)
				tap(vkBack)
				got = noZWJ(read())
				check(strings.HasSuffix(got, " "+want("kitaA")), "edge: Backspace inside a longer word %q, want it to end %q", got, want("kitaA"))
				tap(vkEscape)
				tap(vkSpace)
				press("kitaAbuN")
				tap(vkBack)
				tap(vkBack)
				tap(vkBack)
				got = noZWJ(read())
				check(strings.HasSuffix(got, " "+want("kitaA")), "edge: Backspace inside a longer word %q, want it to end %q", got, want("kitaA"))
				tap(vkEscape)
			}
			cmd.Process.Kill()
		}

		os.WriteFile(out, []byte(strings.Join(log, "\n")+"\n"), 0o644)
		fmt.Println(strings.Join(log, "\n"))
		code := 0
		if fails > 0 {
			code = 1
		}
		result <- code
		pPostMessageW.Call(trayHwnd, wmClose, 0, 0)
	}
	go scenario()
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	removeTray()
	return <-result
}
