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
	if strings.Join(have, ",") != "en.sahih,en.usmani,bn.bengali,bn.bayaan,bn.mujibur,bn.taisirul,bn.zakaria,bn.hoque,ur.usmani" {
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

func TestInsertedAyahUsesArabicMarker(t *testing.T) {
	l := library(t)
	p := Passage{Surah: 1, From: 3, To: 3}
	got := l.Text(p, Arabic, "")
	if !strings.HasSuffix(got, "﴿٣﴾") || strings.Contains(got, "(1:3)") {
		t.Errorf("inserted text = %q, want an Arabic number inside the verse marker", got)
	}
	if got := AyahMarker(255); got != "﴿٢٥٥﴾" {
		t.Errorf("AyahMarker(255) = %q", got)
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

func TestSearch(t *testing.T) {
	l := library(t)
	// A name alone lists sūrahs (exact match first, then names that start with it).
	r := l.Search("baqar", InTranslation, "en.sahih", 6)
	if len(r.Rows) == 0 || r.Rows[0].Kind != KindSurah || r.Rows[0].Surah.Number != 2 {
		t.Fatalf("baqar: %+v", r.Rows)
	}
	n := 0
	for _, row := range l.Search("al", InTranslation, "en.sahih", 6).Rows {
		if row.Kind == KindSurah {
			n++
		}
	}
	if n != 4 {
		t.Errorf("al: %d sūrah rows, want 4", n)
	}
	// A reference gives the passage first, in either mode.
	for _, in := range []SearchIn{InArabic, InTranslation} {
		r = l.Search("baqarah 255-257", in, "en.sahih", 6)
		if len(r.Rows) == 0 || r.Rows[0].Kind != KindPassage || r.Rows[0].Passage.Reference() != "2:255-257" {
			t.Errorf("baqarah 255-257 (%d): %+v", in, r.Rows)
		}
	}
	r = l.Search("2:255", InArabic, "en.sahih", 6)
	if len(r.Rows) != 1 || r.Rows[0].Passage.Reference() != "2:255" || r.Arabic != "" {
		t.Errorf("2:255: %+v", r)
	}
	r = l.Search("2:290", InArabic, "en.sahih", 6)
	if len(r.Rows) != 1 || r.Rows[0].Kind != KindProblem || r.Rows[0].Text != "Al-Baqara has 286 āyāt" {
		t.Errorf("2:290: %+v", r.Rows)
	}
	if r := l.Search("   ", InArabic, "en.sahih", 6); r.Rows != nil {
		t.Errorf("empty: %+v", r)
	}
	// Latin in Arabic mode is read the Qalam way, with full harakat.
	r = l.Search("qul huwa allaahu AaHaduN", InArabic, "en.sahih", 6)
	if r.Arabic != "قُل هُوَ اللَّهُ أَحَدٌ" || r.Total != 1 || len(r.Rows) != 1 || r.Rows[0].Passage.Reference() != "112:1" || r.Rows[0].InTranslation != "" {
		t.Errorf("qul huwa allaahu AaHaduN: %+v", r)
	}
	// "raHmaan" is also the name of sūrah 55, which comes first.
	r = l.Search("raHmaan", InArabic, "en.sahih", 6)
	if r.Rows[0].Kind != KindSurah || r.Rows[0].Surah.Number != 55 || r.Rows[1].Passage.Reference() != "1:1" {
		t.Errorf("raHmaan: %+v", r.Rows)
	}
	// Pasted Arabic with typed Latin after it.
	if r := l.Search("قُلْ huwa", InArabic, "en.sahih", 6); r.Arabic != "قُلْ هُوَ" || r.Total == 0 {
		t.Errorf("mixed: %+v", r)
	}
	// Arabic letters search the Arabic in either mode.
	r = l.Search("يا أيها الذين آمنوا", InTranslation, "en.sahih", 6)
	if r.Total != 89 || r.Rows[0].Passage.Reference() != "2:104" || r.SearchedIn != "the Arabic (harakat optional)" {
		t.Errorf("يا أيها الذين آمنوا: %d %+v", r.Total, r.Rows)
	}
	// Translation mode: the chosen translation if it's English, else Saheeh International.
	r = l.Search("Mercy of Allah", InTranslation, "en.usmani", 3)
	if r.Translation != "en.usmani" || r.Total == 0 || r.Rows[0].InTranslation != "en.usmani" {
		t.Errorf("en.usmani: %+v", r)
	}
	r = l.Search("Mercy of Allah", InTranslation, "bn.hoque", 3)
	if r.Translation != "en.sahih" || r.Total != 41 || r.Rows[0].Passage.Reference() != "2:64" || r.SearchedIn != "English: Saheeh International" {
		t.Errorf("Mercy of Allah: %+v", r)
	}
	// Bangla letters: the chosen Bangla translation, or Muhiuddin Khan.
	if r := l.Search("আল্লাহর", InArabic, "bn.hoque", 3); r.Translation != "bn.hoque" || r.Total == 0 {
		t.Errorf("Bangla: %+v", r)
	}
	if r := l.Search("আল্লাহর", InArabic, "en.sahih", 3); r.Translation != "bn.bengali" || r.Total == 0 {
		t.Errorf("Bangla (English chosen): %+v", r)
	}
}

func TestSearchTranslation(t *testing.T) {
	l := library(t)
	// Case, accents and the order of the words don't matter.
	a, n := l.SearchTranslation("en.sahih", "MERCY allah", 300)
	b, m := l.SearchTranslation("en.sahih", "allāh, mercy", 300)
	if n == 0 || n != m || refs(a) != refs(b) {
		t.Errorf("folding: %d %d", n, m)
	}
	if _, n := l.SearchTranslation("en.sahih", "x", 5); n != 0 {
		t.Errorf("one letter: %d", n)
	}
	if _, n := l.SearchTranslation("xx.none", "mercy", 5); n != 0 {
		t.Errorf("unknown translation: %d", n)
	}
	// Bangla: with or without the hasant and ZWNJ at the end of আল্লাহ্‌.
	_, x := l.SearchTranslation("bn.bengali", "আল্লাহ্‌র", 5)
	_, y := l.SearchTranslation("bn.bengali", "আল্লাহর", 5)
	if x == 0 || x != y {
		t.Errorf("hasant: %d %d", x, y)
	}
	for in, want := range map[string]string{
		"Ṣalāh": "salah", "ÉCOLE": "ecole",
		"য়":         "য়", // য় = য + ়
		"আল্লাহ্‌র": "আল্লাহর",
		"ক্ষ":       "ক্ষ", // ক্ষ keeps its hasant
		"ক্য়া":      "ক্য়া",
	} {
		if got := FoldText(in); got != want {
			t.Errorf("FoldText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBasmalaAndParse(t *testing.T) {
	l := library(t)
	// 95:1 and 97:1 are written بِّسْمِ (with idghām): the basmala still goes.
	if got := l.Arabic(95, 1); Bare(got, false) != "والتين والزيتون" {
		t.Errorf("95:1 = %q", got)
	}
	if got := l.Arabic(97, 1); strings.Contains(Bare(got, false), "بسم") {
		t.Errorf("97:1 = %q", got)
	}
	if got := l.Arabic(9, 1); !strings.HasPrefix(Bare(got, false), "براءه من الله") {
		t.Errorf("9:1 = %q", got)
	}
	if _, n := l.SearchArabic("بسم الله الرحمن الرحيم", false, 5); n != 2 { // 1:1 and 27:30
		t.Errorf("basmala found %d times", n)
	}
	// A text that starts with a mark keeps it (ur.usmani 4:94).
	data := []byte("1|1|ّx\n")
	for i := 1; i < Total; i++ {
		data = append(data, "1|1|y\n"...)
	}
	if got := Parse(data); got == nil || got[0] != "ّx" {
		t.Errorf("Parse: %v", got == nil)
	}
	if l.TranslationText("ur.usmani", 4, 94) == "" {
		t.Error("ur.usmani 4:94 missing")
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
		{"رهمان", true, 70, "2:20 4:90 5:49 15:25 18:6"}, // the same letters first, then similar ones (ٱلرَّحْمَٰنِ)
		{"الرحمن", false, 48, "1:1 1:3 2:163 13:30 17:110"},
		{"رحمن", false, 64, "1:1 1:3 2:163 2:286 7:23"},
		{"صلاة", false, 65, "2:3 2:43 2:45 2:83 2:110"}, // modern spelling finds ٱلصَّلَوٰةَ
		{"ٱلصَّلَوٰةَ", false, 61, "2:3 2:43 2:45 2:83 2:110"},
		{"رَحْمَة", false, 75, "2:157 2:178 3:8 3:107 3:157"}, // with harakat: those matches first
		{"الهمد", true, 30, "8:4 27:82 1:2 6:1 6:45"},
		{"قل هو", true, 25, "2:222 3:165 6:65 13:30 38:67"},
		{"بسم", true, 7, "1:1 2:20 11:41 27:19 27:30"},
		{"بسم", false, 5, "1:1 2:20 11:41 27:19 27:30"},
		{"يا أيها الذين آمنوا", false, 89, "2:104 2:153 2:172 2:178 2:183"}, // يَٰٓأَيُّهَا is one word in the mushaf
		{"يا بني", false, 20, "2:40 2:47 2:122 2:132 5:72"},
		{"آمنوا", false, 268, "2:9 2:13 2:14 2:25 2:26"}, // = ءَامَنُوا۟
		{"التين والزيتون", false, 1, "95:1"},
		{"ر", false, 0, ""},
	}
	for _, c := range cases {
		hits, total := l.SearchArabic(c.query, c.phonetic, 5)
		if total != c.total || refs(hits) != c.first {
			t.Errorf("SearchArabic(%q) = %d [%s], want %d [%s]", c.query, total, refs(hits), c.total, c.first)
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
	want := a1 + " ﴿١﴾\nSay, \"He is Allah, [who is] One,\n\n" + a2 + " ﴿٢﴾\nAllah, the Eternal Refuge."
	if got := l.Text(p, Both, "en.sahih"); got != want {
		t.Errorf("both:\n%s", got)
	}
	if got := l.Text(p, Arabic, "en.sahih"); got != a1+" ﴿١﴾ "+a2+" ﴿٢﴾" {
		t.Errorf("arabic: %s", got)
	}
	if got := l.Text(p, TranslationOnly, "en.sahih"); got != "Say, \"He is Allah, [who is] One, ﴿١﴾ Allah, the Eternal Refuge. ﴿٢﴾" {
		t.Errorf("translation: %s", got)
	}
	p, _ = l.Resolve("2:1")
	if got := l.Text(p, Both, "en.sahih"); got != "الٓمٓ ﴿١﴾\nAlif, Lam, Meem." {
		t.Errorf("2:1: %s", got)
	}
	if got := l.Text(p, Both, "xx.none"); got != "الٓمٓ ﴿١﴾" { // unknown translation: Arabic only
		t.Errorf("unknown translation: %s", got)
	}
	if !strings.HasPrefix(l.Text(p, Both, "bn.hoque"), "الٓمٓ ﴿١﴾\n") {
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
