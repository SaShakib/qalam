package main

import (
	"encoding/json"
	"testing"
)

func TestHotkeys(t *testing.T) {
	var s Settings
	if got := s.hotkeyText("toggle"); got != "Ctrl+Alt+A" {
		t.Errorf("default toggle = %q", got)
	}
	if got := s.hotkeyText("harakat"); got != "none" {
		t.Errorf("default harakat = %q", got)
	}
	// A removed shortcut survives saving as null; a changed one keeps its keys.
	if err := json.Unmarshal([]byte(`{"keys":{"sukun":null,"harakat":{"vk":72,"ctrl":true,"shift":true}}}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.hotkey("sukun") != nil {
		t.Errorf("removed sukun shortcut came back as %v", s.hotkey("sukun"))
	}
	if got := s.hotkeyText("harakat"); got != "Ctrl+Shift+H" {
		t.Errorf("harakat = %q", got)
	}
	data, _ := json.Marshal(s)
	var back Settings
	json.Unmarshal(data, &back)
	if back.hotkey("sukun") != nil || back.hotkeyText("harakat") != "Ctrl+Shift+H" {
		t.Errorf("round trip lost shortcuts: %s", data)
	}
}
