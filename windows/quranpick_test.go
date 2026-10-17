package main

import (
	"testing"

	"qalam/quran"
)

func typeAll(p *quranPick, lib *quran.Library, s string) {
	for _, r := range s {
		p.typeRune(lib, r)
	}
}

func TestQuranPickTyping(t *testing.T) {
	lib, err := quran.Load(quranFS())
	if err != nil {
		t.Fatal(err)
	}
	p := newQuranPick(lib, "")
	if p.passage() != (quran.Passage{Surah: 1, From: 1, To: 1}) {
		t.Fatalf("default: %+v", p)
	}
	typeAll(&p, lib, "2:255-257")
	if p.passage() != (quran.Passage{Surah: 2, From: 255, To: 257}) {
		t.Fatalf("2:255-257 → %+v", p.passage())
	}

	p = newQuranPick(lib, "2:255-257")
	if p.passage().Reference() != "2:255-257" {
		t.Fatalf("last: %+v", p.passage())
	}
	p = newQuranPick(lib, "")
	typeAll(&p, lib, "yasin")
	if p.Surah != 36 {
		t.Fatalf("yasin → %d", p.Surah)
	}
	p = newQuranPick(lib, "")
	typeAll(&p, lib, "125") // too big → starts again from 5
	if p.Surah != 5 {
		t.Fatalf("125 → %d", p.Surah)
	}

	// ↑↓ and moving between the dropdowns
	p = newQuranPick(lib, "")
	p.step(lib, 1)
	p.move(1)
	p.step(lib, 1)
	p.step(lib, 1)
	if p.Surah != 2 || p.From != 3 || p.To != 3 {
		t.Fatalf("steps: %+v", p)
	}
	p.move(1)
	p.step(lib, -1) // To never goes below From
	if p.To != 3 {
		t.Fatalf("to below from: %+v", p)
	}
	p.move(1) // wraps to Sūrah
	if p.Field != fieldSurah {
		t.Fatalf("wrap: %+v", p)
	}
	if !p.typeRune(lib, '/') {
		t.Fatal("/ should open Search")
	}
	p.step(lib, -5) // clamped at Al-Fātiḥa, āyāt reset
	if p.Surah != 1 || p.From != 1 || p.To != 1 {
		t.Fatalf("clamp: %+v", p)
	}
	typeAll(&p, lib, "112")
	p.backspace(lib) // "11" → Hūd
	if p.Surah != 11 {
		t.Fatalf("backspace: %+v", p)
	}
}
