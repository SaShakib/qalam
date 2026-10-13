package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Chrome-based apps (Chrome, Edge, Brave, Opera, Vivaldi, and Electron apps such as VS Code, Slack,
// Discord, Claude) shape text with HarfBuzz, which reorders our harakat and then draws the font's
// built-in "Allah" ligature on top of them (doubled shadda and alif). For those apps the engine puts
// an invisible ZWJ between the two lāms of لله (Options.BlockAllahLigature), the same fix the macOS
// keyboard uses.

// webViewApps are apps that draw their text in a WebView2 (Edge) window but don't ship Chromium's
// .pak files next to their exe.
var webViewApps = map[string]bool{
	"msedgewebview2.exe": true,
	"ms-teams.exe":       true, // Teams (new)
	"olk.exe":            true, // Outlook (new)
	"whatsapp.exe":       true, // WhatsApp desktop
	"whatsapp.root.exe":  true,
	"clipchamp.exe":      true,
}

var (
	chromiumMu    sync.Mutex
	chromiumCache = map[string]bool{}
)

// isChromiumExe reports whether the program at exe renders text with Chromium.
// Electron and CEF apps keep resources.pak next to the exe; Chrome, Edge and Brave keep it
// in a version folder beside it (Application\130.0.6723.117\resources.pak).
func isChromiumExe(exe string) bool {
	if exe == "" {
		return false
	}
	key := strings.ToLower(filepath.Clean(exe))
	chromiumMu.Lock()
	defer chromiumMu.Unlock()
	if known, ok := chromiumCache[key]; ok {
		return known
	}
	found := webViewApps[strings.ToLower(filepath.Base(exe))] || hasPak(filepath.Dir(exe), 1)
	chromiumCache[key] = found
	return found
}

// hasPak looks for Chromium's resource packs in dir, and depth levels of folders below it.
func hasPak(dir string, depth int) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		name := strings.ToLower(e.Name())
		if !e.IsDir() && (name == "resources.pak" || name == "chrome_100_percent.pak") {
			return true
		}
	}
	if depth == 0 {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && startsWithDigit(e.Name()) && hasPak(filepath.Join(dir, e.Name()), depth-1) {
			return true
		}
	}
	return false
}

func startsWithDigit(s string) bool { return s != "" && s[0] >= '0' && s[0] <= '9' }
