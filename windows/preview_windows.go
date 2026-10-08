//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"qalam/engine"
)

// The options panel: an always-on-top box near the caret listing the versions of the word being
// typed (highlighted = what Space will insert), with Sukūn / Harakat buttons. It never takes focus.

var (
	previewHwnd    uintptr
	panelCands     []engine.Candidate
	panelSel       int
	panelLatin     string
	panelRows      []rect
	sukunBtn       rect
	harakatBtn     rect
	arabicFont     uintptr
	latinFont      uintptr
	smallFont      uintptr
	previewBrush   uintptr
	highlightBrush uintptr
	borderBrush    uintptr
	pAddFontMem    = gdi32.NewProc("AddFontMemResourceEx")
	pFrameRect     = user32.NewProc("FrameRect")
	previewFamily  = "Segoe UI"
)

const (
	wmLButtonDown   = 0x0201
	wmMouseActivate = 0x0021
	maNoActivate    = 3
	dtLeft          = 0x0000
)

func scale(px int32) int32 {
	dpi, _, _ := pGetDpiForSystem.Call()
	if dpi == 0 {
		dpi = 96
	}
	return px * int32(dpi) / 96
}

// loadPreviewFont makes the bundled Noto Naskh Arabic available to Qalam only.
func loadPreviewFont() {
	if len(previewFont) == 0 {
		return
	}
	var count uint32
	h, _, _ := pAddFontMem.Call(uintptr(unsafe.Pointer(&previewFont[0])), uintptr(len(previewFont)), 0, uintptr(unsafe.Pointer(&count)))
	if h != 0 {
		previewFamily = "Noto Naskh Arabic"
	}
}

func makeFont(px int32, family string, weight uintptr) uintptr {
	f, _, _ := pCreateFontW.Call(uintptr(-px), 0, 0, 0, weight, 0, 0, 0, 1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(utf16Ptr(family))))
	return f
}

func createPreviewWindow() {
	className := utf16Ptr("QalamPreview")
	wc := wndClassEx{
		LpfnWndProc:   syscall.NewCallback(previewProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	previewHwnd, _, _ = pCreateWindowExW.Call(
		wsExTopmost|wsExToolWindow|wsExNoActivate,
		uintptr(unsafe.Pointer(className)), 0,
		wsPopup|wsBorder, 0, 0, 10, 10, 0, 0, hInstance, 0)

	arabicFont = makeFont(scale(28), previewFamily, 400)
	latinFont = makeFont(scale(13), "Segoe UI", 400)
	smallFont = makeFont(scale(11), "Segoe UI", 600)
	previewBrush, _, _ = pCreateSolidBrush.Call(rgb(255, 253, 245))
	highlightBrush, _, _ = pCreateSolidBrush.Call(rgb(214, 232, 250))
	borderBrush, _, _ = pCreateSolidBrush.Call(rgb(190, 190, 190))
}

func previewProc(hwnd, message, wParam, lParam uintptr) uintptr {
	switch message {
	case wmMouseActivate:
		return maNoActivate // clicking the panel must not take focus from the app
	case wmLButtonDown:
		x, y := int32(int16(lParam&0xFFFF)), int32(int16((lParam>>16)&0xFFFF))
		onPanelClick(x, y)
		return 0
	case wmPaint:
		paintPanel(hwnd)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return r
}

func inRect(x, y int32, r rect) bool { return x >= r.Left && x < r.Right && y >= r.Top && y < r.Bottom }

func onPanelClick(x, y int32) {
	for i, r := range panelRows {
		if inRect(x, y, r) {
			pickCandidate(i)
			return
		}
	}
	if inRect(x, y, sukunBtn) {
		toggleSukun()
	} else if inRect(x, y, harakatBtn) {
		toggleHarakat()
	}
}

func paintPanel(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc rect
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), previewBrush)
	pSetBkMode.Call(hdc, transparent)

	for i, c := range panelCands {
		r := panelRows[i]
		if i == panelSel {
			pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), highlightBrush)
		}
		pad := scale(8)
		num := rect{r.Left + pad, r.Top, r.Left + pad + scale(16), r.Bottom}
		pSelectObject.Call(hdc, latinFont)
		pSetTextColor.Call(hdc, rgb(120, 120, 120))
		drawText(hdc, fmt.Sprint(i+1), &num, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)

		aw, _ := measure(arabicFont, c.Text, dtRTLReading)
		ar := rect{num.Right + pad, r.Top, num.Right + pad + aw, r.Bottom}
		pSelectObject.Call(hdc, arabicFont)
		pSetTextColor.Call(hdc, rgb(20, 20, 20))
		drawText(hdc, c.Text, &ar, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtRTLReading)

		lr := rect{ar.Right + scale(12), r.Top, r.Right - pad, r.Bottom}
		pSelectObject.Call(hdc, latinFont)
		pSetTextColor.Call(hdc, rgb(110, 110, 110))
		drawText(hdc, c.Label, &lr, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	}

	// buttons + typed letters
	pSelectObject.Call(hdc, smallFont)
	pSetTextColor.Call(hdc, rgb(40, 40, 40))
	for _, b := range []struct {
		r    rect
		text string
	}{{sukunBtn, sukunLabel()}, {harakatBtn, harakatLabel()}} {
		r := b.r
		pFrameRect.Call(hdc, uintptr(unsafe.Pointer(&r)), borderBrush)
		drawText(hdc, b.text, &r, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)
	}
	pSetTextColor.Call(hdc, rgb(150, 150, 150))
	tr := rect{harakatBtn.Right + scale(10), sukunBtn.Top, rc.Right - scale(8), sukunBtn.Bottom}
	drawText(hdc, panelLatin+"   ↑↓ choose · Space inserts", &tr, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)

	pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
}

func sukunLabel() string {
	if settings.Sukun == engine.SukunOff {
		return "Sukūn: Off"
	}
	return "Sukūn: Smart"
}

func harakatLabel() string {
	if settings.Harakat == engine.NoHarakat {
		return "Harakat: Off"
	}
	return "Harakat: On"
}

func drawText(hdc uintptr, s string, rc *rect, flags uintptr) {
	u, _ := syscall.UTF16FromString(s)
	pDrawTextW.Call(hdc, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(rc)), flags)
}

func measure(font uintptr, s string, flags uintptr) (int32, int32) {
	hdc, _, _ := pGetDC.Call(previewHwnd)
	defer pReleaseDC.Call(previewHwnd, hdc)
	pSelectObject.Call(hdc, font)
	rc := rect{0, 0, 2000, 200}
	drawText(hdc, s, &rc, flags|dtCalcRect|dtSingleLine|dtNoPrefix)
	return rc.Right - rc.Left, rc.Bottom - rc.Top
}

// showPanel lays out and shows the options (or just the first one if options are switched off).
func showPanel(cands []engine.Candidate, sel int, latin string) {
	if previewHwnd == 0 || len(cands) == 0 {
		return
	}
	if settings.NoOptions {
		cands, sel = cands[sel:sel+1], 0
	}
	panelCands, panelSel, panelLatin = cands, sel, latin

	pad := scale(8)
	rowH := scale(44)
	width := scale(260)
	for _, c := range cands {
		aw, _ := measure(arabicFont, c.Text, dtRTLReading)
		lw, _ := measure(latinFont, c.Label, 0)
		if w := pad + scale(16) + pad + aw + scale(12) + lw + pad*2; w > width {
			width = w
		}
	}
	panelRows = panelRows[:0]
	y := pad / 2
	for range cands {
		panelRows = append(panelRows, rect{pad / 2, y, width - pad/2, y + rowH})
		y += rowH
	}
	btnH := scale(24)
	y += pad / 2
	sw, _ := measure(smallFont, "Sukūn: Smart", 0)
	hw, _ := measure(smallFont, "Harakat: Off", 0)
	sukunBtn = rect{pad, y, pad + sw + scale(16), y + btnH}
	harakatBtn = rect{sukunBtn.Right + scale(6), y, sukunBtn.Right + scale(6) + hw + scale(16), y + btnH}
	height := y + btnH + pad

	x, cy := caretPosition()
	pSetWindowPos.Call(previewHwnd, hwndTopmost, uintptr(x), uintptr(cy), uintptr(width), uintptr(height), swpNoActivate)
	pInvalidateRect.Call(previewHwnd, 0, 1)
	pShowWindow.Call(previewHwnd, swShowNA)
}

func hidePreview() {
	if previewHwnd != 0 {
		pShowWindow.Call(previewHwnd, swHide)
	}
}

// caretPosition is just below the text caret, or near the mouse if the app hides its caret.
func caretPosition() (int32, int32) {
	fg, _, _ := pGetForegroundWindow.Call()
	tid, _, _ := pGetWindowThreadProcessId.Call(fg, 0)
	var gti guiThreadInfo
	gti.CbSize = uint32(unsafe.Sizeof(gti))
	if ok, _, _ := pGetGUIThreadInfo.Call(tid, uintptr(unsafe.Pointer(&gti))); ok != 0 && gti.HwndCaret != 0 {
		pt := point{gti.RcCaret.Left, gti.RcCaret.Bottom}
		pClientToScreen.Call(gti.HwndCaret, uintptr(unsafe.Pointer(&pt)))
		return pt.X, pt.Y + scale(6)
	}
	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return pt.X + scale(12), pt.Y + scale(20)
}
