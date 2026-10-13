package main

import (
	"os"
	"path/filepath"
	"testing"
)

func touch(t *testing.T, path string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestChromiumApps(t *testing.T) {
	dir := t.TempDir()
	// Electron (VS Code, Slack, Discord…): resources.pak beside the exe.
	touch(t, filepath.Join(dir, "Code", "Code.exe"))
	touch(t, filepath.Join(dir, "Code", "resources.pak"))
	// Chrome / Edge / Brave: in a version folder beside the exe.
	touch(t, filepath.Join(dir, "Chrome", "Application", "chrome.exe"))
	touch(t, filepath.Join(dir, "Chrome", "Application", "130.0.6723.117", "resources.pak"))
	// A native app.
	touch(t, filepath.Join(dir, "Office", "WINWORD.EXE"))
	touch(t, filepath.Join(dir, "Office", "Data", "resources.pak")) // not a version folder

	cases := map[string]bool{
		filepath.Join(dir, "Code", "Code.exe"):                    true,
		filepath.Join(dir, "Chrome", "Application", "chrome.exe"): true,
		filepath.Join(dir, "Office", "WINWORD.EXE"):               false,
		filepath.Join(dir, "WindowsApps", "WhatsApp.exe"):         true, // WebView2, known by name
		filepath.Join(dir, "WindowsApps", "Notepad.exe"):          false,
		"": false,
	}
	for exe, want := range cases {
		if got := isChromiumExe(exe); got != want {
			t.Errorf("isChromiumExe(%q) = %v, want %v", exe, got, want)
		}
	}
}
