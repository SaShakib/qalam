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

func TestQuranBoxSearch(t *testing.T) {
	lib, err := quran.Load(quranFS())
	if err != nil {
		t.Fatal(err)
	}
	b := newQuranBox(lib, "2:255-257")
	if b.Pick.passage().Reference() != "2:255-257" || b.Sel != -1 || b.In != quran.InArabic {
		t.Fatalf("opens on the last choice, Arabic search: %+v", b)
	}
	search := func() { b.setResults(lib.Search(b.query(), b.In, "en.sahih", quranResultLimit)) }

	// Latin keys are searched as Arabic, and the field shows the Arabic.
	for _, r := range "qul huwa allaahu AaHaduN" {
		b.typeText(string(r))
	}
	if got := b.shown(); got != "قُل هُوَ اللَّهُ أَحَدٌ" || !b.shownIsArabic() {
		t.Errorf("field shows %q", got)
	}
	search()
	if b.Res.Total != 1 || b.Sel != 0 || b.Pick.passage().Reference() != "112:1" {
		t.Errorf("qul huwa allaahu AaHaduN → %+v, pick %s", b.Res, b.Pick.passage().Reference())
	}
	// Backspace takes off the last Latin key, not an Arabic letter.
	b.backspace()
	if b.Keys != "qul huwa allaahu AaHadu" {
		t.Errorf("backspace: %q", b.Keys)
	}

	// The first result is chosen; ↑↓ go through them (round at the ends) and fill Sūrah / From / To.
	b.clear()
	b.typeText("يا أيها الذين آمنوا")
	search()
	if b.Res.Total != 89 || b.Pick.passage().Reference() != "2:104" {
		t.Fatalf("يا أيها الذين آمنوا → %d, %s", b.Res.Total, b.Pick.passage().Reference())
	}
	b.move(1)
	if b.Pick.passage().Reference() != "2:153" {
		t.Errorf("↓ → %s", b.Pick.passage().Reference())
	}
	b.move(-2)
	if b.Sel != len(b.Res.Rows)-1 {
		t.Errorf("↑ at the top goes to the end: %d", b.Sel)
	}
	if b.Top != len(b.Res.Rows)-quranVisibleRows {
		t.Errorf("the list scrolls to the chosen row: top %d", b.Top)
	}
	b.choose(2)
	if b.Top != 2 || b.Pick.passage().Reference() != "2:172" {
		t.Errorf("click: top %d, %s", b.Top, b.Pick.passage().Reference())
	}

	// A reference fills all three; a sūrah name gives its first āyah.
	b.clear()
	b.typeText("baqarah 255-257")
	search()
	if b.Pick.passage().Reference() != "2:255-257" {
		t.Errorf("baqarah 255-257 → %s", b.Pick.passage().Reference())
	}
	b.clear()
	b.typeText("yasin")
	search()
	if b.Res.Rows[0].Kind != quran.KindSurah || b.Pick.passage().Reference() != "36:1" {
		t.Errorf("yasin → %s", b.Pick.passage().Reference())
	}

	// Translation mode: Latin stays Latin and searches the translation.
	b.clear()
	b.In = quran.InTranslation
	b.typeText("mercy")
	if b.shown() != "mercy" || b.shownIsArabic() {
		t.Errorf("translation field: %q", b.shown())
	}
	search()
	if b.Res.Translation != "en.sahih" || b.Res.Rows[0].InTranslation != "en.sahih" || b.Pick.passage().Reference() != "2:64" {
		t.Errorf("mercy → %+v", b.Res.Rows[:1])
	}

	// Clearing keeps the dropdowns as they are.
	b.clear()
	if b.Keys != "" || b.Res.Rows != nil || b.Sel != -1 || b.Pick.passage().Reference() != "2:64" {
		t.Errorf("clear: %+v", b)
	}
}

func TestQuranBoxTab(t *testing.T) {
	b := newQuranBox(nil, "")
	b.tab(1)
	if !b.OnPick || b.Pick.Field != fieldSurah {
		t.Fatalf("Tab → Sūrah: %+v", b)
	}
	b.tab(1)
	b.tab(1)
	if b.Pick.Field != fieldTo {
		t.Errorf("Tab Tab → To: %+v", b)
	}
	b.tab(1)
	if b.OnPick {
		t.Errorf("Tab after To goes back to the field")
	}
	b.tab(-1)
	if !b.OnPick || b.Pick.Field != fieldTo {
		t.Errorf("Shift+Tab from the field → To: %+v", b)
	}
	b.typeText("x") // typing goes back to the field
	if b.OnPick || b.Keys != "x" {
		t.Errorf("typing: %+v", b)
	}
}
