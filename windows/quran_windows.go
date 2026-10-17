//go:build windows

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"qalam/engine"
	"qalam/quran"
)

// The "insert from the Qur'an" box (Ctrl+Alt+Q), like ⌃⇧Q on the Mac. Choose tab: three dropdowns
// (Sūrah, From āyah, To āyah; a click opens a menu), a preview in the Qur'an font, and Insert.
// Search tab: a word, harakat optional, and the āyāt that have it.
// Like the options panel it never takes focus; the keyboard hook sends it the keys (see quranpick.go).

var (
	quranLib     *quran.Library
	quranLibOnce sync.Once

	quranHwnd   uintptr
	quranOpen   bool
	quranSearch bool      // the Search tab is showing
	quranP      quranPick // the Choose tab
	quranQuery  string    // Search tab
	quranRows   []quran.Suggestion
	quranSel    int
	quranTotal  int
	quranGen    int  // bumped for every new search
	quranMenu   bool // one of the box's menus is open (it has the focus meanwhile)
	quranFg     uintptr
	quranX      int32 // where the box opened (it stays there while open)
	quranY      int32

	// Clickable areas, set by layoutQuran.
	quranTabs     [2]rect
	quranCloseBtn rect
	quranDrops    [3]rect // Sūrah, From, To
	quranPreview  rect
	quranField    rect
	quranRowRects []rect
	quranInfo     rect
	quranModeBtns [3]rect
	quranTransBtn rect
	quranInsert   rect
	quranHint     rect
	quranSepY     int32
	quranW, quranH int32

	quranArabicFont uintptr
	quranNameFont   uintptr
	quranTextFont   uintptr
	quranRefFont    uintptr
	quranTitleFont  uintptr
	quranFamily     = ""
	accentBrush     uintptr
	sepBrush        uintptr

	pSetTimer         = user32.NewProc("SetTimer")
	pKillTimer        = user32.NewProc("KillTimer")
	pMessageBeep      = user32.NewProc("MessageBeep")
	pOpenClipboard    = user32.NewProc("OpenClipboard")
	pCloseClipboard   = user32.NewProc("CloseClipboard")
	pGetClipboardData = user32.NewProc("GetClipboardData")
	pGlobalLock       = kernel32.NewProc("GlobalLock")
	pGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
)

const (
	dtRight       = 0x0002
	dtWordBreak   = 0x0010
	dtEditControl = 0x2000
	dtEndEllipsis = 0x8000
	wmTimer       = 0x0113
	cfUnicodeText = 13
	quranRowLimit = 6
	tpmLeftAlign  = 0x0000
	mfMenuBreak   = 0x0040
)

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
	quranArabicFont = makeFont(scale(26), quranFamily, 400)
	quranNameFont = makeFont(scale(20), quranFamily, 400)
	quranTextFont = makeFont(scale(13), "Segoe UI", 400)
	quranRefFont = makeFont(scale(11), "Segoe UI", 600)
	quranTitleFont = makeFont(scale(14), "Segoe UI", 600)
	accentBrush, _, _ = pCreateSolidBrush.Call(rgb(90, 140, 210))
	sepBrush, _, _ = pCreateSolidBrush.Call(rgb(225, 225, 225))
}

// shown: the āyah as drawn (KFGQPC Hafs needs three marks swapped, see quran.ForKFGQPC).
func shown(s string) string {
	if strings.Contains(strings.ToUpper(quranFamily), "KFGQPC") {
		return quran.ForKFGQPC(s)
	}
	return s
}

// ---------- Keys ----------

func openQuran() {
	lib := quranLibrary()
	if lib == nil || quranHwnd == 0 {
		pMessageBeep.Call(0xFFFFFFFF)
		return
	}
	quranOpen, quranSearch, quranMenu = true, false, false
	quranP = newQuranPick(lib, settings.QuranLast)
	quranQuery, quranRows, quranSel, quranTotal = "", nil, 0, 0
	quranGen++
	quranFg, _, _ = pGetForegroundWindow.Call()
	quranX, quranY = caretPosition()
	tr := settings.quranTranslation()
	go func() { // read the texts and build the search index now, so the first search is quick
		lib.TranslationText(tr, 1, 1)
		lib.SearchArabic("بسم", false, 0)
	}()
	showQuran()
	pSetTimer.Call(quranHwnd, 1, 300, 0)
}

func closeQuran() {
	if !quranOpen {
		return
	}
	quranOpen = false
	quranQuery, quranRows = "", nil
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
	if quranSearch && ctrl && !alt && !win && !shift && vk == 'V' {
		if t := clipboardText(); t != "" {
			quranQuery += strings.Join(strings.Fields(t), " ")
			searchQuran()
		}
		return true
	}
	if ctrl || alt || win { // other shortcuts go to the app
		closeQuran()
		return false
	}
	lib := quranLibrary()
	switch vk {
	case vkEscape:
		closeQuran()
		return true
	case vkReturn:
		runOnUI(insertChosen)
		return true
	}
	if !quranSearch {
		switch vk {
		case vkUp:
			quranP.step(lib, -1)
		case vkDown:
			quranP.step(lib, 1)
		case vkLeft:
			quranP.move(-1)
		case vkRight:
			quranP.move(1)
		case vkTab:
			if shift {
				quranP.move(-1)
			} else {
				quranP.move(1)
			}
		case vkBack:
			quranP.backspace(lib)
		case vkSpace:
			quranP.typeRune(lib, ' ')
		default:
			if r, ok := usKey(vk, shift); ok && quranP.typeRune(lib, r) {
				quranSearch = true
			}
		}
		showQuran()
		return true
	}
	switch vk {
	case vkUp, vkDown:
		if n := len(quranRows); n > 0 {
			step := 1
			if vk == vkUp {
				step = n - 1
			}
			quranSel = (quranSel + step) % n
			showQuran()
		}
	case vkTab:
		quranSearch = false
		showQuran()
	case vkBack:
		if quranQuery == "" {
			quranSearch = false
			showQuran()
			return true
		}
		_, size := utf8.DecodeLastRuneInString(quranQuery)
		quranQuery = quranQuery[:len(quranQuery)-size]
		searchQuran()
	case vkSpace:
		quranQuery += " "
		searchQuran()
	default:
		if r, ok := usKey(vk, shift); ok {
			quranQuery += string(r)
			searchQuran()
		}
	}
	return true
}

// searchQuran looks the word up off the keyboard thread (the first search builds the index) and
// shows the āyāt when they come back, unless something newer was typed by then.
// Latin is read the Qalam way (rahmaan → رحمان), with similar letters counting too.
func searchQuran() {
	lib := quranLibrary()
	quranGen++
	gen, q := quranGen, strings.TrimSpace(quranQuery)
	showQuran() // the typed text at once
	if q == "" {
		quranRows, quranTotal, quranSel = nil, 0, 0
		showQuran()
		return
	}
	go func() {
		ar, phonetic := q, false
		if !hasArabicRune(q) {
			o := engine.DefaultOptions()
			o.Harakat = engine.NoHarakat
			ar, phonetic = engine.Text(q, o), true
		}
		hits, total := lib.SearchArabic(ar, phonetic, quranRowLimit)
		rows := make([]quran.Suggestion, 0, len(hits))
		for _, h := range hits {
			rows = append(rows, quran.Suggestion{Kind: quran.KindHit, Passage: h})
		}
		if len(rows) == 0 {
			rows = []quran.Suggestion{{Kind: quran.KindProblem, Text: "Nothing found for “" + q + "”"}}
		}
		runOnUI(func() {
			if !quranOpen || gen != quranGen {
				return
			}
			quranRows, quranTotal, quranSel = rows, total, 0
			showQuran()
		})
	}()
}

func hasArabicRune(s string) bool {
	for _, r := range s {
		if r >= 0x0600 && r <= 0x06FF {
			return true
		}
	}
	return false
}

// insertChosen: Choose tab → the chosen āyāt (remembered for next time); Search tab → the highlighted āyah.
func insertChosen() {
	if !quranOpen {
		return
	}
	var p quran.Passage
	if quranSearch {
		if quranSel >= len(quranRows) || quranRows[quranSel].Kind != quran.KindHit {
			pMessageBeep.Call(0xFFFFFFFF)
			return
		}
		p = quranRows[quranSel].Passage
	} else {
		p = quranP.passage()
		settings.QuranLast = p.Reference()
		saveSettings(settings)
	}
	insertQuran(p)
}

func insertQuran(p quran.Passage) {
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

// ---------- Window ----------

func quranProc(hwnd, message, wParam, lParam uintptr) uintptr {
	switch message {
	case wmMouseActivate:
		return maNoActivate
	case wmLButtonDown:
		x, y := int32(int16(lParam&0xFFFF)), int32(int16((lParam>>16)&0xFFFF))
		onQuranClick(x, y)
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

func onQuranClick(x, y int32) {
	lib := quranLibrary()
	switch {
	case inRect(x, y, quranCloseBtn):
		closeQuran()
		return
	case inRect(x, y, quranTabs[0]):
		quranSearch = false
	case inRect(x, y, quranTabs[1]):
		quranSearch = true
	case inRect(x, y, quranInsert):
		insertChosen()
		return
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
		if i := quranPopup(quranTransBtn, titles, cur, 0); i >= 0 {
			settings.QuranTranslation = have[i].ID
			saveSettings(settings)
		}
	}
	for i, r := range quranModeBtns {
		if inRect(x, y, r) {
			settings.QuranInsert = string(quran.InsertModes[i])
			saveSettings(settings)
		}
	}
	if !quranSearch {
		for f, r := range quranDrops {
			if !inRect(x, y, r) {
				continue
			}
			quranP.Field, quranP.Typed = f, ""
			n := quranP.ayahs(lib)
			switch f {
			case fieldSurah:
				titles := make([]string, 114)
				for i := range titles {
					s := lib.Surah(i + 1)
					titles[i] = fmt.Sprintf("%d. %s  (%d)", s.Number, s.Name, s.Ayahs)
				}
				if i := quranPopup(r, titles, quranP.Surah-1, 38); i >= 0 {
					quranP.setSurah(i + 1)
				}
			case fieldFrom:
				if i := quranPopup(r, numbers(1, n), quranP.From-1, 30); i >= 0 {
					quranP.setFrom(i + 1)
				}
			case fieldTo:
				if i := quranPopup(r, numbers(quranP.From, n), quranP.To-quranP.From, 30); i >= 0 {
					quranP.To = quranP.From + i
				}
			}
		}
	} else {
		for i, r := range quranRowRects {
			if inRect(x, y, r) {
				quranSel = i
				insertChosen()
				return
			}
		}
	}
	showQuran()
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

// arabicNumber: the āyah-end mark with its number. KFGQPC Hafs circles plain Arabic digits by itself;
// other fonts need ۝ (U+06DD) before them.
func arabicNumber(n int) string {
	d := []rune(fmt.Sprint(n))
	for i := range d {
		d[i] = 0x0660 + d[i] - '0'
	}
	if strings.Contains(strings.ToUpper(quranFamily), "KFGQPC") {
		return string(d)
	}
	return "\u06DD" + string(d)
}

const quranSearchPlaceholder = "Type a word: rahmaan, or paste Arabic with Ctrl+V (harakat optional)"

func quranHintText() string {
	if quranSearch {
		s := "↑↓ choose · Enter insert · Tab back to Choose · Ctrl+V paste · Esc close"
		if h := settings.hotkey("quran"); h != nil {
			s += " · " + h.String() + " opens this box"
		}
		return s
	}
	return "↑↓ change · ←→ or Tab next box · type 36 or yasin, or 2:255-257 · Enter insert · / search · Esc close"
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

func capLines(h, line int32, n int32) int32 {
	if h > line*n {
		return line * n
	}
	return h
}

// previewParts: the first āyāt of the choice (with their numbers), "… and N more", and the translation.
func previewParts() (ar, more, tr string) {
	lib := quranLibrary()
	p := quranP.passage()
	last := p.To
	if last > p.From+2 {
		last = p.From + 2
		more = fmt.Sprintf("… and %d more āyāt", p.To-last)
	}
	var b strings.Builder
	for a := p.From; a <= last; a++ {
		if a > p.From {
			b.WriteString(" ")
		}
		b.WriteString(lib.Arabic(p.Surah, a) + " " + arabicNumber(a))
	}
	if settings.quranMode() != quran.Arabic {
		tr = lib.TranslationText(settings.quranTranslation(), p.Surah, p.From)
	}
	return shown(b.String()), more, tr
}

func modeTitle(m quran.Insert) string {
	switch m {
	case quran.Arabic:
		return "Arabic only"
	case quran.TranslationOnly:
		return "Translation only"
	}
	return "Arabic + translation"
}

// showQuran lays out the box and shows it where it opened (below the caret, moved in at screen edges).
func showQuran() {
	if !quranOpen || quranHwnd == 0 {
		return
	}
	layoutQuran()
	x, cy := quranX, quranY
	width, height := quranW, quranH
	sw, _, _ := pGetSystemMetrics.Call(smCxScreen)
	sh, _, _ := pGetSystemMetrics.Call(smCyScreen)
	if x+width > int32(sw) {
		x = int32(sw) - width
	}
	if x < 0 {
		x = 0
	}
	if cy+height > int32(sh) {
		if above := cy - height - scale(36); above >= 0 {
			cy = above
		} else {
			cy = int32(sh) - height
		}
	}
	pSetWindowPos.Call(quranHwnd, hwndTopmost, uintptr(x), uintptr(cy), uintptr(width), uintptr(height), swpNoActivate)
	pInvalidateRect.Call(quranHwnd, 0, 1)
	pShowWindow.Call(quranHwnd, swShowNA)
}

func layoutQuran() {
	pad := scale(10)
	width := scale(600)
	inner := width - 2*pad - 2*scale(12)
	_, arLine := measure(quranArabicFont, "بِسْمِ", dtRTLReading)
	_, trLine := measure(quranTextFont, "Ag", 0)
	_, refLine := measure(quranRefFont, "Ag", 0)

	// Header: title · [Choose] [Search] [✕]
	y := pad
	hh := scale(26)
	quranCloseBtn = rect{width - pad - scale(28), y, width - pad, y + hh}
	sw, _ := measure(smallFont, "Search", 0)
	cw, _ := measure(smallFont, "Choose", 0)
	quranTabs[1] = rect{quranCloseBtn.Left - scale(6) - sw - scale(20), y, quranCloseBtn.Left - scale(6), y + hh}
	quranTabs[0] = rect{quranTabs[1].Left - scale(4) - cw - scale(20), y, quranTabs[1].Left - scale(4), y + hh}
	y += hh + scale(8)

	quranDrops = [3]rect{}
	quranPreview, quranField = rect{}, rect{}
	quranRowRects = quranRowRects[:0]
	quranInfo = rect{}
	if !quranSearch {
		// Sūrah ▾  From āyah ▾  To āyah ▾ (a label above each)
		top := y + refLine + scale(4)
		dh := scale(30)
		small := scale(110)
		quranDrops[2] = rect{width - pad - small, top, width - pad, top + dh}
		quranDrops[1] = rect{quranDrops[2].Left - scale(10) - small, top, quranDrops[2].Left - scale(10), top + dh}
		quranDrops[0] = rect{pad, top, quranDrops[1].Left - scale(10), top + dh}
		y = top + dh + scale(10)
		ar, more, tr := previewParts()
		h := refLine + scale(4) + capLines(textHeight(quranArabicFont, ar, inner, dtRTLReading|dtRight), arLine, 4)
		if more != "" {
			h += refLine + scale(2)
		}
		if tr != "" {
			h += scale(4) + capLines(textHeight(quranTextFont, tr, inner, 0), trLine, 3)
		}
		quranPreview = rect{pad, y, width - pad, y + h + scale(14)}
		y = quranPreview.Bottom
	} else {
		quranField = rect{pad, y, width - pad, y + scale(32)}
		y = quranField.Bottom + scale(6)
		for i, r := range quranRows {
			h := scale(32)
			if r.Kind == quran.KindHit {
				h = refLine + scale(2) + arLine
				if i == quranSel {
					ar := shown(quranLibrary().Arabic(r.Passage.Surah, r.Passage.From))
					h = refLine + scale(2) + capLines(textHeight(quranArabicFont, ar, inner, dtRTLReading|dtRight), arLine, 3)
				}
				h += scale(10)
			}
			quranRowRects = append(quranRowRects, rect{pad, y, width - pad, y + h})
			y += h + scale(3)
		}
		if quranTotal > len(quranRows) {
			quranInfo = rect{pad + scale(4), y, width - pad, y + scale(18)}
			y = quranInfo.Bottom
		}
	}
	y += scale(6)
	quranSepY = y
	y += scale(8)

	// Bottom: [Arabic only] [Arabic + translation] [Translation only] [translation ▾] ... [Insert]
	bh := scale(26)
	x := pad
	for i, m := range quran.InsertModes {
		w, _ := measure(smallFont, modeTitle(m), 0)
		quranModeBtns[i] = rect{x, y, x + w + scale(16), y + bh}
		x = quranModeBtns[i].Right + scale(4)
	}
	iw, _ := measure(quranTitleFont, "Insert ⏎", 0)
	quranInsert = rect{width - pad - iw - scale(28), y, width - pad, y + bh}
	quranTransBtn = rect{}
	if settings.quranMode() != quran.Arabic {
		quranTransBtn = rect{pad, y + bh + scale(6), quranInsert.Left - scale(8), y + 2*bh + scale(6)}
		y += bh + scale(6)
	}
	y += bh + scale(6)
	quranHint = rect{pad, y, width - pad, y + scale(18)}
	quranW, quranH = width, quranHint.Bottom+pad
}

// chip draws a small button: filled with the accent colour when on.
func chip(hdc uintptr, r rect, text string, on bool, font uintptr) {
	if on {
		pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), accentBrush)
		pSetTextColor.Call(hdc, rgb(255, 255, 255))
	} else {
		pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&r)), borderBrush)
		pSetTextColor.Call(hdc, rgb(40, 40, 40))
	}
	pSelectObject.Call(hdc, font)
	drawText(hdc, text, &r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
}

func paintQuran(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc rect
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), previewBrush)
	pSetBkMode.Call(hdc, transparent)
	lib := quranLibrary()
	if !quranOpen || lib == nil {
		return
	}
	pad := scale(12)
	_, arLine := measure(quranArabicFont, "بِسْمِ", dtRTLReading)
	_, refLine := measure(quranRefFont, "Ag", 0)
	_, trLine := measure(quranTextFont, "Ag", 0)

	// Header
	pSelectObject.Call(hdc, quranTitleFont)
	pSetTextColor.Call(hdc, rgb(20, 20, 20))
	t := rect{scale(10), quranTabs[0].Top, quranTabs[0].Left, quranTabs[0].Bottom}
	drawText(hdc, "Insert from the Qur'an", &t, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	chip(hdc, quranTabs[0], "Choose", !quranSearch, smallFont)
	chip(hdc, quranTabs[1], "Search", quranSearch, smallFont)
	chip(hdc, quranCloseBtn, "✕", false, smallFont)

	if !quranSearch {
		s := lib.Surah(quranP.Surah)
		labels := [3]string{"Sūrah", "From āyah", "To āyah"}
		values := [3]string{
			fmt.Sprintf("%d. %s  ·  %s  ·  %d āyāt", s.Number, s.Name, s.English, s.Ayahs),
			fmt.Sprint(quranP.From), fmt.Sprint(quranP.passage().To),
		}
		for f, r := range quranDrops {
			on := quranP.Field == f
			label := labels[f]
			if on && quranP.Typed != "" {
				label += ": " + quranP.Typed + "▏"
			}
			pSelectObject.Call(hdc, quranRefFont)
			if on {
				pSetTextColor.Call(hdc, rgb(40, 100, 200))
			} else {
				pSetTextColor.Call(hdc, rgb(120, 120, 120))
			}
			l := rect{r.Left + scale(2), r.Top - refLine - scale(3), r.Right, r.Top - scale(2)}
			drawText(hdc, label, &l, dtLeft|dtSingleLine|dtNoPrefix)
			box := r
			pFillRect.Call(hdc, uintptr(unsafe.Pointer(&box)), highlightBrush)
			if on {
				pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&box)), accentBrush)
				in := rect{box.Left + 1, box.Top + 1, box.Right - 1, box.Bottom - 1}
				pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&in)), accentBrush)
			} else {
				pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&box)), borderBrush)
			}
			pSelectObject.Call(hdc, quranTextFont)
			pSetTextColor.Call(hdc, rgb(20, 20, 20))
			v := rect{box.Left + scale(8), box.Top, box.Right - scale(22), box.Bottom}
			drawText(hdc, values[f], &v, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
			a := rect{box.Right - scale(20), box.Top, box.Right - scale(4), box.Bottom}
			pSetTextColor.Call(hdc, rgb(110, 110, 110))
			drawText(hdc, "▾", &a, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
		}

		// Preview card
		r := quranPreview
		pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), highlightBrush)
		in := rect{r.Left + pad, r.Top + scale(7), r.Right - pad, r.Bottom - scale(7)}
		p := quranP.passage()
		ref := s.Name + " " + p.Reference()
		if n := p.To - p.From + 1; n > 1 {
			ref += fmt.Sprintf("  ·  %d āyāt", n)
		}
		pSelectObject.Call(hdc, quranRefFont)
		pSetTextColor.Call(hdc, rgb(120, 120, 120))
		rr := rect{in.Left, in.Top, in.Right, in.Top + refLine}
		drawText(hdc, ref, &rr, dtLeft|dtSingleLine|dtNoPrefix)
		ar, more, tr := previewParts()
		pSelectObject.Call(hdc, quranArabicFont)
		pSetTextColor.Call(hdc, rgb(20, 20, 20))
		top := rr.Bottom + scale(4)
		h := capLines(textHeight(quranArabicFont, ar, in.Right-in.Left, dtRTLReading|dtRight), arLine, 4)
		ab := rect{in.Left, top, in.Right, top + h}
		drawText(hdc, ar, &ab, dtRight|dtRTLReading|dtWordBreak|dtNoPrefix|dtEditControl|dtEndEllipsis)
		y := ab.Bottom
		if more != "" {
			pSelectObject.Call(hdc, quranRefFont)
			pSetTextColor.Call(hdc, rgb(140, 140, 140))
			m := rect{in.Left, y + scale(2), in.Right, y + scale(2) + refLine}
			drawText(hdc, more, &m, dtLeft|dtSingleLine|dtNoPrefix)
			y = m.Bottom
		}
		if tr != "" {
			pSelectObject.Call(hdc, quranTextFont)
			pSetTextColor.Call(hdc, rgb(90, 90, 90))
			th := capLines(textHeight(quranTextFont, tr, in.Right-in.Left, 0), trLine, 3)
			b := rect{in.Left, y + scale(4), in.Right, y + scale(4) + th}
			drawText(hdc, tr, &b, dtLeft|dtWordBreak|dtNoPrefix|dtEditControl|dtEndEllipsis)
		}
	} else {
		// Search field and results
		f := quranField
		pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&f)), accentBrush)
		in := rect{f.Left + 1, f.Top + 1, f.Right - 1, f.Bottom - 1}
		pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&in)), accentBrush)
		field := rect{f.Left + scale(10), f.Top, f.Right - scale(8), f.Bottom}
		pSelectObject.Call(hdc, quranTextFont)
		if quranQuery == "" {
			pSetTextColor.Call(hdc, rgb(160, 160, 160))
			drawText(hdc, quranSearchPlaceholder, &field, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
		} else {
			pSetTextColor.Call(hdc, rgb(20, 20, 20))
			drawText(hdc, quranQuery+"▏", &field, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
		}
		for i, row := range quranRows {
			if i >= len(quranRowRects) {
				break
			}
			r := quranRowRects[i]
			if i == quranSel {
				pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), highlightBrush)
			}
			in := rect{r.Left + pad, r.Top + scale(5), r.Right - pad, r.Bottom - scale(5)}
			if row.Kind != quran.KindHit {
				pSelectObject.Call(hdc, quranTextFont)
				pSetTextColor.Call(hdc, rgb(120, 120, 120))
				t := rect{in.Left, r.Top, in.Right, r.Bottom}
				drawText(hdc, row.Text, &t, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
				continue
			}
			p := row.Passage
			ref := p.Reference()
			if s := lib.Surah(p.Surah); s != nil {
				ref = s.Name + " " + ref
			}
			pSelectObject.Call(hdc, quranRefFont)
			pSetTextColor.Call(hdc, rgb(120, 120, 120))
			t := rect{in.Left, in.Top, in.Right, in.Top + refLine}
			drawText(hdc, ref, &t, dtLeft|dtSingleLine|dtNoPrefix)
			ar := shown(lib.Arabic(p.Surah, p.From))
			pSelectObject.Call(hdc, quranArabicFont)
			pSetTextColor.Call(hdc, rgb(20, 20, 20))
			top := t.Bottom + scale(2)
			if i == quranSel {
				h := capLines(textHeight(quranArabicFont, ar, in.Right-in.Left, dtRTLReading|dtRight), arLine, 3)
				a := rect{in.Left, top, in.Right, top + h}
				drawText(hdc, ar, &a, dtRight|dtRTLReading|dtWordBreak|dtNoPrefix|dtEditControl|dtEndEllipsis)
			} else {
				a := rect{in.Left, top, in.Right, top + arLine}
				drawText(hdc, ar, &a, dtRight|dtRTLReading|dtSingleLine|dtNoPrefix|dtEndEllipsis)
			}
		}
		if quranInfo.Bottom > 0 {
			pSelectObject.Call(hdc, quranRefFont)
			pSetTextColor.Call(hdc, rgb(150, 150, 150))
			t := quranInfo
			drawText(hdc, fmt.Sprintf("%d āyāt found, showing %d. Type more to narrow it down.", quranTotal, len(quranRows)), &t,
				dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
		}
	}

	sep := rect{scale(8), quranSepY, rc.Right - scale(8), quranSepY + 1}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&sep)), sepBrush)
	for i, m := range quran.InsertModes {
		chip(hdc, quranModeBtns[i], modeTitle(m), settings.quranMode() == m, smallFont)
	}
	if quranTransBtn.Right > 0 {
		label := "No translation"
		if t := lib.Translation(settings.quranTranslation()); t != nil {
			label = t.Title()
		}
		chip(hdc, quranTransBtn, label+"   ▾", false, smallFont)
	}
	canInsert := !quranSearch || (len(quranRows) > 0 && quranRows[0].Kind == quran.KindHit)
	chip(hdc, quranInsert, "Insert ⏎", canInsert, quranTitleFont)
	pSelectObject.Call(hdc, quranRefFont)
	pSetTextColor.Call(hdc, rgb(150, 150, 150))
	h := quranHint
	drawText(hdc, quranHintText(), &h, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
}
