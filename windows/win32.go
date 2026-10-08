//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// Thin bindings to the Win32 calls Qalam needs (no third-party packages).

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	pSetWindowsHookExW             = user32.NewProc("SetWindowsHookExW")
	pCallNextHookEx                = user32.NewProc("CallNextHookEx")
	pGetMessageW                   = user32.NewProc("GetMessageW")
	pTranslateMessage              = user32.NewProc("TranslateMessage")
	pDispatchMessageW              = user32.NewProc("DispatchMessageW")
	pPostQuitMessage               = user32.NewProc("PostQuitMessage")
	pDefWindowProcW                = user32.NewProc("DefWindowProcW")
	pRegisterClassExW              = user32.NewProc("RegisterClassExW")
	pCreateWindowExW               = user32.NewProc("CreateWindowExW")
	pDestroyWindow                 = user32.NewProc("DestroyWindow")
	pShowWindow                    = user32.NewProc("ShowWindow")
	pSetWindowPos                  = user32.NewProc("SetWindowPos")
	pGetKeyState                   = user32.NewProc("GetKeyState")
	pSendInput                     = user32.NewProc("SendInput")
	pGetForegroundWindow           = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcessId      = user32.NewProc("GetWindowThreadProcessId")
	pGetGUIThreadInfo              = user32.NewProc("GetGUIThreadInfo")
	pClientToScreen                = user32.NewProc("ClientToScreen")
	pGetCursorPos                  = user32.NewProc("GetCursorPos")
	pBeginPaint                    = user32.NewProc("BeginPaint")
	pEndPaint                      = user32.NewProc("EndPaint")
	pGetClientRect                 = user32.NewProc("GetClientRect")
	pInvalidateRect                = user32.NewProc("InvalidateRect")
	pFillRect                      = user32.NewProc("FillRect")
	pDrawTextW                     = user32.NewProc("DrawTextW")
	pGetDC                         = user32.NewProc("GetDC")
	pReleaseDC                     = user32.NewProc("ReleaseDC")
	pCreatePopupMenu               = user32.NewProc("CreatePopupMenu")
	pAppendMenuW                   = user32.NewProc("AppendMenuW")
	pTrackPopupMenu                = user32.NewProc("TrackPopupMenu")
	pDestroyMenu                   = user32.NewProc("DestroyMenu")
	pSetForegroundWindow           = user32.NewProc("SetForegroundWindow")
	pPostMessageW                  = user32.NewProc("PostMessageW")
	pLoadImageW                    = user32.NewProc("LoadImageW")
	pRegisterWindowMessageW        = user32.NewProc("RegisterWindowMessageW")
	pMessageBoxW                   = user32.NewProc("MessageBoxW")
	pGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
	pSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	pGetDpiForSystem               = user32.NewProc("GetDpiForSystem")
	pFindWindowW                   = user32.NewProc("FindWindowW")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	pCreateMutexW     = kernel32.NewProc("CreateMutexW")
	pCloseHandle      = kernel32.NewProc("CloseHandle")

	pShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	pShellExecuteW    = shell32.NewProc("ShellExecuteW")

	pCreateFontW      = gdi32.NewProc("CreateFontW")
	pSelectObject     = gdi32.NewProc("SelectObject")
	pDeleteObject     = gdi32.NewProc("DeleteObject")
	pSetBkMode        = gdi32.NewProc("SetBkMode")
	pSetTextColor     = gdi32.NewProc("SetTextColor")
	pCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")

	pRegCreateKeyExW  = advapi32.NewProc("RegCreateKeyExW")
	pRegSetValueExW   = advapi32.NewProc("RegSetValueExW")
	pRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	pRegDeleteValueW  = advapi32.NewProc("RegDeleteValueW")
	pRegDeleteKeyW    = advapi32.NewProc("RegDeleteKeyW")
	pRegCloseKey      = advapi32.NewProc("RegCloseKey")
)

const (
	whKeyboardLL     = 13
	wmKeyDown        = 0x0100
	wmSysKeyDown     = 0x0104
	wmDestroy        = 0x0002
	wmClose          = 0x0010
	wmPaint          = 0x000F
	wmCommand        = 0x0111
	wmNull           = 0x0000
	wmApp            = 0x8000
	wmTray           = wmApp + 1
	wmLButtonUp      = 0x0202
	wmRButtonUp      = 0x0205
	llkhfExtended    = 0x01
	llkhfInjected    = 0x10
	inputKeyboard    = 1
	keyExtended      = 0x0001
	keyUp            = 0x0002
	keyUnicode       = 0x0004
	swHide           = 0
	swShowNA         = 8
	swShowNormal     = 1
	wsPopup          = 0x80000000
	wsBorder         = 0x00800000
	wsExTopmost      = 0x00000008
	wsExToolWindow   = 0x00000080
	wsExNoActivate   = 0x08000000
	swpNoActivate    = 0x0010
	hwndTopmost      = ^uintptr(0) // (HWND)-1
	nimAdd           = 0
	nimModify        = 1
	nimDelete        = 2
	nifMessage       = 0x01
	nifIcon          = 0x02
	nifTip           = 0x04
	nifInfo          = 0x10
	niifInfo         = 0x01
	mfString         = 0x0000
	mfChecked        = 0x0008
	mfSeparator      = 0x0800
	mfPopup          = 0x0010
	smCxScreen       = 0
	smCyScreen       = 1
	tpmReturnCmd     = 0x0100
	tpmRightButton   = 0x0002
	imageIcon        = 1
	lrLoadFromFile   = 0x0010
	smCxSmIcon       = 49
	smCySmIcon       = 50
	dtCenter         = 0x0001
	dtVCenter        = 0x0004
	dtSingleLine     = 0x0020
	dtNoPrefix       = 0x0800
	dtCalcRect       = 0x0400
	dtRTLReading     = 0x20000
	transparent      = 1
	mbYesNo          = 0x04
	mbIconQuestion   = 0x20
	mbIconInfo       = 0x40
	idYes            = 6
	hkeyCurrentUser  = 0x80000001
	keyAllAccess     = 0xF003F
	regSZ            = 1
	regDWORD         = 4
	errAlreadyExists = 183
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }

type msg struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type kbdllHook struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// INPUT with the KEYBDINPUT member (64-bit layout, 40 bytes).
type input struct {
	Type        uint32
	_           uint32
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	_           uint32
	DwExtraInfo uintptr
	_           [8]byte
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type guiThreadInfo struct {
	CbSize        uint32
	Flags         uint32
	HwndActive    uintptr
	HwndFocus     uintptr
	HwndCapture   uintptr
	HwndMenuOwner uintptr
	HwndMoveSize  uintptr
	HwndCaret     uintptr
	RcCaret       rect
}

type paintStruct struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     rect
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func copyUTF16(dst []uint16, s string) {
	u, _ := syscall.UTF16FromString(s)
	if len(u) > len(dst) {
		u = u[:len(dst)]
		u[len(u)-1] = 0
	}
	copy(dst, u)
}

func keyDown(vk uintptr) bool {
	r, _, _ := pGetKeyState.Call(vk)
	return int16(r) < 0
}

func messageBox(text, title string, flags uintptr) int {
	r, _, _ := pMessageBoxW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(unsafe.Pointer(utf16Ptr(title))), flags)
	return int(r)
}

func shellOpen(path string) {
	pShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16Ptr("open"))), uintptr(unsafe.Pointer(utf16Ptr(path))), 0, 0, swShowNormal)
}

func rgb(r, g, b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }
