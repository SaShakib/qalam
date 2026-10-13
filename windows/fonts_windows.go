//go:build windows

package main

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

//go:embed assets/fonts
var fontFiles embed.FS

var (
	fonts       = loadBundledFonts(fontFiles, "assets/fonts")
	previewFont = func() []byte {
		if f := typingFont(fonts); f != nil {
			return f.Data
		}
		return nil
	}()
	pAddFontResourceW = gdi32.NewProc("AddFontResourceW")
)

const (
	fontsKey      = `Software\Microsoft\Windows NT\CurrentVersion\Fonts`
	hwndBroadcast = 0xFFFF
	wmFontChange  = 0x001D
)

// userFontsDir is where Windows 10 (1809+) and 11 keep fonts installed for one user, without admin rights.
func userFontsDir() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "Fonts")
}

// fontInstalled: the font is installed for this user or for everyone.
func fontInstalled(f bundledFont) bool {
	if fileExists(filepath.Join(userFontsDir(), f.File)) || fileExists(filepath.Join(os.Getenv("WINDIR"), "Fonts", f.File)) {
		return true
	}
	// Installed under another file name (for example from the font's own download): look for its name.
	k, err := regCreate(fontsKey)
	if err != nil {
		return false
	}
	defer pRegCloseKey.Call(k)
	r, _, _ := pRegQueryValueExW.Call(k, uintptr(unsafe.Pointer(utf16Ptr(f.Name+" (TrueType)"))), 0, 0, 0, 0)
	return r == 0
}

func allFontsInstalled() bool {
	for _, f := range fonts {
		if !fontInstalled(f) {
			return false
		}
	}
	return true
}

// installFonts installs the bundled fonts for this user, so every app can use them. No admin rights needed.
func installFonts() error {
	dir := userFontsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	k, err := regCreate(fontsKey)
	if err != nil {
		return err
	}
	defer pRegCloseKey.Call(k)
	changed := false
	for _, f := range fonts {
		if fontInstalled(f) {
			continue
		}
		dest := filepath.Join(dir, f.File)
		if err := os.WriteFile(dest, f.Data, 0o644); err != nil {
			return err
		}
		kind := " (TrueType)"
		if strings.EqualFold(filepath.Ext(f.File), ".otf") {
			kind = " (OpenType)"
		}
		regSetString(k, f.Name+kind, dest)
		pAddFontResourceW.Call(uintptr(unsafe.Pointer(utf16Ptr(dest))))
		changed = true
	}
	if changed {
		pPostMessageW.Call(hwndBroadcast, wmFontChange, 0, 0)
	}
	return nil
}
