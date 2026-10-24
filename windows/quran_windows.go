//go:build windows

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"qalam/engine"
	"qalam/quran"
)

// The "insert from the Qur'an" box (Ctrl+Alt+Q), like ⌃⇧Q on the Mac:
//
//	Insert from the Qur'an                                    ✕
//	[ search field: part of an āyah, a sūrah, 2:255 ] [Arabic|Translation]
//	results (up to 5 rows, then it scrolls)          ← only while searching
//	Sūrah ▾   From āyah ▾   To āyah ▾
//	preview of the āyāt (Qur'an font, translation under each)
//	[Arabic only|Arabic + translation|Translation only]   Cancel  Insert
//	Translation ▾                                          ← unless Arabic only
//
// Latin letters typed in Arabic mode are written by Qalam (qul → قُلْ) and the field shows the Arabic.
// A result fills Sūrah / From / To. The box grows downward only when it needs room (its top stays).
// Like the options panel it never takes focus; the keyboard hook sends it the keys (quranpick.go has
// the state, so it can be tested anywhere).

var (
	quranLib     *quran.Library
	quranLibOnce sync.Once

	quranHwnd uintptr
	quranOpen bool
	quranB    quranBox
	quranGen  int  // bumped for every new search
	quranBusy bool // a search is running
	quranMenu bool // one of the box's menus is open (it has the focus meanwhile)
	quranFg   uintptr
	quranWork rect   // the work area of the monitor the box opened on
	quranTop  int32  // the box's top edge (fixed while it's open)
	quranDPI  uint32 // the fonts were made for this DPI

	// Clickable areas, set by layoutQuran.
	quranCloseBtn  rect
	quranField     rect
	quranSwitch    [2]rect // Arabic | Translation
	quranCaption   rect
	quranList      rect
	quranRowRects  []rect // the rows showing, from quranB.Top
	quranDrops     [3]rect
	quranPreview   rect
	quranShownAyat int // āyāt that fit in the preview
	quranSepY      int32
	quranModeBtns  [3]rect
	quranCancelBtn rect
	quranInsertBtn rect
	quranTransBtn  rect
	quranHint      rect
	quranW, quranH int32

	qFontTitle, qFontUI, qFontSmall, qFontBold, qFontButton uintptr // Segoe UI
	qFontBangla                                             uintptr // Nirmala UI, for Bangla translations
	qFontField, qFontRow, qFontPreview                      uintptr // the Qur'an font
	quranFamily                                             = ""

	qBrushBg, qBrushCard, qBrushBorder, qBrushAccent, qBrushSelected, qBrushSep, qBrushWhite uintptr

	pSetTimer         = user32.NewProc("SetTimer")
	pKillTimer        = user32.NewProc("KillTimer")
	pMessageBeep      = user32.NewProc("MessageBeep")
	pOpenClipboard    = user32.NewProc("OpenClipboard")
	pCloseClipboard   = user32.NewProc("CloseClipboard")
	pGetClipboardData = user32.NewProc("GetClipboardData")
	pScreenToClient   = user32.NewProc("ScreenToClient")
	pMonitorFromPoint = user32.NewProc("MonitorFromPoint")
	pGetMonitorInfoW  = user32.NewProc("GetMonitorInfoW")
	pGlobalLock       = kernel32.NewProc("GlobalLock")
	pGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	pGetDpiForMonitor = syscall.NewLazyDLL("shcore.dll").NewProc("GetDpiForMonitor")
)

const (
	dtRight         = 0x0002
	dtWordBreak     = 0x0010
	dtEditControl   = 0x2000
	dtEndEllipsis   = 0x8000
	wmTimer         = 0x0113
	wmLButtonDblClk = 0x0203
	wmMouseWheel    = 0x020A
	wmQuranInsert   = wmApp + 10
	csDblClks       = 0x0008
	cfUnicodeText   = 13
	tpmLeftAlign    = 0x0000
	mfMenuBreak     = 0x0040
	quranWidth      = 640 // logical pixels, like the Mac box
	quranMaxPreview = 260 // the preview card's height before "… and N more"
	quranMaxAyat    = 20  // āyāt shown in the preview at most (all are inserted)
)

type monitorInfo struct {
	CbSize    uint32
	RcMonitor rect
	RcWork    rect
	DwFlags   uint32
}

func quranLibrary() *quran.Library {
	quranLibOnce.Do(func() {
		quranLib, _ = quran.Load(quranFS())
	})
	return quranLib
}

// createQuranWindow makes the box (hidden) and loads KFGQPC Hafs for the āyāt.
func createQuranWindow() {
	className := utf16Ptr("QalamQuran")
	wc := wndClassEx{
		Style:         csDblClks,
		LpfnWndProc:   syscall.NewCallback(quranProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	quranHwnd, _, _ = pCreateWindowExW.Call(
		wsExTopmost|wsExToolWindow|wsExNoActivate,
		uintptr(unsafe.Pointer(className)), 0,
		wsPopup|wsBorder, 0, 0, 10, 10, 0, 0, hInstance, 0)

	quranFamily = previewFamily
	for _, f := range fonts {
		if isQuranFont(f) {
			var count uint32
			if h, _, _ := pAddFontMem.Call(uintptr(unsafe.Pointer(&f.Data[0])), uintptr(len(f.Data)), 0, uintptr(unsafe.Pointer(&count))); h != 0 {
				quranFamily = f.Name
			}
			break
		}
	}
	brush := func(r, g, b byte) uintptr { h, _, _ := pCreateSolidBrush.Call(rgb(r, g, b)); return h }
	qBrushBg = brush(255, 255, 255)
	qBrushCard = brush(246, 247, 249)
	qBrushBorder = brush(206, 210, 216)
	qBrushAccent = brush(46, 110, 200)
	qBrushSelected = brush(221, 233, 250)
	qBrushSep = brush(228, 230, 234)
	qBrushWhite = brush(255, 255, 255)
	makeQuranFonts(dpiForSystem())
}

func dpiForSystem() uint32 {
	dpi, _, _ := pGetDpiForSystem.Call()
	if dpi == 0 {
		dpi = 96
	}
	return uint32(dpi)
}

// makeQuranFonts makes the box's fonts for a DPI (again when it opens on a monitor with another one).
func makeQuranFonts(dpi uint32) {
	if dpi == quranDPI {
		return
	}
	for _, f := range []uintptr{qFontTitle, qFontUI, qFontSmall, qFontBold, qFontButton, qFontBangla, qFontField, qFontRow, qFontPreview} {
		if f != 0 {
			pDeleteObject.Call(f)
		}
	}
	quranDPI = dpi
	qFontTitle = makeFont(qs(15), "Segoe UI", 600)
	qFontUI = makeFont(qs(14), "Segoe UI", 400)
	qFontSmall = makeFont(qs(12), "Segoe UI", 400)
	qFontBold = makeFont(qs(12), "Segoe UI", 600)
	qFontButton = makeFont(qs(13), "Segoe UI", 400)
	qFontBangla = makeFont(qs(15), "Nirmala UI", 400)
	qFontField = makeFont(qs(22), quranFamily, 400)
	qFontRow = makeFont(qs(20), quranFamily, 400)
	qFontPreview = makeFont(qs(25), quranFamily, 400)
}

// qs scales logical pixels to the DPI of the monitor the box is on.
func qs(px int32) int32 {
	if quranDPI == 0 {
		return scale(px)
	}
	return px * int32(quranDPI) / 96
}

// shown: the āyah as drawn (KFGQPC Hafs needs three marks swapped, see quran.ForKFGQPC).
func shown(s string) string {
	if isKFGQPC() {
		return quran.ForKFGQPC(s)
	}
	return s
}

func isKFGQPC() bool { return strings.Contains(strings.ToUpper(quranFamily), "KFGQPC") }

// ---------- Opening, closing, keys ----------

func openQuran() {
	lib := quranLibrary()
	if lib == nil || quranHwnd == 0 {
		pMessageBeep.Call(0xFFFFFFFF)
		return
	}
	quranOpen, quranMenu = true, false
	quranB = newQuranBox(lib, settings.QuranLast)
	quranGen++
	quranFg, _, _ = pGetForegroundWindow.Call()
	placeQuran()
	tr := settings.quranTranslation()
	go func() { // read the texts and build the search index now, so the first search is quick
		lib.TranslationText(tr, 1, 1)
		lib.SearchArabic("بسم", false, 0)
	}()
	showQuran()
	pSetTimer.Call(quranHwnd, 1, 300, 0)
}

// placeQuran picks the monitor (where the caret is, else the mouse) and the box's top edge: a fifth
// of the way down, so it has room to grow downward.
func placeQuran() {
	x, y := caretPosition()
	mon, _, _ := pMonitorFromPoint.Call(uintptr(uint32(x))|uintptr(uint32(y))<<32, 2) // MONITOR_DEFAULTTONEAREST
	mi := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	if mon != 0 {
		if ok, _, _ := pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi))); ok == 0 {
			mon = 0
		}
	}
	if mon == 0 {
		sw, _, _ := pGetSystemMetrics.Call(smCxScreen)
		sh, _, _ := pGetSystemMetrics.Call(smCyScreen)
		mi.RcWork = rect{0, 0, int32(sw), int32(sh)}
	}
	quranWork = mi.RcWork
	dpi := dpiForSystem()
	if mon != 0 && pGetDpiForMonitor.Find() == nil {
		var dx, dy uint32
		if r, _, _ := pGetDpiForMonitor.Call(mon, 0, uintptr(unsafe.Pointer(&dx)), uintptr(unsafe.Pointer(&dy))); r == 0 && dx > 0 {
			dpi = dx // MDT_EFFECTIVE_DPI
		}
	}
	makeQuranFonts(dpi)
	quranTop = quranWork.Top + (quranWork.Bottom-quranWork.Top)/5
}

func closeQuran() {
	if !quranOpen {
		return
	}
	quranOpen = false
	quranB = quranBox{Sel: -1}
	quranGen++
	pKillTimer.Call(quranHwnd, 1)
	pShowWindow.Call(quranHwnd, swHide)
}

// quranKey handles a key while the box is open; true = swallow it.
func quranKey(vk uint32, ctrl, leftAlt, alt, shift, win bool) bool {
	if quranMenu || isModifier(vk) { // an open menu gets its keys
		return false
	}
	if h := settings.hotkey("quran"); h != nil && h.VK == vk && h.Ctrl == ctrl && h.Alt == leftAlt && h.Alt == alt && h.Shift == shift {
		closeQuran() // the shortcut again closes it
		return true
	}
	if ctrl && !alt && !win && !shift && vk == 'V' {
		if t := clipboardText(); t != "" {
			quranB.typeText(strings.Join(strings.Fields(t), " "))
			searchQuran()
		}
		return true
	}
	if ctrl && !alt && !win && vk == vkBack { // Ctrl+Backspace empties the field
		quranB.clear()
		searchQuran()
		return true
	}
	if ctrl || alt || win { // other shortcuts go to the app
		closeQuran()
		return false
	}
	lib := quranLibrary()
	switch vk {
	case vkEscape: // clears the search first, then closes
		if quranB.Keys != "" {
			quranB.clear()
			searchQuran()
		} else {
			closeQuran()
		}
		return true
	case vkReturn:
		// Not from inside the keyboard hook: it has to return quickly, and the text is typed with SendInput.
		pPostMessageW.Call(quranHwnd, wmQuranInsert, 0, 0)
		return true
	case vkTab:
		if shift {
			quranB.tab(-1)
		} else {
			quranB.tab(1)
		}
		showQuran()
		return true
	}
	if quranB.OnPick {
		p := &quranB.Pick
		switch vk {
		case vkUp:
			p.step(lib, -1)
		case vkDown:
			p.step(lib, 1)
		case vkLeft:
			p.move(-1)
		case vkRight:
			p.move(1)
		case vkBack:
			p.backspace(lib)
		case vkSpace:
			p.typeRune(lib, ' ')
		default:
			if r, ok := usKey(vk, shift); ok && p.typeRune(lib, r) {
				quranB.OnPick = false // "/" goes back to the search field
			}
		}
		showQuran()
		return true
	}
	switch vk {
	case vkUp:
		quranB.move(-1)
		showQuran()
	case vkDown:
		quranB.move(1)
		showQuran()
	case vkBack:
		if quranB.backspace() {
			searchQuran()
		}
	case vkSpace:
		quranB.typeText(" ")
		searchQuran()
	default:
		if r, ok := usKey(vk, shift); ok {
			quranB.typeText(string(r))
			searchQuran()
		}
	}
	return true
}

// searchQuran searches off the keyboard thread (the first search builds the index) and shows the
// results when they come back, unless something newer was typed by then. The field shows at once.
func searchQuran() {
	lib := quranLibrary()
	quranGen++
	gen, q, in, tr := quranGen, quranB.query(), quranB.In, settings.quranTranslation()
	quranBusy = q != ""
	if q == "" {
		quranB.setResults(quran.Results{})
		showQuran()
		return
	}
	showQuran()
	go func() {
		res := lib.Search(q, in, tr, quranResultLimit)
		runOnUI(func() {
			if !quranOpen || gen != quranGen {
				return
			}
			quranB.setResults(res)
			quranBusy = false
			showQuran()
		})
	}()
}

// insertChosen types the chosen āyāt into the app you were in, and remembers them for next time.
func insertChosen() {
	if !quranOpen {
		return
	}
	p := quranB.Pick.passage()
	settings.QuranLast = p.Reference()
	saveSettings(settings)
	text := quranLibrary().Text(p, settings.quranMode(), settings.quranTranslation())
	closeQuran()
	if typingOptions().BlockAllahLigature {
		text = engine.BlockAllahLigature(text)
	}
	sendLines(text)
}

// sendLines types text with Shift+Enter between lines: a new line in editors, and in chat apps a
// new line instead of sending the message.
func sendLines(text string) {
	inputs := []input{}
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			inputs = append(inputs,
				input{Type: inputKeyboard, WVk: vkShift},
				input{Type: inputKeyboard, WVk: vkReturn},
				input{Type: inputKeyboard, WVk: vkReturn, DwFlags: keyUp},
				input{Type: inputKeyboard, WVk: vkShift, DwFlags: keyUp})
		}
		for _, u := range utf16(line) {
			inputs = append(inputs,
				input{Type: inputKeyboard, WScan: u, DwFlags: keyUnicode},
				input{Type: inputKeyboard, WScan: u, DwFlags: keyUnicode | keyUp})
		}
	}
	if len(inputs) > 0 {
		pSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(input{}))
	}
}

func clipboardText() string {
	if ok, _, _ := pOpenClipboard.Call(0); ok == 0 {
		return ""
	}
	defer pCloseClipboard.Call()
	h, _, _ := pGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return ""
	}
	p, _, _ := pGlobalLock.Call(h)
	if p == 0 {
		return ""
	}
	defer pGlobalUnlock.Call(h)
	var u []uint16
	for i := uintptr(0); i < 4096; i++ {
		c := *(*uint16)(unsafe.Pointer(p + 2*i))
		if c == 0 {
			break
		}
		u = append(u, c)
	}
	return syscall.UTF16ToString(u)
}

// ---------- Mouse ----------

func quranProc(hwnd, message, wParam, lParam uintptr) uintptr {
	switch message {
	case wmMouseActivate:
		return maNoActivate
	case wmLButtonDown, wmLButtonDblClk:
		x, y := int32(int16(lParam&0xFFFF)), int32(int16((lParam>>16)&0xFFFF))
		onQuranClick(x, y, message == wmLButtonDblClk)
		return 0
	case wmMouseWheel:
		// Windows sends the wheel to the window under the mouse ("scroll inactive windows", on by default).
		pt := point{int32(int16(lParam & 0xFFFF)), int32(int16((lParam >> 16) & 0xFFFF))}
		pScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
		onQuranWheel(pt.X, pt.Y, int32(int16((wParam>>16)&0xFFFF)))
		return 0
	case wmQuranInsert:
		insertChosen()
		return 0
	case wmTimer:
		// Another window came to the front (a click elsewhere, Alt+Tab): close. Not while a menu is open.
		if fg, _, _ := pGetForegroundWindow.Call(); quranOpen && !quranMenu && fg != quranFg {
			closeQuran()
		}
		return 0
	case wmPaint:
		paintQuran(hwnd)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return r
}

func onQuranClick(x, y int32, double bool) {
	lib := quranLibrary()
	b := &quranB
	switch {
	case inRect(x, y, quranCloseBtn), inRect(x, y, quranCancelBtn):
		closeQuran()
		return
	case inRect(x, y, quranInsertBtn):
		insertChosen()
		return
	case inRect(x, y, quranField):
		b.OnPick = false
	case inRect(x, y, quranTransBtn):
		have := lib.Available()
		titles := make([]string, len(have))
		cur := 0
		for i, t := range have {
			titles[i] = t.Title()
			if t.ID == settings.quranTranslation() {
				cur = i
			}
		}
		if i := quranPopup(quranTransBtn, titles, cur, 0); i >= 0 && have[i].ID != settings.quranTranslation() {
			settings.QuranTranslation = have[i].ID
			saveSettings(settings)
			if b.Res.Translation != "" {
				searchQuran() // a translation search looks in the chosen translation when it can
			}
		}
	}
	for i, r := range quranSwitch {
		if inRect(x, y, r) && b.In != quran.SearchIn(i) {
			b.In = quran.SearchIn(i)
			searchQuran()
		}
	}
	for i, r := range quranModeBtns {
		if inRect(x, y, r) {
			settings.QuranInsert = string(quran.InsertModes[i])
			saveSettings(settings)
		}
	}
	for i, r := range quranRowRects {
		if inRect(x, y, r) {
			b.choose(b.Top + i)
			if double && b.Res.Rows[b.Top+i].Kind != quran.KindProblem {
				insertChosen()
				return
			}
		}
	}
	for f, r := range quranDrops {
		if !inRect(x, y, r) {
			continue
		}
		p := &b.Pick
		b.OnPick, p.Field, p.Typed = true, f, ""
		n := p.ayahs(lib)
		switch f {
		case fieldSurah:
			titles := make([]string, 114)
			for i := range titles {
				titles[i] = surahTitle(lib.Surah(i + 1))
			}
			if i := quranPopup(r, titles, p.Surah-1, 38); i >= 0 {
				p.setSurah(i + 1)
			}
		case fieldFrom:
			if i := quranPopup(r, numbers(1, n), p.From-1, 30); i >= 0 {
				p.setFrom(i + 1)
			}
		case fieldTo:
			if i := quranPopup(r, numbers(p.From, n), p.passage().To-p.From, 30); i >= 0 {
				p.To = p.From + i
			}
		}
	}
	showQuran()
}

// onQuranWheel scrolls the results under the mouse (the selection stays).
func onQuranWheel(x, y, delta int32) {
	if !inRect(x, y, quranList) {
		return
	}
	b := &quranB
	step := 1
	if delta > 0 {
		step = -1
	}
	b.Top = clamp(b.Top+step, 0, max(0, len(b.Res.Rows)-quranVisibleRows))
	showQuran()
}

// surahTitle: "2. Al-Baqara · البقرة" (the dropdown and its menu).
func surahTitle(s *quran.Surah) string {
	return fmt.Sprintf("%d. %s  ·  %s", s.Number, s.Name, s.Arabic)
}

func numbers(from, to int) []string {
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, fmt.Sprint(i))
	}
	return out
}

// quranPopup shows a menu under r (long lists in columns of `column`) and returns the index picked, or -1.
// A menu needs the focus while it's open; the app you were typing in gets it back afterwards.
func quranPopup(r rect, titles []string, current, column int) int {
	menu, _, _ := pCreatePopupMenu.Call()
	defer pDestroyMenu.Call(menu)
	for i, t := range titles {
		flags := uintptr(mfString)
		if i == current {
			flags |= mfChecked
		}
		if column > 0 && i > 0 && i%column == 0 {
			flags |= mfMenuBreak
		}
		pAppendMenuW.Call(menu, flags, uintptr(i+1), uintptr(unsafe.Pointer(utf16Ptr(t))))
	}
	pt := point{X: r.Left, Y: r.Bottom}
	pClientToScreen.Call(quranHwnd, uintptr(unsafe.Pointer(&pt)))
	quranMenu = true
	pSetForegroundWindow.Call(quranHwnd)
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmReturnCmd|tpmLeftAlign, uintptr(pt.X), uintptr(pt.Y), 0, quranHwnd, 0)
	pPostMessageW.Call(quranHwnd, wmNull, 0, 0)
	pSetForegroundWindow.Call(quranFg)
	quranMenu = false
	return int(cmd) - 1
}

// ---------- Text helpers ----------

// arabicNumber: the āyah-end mark with its number. KFGQPC Hafs circles plain Arabic digits by itself;
// other fonts need ۝ (U+06DD) before them.
func arabicNumber(n int) string {
	d := []rune(fmt.Sprint(n))
	for i := range d {
		d[i] = 0x0660 + d[i] - '0'
	}
	if isKFGQPC() {
		return string(d)
	}
	return "۝" + string(d)
}

// textHeight is the height of s drawn in font, wrapped to width.
func textHeight(font uintptr, s string, width int32, flags uintptr) int32 {
	hdc, _, _ := pGetDC.Call(quranHwnd)
	defer pReleaseDC.Call(quranHwnd, hdc)
	pSelectObject.Call(hdc, font)
	rc := rect{0, 0, width, 10000}
	drawText(hdc, s, &rc, flags|dtCalcRect|dtWordBreak|dtNoPrefix)
	return rc.Bottom - rc.Top
}

func lineHeight(font uintptr) int32 {
	_, h := measure(font, "Agبِسْمِ", 0)
	return h
}

func textWidth(font uintptr, s string, flags uintptr) int32 {
	w, _ := measure(font, s, flags)
	return w
}

// capLines keeps whole lines: at most what fits in limit.
func capLines(h, line, limit int32) int32 {
	if h <= limit || line <= 0 {
		return h
	}
	return max(line, limit/line*line)
}

// transStyle: the font and direction for a translation (Bangla in Nirmala UI, Urdu right to left).
func transStyle(id string) (uintptr, uintptr) {
	switch {
	case strings.HasPrefix(id, "bn."):
		return qFontBangla, 0
	case strings.HasPrefix(id, "ur."):
		return qFontUI, dtRight | dtRTLReading
	}
	return qFontUI, 0
}

// The preview: āyah a in the chosen mode (Arabic with its number; translation with its reference).
func previewTexts(p quran.Passage, a int) (ar, tr string) {
	lib := quranLibrary()
	mode := settings.quranMode()
	if mode != quran.TranslationOnly {
		ar = shown(lib.Arabic(p.Surah, a)) + " " + arabicNumber(a)
	}
	if mode != quran.Arabic {
		if t := lib.TranslationText(settings.quranTranslation(), p.Surah, a); t != "" {
			tr = fmt.Sprintf("%s (%d:%d)", t, p.Surah, a)
		} else if mode == quran.TranslationOnly {
			ar = shown(lib.Arabic(p.Surah, a)) + " " + arabicNumber(a)
		}
	}
	return
}

func ayahHeight(p quran.Passage, a int, width int32) (arH, trH int32) {
	ar, tr := previewTexts(p, a)
	if ar != "" {
		arH = textHeight(qFontPreview, ar, width, dtRTLReading|dtRight)
	}
	if tr != "" {
		f, dir := transStyle(settings.quranTranslation())
		trH = textHeight(f, tr, width, dir)
	}
	return
}

const (
	quranPlaceholderArabic = "Part of an āyah (qul huwa allaahu), a sūrah, or 2:255"
	quranPlaceholderTrans  = "A word in the translation, a sūrah name, or 2:255"
)

func quranHintText() string {
	if quranB.OnPick {
		return "↑↓ change · type 36, yasin or 2:255-257 · Tab next · / back to the search · Enter insert · Esc close"
	}
	if quranB.Keys != "" {
		return "↑↓ choose · Backspace removes the last key · Tab: Sūrah, From, To · Enter insert · Esc clear"
	}
	s := "Type to search · Tab: Sūrah, From, To · Enter insert · Esc close"
	if h := settings.hotkey("quran"); h != nil {
		s += " · " + h.String() + " opens this box"
	}
	return s
}

// quranCaptionText is the line over the results.
func quranCaptionText() string {
	b := &quranB
	if b.query() == "" {
		return ""
	}
	typed := ""
	if b.In == quran.InArabic && hasLatinLetter(b.Keys) && quran.Script(b.Keys) != "bangla" {
		typed = "“" + b.query() + "”: "
	}
	r := b.Res
	switch {
	case quranBusy && len(r.Rows) == 0:
		return typed + "searching…"
	case r.Total == 0 && len(r.Rows) == 0:
		if b.In == quran.InArabic && quran.Script(b.Keys) == "latin" {
			return typed + "nothing found. An English word? Click Translation."
		}
		return typed + "nothing found."
	case r.Total == 0:
		return typed + "↑↓ or click to choose, double-click inserts."
	}
	s := fmt.Sprintf("%s%d āyāt in %s", typed, r.Total, r.SearchedIn)
	if r.Total > quranResultLimit {
		s += fmt.Sprintf(" (first %d)", quranResultLimit)
	}
	return s + ". ↑↓ or click to choose."
}

// ---------- Layout ----------

// showQuran lays out the box and shows it: centred on its monitor, its top edge fixed, moved up
// only if the bottom would leave the screen.
func showQuran() {
	if !quranOpen || quranHwnd == 0 {
		return
	}
	layoutQuran()
	w := quranWork
	x := w.Left + (w.Right-w.Left-quranW)/2
	y := quranTop
	if y+quranH > w.Bottom-qs(8) {
		y = max(w.Top, w.Bottom-qs(8)-quranH)
	}
	pSetWindowPos.Call(quranHwnd, hwndTopmost, uintptr(x), uintptr(y), uintptr(quranW), uintptr(quranH), swpNoActivate)
	pInvalidateRect.Call(quranHwnd, 0, 1)
	pShowWindow.Call(quranHwnd, swShowNA)
}

// resultRowHeight: rows are sized from the font they use, so no text is cut.
func resultRowHeight(r quran.Suggestion) int32 {
	switch {
	case r.Kind == quran.KindHit && r.InTranslation == "":
		return max(qs(38), lineHeight(qFontRow)+qs(4))
	case r.Kind == quran.KindSurah:
		return max(qs(34), lineHeight(qFontUI)+qs(12))
	}
	return max(qs(32), lineHeight(qFontUI)+qs(10))
}

func layoutQuran() {
	b := &quranB
	pad := qs(16)
	width := qs(quranWidth)
	y := qs(12)

	// Header: title · ✕
	quranCloseBtn = rect{width - pad - qs(28), y, width - pad + qs(4), y + qs(28)}
	y += qs(28) + qs(8)

	// Search field · [Arabic | Translation]
	fh := max(qs(38), lineHeight(qFontField)+qs(4))
	sw := qs(184)
	quranSwitch[1] = rect{width - pad - sw/2, y, width - pad, y + fh}
	quranSwitch[0] = rect{width - pad - sw, y, quranSwitch[1].Left, y + fh}
	quranField = rect{pad, y, quranSwitch[0].Left - qs(8), y + fh}
	y += fh

	// Results
	quranCaption, quranList = rect{}, rect{}
	quranRowRects = quranRowRects[:0]
	if b.query() != "" {
		y += qs(6)
		quranCaption = rect{pad, y, width - pad, y + lineHeight(qFontSmall)}
		y = quranCaption.Bottom + qs(4)
		if n := len(b.Res.Rows); n > 0 {
			b.Top = clamp(b.Top, 0, max(0, n-quranVisibleRows))
			ly := y + qs(2)
			for i := b.Top; i < n && i < b.Top+quranVisibleRows; i++ {
				h := resultRowHeight(b.Res.Rows[i])
				quranRowRects = append(quranRowRects, rect{pad + qs(2), ly, width - pad - qs(2), ly + h})
				ly += h
			}
			quranList = rect{pad, y, width - pad, ly + qs(2)}
			y = quranList.Bottom
		}
	}
	y += qs(12)

	// Sūrah ▾  From āyah ▾  To āyah ▾ (a label above each)
	top := y + lineHeight(qFontSmall) + qs(3)
	dh := qs(30)
	quranDrops[0] = rect{pad, top, pad + qs(300), top + dh}
	quranDrops[1] = rect{quranDrops[0].Right + qs(12), top, quranDrops[0].Right + qs(12) + qs(92), top + dh}
	quranDrops[2] = rect{quranDrops[1].Right + qs(12), top, quranDrops[1].Right + qs(12) + qs(92), top + dh}
	y = top + dh + qs(12)

	// The bottom part's height, so the preview can be capped to what the screen has room for.
	bh := qs(30)
	bottomH := qs(10) + qs(1) + qs(10) + bh + qs(10) + lineHeight(qFontSmall) + qs(12)
	if settings.quranMode() != quran.Arabic {
		bottomH += qs(28) + qs(8)
	}

	// Preview card: a reference line, then the āyāt that fit.
	p := b.Pick.passage()
	refH := lineHeight(qFontBold)
	inner := width - 2*pad - 2*qs(12)
	room := quranWork.Bottom - quranWork.Top - qs(16) - (y + refH + qs(4)) - bottomH - 2*qs(10)
	limit := min(qs(quranMaxPreview), max(qs(90), room))
	used := int32(0)
	quranShownAyat = 0
	last := min(p.To, p.From+quranMaxAyat-1)
	moreH := lineHeight(qFontSmall) + qs(4)
	for a := p.From; a <= last; a++ {
		arH, trH := ayahHeight(p, a, inner)
		h := arH + trH
		if arH > 0 && trH > 0 {
			h += qs(4)
		}
		if quranShownAyat > 0 {
			h += qs(10)
		}
		reserve := int32(0)
		if a < p.To {
			reserve = moreH
		}
		if used+h+reserve > limit {
			if quranShownAyat == 0 { // one long āyah: as many whole lines as fit
				used = limit - reserve
				quranShownAyat = 1
			}
			break
		}
		used += h
		quranShownAyat++
	}
	if p.From+quranShownAyat-1 < p.To {
		used += moreH
	}
	quranPreview = rect{pad, y + refH + qs(4), width - pad, y + refH + qs(4) + used + 2*qs(10)}
	y = quranPreview.Bottom + qs(10)

	// Separator, then [mode] … [Cancel] [Insert], the translation, and the keys.
	quranSepY = y
	y += qs(1) + qs(10)
	x := pad
	for i, m := range quran.InsertModes {
		w := textWidth(qFontButton, m.Title(), 0) + qs(24)
		quranModeBtns[i] = rect{x, y, x + w, y + bh}
		x += w
	}
	iw := max(qs(84), textWidth(qFontButton, "Insert", 0)+qs(32))
	cw := max(qs(76), textWidth(qFontButton, "Cancel", 0)+qs(28))
	quranInsertBtn = rect{width - pad - iw, y, width - pad, y + bh}
	quranCancelBtn = rect{quranInsertBtn.Left - qs(8) - cw, y, quranInsertBtn.Left - qs(8), y + bh}
	y += bh
	quranTransBtn = rect{}
	if settings.quranMode() != quran.Arabic {
		y += qs(8)
		lw := textWidth(qFontSmall, "Translation", 0) + qs(10)
		quranTransBtn = rect{pad + lw, y, pad + lw + qs(400), y + qs(28)}
		y += qs(28)
	}
	y += qs(10)
	quranHint = rect{pad, y, width - pad, y + lineHeight(qFontSmall)}
	quranW, quranH = width, quranHint.Bottom+qs(12)
}

// ---------- Painting ----------

func fill(hdc uintptr, r rect, brush uintptr) {
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
}

func frame(hdc uintptr, r rect, brush uintptr, thick bool) {
	pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
	if thick {
		in := rect{r.Left + 1, r.Top + 1, r.Right - 1, r.Bottom - 1}
		pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&in)), brush)
	}
}

func text(hdc uintptr, s string, r rect, font uintptr, color uintptr, flags uintptr) {
	pSelectObject.Call(hdc, font)
	pSetTextColor.Call(hdc, color)
	drawText(hdc, s, &r, flags|dtNoPrefix)
}

var (
	colText   = rgb(28, 30, 34)
	colSoft   = rgb(96, 101, 110)
	colFaint  = rgb(140, 145, 152)
	colAccent = rgb(46, 110, 200)
	colWhite  = rgb(255, 255, 255)
)

// segmented draws joined buttons with the chosen one filled.
func segmented(hdc uintptr, parts []rect, titles []string, on int) {
	for i, r := range parts {
		if i == on {
			fill(hdc, r, qBrushAccent)
			text(hdc, titles[i], r, qFontButton, colWhite, dtCenter|dtVCenter|dtSingleLine|dtEndEllipsis)
		} else {
			fill(hdc, r, qBrushWhite)
			text(hdc, titles[i], r, qFontButton, colText, dtCenter|dtVCenter|dtSingleLine|dtEndEllipsis)
		}
	}
	all := rect{parts[0].Left, parts[0].Top, parts[len(parts)-1].Right, parts[0].Bottom}
	frame(hdc, all, qBrushBorder, false)
	for i := 1; i < len(parts); i++ {
		if i != on && i-1 != on {
			fill(hdc, rect{parts[i].Left, all.Top + qs(6), parts[i].Left + 1, all.Bottom - qs(6)}, qBrushBorder)
		}
	}
}

func button(hdc uintptr, r rect, title string, primary bool) {
	if primary {
		fill(hdc, r, qBrushAccent)
		text(hdc, title, r, qFontButton, colWhite, dtCenter|dtVCenter|dtSingleLine)
		return
	}
	fill(hdc, r, qBrushWhite)
	frame(hdc, r, qBrushBorder, false)
	text(hdc, title, r, qFontButton, colText, dtCenter|dtVCenter|dtSingleLine)
}

// dropdown: a white box with the value and ▾; the focused one has an accent frame.
func dropdown(hdc uintptr, r rect, value string, focused bool) {
	fill(hdc, r, qBrushWhite)
	if focused {
		frame(hdc, r, qBrushAccent, true)
	} else {
		frame(hdc, r, qBrushBorder, false)
	}
	text(hdc, value, rect{r.Left + qs(10), r.Top, r.Right - qs(24), r.Bottom}, qFontUI, colText, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
	text(hdc, "▾", rect{r.Right - qs(22), r.Top, r.Right - qs(6), r.Bottom}, qFontUI, colSoft, dtCenter|dtVCenter|dtSingleLine)
}

func paintQuran(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc rect
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	fill(hdc, rc, qBrushBg)
	pSetBkMode.Call(hdc, transparent)
	lib := quranLibrary()
	if !quranOpen || lib == nil {
		return
	}
	b := &quranB
	pad := qs(16)

	// Header
	text(hdc, "Insert from the Qur'an", rect{pad, quranCloseBtn.Top, quranCloseBtn.Left, quranCloseBtn.Bottom}, qFontTitle, colText, dtLeft|dtVCenter|dtSingleLine)
	text(hdc, "✕", quranCloseBtn, qFontUI, colSoft, dtCenter|dtVCenter|dtSingleLine)

	paintField(hdc)
	segmented(hdc, quranSwitch[:], []string{"Arabic", "Translation"}, int(b.In))

	// Results
	if quranCaption.Bottom > 0 {
		text(hdc, quranCaptionText(), quranCaption, qFontSmall, colSoft, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
	}
	if quranList.Bottom > 0 {
		fill(hdc, quranList, qBrushCard)
		frame(hdc, quranList, qBrushBorder, false)
		for i, r := range quranRowRects {
			paintRow(hdc, r, b.Res.Rows[b.Top+i], b.Top+i == b.Sel)
		}
		if n := len(b.Res.Rows); n > quranVisibleRows { // a thin scroll bar
			track := rect{quranList.Right - qs(5), quranList.Top + qs(3), quranList.Right - qs(2), quranList.Bottom - qs(3)}
			th := track.Bottom - track.Top
			h := max(qs(16), th*quranVisibleRows/int32(n))
			t := track.Top + (th-h)*int32(b.Top)/int32(n-quranVisibleRows)
			fill(hdc, rect{track.Left, t, track.Right, t + h}, qBrushBorder)
		}
	}

	// Sūrah / From / To
	p := b.Pick.passage()
	s := lib.Surah(p.Surah)
	labels := [3]string{"Sūrah", "From āyah", "To āyah"}
	values := [3]string{surahTitle(s), fmt.Sprint(p.From), fmt.Sprint(p.To)}
	for f, r := range quranDrops {
		on := b.OnPick && b.Pick.Field == f
		label, col := labels[f], colSoft
		if on {
			col = colAccent
			if b.Pick.Typed != "" {
				label += ":  " + b.Pick.Typed
			}
		}
		lh := lineHeight(qFontSmall)
		text(hdc, label, rect{r.Left + qs(1), r.Top - qs(3) - lh, r.Right + qs(40), r.Top - qs(3)}, qFontSmall, col, dtLeft|dtSingleLine)
		dropdown(hdc, r, values[f], on)
	}
	if s.Ayahs > 1 {
		info := rect{quranDrops[2].Right + qs(12), quranDrops[2].Top, rc.Right - pad, quranDrops[2].Bottom}
		text(hdc, fmt.Sprintf("of %d", s.Ayahs), info, qFontSmall, colFaint, dtLeft|dtVCenter|dtSingleLine)
	}

	paintPreview(hdc, p, s)

	// Bottom
	fill(hdc, rect{pad, quranSepY, rc.Right - pad, quranSepY + qs(1)}, qBrushSep)
	titles := make([]string, len(quran.InsertModes))
	on := 0
	for i, m := range quran.InsertModes {
		titles[i] = m.Title()
		if settings.quranMode() == m {
			on = i
		}
	}
	segmented(hdc, quranModeBtns[:], titles, on)
	button(hdc, quranCancelBtn, "Cancel", false)
	button(hdc, quranInsertBtn, "Insert", true)
	if quranTransBtn.Bottom > 0 {
		text(hdc, "Translation", rect{pad, quranTransBtn.Top, quranTransBtn.Left, quranTransBtn.Bottom}, qFontSmall, colSoft, dtLeft|dtVCenter|dtSingleLine)
		label := "No translation"
		if t := lib.Translation(settings.quranTranslation()); t != nil {
			label = t.Title()
		}
		dropdown(hdc, quranTransBtn, label, false)
	}
	text(hdc, quranHintText(), quranHint, qFontSmall, colFaint, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
}

// paintField: the search field. Arabic is drawn right to left in the Qur'an font; when it's longer
// than the field, its end (what was typed last) stays in sight. The caret is a thin bar.
func paintField(hdc uintptr) {
	b := &quranB
	f := quranField
	fill(hdc, f, qBrushWhite)
	if b.OnPick {
		frame(hdc, f, qBrushBorder, false)
	} else {
		frame(hdc, f, qBrushAccent, true)
	}
	in := rect{f.Left + qs(10), f.Top + 2, f.Right - qs(10), f.Bottom - 2}
	s := b.shown()
	if s == "" {
		ph := quranPlaceholderArabic
		if b.In == quran.InTranslation {
			ph = quranPlaceholderTrans
		}
		text(hdc, ph, in, qFontUI, colFaint, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
		if !b.OnPick {
			fill(hdc, rect{in.Left - qs(3), in.Top + qs(7), in.Left - qs(3) + qs(2), in.Bottom - qs(7)}, qBrushAccent)
		}
		return
	}
	caretX := int32(0)
	if b.shownIsArabic() {
		w := textWidth(qFontField, s, dtRTLReading)
		flags := uintptr(dtRight | dtRTLReading | dtVCenter | dtSingleLine)
		caretX = in.Right - w - qs(3)
		if w > in.Right-in.Left {
			flags = dtLeft | dtRTLReading | dtVCenter | dtSingleLine
			caretX = in.Left - qs(3)
		}
		text(hdc, s, in, qFontField, colText, flags)
	} else {
		w := textWidth(qFontUI, s, 0)
		flags := uintptr(dtLeft | dtVCenter | dtSingleLine)
		caretX = in.Left + w + qs(1)
		if w > in.Right-in.Left {
			flags = dtRight | dtVCenter | dtSingleLine
			caretX = in.Right + qs(1)
		}
		text(hdc, s, in, qFontUI, colText, flags)
	}
	if !b.OnPick {
		fill(hdc, rect{caretX, in.Top + qs(7), caretX + qs(2), in.Bottom - qs(7)}, qBrushAccent)
	}
}

// paintRow: one result. Left: its reference (or "Sūrah"); then the āyah (Arabic, right to left, one
// line), the translation it was found in, or the sūrah's name.
func paintRow(hdc uintptr, r rect, row quran.Suggestion, selected bool) {
	lib := quranLibrary()
	if selected {
		fill(hdc, r, qBrushSelected)
		fill(hdc, rect{r.Left, r.Top, r.Left + qs(3), r.Bottom}, qBrushAccent)
	}
	label := rect{r.Left + qs(10), r.Top, r.Left + qs(10) + qs(58), r.Bottom}
	body := rect{label.Right + qs(8), r.Top, r.Right - qs(12), r.Bottom}
	switch row.Kind {
	case quran.KindProblem:
		text(hdc, row.Text, rect{label.Left, r.Top, body.Right, r.Bottom}, qFontUI, colSoft, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
	case quran.KindSurah:
		s := row.Surah
		text(hdc, "Sūrah", label, qFontBold, colSoft, dtLeft|dtVCenter|dtSingleLine)
		aw := textWidth(qFontUI, s.Arabic, dtRTLReading)
		text(hdc, s.Arabic, body, qFontUI, colText, dtRight|dtRTLReading|dtVCenter|dtSingleLine)
		body.Right -= aw + qs(12)
		text(hdc, fmt.Sprintf("%d. %s  ·  %d āyāt", s.Number, s.Name, s.Ayahs), body, qFontUI, colText, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
	case quran.KindPassage:
		p := row.Passage
		text(hdc, p.Reference(), label, qFontBold, colSoft, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
		t := p.Reference()
		if s := lib.Surah(p.Surah); s != nil {
			t = s.Name + " " + t
			if p.To > p.From {
				t += fmt.Sprintf("  ·  %d āyāt", p.To-p.From+1)
			}
		}
		text(hdc, t, body, qFontUI, colText, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
	case quran.KindHit:
		p := row.Passage
		text(hdc, p.Reference(), label, qFontBold, colSoft, dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
		if row.InTranslation == "" {
			text(hdc, shown(lib.Arabic(p.Surah, p.From)), body, qFontRow, colText, dtRight|dtRTLReading|dtVCenter|dtSingleLine|dtEndEllipsis)
		} else {
			f, dir := transStyle(row.InTranslation)
			text(hdc, lib.TranslationText(row.InTranslation, p.Surah, p.From), body, f, colText, dir|dtLeft|dtVCenter|dtSingleLine|dtEndEllipsis)
		}
	}
}

// paintPreview: "Al-Baqara 2:255-257 · 3 āyāt", then the āyāt in a card, as many as layoutQuran fitted.
func paintPreview(hdc uintptr, p quran.Passage, s *quran.Surah) {
	pad := qs(16)
	ref := s.Name + " " + p.Reference()
	if n := p.To - p.From + 1; n > 1 {
		ref += fmt.Sprintf("  ·  %d āyāt", n)
	}
	refH := lineHeight(qFontBold)
	text(hdc, ref, rect{pad + qs(1), quranPreview.Top - qs(4) - refH, quranPreview.Right, quranPreview.Top - qs(4)}, qFontBold, colSoft, dtLeft|dtSingleLine|dtEndEllipsis)
	card := quranPreview
	fill(hdc, card, qBrushCard)
	frame(hdc, card, qBrushBorder, false)
	in := rect{card.Left + qs(12), card.Top + qs(10), card.Right - qs(12), card.Bottom - qs(10)}
	shownTo := p.From + quranShownAyat - 1
	moreH := int32(0)
	if shownTo < p.To {
		moreH = lineHeight(qFontSmall) + qs(4)
	}
	y := in.Top
	tf, dir := transStyle(settings.quranTranslation())
	arLine, trLine := lineHeight(qFontPreview), lineHeight(tf)
	for a := p.From; a <= shownTo; a++ {
		if a > p.From {
			y += qs(10)
		}
		ar, tr := previewTexts(p, a)
		arH, trH := ayahHeight(p, a, in.Right-in.Left)
		bottom := in.Bottom - moreH
		if ar != "" {
			h := capLines(arH, arLine, bottom-y)
			r := rect{in.Left, y, in.Right, y + h}
			flags := uintptr(dtRight | dtRTLReading | dtWordBreak | dtEditControl)
			if h < arH {
				flags |= dtEndEllipsis
			}
			text(hdc, ar, r, qFontPreview, colText, flags)
			y += h
			if tr != "" {
				y += qs(4)
			}
		}
		if tr != "" && y < bottom {
			h := capLines(trH, trLine, bottom-y)
			r := rect{in.Left, y, in.Right, y + h}
			flags := dir | dtWordBreak | dtEditControl
			if h < trH {
				flags |= dtEndEllipsis
			}
			text(hdc, tr, r, tf, colSoft, flags)
			y += h
		}
	}
	if moreH > 0 {
		r := rect{in.Left, in.Bottom - moreH + qs(4), in.Right, in.Bottom}
		text(hdc, fmt.Sprintf("… and %d more āyāt (all of them are inserted)", p.To-shownTo), r, qFontSmall, colFaint, dtLeft|dtSingleLine)
	}
}
