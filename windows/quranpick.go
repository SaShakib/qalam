package main

import (
	"strconv"
	"unicode"

	"qalam/quran"
)

// quranPick is the Qur'an box's Choose tab: three dropdowns (Sūrah, From āyah, To āyah), the same as
// on the Mac. Keys: ↑↓ change the highlighted one, ←→/Tab move, typing 36 or yasin jumps to a sūrah,
// 2:255-257 fills all three. Kept apart from the window code so it can be tested anywhere.
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

// typeRune handles a typed character; true when it means "open Search" ("/").
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
