//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"qalam/engine"
)

const processQueryLimitedInformation = 0x1000

var (
	lastFgPID      uintptr
	lastFgChromium bool
)

// foregroundIsChromium reports whether the app being typed into renders text with Chromium.
func foregroundIsChromium() bool {
	fg, _, _ := pGetForegroundWindow.Call()
	if fg == 0 {
		return false
	}
	var pid uint32
	pGetWindowThreadProcessId.Call(fg, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return false
	}
	if uintptr(pid) == lastFgPID {
		return lastFgChromium
	}
	lastFgPID, lastFgChromium = uintptr(pid), isChromiumExe(processPath(pid))
	return lastFgChromium
}

// processPath is the full path of a process's exe, or "" if Windows won't say.
func processPath(pid uint32) string {
	h, _, _ := pOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if h == 0 {
		return ""
	}
	defer pCloseHandle.Call(h)
	buf := make([]uint16, 1024)
	size := uint32(len(buf))
	if ok, _, _ := pQueryFullProcessImageNameW.Call(h, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size))); ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:size])
}

// typingOptions are the settings for the next keystroke, with the Allah-ligature guard
// switched on for Chrome-based apps (it is never saved in settings.json).
func typingOptions() engine.Options {
	o := settings.Options
	o.BlockAllahLigature = foregroundIsChromium()
	return o
}
