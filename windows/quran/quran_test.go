package quran

import (
	"os"
	"strings"
	"testing"
)

// The expected results are what the Mac version gives (engine/Quran.swift, via `qalam ayah` and
// `qalam search … --trans none`), so both keyboards find the same āyāt.

var lib *Library

func library(t *testing.T) *Library {
	t.Helper()
	if lib == nil {
		l, err := Load(os.DirFS("../../quran"))
		if err != nil {
			t.Fatal(err)
		}
		lib = l
	}
	return lib
}

func TestLoad(t *testing.T) {
	l := library(t)
	if len(l.Surahs) != 114 || l.Surahs[1].Name != "Al-Baqara" || l.Surahs[1].Ayahs != 286 {
		t.Fatalf("sūrahs: %d, %+v", len(l.Surahs), l.Surahs[1])
	}
	var have []string
	for _, tr := range l.Available() {
		have = append(have, tr.ID)
	}
	if strings.Join(have, ",") != "en.sahih,bn.bengali,bn.hoque" {
		t.Errorf("available translations: %v", have)
	}
	if got := l.Translation("bn.bengali").Title(); got != "Bangla: Muhiuddin Khan (মুহিউদ্দীন খান)" {
		t.Errorf("title: %s", got)
	}
	if got := l.Translation("en.sahih").Title(); got != "English: Saheeh International" {
		t.Errorf("title: %s", got)
	}
	if got := l.Arabic(2, 1); got != "الٓمٓ" { // the basmala before āyah 1 is left out
		t.Errorf("2:1 = %q", got)
	}
	if got := l.Arabic(1, 1); !strings.HasPrefix(got, "بِسْمِ") || l.Arabic(1, 2) == "" {
		t.Errorf("1:1 = %q", got)
	}
	if l.Arabic(2, 287) != "" || l.Arabic(115, 1) != "" {
		t.Error("out of range should be empty")
	}
}

func TestResolve(t *testing.T) {
	l := library(t)
	cases := map[string]string{
		"2:255":         "2:255",
		"2:255-257":     "2:255-257",
		"2 255":         "2:255",
		"2.255":         "2:255",
		"baqarah 255":   "2:255",
		"Al-Baqara 255": "2:255",
		"al imran 3-5":  "3:3-5",
		"aal imraan 3":  "3:3",
		"yaseen 1":      "36:1",
		"yasin 1":       "36:1",
		"faatiha 2":     "1:2",
		"fatiha 2":      "1:2",
		"112":           "112:1-4",
		"cow 3":         "2:3",
		"البقرة 5":      "2:5",
		"بقرة 5":        "2:5",
		"ikhlas":        "112:1-4",
		"2:285-300":     "2:285-286",
		"2:255–257":     "2:255-257", // en dash
		"2:255-":        "2:255",
		"2:257-255":     "2:257",
		"  2:255  ":     "2:255",
	}
	for q, want := range cases {
		p, ok := l.Resolve(q)
		if !ok || p.Reference() != want {
			t.Errorf("Resolve(%q) = %v %v, want %s", q, p.Reference(), ok, want)
		}
	}
	for _, q := range []string{"2:290", "115:1", "xyzzy 3", ""} {
		if p, ok := l.Resolve(q); ok {
			t.Errorf("Resolve(%q) = %s, want nothing", q, p.Reference())
		}
	}
}

func TestFold(t *testing.T) {
	cases := map[string]string{
		"Al-Baqara": "albaqara", "al baqarah": "albaqara", "albaqara": "albaqara",
		"Yaseen": "iasin", "yasin": "iasin", "Faatiha": "fatiha", "Al-Faatiha": "alfatiha",
		"البقرة": "بقرة", "الْبَقَرَة": "بقرة", "آل عمران": "عمران",
	}
	for in, want := range cases {
		if got := Fold(in); got != want {
			t.Errorf("Fold(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSuggestions(t *testing.T) {
	l := library(t)
	// A name alone lists sūrahs (exact match first, then names that start with it).
	rows, _ := l.Lookup("baqar", 6)
	if len(rows) == 0 || rows[0].Kind != KindSurah || rows[0].Surah.Number != 2 {
		t.Fatalf("baqar: %+v", rows)
	}
	rows, _ = l.Lookup("al", 6)
	n := 0
	for _, r := range rows {
		if r.Kind == KindSurah {
			n++
		}
	}
	if n != 4 {
		t.Errorf("al: %d sūrah rows, want 4", n)
	}
	// A reference gives the passage.
	rows, _ = l.Lookup("baqarah 255", 6)
	if len(rows) == 0 || rows[0].Kind != KindPassage || rows[0].Passage.Reference() != "2:255" {
		t.Errorf("baqarah 255: %+v", rows)
	}
	// Too far: a problem row.
	rows, _ = l.Lookup("2:290", 6)
	if len(rows) != 1 || rows[0].Kind != KindProblem || rows[0].Text != "Al-Baqara has 286 āyāt" {
		t.Errorf("2:290: %+v", rows)
	}
	// Nothing at all.
	rows, _ = l.Lookup("xyzzyx", 6)
	if len(rows) != 1 || rows[0].Kind != KindProblem {
		t.Errorf("xyzzyx: %+v", rows)
	}
	if rows, _ := l.Lookup("   ", 6); rows != nil {
		t.Errorf("empty: %+v", rows)
	}
	// Search: Arabic, and Latin read through Qalam.
	// "rahmaan" is also the name of sūrah 55, which comes first.
	rows, total := l.Lookup("rahmaan", 6)
	if total != 58 || len(rows) != 6 || rows[0].Surah.Number != 55 || rows[1].Kind != KindHit || rows[1].Passage.Reference() != "1:1" {
		t.Errorf("rahmaan: %d %+v", total, rows)
	}
	rows, total = l.Lookup("الرحمن", 6)
	if total != 50 || rows[0].Surah.Number != 55 || rows[1].Kind != KindHit {
		t.Errorf("الرحمن: %d %+v", total, rows)
	}
}

func refs(ps []Passage) string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Reference())
	}
	return strings.Join(out, " ")
}

func TestSearchArabic(t *testing.T) {
	l := library(t)
	cases := []struct {
		query    string
		phonetic bool
		total    int
		first    string
	}{
		{"رهمان", true, 58, "1:1 1:3 2:163 13:30 17:110"}, // "rahmaan" through Qalam: similar letters count
		{"الرحمن", false, 50, "1:1 1:3 2:163 13:30 17:110"},
		{"رحمن", false, 66, "1:1 1:3 2:163 2:286 7:23"},
		{"صلاة", false, 65, "2:3 2:43 2:45 2:83 2:110"}, // modern spelling finds ٱلصَّلَوٰةَ
		{"ٱلصَّلَوٰةَ", false, 61, "2:3 2:43 2:45 2:83 2:110"},
		{"رَحْمَة", false, 75, "2:157 2:178 3:8 3:107 3:157"}, // with harakat: those matches first
		{"الهمد", true, 28, "1:2 6:1 6:45 7:43 9:112"},
		{"قل هو", true, 11, "2:222 3:165 6:65 13:30 38:67"},
		{"بسم", true, 9, "1:1 2:20 11:41 27:19 27:30"},
		{"ر", false, 0, ""},
	}
	for _, c := range cases {
		hits, total := l.SearchArabic(c.query, c.phonetic, 5)
		if total != c.total || refs(hits) != c.first {
			t.Errorf("SearchArabic(%q) = %d [%s], want %d [%s]", c.query, total, refs(hits), c.total, c.first)
		}
	}
}

func TestPhoneticQuery(t *testing.T) {
	// The Latin is read without harakat, as the Mac does.
	for in, want := range map[string]string{"alhamdu": "الهمد", "qul huwa": "قل هو", "bismi": "بسم"} {
		rows, _ := library(t).Lookup(in, 1)
		_ = rows
		hits, _ := library(t).SearchArabic(want, true, 1)
		if len(rows) == 0 || len(hits) == 0 || rows[0].Passage != hits[0] {
			t.Errorf("%s: rows %+v, want first hit %+v", in, rows, hits)
		}
	}
}

func TestText(t *testing.T) {
	l := library(t)
	p, _ := l.Resolve("112:1-2")
	a1, a2 := l.Arabic(112, 1), l.Arabic(112, 2)
	if !strings.HasPrefix(a1, "قُلْ هُوَ") {
		t.Errorf("112:1 = %s", a1)
	}
	want := a1 + "\nSay, \"He is Allah, [who is] One, (112:1)\n\n" + a2 + "\nAllah, the Eternal Refuge. (112:2)"
	if got := l.Text(p, Both, "en.sahih"); got != want {
		t.Errorf("both:\n%s", got)
	}
	if got := l.Text(p, Arabic, "en.sahih"); got != a1+" (112:1) "+a2+" (112:2)" {
		t.Errorf("arabic: %s", got)
	}
	if got := l.Text(p, TranslationOnly, "en.sahih"); got != "Say, \"He is Allah, [who is] One, (112:1) Allah, the Eternal Refuge. (112:2)" {
		t.Errorf("translation: %s", got)
	}
	p, _ = l.Resolve("2:1")
	if got := l.Text(p, Both, "en.sahih"); got != "الٓمٓ\nAlif, Lam, Meem. (2:1)" {
		t.Errorf("2:1: %s", got)
	}
	if got := l.Text(p, Both, "xx.none"); got != "الٓمٓ (2:1)" { // unknown translation: Arabic only
		t.Errorf("unknown translation: %s", got)
	}
	if !strings.HasPrefix(l.Text(p, Both, "bn.hoque"), "الٓمٓ\n") {
		t.Error("bn.hoque missing")
	}
}

func TestForKFGQPC(t *testing.T) {
	if got := ForKFGQPC("a۟bۣc۫d"); got != "a۠bۜc۬d" {
		t.Errorf("%q", got)
	}
}

func TestGzipFiles(t *testing.T) {
	// The Windows app embeds the texts gzipped (assets/quran); they must read the same.
	l, err := Load(os.DirFS("../assets/quran"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"quran-uthmani", "en.sahih", "bn.bengali", "bn.hoque"} {
		a, b := l.lines(id), library(t).lines(id)
		if len(a) != Total || strings.Join(a, "\n") != strings.Join(b, "\n") {
			t.Errorf("assets/quran/%s.txt.gz differs from quran/%s.txt (run `make guide`)", id, id)
		}
	}
}
