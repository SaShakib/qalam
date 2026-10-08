//go:build windows

// Qalam for Windows: a phonetic Arabic keyboard that lives in the system tray.
// Ctrl + Left Alt + A (or clicking the tray icon) switches Arabic typing on and off; the shortcuts can be changed from the tray menu.
package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"qalam/engine"
)

//go:embed assets/qalam-on.ico
var iconOn []byte

//go:embed assets/qalam-off.ico
var iconOff []byte

//go:embed assets/guide.html
var guideHTML []byte

//go:embed assets/NotoNaskhArabic.ttf
var previewFont []byte

var version = "dev" // set at build time: -X main.version=…

var (
	settings  Settings
	composer  engine.Composer
	hInstance uintptr
	trayHwnd  uintptr
	hookProc  = syscall.NewCallback(keyboardHook)
)

func main() {
	runtime.LockOSThread()

	updated := false
	for _, a := range os.Args[1:] {
		switch a {
		case "--uninstall":
			uninstall()
			return
		case "--updated":
			updated = true
		}
	}

	pSetProcessDpiAwarenessContext.Call(^uintptr(3)) // PER_MONITOR_AWARE_V2 (-4)
	settings = loadSettings()

	// First run from a download: offer to install for this user.
	if !updated && !runningFromInstallDir() && !settings.AskedInstall {
		settings.AskedInstall = true
		saveSettings(settings)
		if messageBox("Install Qalam for this user?\n\n"+
			"Yes: copy it to your Programs folder, add it to the Start menu, and start it with Windows.\n"+
			"No: just run it from here this time.", "Qalam", mbYesNo|mbIconQuestion) == idYes {
			if err := install(); err == nil {
				return // the installed copy is now running
			} else {
				messageBox("Could not install: "+err.Error()+"\nQalam will run from here instead.", "Qalam", mbIconInfo)
			}
		}
	}

	// One copy at a time. After an update, wait for the old copy to finish quitting.
	var mutex uintptr
	for attempt := 0; ; attempt++ {
		h, _, err := pCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(utf16Ptr("Qalam.SingleInstance"))))
		if errno, ok := err.(syscall.Errno); !ok || errno != errAlreadyExists {
			mutex = h
			break
		}
		pCloseHandle.Call(h)
		if !updated || attempt >= 20 {
			messageBox("Qalam is already running. Look for the ق icon in the system tray (bottom right).", "Qalam", mbIconInfo)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	defer pCloseHandle.Call(mutex)

	hInstance, _, _ = pGetModuleHandleW.Call(0)
	loadPreviewFont()
	createPreviewWindow()
	createTray()

	hook, _, _ := pSetWindowsHookExW.Call(whKeyboardLL, hookProc, hInstance, 0)
	if hook == 0 {
		messageBox("Qalam could not start its keyboard hook.", "Qalam", mbIconInfo)
		return
	}

	if updated {
		go cleanupAfterUpdate()
		balloon("Qalam updated", "You now have version "+version+".")
	}
	startUpdateLoop()

	if !settings.FirstRunDone {
		settings.FirstRunDone = true
		settings.Enabled = true
		saveSettings(settings)
		updateTray()
		openGuide()
		balloon("Qalam is on", "Type Arabic anywhere. "+settings.hotkeyText("toggle")+" switches between Arabic and English (change it: right-click the tray icon → Shortcuts).")
	}

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
}

func appDataDir() string {
	dir := filepath.Join(os.Getenv("APPDATA"), "Qalam")
	os.MkdirAll(dir, 0o755)
	return dir
}

func openGuide() {
	path := filepath.Join(appDataDir(), "guide.html")
	if err := os.WriteFile(path, guideHTML, 0o644); err == nil {
		shellOpen(path)
	}
}

func toggle() {
	settings.Enabled = !settings.Enabled
	composer = engine.Composer{}
	hidePreview()
	saveSettings(settings)
	updateTray()
}

func exePath() string {
	p, _ := os.Executable()
	return p
}

func runningFromInstallDir() bool {
	return strings.EqualFold(filepath.Clean(exePath()), filepath.Clean(installedExe()))
}
