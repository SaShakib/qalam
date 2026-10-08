package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"qalam/engine"
)

// Settings are stored in %APPDATA%\Qalam\settings.json.
type Settings struct {
	engine.Options
	Enabled      bool              `json:"enabled"`
	FirstRunDone bool              `json:"firstRunDone"`
	AskedInstall bool              `json:"askedInstall"`
	NoAutoUpdate bool              `json:"noAutoUpdate"`
	NoOptions    bool              `json:"noOptions"`    // hide the options panel (show only the word)
	PrefersPlain bool              `json:"prefersPlain"` // last pick was a version without harakat
	Learned      map[string]string `json:"learned"`      // typed word → the option picked for it
	// Shortcuts the user changed: "toggle", "sukun", "harakat". A missing name uses the default;
	// null means the user removed it.
	Keys map[string]*Hotkey `json:"keys,omitempty"`
}

// Hotkey is a shortcut: Ctrl and/or (left) Alt, optional Shift, and one key.
type Hotkey struct {
	VK    uint32 `json:"vk"`
	Ctrl  bool   `json:"ctrl,omitempty"`
	Alt   bool   `json:"alt,omitempty"`
	Shift bool   `json:"shift,omitempty"`
}

// hotkeyNames lists the shortcuts in menu order, with their titles.
var hotkeyNames = []struct{ Name, Title string }{
	{"toggle", "Arabic on / off"},
	{"sukun", "Sukūn off / on"},
	{"harakat", "Harakat off / on"},
}

var hotkeyDefaults = map[string]*Hotkey{
	"toggle":  {VK: 'A', Ctrl: true, Alt: true},
	"sukun":   {VK: 'O', Ctrl: true, Alt: true},
	"harakat": nil,
}

// hotkey is the shortcut in use for name, or nil if there is none.
func (s Settings) hotkey(name string) *Hotkey {
	if h, ok := s.Keys[name]; ok {
		return h
	}
	return hotkeyDefaults[name]
}

// hotkeyText is "Ctrl+Alt+O", or "none".
func (s Settings) hotkeyText(name string) string {
	if h := s.hotkey(name); h != nil {
		return h.String()
	}
	return "none"
}

func (h Hotkey) String() string {
	out := ""
	if h.Ctrl {
		out += "Ctrl+"
	}
	if h.Alt {
		out += "Alt+"
	}
	if h.Shift {
		out += "Shift+"
	}
	return out + keyName(h.VK)
}

func keyName(vk uint32) string {
	switch {
	case vk >= 'A' && vk <= 'Z', vk >= '0' && vk <= '9':
		return string(rune(vk))
	case vk >= 0x70 && vk <= 0x7B:
		return fmt.Sprintf("F%d", vk-0x6F)
	}
	names := map[uint32]string{
		0x20: "Space", 0x08: "Backspace", 0x09: "Tab", 0x0D: "Enter", 0x2E: "Delete",
		0x25: "Left", 0x26: "Up", 0x27: "Right", 0x28: "Down",
		0xBA: ";", 0xBB: "=", 0xBC: ",", 0xBD: "-", 0xBE: ".", 0xBF: "/", 0xC0: "`",
		0xDB: "[", 0xDC: "\\", 0xDD: "]", 0xDE: "'",
	}
	if n, ok := names[vk]; ok {
		return n
	}
	return fmt.Sprintf("Key%02X", vk)
}

func settingsPath() string {
	return filepath.Join(os.Getenv("APPDATA"), "Qalam", "settings.json")
}

func loadSettings() Settings {
	s := Settings{Options: engine.DefaultOptions(), Enabled: true}
	if data, err := os.ReadFile(settingsPath()); err == nil {
		json.Unmarshal(data, &s)
	}
	return s
}

func saveSettings(s Settings) {
	os.MkdirAll(filepath.Dir(settingsPath()), 0o755)
	if data, err := json.MarshalIndent(s, "", "  "); err == nil {
		os.WriteFile(settingsPath(), data, 0o644)
	}
}
