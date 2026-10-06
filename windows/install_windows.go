//go:build windows

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Per-user install: no admin rights needed.
//   exe:        %LOCALAPPDATA%\Programs\Qalam\Qalam.exe
//   Start menu: %APPDATA%\Microsoft\Windows\Start Menu\Programs\Qalam.lnk
//   startup:    HKCU\...\Run\Qalam
//   uninstall:  HKCU\...\Uninstall\Qalam  (shows in Settings → Apps)

const (
	runKey       = `Software\Microsoft\Windows\CurrentVersion\Run`
	uninstallKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\Qalam`
)

func installDir() string   { return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Qalam") }
func installedExe() string { return filepath.Join(installDir(), "Qalam.exe") }
func shortcutPath() string {
	return filepath.Join(os.Getenv("APPDATA"), `Microsoft\Windows\Start Menu\Programs`, "Qalam.lnk")
}

func install() error {
	if err := os.MkdirAll(installDir(), 0o755); err != nil {
		return err
	}
	// Stop an older installed copy first, so its file can be replaced.
	closeRunningCopy()
	if err := copyFile(exePath(), installedExe()); err != nil {
		return err
	}
	exe := installedExe()
	setStartup(true)
	if k, err := regCreate(uninstallKey); err == nil {
		regSetString(k, "DisplayName", "Qalam (phonetic Arabic keyboard)")
		regSetString(k, "DisplayVersion", version)
		regSetString(k, "Publisher", "Qalam")
		regSetString(k, "DisplayIcon", exe)
		regSetString(k, "InstallLocation", installDir())
		regSetString(k, "UninstallString", fmt.Sprintf(`"%s" --uninstall`, exe))
		regSetDWORD(k, "NoModify", 1)
		regSetDWORD(k, "NoRepair", 1)
		pRegCloseKey.Call(k)
	}
	makeShortcut(exe)
	cmd := exec.Command(exe)
	return cmd.Start()
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func makeShortcut(target string) {
	script := fmt.Sprintf(`$s=(New-Object -ComObject WScript.Shell).CreateShortcut('%s');$s.TargetPath='%s';$s.Description='Qalam phonetic Arabic keyboard';$s.Save()`,
		shortcutPath(), target)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()
}

// closeRunningCopy asks another running Qalam to quit (via its tray window).
func closeRunningCopy() {
	if h, _, _ := pFindWindowW.Call(uintptr(unsafe.Pointer(utf16Ptr("QalamTray"))), 0); h != 0 {
		pPostMessageW.Call(h, wmClose, 0, 0)
	}
}

func startUninstall() {
	cmd := exec.Command(installedExe(), "--uninstall")
	cmd.Start()
}

func uninstall() {
	if messageBox("Remove Qalam from this computer?", "Qalam", mbYesNo|mbIconQuestion) != idYes {
		return
	}
	closeRunningCopy()
	setStartup(false)
	pRegDeleteKeyW.Call(hkeyCurrentUser, uintptr(unsafe.Pointer(utf16Ptr(uninstallKey))))
	os.Remove(shortcutPath())
	os.RemoveAll(filepath.Join(os.Getenv("APPDATA"), "Qalam"))
	messageBox("Qalam has been removed.", "Qalam", mbIconInfo)
	// The running exe can't delete itself: remove the folder a moment after this process exits.
	cmd := exec.Command("cmd", "/c", "ping 127.0.0.1 -n 3 > nul & rmdir /s /q \""+installDir()+"\"")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

// MARK: start with Windows

func startupEnabled() bool {
	k, err := regCreate(runKey)
	if err != nil {
		return false
	}
	defer pRegCloseKey.Call(k)
	r, _, _ := pRegQueryValueExW.Call(k, uintptr(unsafe.Pointer(utf16Ptr("Qalam"))), 0, 0, 0, 0)
	return r == 0
}

func setStartup(on bool) {
	k, err := regCreate(runKey)
	if err != nil {
		return
	}
	defer pRegCloseKey.Call(k)
	if on {
		exe := exePath()
		if runningFromInstallDir() || fileExists(installedExe()) {
			exe = installedExe()
		}
		regSetString(k, "Qalam", `"`+exe+`"`)
	} else {
		pRegDeleteValueW.Call(k, uintptr(unsafe.Pointer(utf16Ptr("Qalam"))))
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// MARK: registry helpers

func regCreate(path string) (uintptr, error) {
	var k uintptr
	r, _, _ := pRegCreateKeyExW.Call(hkeyCurrentUser, uintptr(unsafe.Pointer(utf16Ptr(path))), 0, 0, 0,
		keyAllAccess, 0, uintptr(unsafe.Pointer(&k)), 0)
	if r != 0 {
		return 0, syscall.Errno(r)
	}
	return k, nil
}

func regSetString(k uintptr, name, value string) {
	u, _ := syscall.UTF16FromString(value)
	pRegSetValueExW.Call(k, uintptr(unsafe.Pointer(utf16Ptr(name))), 0, regSZ,
		uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)*2))
}

func regSetDWORD(k uintptr, name string, value uint32) {
	pRegSetValueExW.Call(k, uintptr(unsafe.Pointer(utf16Ptr(name))), 0, regDWORD,
		uintptr(unsafe.Pointer(&value)), 4)
}
