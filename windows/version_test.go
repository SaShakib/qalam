package main

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.2.1", "1.2.0", true},
		{"1.10.0", "1.9.9", true},
		{"1.2", "1.2.0", false},
		{"1.2.0", "1.2.0", false},
		{"1.1.9", "1.2.0", false},
		{"2.0.0", "1.99.99", true},
	}
	for _, c := range cases {
		if got := isNewer(c.a, c.b); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v", c.a, c.b, got)
		}
	}
}

func TestUSKeys(t *testing.T) {
	cases := []struct {
		vk    uint32
		shift bool
		want  rune
	}{
		{'A', false, 'a'}, {'A', true, 'A'}, {'H', true, 'H'}, {'1', true, '!'},
		{0xDE, false, '\''}, {0xDE, true, '"'}, {0xBF, false, '/'}, {0xC0, true, '~'},
	}
	for _, c := range cases {
		if got, ok := usKey(c.vk, c.shift); !ok || got != c.want {
			t.Errorf("usKey(%#x, %v) = %q", c.vk, c.shift, got)
		}
	}
	if _, ok := usKey(0x70, false); ok { // F1
		t.Errorf("F1 should not type a character")
	}
}

func TestSettingsFile(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	s := loadSettings()
	if !s.Enabled || !s.SpellingWords || s.SetupDone {
		t.Fatalf("defaults: %+v", s)
	}
	s.SetupDone = true
	s.NoOptions = true
	s.Learned = map[string]string{"kataba": "كتب"}
	saveSettings(s)
	back := loadSettings()
	if !back.SetupDone || !back.NoOptions || back.Learned["kataba"] != "كتب" {
		t.Errorf("round trip: %+v", back)
	}
}
