package main

import "testing"

func TestInlineEdit(t *testing.T) {
	cases := []struct {
		old, new string
		back     int
		add      string
	}{
		{"", "كَ", 0, "كَ"},
		{"كَ", "كَتَ", 0, "تَ"},           // typing on adds to the word
		{"كَتَ", "كَتْ", 1, "ْ"},          // only the last mark changes
		{"كَتَبَ", "", 4 + 2, ""},         // the word is gone (Backspace to nothing, Esc)
		{"كَتَبَ", "kataba", 6, "kataba"}, // Esc gives the Latin back
		{"same", "same", 0, ""},
	}
	for _, c := range cases {
		back, add := inlineEdit(c.old, c.new)
		if back != c.back || add != c.add {
			t.Errorf("inlineEdit(%q, %q) = %d %q, want %d %q", c.old, c.new, back, add, c.back, c.add)
		}
	}
}
