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

// The "insert from the Qur'an" box (Ctrl+Alt+Q), like ⌃⇧Q on the Mac: what you type shows in its
// field, the list under it has sūrahs, the passage, or search results in the Qur'an font.
// Like the options panel it never takes focus; the keyboard hook sends it the keys.
//   ↑↓ choose · Enter inserts · Esc closes · Tab: Arabic / Arabic + translation / translation ·
//   Shift+Tab: next translation · Ctrl+V pastes (for Arabic search).

var (
	quranLib     *quran.Library
	quranLibOnce sync.Once

	quranHwnd    uintptr
	quranOpen    bool
	quranQuery   string
	quranRows    []quran.Suggestion
	quranSel     int
	quranTotal   int
	quranGen     int  // bumped for every new query
	quranShown   int  // the query whose rows are in quranRows
	quranPickNow bool // Enter was pressed before the rows for the query came back
	quranFg      uintptr

	quranRowRects []rect
	quranModeBtn  rect
	quranTransBtn rect
	quranField    rect
	quranInfo     rect
	quranHint     rect
	quranSepY     int32

	quranArabicFont uintptr
	quranNameFont   uintptr
	quranTextFont   uintptr
	quranRefFont    uintptr
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
	quranOpen = true
	quranQuery, quranRows, quranSel, quranTotal, quranPickNow = "", nil, 0, 0, false
	quranGen++
	quranShown = quranGen
	quranFg, _, _ = pGetForegroundWindow.Call()
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
	quranQuery, quranRows, quranPickNow = "", nil, false
	quranGen++
	pKillTimer.Call(quranHwnd, 1)
	pShowWindow.Call(quranHwnd, swHide)
}

// quranKey handles a key while the box is open; true = swallow it.
func quranKey(vk uint32, ctrl, leftAlt, alt, shift, win bool) bool {
	if isModifier(vk) {
		return false
	}
	if h := settings.hotkey("quran"); h != nil && h.VK == vk && h.Ctrl == ctrl && h.Alt == leftAlt && h.Alt == alt && h.Shift == shift {
		closeQuran() // the shortcut again closes it
		return true
	}
	if ctrl && !alt && !win && !shift && vk == 'V' {
		if t := clipboardText(); t != "" {
			quranQuery += strings.Join(strings.Fields(t), " ")
			refreshQuran()
		}
		return true
	}
	if ctrl || alt || win { // other shortcuts go to the app
		closeQuran()
		return false
	}
	switch vk {
	case vkEscape:
		closeQuran()
	case vkReturn:
		if quranShown == quranGen {
			runOnUI(func() { pickQuran(quranSel) })
		} else {
			quranPickNow = true
		}
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
		if shift {
			cycleQuranTranslation()
		} else {
			cycleQuranMode()
		}
		showQuran()
	case vkBack:
		if quranQuery == "" {
			closeQuran()
			return true
		}
		_, size := utf8.DecodeLastRuneInString(quranQuery)
		quranQuery = quranQuery[:len(quranQuery)-size]
		refreshQuran()
	case vkSpace:
		quranQuery += " "
		refreshQuran()
	default:
		if r, ok := usKey(vk, shift); ok {
			quranQuery += string(r)
			refreshQuran()
		}
	}
	return true
}

// refreshQuran looks the query up off the keyboard thread (the first search builds the index)
// and shows the new rows when they come back, unless something newer was typed by then.
func refreshQuran() {
	lib := quranLibrary()
	quranGen++
	gen, q := quranGen, quranQuery
	showQuran() // the typed text at once
	go func() {
		rows, total := lib.Lookup(q, quranRowLimit)
		runOnUI(func() {
			if !quranOpen || gen != quranGen {
				return
			}
			quranRows, quranTotal, quranSel, quranShown = rows, total, 0, gen
			if quranPickNow {
				quranPickNow = false
				pickQuran(0)
				return
			}
			showQuran()
		})
	}()
}

// pickQuran: a sūrah row fills in "2:" for the āyah number; a passage or search result is inserted.
func pickQuran(i int) {
	lib := quranLibrary()
	if !quranOpen || lib == nil {
		return
	}
	if i >= len(quranRows) {
		if p, ok := lib.Resolve(quranQuery); ok {
			insertQuran(p)
		}
		return
	}
	switch r := quranRows[i]; r.Kind {
	case quran.KindSurah:
		quranQuery = fmt.Sprintf("%d:", r.Surah.Number)
		refreshQuran()
	case quran.KindPassage, quran.KindHit:
		insertQuran(r.Passage)
	default:
		pMessageBeep.Call(0xFFFFFFFF)
	}
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

func cycleQuranMode() {
	m := settings.quranMode()
	for i, x := range quran.InsertModes {
		if x == m {
			settings.QuranInsert = string(quran.InsertModes[(i+1)%len(quran.InsertModes)])
		}
	}
	saveSettings(settings)
}

// cycleQuranTranslation: the next bundled translation.
func cycleQuranTranslation() {
	have := quranLibrary().Available()
	if len(have) == 0 {
		return
	}
	next := 0
	for i, t := range have {
		if t.ID == settings.quranTranslation() {
			next = (i + 1) % len(have)
		}
	}
	settings.QuranTranslation = have[next].ID
	if settings.quranMode() == quran.Arabic {
		settings.QuranInsert = string(quran.Both)
	}
	saveSettings(settings)
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
		// Another window came to the front (a click elsewhere, Alt+Tab): close.
		if fg, _, _ := pGetForegroundWindow.Call(); quranOpen && fg != quranFg {
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
	for i, r := range quranRowRects {
		if inRect(x, y, r) {
			pickQuran(i)
			return
		}
	}
	if inRect(x, y, quranModeBtn) {
		cycleQuranMode()
		showQuran()
	} else if inRect(x, y, quranTransBtn) {
		cycleQuranTranslation()
		showQuran()
	}
}

func quranModeLabel() string {
	switch settings.quranMode() {
	case quran.Arabic:
		return "Insert: Arabic only   Tab"
	case quran.TranslationOnly:
		return "Insert: translation only   Tab"
	}
	return "Insert: Arabic + translation   Tab"
}

func quranTransLabel() string {
	if t := quranLibrary().Translation(settings.quranTranslation()); t != nil {
		return t.Title() + "   Shift+Tab"
	}
	return "No translation   Shift+Tab"
}

func quranHintText() string {
	s := "↑↓ choose · Enter insert · Esc close · Ctrl+V paste"
	if h := settings.hotkey("quran"); h != nil {
		s += " · " + h.String() + " opens this box"
	}
	return s
}

const quranPlaceholder = "2:255 · 2:255-257 · baqarah 255 · or search: rahmaan"

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

// rowParts gives a passage row's āyah and (when shown) translation.
func rowParts(r quran.Suggestion, selected bool) (string, string) {
	lib := quranLibrary()
	ar := shown(lib.Arabic(r.Passage.Surah, r.Passage.From))
	tr := ""
	if selected && settings.quranMode() != quran.Arabic {
		tr = lib.TranslationText(settings.quranTranslation(), r.Passage.Surah, r.Passage.From)
	}
	return ar, tr
}

// showQuran lays out the box and shows it near the caret.
func showQuran() {
	if !quranOpen || quranHwnd == 0 {
		return
	}
	pad := scale(8)
	width := scale(580)
	inner := width - 2*pad - 2*scale(10)
	_, arLine := measure(quranArabicFont, "بِسْمِ", dtRTLReading)
	_, trLine := measure(quranTextFont, "Ag", 0)
	_, refLine := measure(quranRefFont, "Ag", 0)

	y := pad
	quranField = rect{pad, y, width - pad, y + scale(32)}
	y = quranField.Bottom + scale(6)

	quranRowRects = quranRowRects[:0]
	for i, r := range quranRows {
		var h int32
		switch r.Kind {
		case quran.KindSurah, quran.KindProblem:
			h = scale(32)
		default:
			ar, tr := rowParts(r, i == quranSel)
			h = refLine + scale(2) + arLine
			if i == quranSel {
				h = refLine + scale(2) + capLines(textHeight(quranArabicFont, ar, inner, dtRTLReading|dtRight), arLine, 4)
				if tr != "" {
					h += scale(4) + capLines(textHeight(quranTextFont, tr, inner, 0), trLine, 3)
				}
			}
			h += scale(10)
		}
		quranRowRects = append(quranRowRects, rect{pad, y, width - pad, y + h})
		y += h + scale(2)
	}
	quranInfo = rect{}
	if quranTotal > len(quranRows) {
		quranInfo = rect{pad + scale(4), y, width - pad, y + scale(18)}
		y = quranInfo.Bottom
	}
	y += scale(4)
	quranSepY = y
	y += scale(8)
	btnH := scale(24)
	mw, _ := measure(smallFont, quranModeLabel(), 0)
	quranModeBtn = rect{pad, y, pad + mw + scale(16), y + btnH}
	quranTransBtn = rect{}
	if settings.quranMode() != quran.Arabic {
		tw, _ := measure(smallFont, quranTransLabel(), 0)
		quranTransBtn = rect{quranModeBtn.Right + scale(6), y, quranModeBtn.Right + scale(6) + tw + scale(16), y + btnH}
	}
	y += btnH + scale(4)
	quranHint = rect{pad, y, width - pad, y + scale(18)}
	height := quranHint.Bottom + pad

	// Below the caret; above it, or moved in, when the screen ends.
	x, cy := caretPosition()
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
	pad := scale(10)

	// The field: "Qur'an", then what's typed (or the placeholder).
	f := quranField
	pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&f)), accentBrush)
	pSelectObject.Call(hdc, quranRefFont)
	pSetTextColor.Call(hdc, rgb(110, 110, 110))
	lw, _ := measure(quranRefFont, "Qur'an", 0)
	label := rect{f.Left + scale(8), f.Top, f.Left + scale(8) + lw, f.Bottom}
	drawText(hdc, "Qur'an", &label, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	field := rect{label.Right + scale(10), f.Top, f.Right - scale(8), f.Bottom}
	pSelectObject.Call(hdc, quranTextFont)
	if quranQuery == "" {
		pSetTextColor.Call(hdc, rgb(160, 160, 160))
		drawText(hdc, quranPlaceholder, &field, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
	} else {
		pSetTextColor.Call(hdc, rgb(20, 20, 20))
		drawText(hdc, quranQuery+"▏", &field, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	}

	_, arLine := measure(quranArabicFont, "بِسْمِ", dtRTLReading)
	_, refLine := measure(quranRefFont, "Ag", 0)
	_, trLine := measure(quranTextFont, "Ag", 0)
	for i, row := range quranRows {
		if i >= len(quranRowRects) {
			break
		}
		r := quranRowRects[i]
		selected := i == quranSel
		if selected {
			pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), highlightBrush)
		}
		in := rect{r.Left + pad, r.Top + scale(5), r.Right - pad, r.Bottom - scale(5)}
		switch row.Kind {
		case quran.KindSurah:
			s := row.Surah
			pSelectObject.Call(hdc, quranTextFont)
			pSetTextColor.Call(hdc, rgb(30, 30, 30))
			t := rect{in.Left, r.Top, in.Right - scale(120), r.Bottom}
			drawText(hdc, fmt.Sprintf("%d  %s  ·  %s  ·  %d āyāt  ·  %s", s.Number, s.Name, s.English, s.Ayahs, s.Type), &t,
				dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
			pSelectObject.Call(hdc, quranNameFont)
			name := rect{in.Right - scale(120), r.Top, in.Right, r.Bottom}
			drawText(hdc, s.Arabic, &name, dtRight|dtVCenter|dtSingleLine|dtNoPrefix|dtRTLReading)
		case quran.KindProblem:
			pSelectObject.Call(hdc, quranTextFont)
			pSetTextColor.Call(hdc, rgb(120, 120, 120))
			t := rect{in.Left, r.Top, in.Right, r.Bottom}
			drawText(hdc, row.Text, &t, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
		default:
			p := row.Passage
			ref := p.Reference()
			if s := lib.Surah(p.Surah); s != nil {
				ref += "  " + s.Name
			}
			if n := p.To - p.From + 1; n > 1 {
				ref += fmt.Sprintf("  ·  %d āyāt", n)
			}
			pSelectObject.Call(hdc, quranRefFont)
			pSetTextColor.Call(hdc, rgb(120, 120, 120))
			t := rect{in.Left, in.Top, in.Right, in.Top + refLine}
			drawText(hdc, ref, &t, dtLeft|dtSingleLine|dtNoPrefix)

			ar, tr := rowParts(row, selected)
			pSelectObject.Call(hdc, quranArabicFont)
			pSetTextColor.Call(hdc, rgb(20, 20, 20))
			top := t.Bottom + scale(2)
			if selected {
				h := capLines(textHeight(quranArabicFont, ar, in.Right-in.Left, dtRTLReading|dtRight), arLine, 4)
				a := rect{in.Left, top, in.Right, top + h}
				drawText(hdc, ar, &a, dtRight|dtRTLReading|dtWordBreak|dtNoPrefix|dtEditControl|dtEndEllipsis)
				if tr != "" {
					pSelectObject.Call(hdc, quranTextFont)
					pSetTextColor.Call(hdc, rgb(90, 90, 90))
					th := capLines(textHeight(quranTextFont, tr, in.Right-in.Left, 0), trLine, 3)
					b := rect{in.Left, a.Bottom + scale(4), in.Right, a.Bottom + scale(4) + th}
					drawText(hdc, tr, &b, dtLeft|dtWordBreak|dtNoPrefix|dtEditControl|dtEndEllipsis)
				}
			} else {
				a := rect{in.Left, top, in.Right, top + arLine}
				drawText(hdc, ar, &a, dtRight|dtRTLReading|dtSingleLine|dtNoPrefix|dtEndEllipsis)
			}
		}
	}

	if quranInfo.Bottom > 0 {
		pSelectObject.Call(hdc, quranRefFont)
		pSetTextColor.Call(hdc, rgb(150, 150, 150))
		t := quranInfo
		drawText(hdc, fmt.Sprintf("%d āyāt found, showing %d. Type more to narrow it down.", quranTotal, len(quranRows)), &t,
			dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	}
	sep := rect{scale(8), quranSepY, rc.Right - scale(8), quranSepY + 1}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&sep)), sepBrush)

	pSelectObject.Call(hdc, smallFont)
	pSetTextColor.Call(hdc, rgb(40, 40, 40))
	for _, b := range []struct {
		r    rect
		text string
	}{{quranModeBtn, quranModeLabel()}, {quranTransBtn, quranTransLabel()}} {
		if b.r.Right == 0 {
			continue
		}
		r := b.r
		pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&r)), borderBrush)
		drawText(hdc, b.text, &r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
	}
	pSelectObject.Call(hdc, quranRefFont)
	pSetTextColor.Call(hdc, rgb(150, 150, 150))
	h := quranHint
	drawText(hdc, quranHintText(), &h, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
}
