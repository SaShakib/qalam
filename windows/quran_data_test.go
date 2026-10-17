package main

import (
	"testing"

	"qalam/quran"
)

func TestEmbeddedQuran(t *testing.T) {
	l, err := quran.Load(quranFS())
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Available()) != 3 {
		t.Errorf("bundled translations: %v", l.Available())
	}
	p, ok := l.Resolve("2:255")
	if !ok || l.Arabic(p.Surah, p.From) == "" || l.TranslationText("bn.hoque", 2, 255) == "" {
		t.Error("2:255 missing from the embedded files")
	}
}

func TestQuranMode(t *testing.T) {
	var s Settings
	if s.quranMode() != quran.Both || s.quranTranslation() != "en.sahih" {
		t.Errorf("defaults: %s %s", s.quranMode(), s.quranTranslation())
	}
	s.QuranInsert = "translation"
	if s.quranMode() != quran.TranslationOnly {
		t.Errorf("mode: %s", s.quranMode())
	}
	s.QuranInsert = "bogus"
	if s.quranMode() != quran.Both {
		t.Errorf("bad value: %s", s.quranMode())
	}
}
