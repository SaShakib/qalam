package main

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"qalam/quran"
)

// quranPick is the Qur'an box's three dropdowns (Sūrah, From āyah, To āyah), the same as on the Mac.
// When Tab has moved to them: ↑↓ change the highlighted one, ←→/Tab move, typing 36 or yasin jumps to
// a sūrah, 2:255-257 fills all three. Kept apart from the window code so it can be tested anywhere.
type quranPick struct {
	Surah, From, To int
	Field           int    // 0 Sūrah, 1 From, 2 To
	Typed           string // what's typed into the highlighted dropdown
}

const (
	fieldSurah = iota
	fieldFrom
	fieldTo
)

// newQuranPick opens on the last choice ("2:255-257"), or Al-Fātiḥa 1.
func newQuranPick(lib *quran.Library, last string) quranPick {
	p := quranPick{Surah: 1, From: 1, To: 1}
	if lib != nil && last != "" {
		if r, ok := lib.Resolve(last); ok {
			p.Surah, p.From, p.To = r.Surah, r.From, r.To
		}
	}
	return p
}

func (p quranPick) passage() quran.Passage {
	to := p.To
	if to < p.From {
		to = p.From
	}
	return quran.Passage{Surah: p.Surah, From: p.From, To: to}
}

func (p *quranPick) ayahs(lib *quran.Library) int {
	if s := lib.Surah(p.Surah); s != nil {
		return s.Ayahs
	}
	return 1
}

func (p *quranPick) setSurah(n int) {
	if n == p.Surah {
		return
	}
	p.Surah, p.From, p.To = n, 1, 1
}

func (p *quranPick) setFrom(n int) {
	p.From = n
	if p.To < n {
		p.To = n
	}
}

// step: ↑ (-1) or ↓ (+1) on the highlighted dropdown.
func (p *quranPick) step(lib *quran.Library, d int) {
	p.Typed = ""
	n := p.ayahs(lib)
	switch p.Field {
	case fieldSurah:
		p.setSurah(clamp(p.Surah+d, 1, 114))
	case fieldFrom:
		p.setFrom(clamp(p.From+d, 1, n))
	case fieldTo:
		p.To = clamp(p.To+d, p.From, n)
	}
}

// move: → / Tab (+1) or ← / Shift+Tab (-1) to the next dropdown.
func (p *quranPick) move(d int) {
	p.Field = (p.Field + d + 3) % 3
	p.Typed = ""
}

// typeRune handles a typed character; true when it means "back to the search field" ("/").
func (p *quranPick) typeRune(lib *quran.Library, r rune) bool {
	switch {
	case r == '/':
		return true
	case r == ':' || r == '-' || r == ' ': // 2:255-257 fills all three
		if p.Field != fieldTo {
			p.Field++
			p.Typed = ""
		}
		return false
	case p.Field != fieldSurah && !unicode.IsDigit(r):
		return false
	}
	p.Typed += string(r)
	p.apply(lib)
	return false
}

func (p *quranPick) backspace(lib *quran.Library) {
	if r := []rune(p.Typed); len(r) > 0 {
		p.Typed = string(r[:len(r)-1])
		p.apply(lib)
	}
}

// apply what's typed: a number, or (Sūrah) a name like "yasin". A number that is too big starts
// again from its last digit (Sūrah: 1, 2, then 5 → 125 is too big → 5).
func (p *quranPick) apply(lib *quran.Library) {
	if p.Typed == "" {
		return
	}
	n := p.ayahs(lib)
	num, err := strconv.Atoi(p.Typed)
	switch p.Field {
	case fieldSurah:
		if err != nil {
			if q := lib.ParseQuery(p.Typed); len(q.Surahs) > 0 {
				p.setSurah(q.Surahs[0].Number)
			}
			return
		}
		if num < 1 || num > 114 {
			p.Typed = p.Typed[len(p.Typed)-1:]
			num, _ = strconv.Atoi(p.Typed)
		}
		if num >= 1 && num <= 114 {
			p.setSurah(num)
		}
	case fieldFrom, fieldTo:
		if err != nil {
			return
		}
		if num < 1 || num > n {
			p.Typed = p.Typed[len(p.Typed)-1:]
			num, _ = strconv.Atoi(p.Typed)
		}
		num = clamp(num, 1, n)
		if p.Field == fieldFrom {
			p.setFrom(num)
		} else {
			p.To = clamp(num, p.From, n)
		}
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// quranBox is the whole box: the search field on top, the results under it, the dropdowns, and
// where the keys go (the field, or the dropdowns after Tab).
type quranBox struct {
	Pick   quranPick
	OnPick bool           // Tab moved the keys to the dropdowns (Pick.Field is the one)
	Keys   string         // what's typed in the field: Latin keys, or pasted text
	In     quran.SearchIn // where Latin letters are searched
	Res    quran.Results
	Sel    int // the chosen result, -1 = none
	Top    int // the first result row showing (the list scrolls)
}

const quranVisibleRows = 5 // results shown before the list scrolls
const quranResultLimit = 200

func newQuranBox(lib *quran.Library, last string) quranBox {
	return quranBox{Pick: newQuranPick(lib, last), Sel: -1}
}

func hasLatinLetter(s string) bool {
	for _, r := range s {
		if r < 0x80 && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// shown is what the field shows: in Arabic mode Latin letters as Qalam writes them (qul → قُلْ).
func (b *quranBox) shown() string {
	if b.In == quran.InArabic && hasLatinLetter(b.Keys) && quran.Script(b.Keys) != "bangla" {
		return quran.PhoneticArabic(b.Keys)
	}
	return b.Keys
}

// shownIsArabic: the field's text is drawn right to left in the Qur'an font.
func (b *quranBox) shownIsArabic() bool { return quran.Script(b.shown()) == "arabic" }

func (b *quranBox) query() string { return strings.Trim(b.Keys, " \t") }

// typeText adds typed or pasted text to the field (the keys go back to it).
func (b *quranBox) typeText(s string) {
	b.OnPick = false
	b.Keys += s
}

// backspace removes the last key typed (or the last letter pasted).
func (b *quranBox) backspace() bool {
	if b.Keys == "" {
		return false
	}
	_, size := utf8.DecodeLastRuneInString(b.Keys)
	b.Keys = b.Keys[:len(b.Keys)-size]
	return true
}

// clear empties the field and the results; the dropdowns keep what they show.
func (b *quranBox) clear() {
	b.Keys, b.Res, b.Sel, b.Top = "", quran.Results{}, -1, 0
}

// setResults shows a search's rows and chooses the first.
func (b *quranBox) setResults(res quran.Results) {
	b.Res, b.Sel, b.Top = res, -1, 0
	if len(res.Rows) > 0 {
		b.choose(0)
	}
}

// move: ↑ (-1) or ↓ (+1) through the results, round at the ends.
func (b *quranBox) move(d int) {
	if n := len(b.Res.Rows); n > 0 {
		b.choose(((b.Sel+d)%n + n) % n)
	}
}

// choose a result: it fills Sūrah / From / To, and the list scrolls to show it.
func (b *quranBox) choose(i int) {
	if i < 0 || i >= len(b.Res.Rows) {
		return
	}
	b.Sel = i
	if i < b.Top {
		b.Top = i
	} else if i >= b.Top+quranVisibleRows {
		b.Top = i - quranVisibleRows + 1
	}
	r := b.Res.Rows[i]
	switch r.Kind {
	case quran.KindSurah:
		b.Pick.Surah, b.Pick.From, b.Pick.To = r.Surah.Number, 1, 1
	case quran.KindPassage:
		b.Pick.Surah, b.Pick.From, b.Pick.To = r.Passage.Surah, r.Passage.From, r.Passage.To
	case quran.KindHit:
		b.Pick.Surah, b.Pick.From, b.Pick.To = r.Passage.Surah, r.Passage.From, r.Passage.From
	}
	b.Pick.Typed = ""
}

// tab: Tab (+1) / Shift+Tab (-1) goes field → Sūrah → From → To → field.
func (b *quranBox) tab(d int) {
	pos := 0
	if b.OnPick {
		pos = 1 + b.Pick.Field
	}
	pos = (pos + d + 4) % 4
	b.OnPick = pos > 0
	if b.OnPick {
		b.Pick.Field = pos - 1
	}
	b.Pick.Typed = ""
}
