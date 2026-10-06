package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"qalam/engine"
)

// Settings are stored in %APPDATA%\Qalam\settings.json.
type Settings struct {
	engine.Options
	Enabled      bool `json:"enabled"`
	FirstRunDone bool `json:"firstRunDone"`
	AskedInstall bool `json:"askedInstall"`
	NoAutoUpdate bool `json:"noAutoUpdate"`
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
