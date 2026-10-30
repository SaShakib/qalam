//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"github.com/jchv/go-webview2/webviewloader"
)

// The app window: the sidebar with Home, the Qur'an, Letters, Practice, Write and Settings
// (assets/app.html, served on 127.0.0.1 by setup_server.go). It runs as a second Qalam.exe,
// `Qalam.exe --window <url>`, in a WebView2. The keyboard hook stays in the tray process, so the
// window can never hold up typing, and closing it frees its memory. Without the WebView2 runtime
// (it comes with Windows 11 and with Edge on Windows 10) the page opens in Edge's app mode, then
// in the default browser.

const (
	wmShowApp   = wmApp + 3
	exitNoView2 = 3
)

var (
	setupSrv *setupServer
	appMu    sync.Mutex
	appCmd   *exec.Cmd
)

func appServer() *setupServer {
	if setupSrv != nil {
		return setupSrv
	}
	srv := newSetupServer(appHost{}, setupHTML, previewFont)
	srv.shell, srv.guide, srv.quran = appHTML, guidePage(guideHTML), quranLibrary
	for _, f := range fonts {
		if isQuranFont(f) {
			srv.quranFont = f.Data
		}
	}
	if err := srv.listen(); err != nil {
		messageBox("Could not open Qalam's window: "+err.Error(), "Qalam", mbIconInfo)
		return nil
	}
	setupSrv = srv
	return srv
}

func appWindowPID() uint32 {
	appMu.Lock()
	defer appMu.Unlock()
	if appCmd != nil && appCmd.Process != nil {
		return uint32(appCmd.Process.Pid)
	}
	return 0
}

// isOwnWindow: the window belongs to Qalam (the app window, or the tray process's own).
func isOwnWindow(hwnd uintptr) bool {
	if hwnd == 0 || hwnd == trayHwnd || hwnd == previewHwnd || hwnd == quranHwnd {
		return true
	}
	var pid uint32
	pGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return pid != 0 && pid == appWindowPID()
}

func appWindowHandle() uintptr {
	h, _, _ := pFindWindowW.Call(uintptr(unsafe.Pointer(utf16Ptr("webview"))), uintptr(unsafe.Pointer(utf16Ptr("Qalam"))))
	return h
}

// openApp shows the window at a page ("" = Home, "setup", "settings", "quran", "practice"…).
func openApp(page string) {
	srv := appServer()
	if srv == nil {
		return
	}
	if appWindowPID() != 0 {
		if page != "" {
			srv.requestGoto(page)
		}
		if h := appWindowHandle(); h != 0 {
			if iconic, _, _ := pIsIconic.Call(h); iconic != 0 {
				pShowWindow.Call(h, swRestore)
			}
			pSetForegroundWindow.Call(h)
		}
		return
	}
	cmd := exec.Command(exePath(), "--window", srv.url(page))
	if err := cmd.Start(); err != nil {
		openAppFallback(srv.url(page))
		return
	}
	appMu.Lock()
	appCmd = cmd
	appMu.Unlock()
	started := time.Now()
	go func() {
		err := cmd.Wait()
		appMu.Lock()
		if appCmd == cmd {
			appCmd = nil
		}
		appMu.Unlock()
		// It stopped at once with an error: no WebView2 on this PC.
		if err != nil && time.Since(started) < 10*time.Second {
			runOnUI(func() { openAppFallback(srv.url(page)) })
		}
	}()
}

// openAppFallback: Edge's app mode has no address bar; failing that, the default browser.
func openAppFallback(url string) {
	args := "--app=" + url + " --window-size=1120,780"
	r, _, _ := pShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16Ptr("open"))), uintptr(unsafe.Pointer(utf16Ptr("msedge.exe"))),
		uintptr(unsafe.Pointer(utf16Ptr(args))), 0, swShowNormal)
	if r <= 32 {
		shellOpen(url)
	}
}

func closeApp() {
	appMu.Lock()
	defer appMu.Unlock()
	if appCmd != nil && appCmd.Process != nil {
		appCmd.Process.Kill()
	}
}

// showRunningCopy asks the Qalam already running to open its window.
func showRunningCopy() bool {
	h, _, _ := pFindWindowW.Call(uintptr(unsafe.Pointer(utf16Ptr("QalamTray"))), 0)
	if h == 0 {
		return false
	}
	pAllowSetForegroundWindow.Call(^uintptr(0)) // ASFW_ANY: its window may come to the front
	pPostMessageW.Call(h, wmShowApp, 0, 0)
	return true
}

// runWindow is the window process (`Qalam.exe --window <url>`).
func runWindow(url string) {
	pSetProcessDpiAwarenessContext.Call(^uintptr(3)) // PER_MONITOR_AWARE_V2 (-4)
	if v, err := webviewloader.GetInstalledVersion(); err != nil || v == "" {
		os.Exit(exitNoView2)
	}
	dpi, _, _ := pGetDpiForSystem.Call()
	if dpi == 0 {
		dpi = 96
	}
	px := func(n uint) uint { return n * uint(dpi) / 96 }
	sw, _, _ := pGetSystemMetrics.Call(smCxScreen)
	sh, _, _ := pGetSystemMetrics.Call(smCyScreen)
	width, height := min(px(1120), uint(sw)*9/10), min(px(780), uint(sh)*85/100) // small screens: it must fit
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		DataPath:  filepath.Join(os.Getenv("LOCALAPPDATA"), "Qalam", "WebView2"),
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title: "Qalam", Width: width, Height: height, Center: true,
		},
	})
	if w == nil {
		os.Exit(exitNoView2)
	}
	defer w.Destroy()
	w.SetSize(int(min(px(760), width)), int(min(px(520), height)), webview2.HintMin)
	setWindowIcon(uintptr(w.Window()))
	w.Navigate(url)
	w.Run()
}

// setWindowIcon gives the window and its taskbar button the ق.
func setWindowIcon(hwnd uintptr) {
	const wmSetIcon = 0x0080
	path := filepath.Join(appDataDir(), "qalam-window.ico")
	if os.WriteFile(path, iconOn, 0o644) != nil {
		return
	}
	for i, size := range []uintptr{16, 32} {
		size = size * uintptr(scale(96)) / 96
		h, _, _ := pLoadImageW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(path))), imageIcon, size, size, lrLoadFromFile)
		if h != 0 {
			pSendMessageW.Call(hwnd, wmSetIcon, uintptr(i), h)
		}
	}
}
