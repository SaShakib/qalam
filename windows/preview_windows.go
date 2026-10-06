//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// The preview: a small always-on-top box near the caret showing the word being typed.

var (
	previewHwnd   uintptr
	previewArabic string
	previewLatin  string
	arabicFont    uintptr
	latinFont     uintptr
	previewBrush  uintptr
	pAddFontMem   = gdi32.NewProc("AddFontMemResourceEx")
	previewFamily = "Segoe UI"
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

	arabicFont, _, _ = pCreateFontW.Call(uintptr(-scale(34)), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(utf16Ptr(previewFamily))))
	latinFont, _, _ = pCreateFontW.Call(uintptr(-scale(14)), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(utf16Ptr("Consolas"))))
	previewBrush, _, _ = pCreateSolidBrush.Call(rgb(255, 253, 240))
}

func previewProc(hwnd, message, wParam, lParam uintptr) uintptr {
	if message == wmPaint {
		var ps paintStruct
		hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		var rc rect
		pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
		pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), previewBrush)
		pSetBkMode.Call(hdc, transparent)

		split := rc.Bottom - scale(22)
		top := rect{rc.Left, rc.Top, rc.Right, split}
		bottom := rect{rc.Left, split, rc.Right, rc.Bottom}

		pSelectObject.Call(hdc, arabicFont)
		pSetTextColor.Call(hdc, rgb(20, 20, 20))
		drawText(hdc, previewArabic, &top, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix|dtRTLReading)

		pSelectObject.Call(hdc, latinFont)
		pSetTextColor.Call(hdc, rgb(120, 120, 120))
		drawText(hdc, previewLatin, &bottom, dtCenter|dtVCenter|dtSingleLine|dtNoPrefix)

		pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return r
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

func showPreview(arabic, latin string) {
	if previewHwnd == 0 {
		return
	}
	previewArabic, previewLatin = arabic, latin
	aw, ah := measure(arabicFont, arabic, dtRTLReading)
	lw, _ := measure(latinFont, latin, 0)
	w := aw
	if lw > w {
		w = lw
	}
	w += scale(28)
	h := ah + scale(22) + scale(10)

	x, y := caretPosition()
	pSetWindowPos.Call(previewHwnd, hwndTopmost, uintptr(x), uintptr(y), uintptr(w), uintptr(h), swpNoActivate)
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
