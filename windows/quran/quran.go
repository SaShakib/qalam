// Package quran is the Windows port of engine/Quran.swift: the whole Qur'an (Tanzil Uthmani text,
// full harakat) and its bundled translations, for the "insert āyah" box (Ctrl+Alt+Q).
// Reference parsing, sūrah-name folding, Arabic search and the inserted text match the Mac version
// (quran_test.go checks the same cases). Searching inside translations is left out on Windows.
package quran

import (
	"bufio"
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
	KindHit                 // a search match in the Arabic
	KindProblem             // e.g. "Al-Baqara has 286 āyāt"
)

// Suggestion is one row of the dropdown under the query.
type Suggestion struct {
	Kind    Kind
	Surah   Surah   // KindSurah
	Passage Passage // KindPassage, KindHit
	Text    string  // KindProblem
}

type Library struct {
	Surahs  []Surah
	Catalog []Translation // every translation in translations.tsv (only bundled ones have files here)
	fsys    fs.FS
	start   []int // index of each sūrah's first āyah (1-based sūrah)

	mu    sync.Mutex
	texts map[string][]string

	searchOnce  sync.Once
	bareArabic  [][2]string
	looseArabic [][2]string
	lightArabic []string
	keys        [][]string
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
		return strings.TrimPrefix(t[i], l.basmala())
	}
	return t[i]
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

// basmala is āyah 1:1 and a space: Tanzil puts it before āyah 1 of the other sūrahs (except 9).
func (l *Library) basmala() string {
	if t := l.lines("quran-uthmani"); t != nil {
		return t[0] + " "
	}
	return ""
}

// Parse turns "sura|aya|text" lines in order into the 6236 texts; nil if it isn't a whole Qur'an.
func Parse(data []byte) []string {
	out := make([]string, 0, Total)
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line[0] < '0' || line[0] > '9' {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		out = append(out, strings.TrimSpace(parts[2]))
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

// Suggestions is the dropdown for what's typed so far: a reference ("2:255", "baqarah 255"), sūrah
// names, and Arabic search results. arabic is the query as Arabic when it was typed in Latin ("" if not).
func (l *Library) Suggestions(raw, arabic string, limit int) []Suggestion {
	trimmed := strings.Trim(raw, " \t")
	if trimmed == "" {
		return nil
	}
	q := l.ParseQuery(raw)
	if len(q.Surahs) == 0 || q.From == 0 {
		// Not a reference: sūrah names first, then āyāt that contain the words.
		var rows []Suggestion
		if q.From == 0 {
			for i, s := range q.Surahs {
				if i == 4 {
					break
				}
				rows = append(rows, Suggestion{Kind: KindSurah, Surah: s})
			}
		}
		ar := arabic
		typedArabic := hasArabic(trimmed)
		if typedArabic {
			ar = trimmed
		}
		if ar != "" {
			hits, _ := l.SearchArabic(ar, !typedArabic, limit)
			for _, p := range hits {
				rows = append(rows, Suggestion{Kind: KindHit, Passage: p})
			}
		}
		if len(rows) == 0 {
			rows = []Suggestion{{Kind: KindProblem, Text: "Nothing found for “" + trimmed + "”"}}
		}
		if len(rows) > limit {
			rows = rows[:limit]
		}
		return rows
	}
	// A clear name or number with an āyah: that passage (other sūrahs with the same start below it).
	var rows []Suggestion
	for i, s := range q.Surahs {
		if i == 4 {
			break
		}
		if p, ok := passage(s, q.From, q.To); ok {
			rows = append(rows, Suggestion{Kind: KindPassage, Passage: p})
		} else if i == 0 {
			rows = append(rows, Suggestion{Kind: KindProblem, Text: fmt.Sprintf("%s has %d āyāt", s.Name, s.Ayahs)})
		}
	}
	return rows
}

// Lookup is what the box shows for a query: the rows, and how many āyāt the search found in all
// (0 when the rows aren't search results). Latin that isn't a reference is also read as Arabic
// through Qalam (rahmaan → رحمان), with similar letters allowed.
func (l *Library) Lookup(raw string, limit int) ([]Suggestion, int) {
	phonetic := ""
	if hasLatinLetter(raw) {
		o := engine.DefaultOptions()
		o.Harakat = engine.NoHarakat
		phonetic = engine.Text(raw, o)
	}
	rows := l.Suggestions(raw, phonetic, limit)
	for _, r := range rows {
		if r.Kind == KindHit {
			ar := phonetic
			if hasArabic(raw) || ar == "" {
				ar = raw
			}
			_, total := l.SearchArabic(ar, phonetic != "", 0)
			return rows, total
		}
	}
	return rows, 0
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
	return strings.ReplaceAll(b.String(), "اا", "ا")
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
		basmala := l.basmala()
		for i, s := range t {
			if su, a := l.position(i); a == 1 && su != 1 {
				s = strings.TrimPrefix(s, basmala)
			}
			b := [2]string{Bare(s, false), Bare(s, true)}
			l.bareArabic = append(l.bareArabic, b)
			l.looseArabic = append(l.looseArabic, [2]string{Loose(b[0]), Loose(b[1])})
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
	// Order: your harakat too → the same letters → (typed in Latin) similar letters.
	var exact, letters, similar []int
	for i, t := range l.bareArabic {
		if strings.Contains(t[0], q) || strings.Contains(t[1], q) || strings.Contains(t[1], qa) {
			if typed != "" && strings.Contains(l.lightArabic[i], typed) {
				exact = append(exact, i)
			} else {
				letters = append(letters, i)
			}
		} else if phonetic && runeCount(ql) >= 3 && (strings.Contains(l.looseArabic[i][0], ql) || strings.Contains(l.looseArabic[i][1], ql)) {
			similar = append(similar, i)
		}
	}
	all := append(append(exact, letters...), similar...)
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
