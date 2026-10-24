// Package quran is the Windows port of engine/Quran.swift: the whole Qur'an (Tanzil Uthmani text,
// full harakat) and its bundled translations, for the "insert āyah" box (Ctrl+Alt+Q).
// Reference parsing, sūrah-name folding, the Arabic and translation search and the inserted text
// match the Mac version (quran_test.go checks the same cases).
package quran

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"qalam/engine"
)

const Total = 6236

type Surah struct {
	Number  int
	Ayahs   int
	Arabic  string
	Name    string // Tanzil transliteration, e.g. "Al-Baqara"
	English string // "The Cow"
	Type    string // Meccan / Medinan
}

type Translation struct {
	ID         string // Tanzil id, e.g. "en.sahih"
	Language   string
	Name       string
	Translator string
	Bundled    bool
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 0x7F {
			return false
		}
	}
	return true
}

// Title: "English: Saheeh International", "Bangla: Muhiuddin Khan (মুহিউদ্দীন খান)".
func (t Translation) Title() string {
	if t.Name == t.Translator || isASCII(t.Name) && strings.HasPrefix(t.Translator, t.Name) {
		return t.Language + ": " + t.Name
	}
	return t.Language + ": " + t.Translator + " (" + t.Name + ")"
}

// Insert is what the keyboard puts in: Arabic only, Arabic with a translation, or the translation only.
type Insert string

const (
	Arabic          Insert = "arabic"
	Both            Insert = "both"
	TranslationOnly Insert = "translation"
)

var InsertModes = []Insert{Arabic, Both, TranslationOnly}

func (m Insert) Title() string {
	switch m {
	case Arabic:
		return "Arabic only"
	case TranslationOnly:
		return "Translation only"
	}
	return "Arabic + translation"
}

// Passage is a run of āyāt inside one sūrah.
type Passage struct{ Surah, From, To int }

func (p Passage) Reference() string {
	if p.From == p.To {
		return fmt.Sprintf("%d:%d", p.Surah, p.From)
	}
	return fmt.Sprintf("%d:%d-%d", p.Surah, p.From, p.To)
}

type Kind int

const (
	KindSurah   Kind = iota // pick → the query becomes "2:" so the āyah number comes next
	KindPassage             // pick → insert
	KindHit                 // a search match (in the Arabic, or in the translation InTranslation)
	KindProblem             // e.g. "Al-Baqara has 286 āyāt"
)

// Suggestion is one row of the dropdown under the query.
type Suggestion struct {
	Kind    Kind
	Surah   Surah   // KindSurah
	Passage Passage // KindPassage, KindHit
	Text    string  // KindProblem
	// KindHit: the translation the words were found in ("" = found in the Arabic).
	InTranslation string
}

type Library struct {
	Surahs  []Surah
	Catalog []Translation // every translation in translations.tsv (only bundled ones have files here)
	fsys    fs.FS
	start   []int // index of each sūrah's first āyah (1-based sūrah)

	mu    sync.Mutex
	texts map[string][]string

	searchOnce   sync.Once
	bareArabic   [][2]string
	looseArabic  [][2]string
	joinedArabic [][2]string
	joinedLoose  []string
	lightArabic  []string
	keys         [][]string
	folded       map[string][]string // translation id → its texts folded for searching
}

// Load reads surahs.tsv, translations.tsv and, when needed, "<id>.txt" or "<id>.txt.gz" from fsys.
func Load(fsys fs.FS) (*Library, error) {
	data, err := fs.ReadFile(fsys, "surahs.tsv")
	if err != nil {
		return nil, err
	}
	l := &Library{fsys: fsys, texts: map[string][]string{}}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		c := strings.Split(strings.TrimRight(line, "\r"), "\t")
		if len(c) < 6 {
			continue
		}
		n, e1 := strconv.Atoi(c[0])
		a, e2 := strconv.Atoi(c[1])
		if e1 != nil || e2 != nil {
			continue
		}
		l.Surahs = append(l.Surahs, Surah{n, a, c[2], c[3], c[4], c[5]})
	}
	if len(l.Surahs) != 114 {
		return nil, fmt.Errorf("surahs.tsv: %d sūrahs", len(l.Surahs))
	}
	l.start = []int{0, 0}
	for _, s := range l.Surahs[:113] {
		l.start = append(l.start, l.start[len(l.start)-1]+s.Ayahs)
	}
	if t, err := fs.ReadFile(fsys, "translations.tsv"); err == nil {
		for _, line := range strings.Split(string(t), "\n") {
			if strings.HasPrefix(line, "#") {
				continue
			}
			c := strings.FieldsFunc(strings.TrimRight(line, "\r"), func(r rune) bool { return r == '\t' })
			if len(c) >= 5 {
				l.Catalog = append(l.Catalog, Translation{c[0], c[1], c[2], c[3], c[4] == "yes"})
			}
		}
	}
	return l, nil
}

// Available: translations whose file is here, in catalog order.
func (l *Library) Available() []Translation {
	var out []Translation
	for _, t := range l.Catalog {
		if l.hasFile(t.ID) {
			out = append(out, t)
		}
	}
	return out
}

func (l *Library) Translation(id string) *Translation {
	for i := range l.Catalog {
		if l.Catalog[i].ID == id {
			return &l.Catalog[i]
		}
	}
	return nil
}

func (l *Library) Surah(n int) *Surah {
	if n < 1 || n > 114 {
		return nil
	}
	return &l.Surahs[n-1]
}

// Arabic is one āyah with full harakat (the basmala Tanzil puts before āyah 1 is left out).
func (l *Library) Arabic(surah, ayah int) string {
	t, i := l.lines("quran-uthmani"), l.index(surah, ayah)
	if t == nil || i < 0 {
		return ""
	}
	if ayah == 1 && surah != 1 {
		return dropBasmala(t[i])
	}
	return t[i]
}

// dropBasmala: Tanzil writes the basmala in front of āyah 1 (sometimes spelled بِّسْمِ, with idghām,
// as in 95:1), so it's matched by its letters.
func dropBasmala(s string) string {
	words := strings.Fields(s)
	if len(words) <= 4 || Bare(strings.Join(words[:4], " "), false) != "بسم الله الرحمن الرحيم" {
		return s
	}
	return strings.Join(words[4:], " ")
}

func (l *Library) TranslationText(id string, surah, ayah int) string {
	t, i := l.lines(id), l.index(surah, ayah)
	if t == nil || i < 0 {
		return ""
	}
	return t[i]
}

func (l *Library) index(surah, ayah int) int {
	s := l.Surah(surah)
	if s == nil || ayah < 1 || ayah > s.Ayahs {
		return -1
	}
	return l.start[surah] + ayah - 1
}

func (l *Library) hasFile(id string) bool {
	for _, name := range []string{id + ".txt", id + ".txt.gz"} {
		if _, err := fs.Stat(l.fsys, name); err == nil {
			return true
		}
	}
	return false
}

func (l *Library) lines(id string) []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t, ok := l.texts[id]; ok {
		return t
	}
	var t []string
	if data, err := fs.ReadFile(l.fsys, id+".txt"); err == nil {
		t = Parse(data)
	} else if data, err := fs.ReadFile(l.fsys, id+".txt.gz"); err == nil {
		if z, err := gzip.NewReader(bytes.NewReader(data)); err == nil {
			if raw, err := io.ReadAll(z); err == nil {
				t = Parse(raw)
			}
		}
	}
	l.texts[id] = t
	return t
}

// Parse turns "sura|aya|text" lines in order into the 6236 texts; nil if it isn't a whole Qur'an.
// It splits on bytes, so a text that starts with a mark (ur.usmani 4:94) keeps it.
func Parse(data []byte) []string {
	out := make([]string, 0, Total)
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 || line[0] < '0' || line[0] > '9' {
			continue
		}
		parts := bytes.SplitN(line, []byte("|"), 3)
		if len(parts) != 3 {
			continue
		}
		out = append(out, strings.TrimSpace(string(parts[2])))
	}
	if len(out) != Total {
		return nil
	}
	return out
}

// ---------- Query: "2:255", "2:255-257", "2 255", "baqarah 255", "al imran 3-5", "cow", "2" ----------

type Query struct {
	Surahs   []Surah // what the sūrah part matches (best first)
	From, To int     // 0 = not given
}

// trailing "255" or "255-257" (or "255-") is the āyah part, if something comes before it
var ayahPart = regexp.MustCompile(`[\s:.]\s*(\d+)\s*(-\s*(\d*))?\s*$`)

func (l *Library) ParseQuery(raw string) Query {
	q := strings.ReplaceAll(strings.Trim(raw, " \t"), "–", "-")
	var out Query
	if m := ayahPart.FindStringSubmatchIndex(q); m != nil {
		out.From, _ = strconv.Atoi(q[m[2]:m[3]])
		if m[6] >= 0 && m[7] > m[6] {
			out.To, _ = strconv.Atoi(q[m[6]:m[7]])
		}
		q = q[:m[0]]
	}
	out.Surahs = l.matchSurahs(strings.Trim(q, " :.-"))
	return out
}

// SearchIn is where Latin letters are searched: the Arabic (read the Qalam way) or the translation.
type SearchIn int

const (
	InArabic SearchIn = iota
	InTranslation
)

// Results is what the box lists for a query (like the Mac box's search).
type Results struct {
	Rows        []Suggestion
	Total       int    // āyāt found in all (Rows holds at most the limit)
	SearchedIn  string // "the Arabic", "English: Saheeh International"…
	Arabic      string // the Arabic searched for, when Latin was read the Qalam way
	Translation string // the translation searched ("" = the Arabic)
}

// Script: "arabic", "bangla" or "latin".
func Script(s string) string {
	for _, r := range s {
		if r >= 0x0600 && r <= 0x06FF || r >= 0x0750 && r <= 0x077F {
			return "arabic"
		}
	}
	for _, r := range s {
		if r >= 0x0980 && r <= 0x09FF {
			return "bangla"
		}
	}
	return "latin"
}

// SearchTranslationFor is the translation a Latin or Bangla search looks in: the chosen one if it's
// in that language, otherwise Saheeh International / Muhiuddin Khan.
func (l *Library) SearchTranslationFor(script, chosen string) string {
	t := l.Translation(chosen)
	if script == "bangla" {
		if t != nil && t.Language == "Bangla" {
			return chosen
		}
		return "bn.bengali"
	}
	if t != nil && t.Language == "English" {
		return chosen
	}
	return "en.sahih"
}

// PhoneticArabic is Latin read the Qalam way, with full harakat (qul huwa → قُلْ هُوَ); Arabic letters
// in it stay as they are.
func PhoneticArabic(latin string) string {
	return engine.Text(latin, engine.DefaultOptions())
}

func onlyReference(s string) bool {
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r == ':' || r == '-' || r == ' ') {
			return false
		}
	}
	return true
}

// Search: a reference ("2:255", "baqarah 255-257") first, then sūrah names, then āyāt. Arabic letters
// search the Arabic; Latin searches the Arabic (read the Qalam way) or the translation, as in says;
// Bangla letters search a Bangla translation.
func (l *Library) Search(raw string, in SearchIn, translation string, limit int) Results {
	q := strings.Trim(raw, " \t")
	var res Results
	if q == "" {
		return res
	}
	digits := strings.ContainsAny(q, "0123456789")
	if digits {
		if p, ok := l.Resolve(q); ok {
			if _, err := strconv.Atoi(q); err == nil || l.ParseQuery(q).From != 0 {
				res.Rows = append(res.Rows, Suggestion{Kind: KindPassage, Passage: p})
			}
		} else if pq := l.ParseQuery(q); len(pq.Surahs) > 0 && pq.From != 0 {
			s := pq.Surahs[0]
			res.Rows = append(res.Rows, Suggestion{Kind: KindProblem, Text: fmt.Sprintf("%s has %d āyāt", s.Name, s.Ayahs)})
		}
	} else {
		for i, s := range l.ParseQuery(q).Surahs {
			if i == 4 {
				break
			}
			res.Rows = append(res.Rows, Suggestion{Kind: KindSurah, Surah: s})
		}
	}
	var hits []Passage
	script := Script(q)
	switch {
	case script == "arabic" && !hasLatinLetter(q):
		hits, res.Total = l.SearchArabic(q, false, limit)
		res.SearchedIn = "the Arabic (harakat optional)"
	case script == "arabic": // pasted Arabic and typed Latin: the Latin read the Qalam way
		res.Arabic = PhoneticArabic(q)
		hits, res.Total = l.SearchArabic(res.Arabic, true, limit)
		res.SearchedIn = "the Arabic"
	case script == "latin" && in == InArabic && !onlyReference(q):
		res.Arabic = PhoneticArabic(q)
		hits, res.Total = l.SearchArabic(res.Arabic, true, limit)
		res.SearchedIn = "the Arabic"
	case script == "latin" && onlyReference(q):
	default:
		res.Translation = l.SearchTranslationFor(script, translation)
		hits, res.Total = l.SearchTranslation(res.Translation, q, limit)
		res.SearchedIn = res.Translation
		if t := l.Translation(res.Translation); t != nil {
			res.SearchedIn = t.Title()
		}
	}
	for _, p := range hits {
		res.Rows = append(res.Rows, Suggestion{Kind: KindHit, Passage: p, InTranslation: res.Translation})
	}
	return res
}

// Resolve is the passage a query means ("112" = the whole sūrah).
func (l *Library) Resolve(raw string) (Passage, bool) {
	q := l.ParseQuery(raw)
	if len(q.Surahs) == 0 {
		return Passage{}, false
	}
	s := q.Surahs[0]
	if q.From == 0 {
		return Passage{s.Number, 1, s.Ayahs}, true
	}
	return passage(s, q.From, q.To)
}

func passage(s Surah, from, to int) (Passage, bool) {
	if from < 1 || from > s.Ayahs {
		return Passage{}, false
	}
	end := to
	if end < from {
		end = from
	}
	if end > s.Ayahs {
		end = s.Ayahs
	}
	return Passage{s.Number, from, end}, true
}

func (l *Library) matchSurahs(spec string) []Surah {
	if spec == "" {
		return nil
	}
	if n, err := strconv.Atoi(spec); err == nil {
		if s := l.Surah(n); s != nil {
			return []Surah{*s}
		}
		return nil
	}
	key := Fold(spec)
	if key == "" {
		return nil
	}
	l.prepare()
	var exact, prefix []Surah
	for _, s := range l.Surahs {
		keys := l.keys[s.Number]
		found, starts := false, false
		for _, k := range keys {
			if k == key {
				found = true
			} else if strings.HasPrefix(k, key) {
				starts = true
			}
		}
		if found {
			exact = append(exact, s)
		} else if starts {
			prefix = append(prefix, s)
		}
	}
	return append(exact, prefix...)
}

func surahKeys(s Surah) []string {
	k := []string{Fold(s.Name), Fold(s.English), Fold(s.Arabic)}
	parts := strings.Split(s.Name, "-")
	for i := 1; i < len(parts); i++ {
		k = append(k, Fold(strings.Join(parts[i:], ""))) // "Baqara", "Imraan"
	}
	if strings.HasPrefix(s.English, "The ") {
		k = append(k, Fold(s.English[4:]))
	}
	var out []string
	for _, x := range k {
		if x != "" {
			out = append(out, x)
		}
	}
	return out
}

func hasArabic(s string) bool {
	for _, r := range s {
		if r >= 0x0600 && r <= 0x06FF {
			return true
		}
	}
	return false
}

func hasLatinLetter(s string) bool {
	for _, r := range s {
		if r < 0x80 && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

var foldPairs = [][2]string{{"aa", "a"}, {"ee", "i"}, {"ii", "i"}, {"oo", "u"}, {"uu", "u"}, {"ou", "u"}, {"y", "i"}, {"w", "u"}}

// Fold is the loose spelling of a sūrah name: "Al-Baqara", "al baqarah", "albaqara" → "albaqara";
// "Yaseen" = "yasin"; "Faatiha" = "fatiha"; Arabic: bare letters without ال.
func Fold(s string) string {
	if hasArabic(s) {
		var b strings.Builder
		for _, r := range s {
			if r >= 0x0621 && r <= 0x064A {
				switch r {
				case 'أ', 'إ', 'آ':
					r = 'ا'
				}
				b.WriteRune(r)
			}
		}
		return strings.TrimPrefix(b.String(), "ال")
	}
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	t := b.String()
	for _, p := range foldPairs {
		t = strings.ReplaceAll(t, p[0], p[1])
	}
	out := []byte{}
	for i := 0; i < len(t); i++ {
		if len(out) == 0 || out[len(out)-1] != t[i] { // doubled letters count once
			out = append(out, t[i])
		}
	}
	if len(out) > 3 && out[len(out)-1] == 'h' { // baqarah = baqara
		out = out[:len(out)-1]
	}
	return string(out)
}

// ---------- Search: the whole Qur'an, with or without harakat ----------

// Bare keeps letters only: harakat, small Qur'anic marks and tatweel removed; ٱ أ إ آ → ا, ى → ي,
// ة → ه. With daggerAlif the dagger alif (and Uthmani وٰ / ىٰ) is written ا, so رحمان and صلاة
// find ٱلرَّحْمَٰنِ and ٱلصَّلَوٰةَ.
func Bare(s string, daggerAlif bool) string {
	if daggerAlif {
		s = strings.NewReplacer("وٰ", "ا", "ىٰ", "ا").Replace(s)
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == 0x0670:
			if daggerAlif {
				b.WriteRune('ا')
			}
		case r >= 0x064B && r <= 0x065F, r >= 0x06D6 && r <= 0x06ED, r == 0x0640, r >= 0x08F0 && r <= 0x08FF, r == 0x200C, r == 0x200D:
		case r == 0x0671, r == 0x0623, r == 0x0625, r == 0x0622:
			b.WriteRune('ا')
		case r == 0x0649:
			b.WriteRune('ي')
		case r == 0x0629:
			b.WriteRune('ه')
		default:
			b.WriteRune(r)
		}
	}
	// ءا (the mushaf's ءَامَنُوا۟) = آ = ا; a doubled alif counts once.
	return strings.ReplaceAll(strings.ReplaceAll(b.String(), "ءا", "ا"), "اا", "ا")
}

// joined is a bare text without spaces: the mushaf joins some words (يَٰٓأَيُّهَا, يَٰبَنِىٓ),
// people type them apart (يا أيها).
func joined(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "اا", "ا")
}

// For Latin typed phonetically: letters an English typist mixes up count as one.
var looseMap = map[rune]rune{'ح': 'ه', 'خ': 'ه', 'ص': 'س', 'ث': 'س', 'ط': 'ت', 'ض': 'د', 'ظ': 'ز', 'ذ': 'ز',
	'ق': 'ك', 'ع': 'ا', 'ء': 'ا', 'ؤ': 'و', 'ئ': 'ي', 'غ': 'ج'}

func Loose(bare string) string {
	out := []rune{}
	for _, c := range bare {
		m := c
		if x, ok := looseMap[c]; ok {
			m = x
		}
		if !(m == 'ا' && len(out) > 0 && out[len(out)-1] == 'ا') {
			out = append(out, m)
		}
	}
	return string(out)
}

// light keeps harakat and evens out spelling variants (ۡ = ْ, ٱ = ا, no small marks).
func light(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == 0x06E1:
			b.WriteRune(0x0652)
		case r == 0x0671:
			b.WriteRune('ا')
		case r >= 0x06D6 && r <= 0x06E0, r >= 0x06E2 && r <= 0x06ED, r == 0x0640, r == 0x200C, r == 0x200D:
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func hasHarakat(s string) bool {
	for _, r := range s {
		if r >= 0x064B && r <= 0x0652 {
			return true
		}
	}
	return false
}

func runeCount(s string) int { return len([]rune(s)) }

// prepare builds the search texts and the sūrah-name keys once.
func (l *Library) prepare() {
	l.searchOnce.Do(func() {
		l.keys = [][]string{nil}
		for _, s := range l.Surahs {
			l.keys = append(l.keys, surahKeys(s))
		}
		t := l.lines("quran-uthmani")
		for i, s := range t {
			if su, a := l.position(i); a == 1 && su != 1 {
				s = dropBasmala(s)
			}
			b := [2]string{Bare(s, false), Bare(s, true)}
			lo := [2]string{Loose(b[0]), Loose(b[1])}
			l.bareArabic = append(l.bareArabic, b)
			l.looseArabic = append(l.looseArabic, lo)
			l.joinedArabic = append(l.joinedArabic, [2]string{joined(b[0]), joined(b[1])})
			l.joinedLoose = append(l.joinedLoose, joined(lo[1]))
			l.lightArabic = append(l.lightArabic, light(t[i]))
		}
	})
}

func (l *Library) position(i int) (int, int) {
	lo, hi := 1, 114
	for lo < hi {
		m := (lo + hi + 1) / 2
		if l.start[m] <= i {
			lo = m
		} else {
			hi = m - 1
		}
	}
	return lo, i - l.start[lo] + 1
}

// SearchArabic finds āyāt whose Arabic contains query (harakat optional: matches with your harakat
// come first). phonetic: the query came from Latin letters, so similar letters match too.
// It returns up to limit passages and the number of āyāt found.
func (l *Library) SearchArabic(query string, phonetic bool, limit int) ([]Passage, int) {
	q := Bare(query, false)
	if runeCount(q) < 2 {
		return nil, 0
	}
	l.prepare()
	qa := Bare(query, true)
	ql := Loose(qa)
	typed := ""
	if hasHarakat(query) {
		typed = light(query)
	}
	qj, qja, qlj := joined(q), joined(qa), joined(ql)
	byJoined, looseJoined := runeCount(qj) >= 4, runeCount(qlj) >= 4
	// Order: your harakat too → the same letters → the same letters with the mushaf's spacing →
	// (typed in Latin) similar letters, e.g. rahmaan → ٱلرَّحْمَٰنِ.
	// strings.Contains works on bytes, so it matches inside a cluster of marks too.
	var exact, letters, spacing, similar []int
	for i, t := range l.bareArabic {
		switch {
		case strings.Contains(t[0], q) || strings.Contains(t[1], q) || strings.Contains(t[1], qa):
			if typed != "" && strings.Contains(l.lightArabic[i], typed) {
				exact = append(exact, i)
			} else {
				letters = append(letters, i)
			}
		case byJoined && (strings.Contains(l.joinedArabic[i][0], qj) || strings.Contains(l.joinedArabic[i][1], qja)):
			spacing = append(spacing, i)
		case phonetic && runeCount(ql) >= 3 && (strings.Contains(l.looseArabic[i][0], ql) || strings.Contains(l.looseArabic[i][1], ql) ||
			looseJoined && strings.Contains(l.joinedLoose[i], qlj)):
			similar = append(similar, i)
		}
	}
	return l.passages(append(append(append(exact, letters...), spacing...), similar...), limit)
}

// passages turns line indexes into one-āyah passages (at most limit), and gives how many there were.
func (l *Library) passages(all []int, limit int) ([]Passage, int) {
	var hits []Passage
	for _, i := range all {
		if len(hits) >= limit {
			break
		}
		s, a := l.position(i)
		hits = append(hits, Passage{s, a, a})
	}
	return hits, len(all)
}

// SearchTranslation finds āyāt whose translation contains every word of query (case and accents
// ignored, Bangla written with or without a hasant at the end of a word).
func (l *Library) SearchTranslation(id, query string, limit int) ([]Passage, int) {
	words := strings.FieldsFunc(FoldText(query), func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) })
	if len(words) == 0 || runeCount(strings.Join(words, "")) < 2 {
		return nil, 0
	}
	l.mu.Lock()
	folded, ok := l.folded[id]
	l.mu.Unlock()
	if !ok {
		t := l.lines(id)
		if t == nil {
			return nil, 0
		}
		folded = make([]string, len(t))
		for i, s := range t {
			folded[i] = FoldText(s)
		}
		l.mu.Lock()
		if l.folded == nil {
			l.folded = map[string][]string{}
		}
		l.folded[id] = folded
		l.mu.Unlock()
	}
	var all []int
	for i, t := range folded {
		every := true
		for _, w := range words {
			if !strings.Contains(t, w) {
				every = false
				break
			}
		}
		if every {
			all = append(all, i)
		}
	}
	return l.passages(all, limit)
}

// FoldText evens out case, accents (ā = a, Ṣ = s), joiners and Bangla spellings: the nukta letters
// and the two-part vowel signs as their parts (য় = য + ়), a hasant not followed by a consonant
// dropped (আল্লাহ্‌র = আল্লাহর). Like the Mac's decomposed, lowercased text.
func FoldText(s string) string {
	in := []rune(strings.ToLower(s))
	var b strings.Builder
	for i, r := range in {
		switch {
		case r >= 0x0300 && r <= 0x036F, r == 0x200C, r == 0x200D: // accents, joiners
		case r == 0x09CD && (i+1 == len(in) || !bengaliConsonant(in[i+1])):
		case r == 0x09DC:
			b.WriteString("\u09A1\u09BC")
		case r == 0x09DD:
			b.WriteString("\u09A2\u09BC")
		case r == 0x09DF:
			b.WriteString("\u09AF\u09BC")
		case r == 0x09CB:
			b.WriteString("\u09C7\u09BE")
		case r == 0x09CC:
			b.WriteString("\u09C7\u09D7")
		default:
			if base, ok := accentBase(r); ok {
				b.WriteByte(base)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// ForKFGQPC: the KFGQPC Hafs font draws three Tanzil marks as a dotted circle, so they're swapped
// for the forms it has: small rounded zero (U+06DF → U+06E0), small low seen (U+06E3 → U+06DC),
// ishmām (U+06EB → U+06EC).
func ForKFGQPC(s string) string {
	return strings.NewReplacer("۟", "۠", "ۣ", "ۜ", "۫", "۬").Replace(s)
}

// Text is what gets inserted: "ٱللَّهُ لَآ إِلَٰهَ … (2:255)", then the translation on the next line;
// each āyah has its reference.
func (l *Library) Text(p Passage, mode Insert, translation string) string {
	trans := ""
	if mode != Arabic && l.Translation(translation) != nil {
		trans = translation
	}
	var out []string
	for a := p.From; a <= p.To; a++ {
		ref := fmt.Sprintf("(%d:%d)", p.Surah, a)
		ar := l.Arabic(p.Surah, a)
		tr := ""
		if trans != "" {
			tr = l.TranslationText(trans, p.Surah, a)
		}
		switch {
		case mode == Arabic:
			out = append(out, ar+" "+ref)
		case mode == TranslationOnly && tr != "":
			out = append(out, tr+" "+ref)
		case mode == TranslationOnly:
			out = append(out, ar+" "+ref)
		case tr != "":
			out = append(out, ar+"\n"+tr+" "+ref)
		default:
			out = append(out, ar+" "+ref)
		}
	}
	sep := " "
	if mode == Both {
		sep = "\n\n"
	}
	return strings.Join(out, sep)
}

func bengaliConsonant(r rune) bool {
	return r >= 0x0995 && r <= 0x09B9 || r == 0x09DC || r == 0x09DD || r == 0x09DF
}

// The plain letter of each accented Latin letter (À–ɏ, Ḁ–ỿ), "." where there is none; made from
// Unicode's decompositions, as the Mac's decomposedStringWithCanonicalMapping does.
const (
	accentsLatin = "aaaaaa.ceeeeiiii.nooooo..uuuuy..aaaaaa.ceeeeiiii.nooooo..uuuuy.y" +
		"aaaaaaccccccccdd..eeeeeeeeeegggggggghh..iiiiiiiii...jjkk.llllll." +
		"...nnnnnn...oooooo..rrrrrrsssssssstttt..uuuuuuuuuuuuwwyyyzzzzzz." +
		"................................oo.............uu..............." +
		".............aaiioouuuuuuuuuu.aaaa....ggkkoooo..j...gg..nnaa...." +
		"aaaaeeeeiiiioooorrrruuuusstt..hh......aaeeooooooooyy............" +
		"................"
	accentsExtra = "aabbbbbbccddddddddddeeeeeeeeeeffgghhhhhhhhhhiiiikkkkkkllllllllmm" +
		"mmmmnnnnnnnnoooooooopppprrrrrrrrssssssssssttttttttuuuuuuuuuuvvvv" +
		"wwwwwwwwwwxxxxyyzzzzzzhtwy......aaaaaaaaaaaaaaaaaaaaaaaaeeeeeeee" +
		"eeeeeeeeiiiioooooooooooooooooooooooouuuuuuuuuuuuuuyyyyyyyy......"
)

func accentBase(r rune) (byte, bool) {
	var c byte = '.'
	switch {
	case r >= 0x00C0 && r <= 0x024F:
		c = accentsLatin[r-0x00C0]
	case r >= 0x1E00 && r <= 0x1EFF:
		c = accentsExtra[r-0x1E00]
	}
	return c, c != '.'
}
