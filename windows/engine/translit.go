package engine

import "strings"

type kind int

const (
	kConsonant kind = iota
	kHamza
	kAlif
	kMadda
	kWasl
	kTaMarbuta
	kMaqsura
	kArticleLam
	kTatweel
	kLiteral
	kRaw
)

type unit struct {
	kind        kind
	base        string
	vowel       V
	tanween     V
	shadda      bool
	sukun       bool
	marks       []string
	forcedSeat  string
	chunkStart  bool
	wordStart   bool
	isMadd      bool
	assimilated bool
}

func (u *unit) hasDagger() bool {
	for _, m := range u.marks {
		if m == dagger {
			return true
		}
	}
	return false
}

func (u *unit) blocksSukun() bool {
	for _, m := range u.marks {
		if sukunBlockers[m] {
			return true
		}
	}
	return false
}

func (u *unit) isBare() bool {
	return u.vowel == vNone && u.tanween == vNone && !u.shadda && !u.sukun && len(u.marks) == 0
}

func (u *unit) takesVowel() bool {
	return (u.kind == kConsonant || u.kind == kHamza || u.kind == kTaMarbuta) &&
		u.vowel == vNone && u.tanween == vNone && !u.sukun
}

// Word transliterates one word (what sits between spaces).
func Word(input string, o Options) string {
	if input == "" {
		return ""
	}
	if o.SpellingWords && !strings.Contains(input, "\\") {
		if s, ok := lookupSpelling(input, o); ok {
			return finish(s, o)
		}
	}
	units := buildUnits(tokenize(input, o))
	applyArticle(units)
	markMadd(units)
	units = seatHamzas(units)
	units = insertTanweenAlif(units)
	if o.Harakat == Full {
		autoSukun(units)
	}
	return finish(render(units, o), o)
}

// Text transliterates running text: words, spaces, punctuation, numbers.
func Text(s string, o Options) string {
	var c Composer
	var b strings.Builder
	for _, r := range s {
		var key Key
		switch {
		case r == ' ':
			key = Key{Kind: KeySpace}
		case r == '\n' || r == '\r':
			key = Key{Kind: KeyOther}
		default:
			key = Key{Kind: KeyChar, Char: r}
		}
		consumed, actions := c.Handle(key, o)
		for _, a := range actions {
			if a.Commit {
				b.WriteString(a.Text)
			}
		}
		if !consumed {
			b.WriteRune(r)
		}
	}
	for _, a := range c.Flush(o) {
		b.WriteString(a.Text)
	}
	return b.String()
}

// Accepts reports whether a key continues (or starts) a word being composed.
func Accepts(r rune, buffer string) bool {
	if isASCIILetter(r) {
		return true
	}
	switch r {
	case '\'', '\\', '_':
		return true
	case '`', '^', '=', '~', '-':
		return buffer != ""
	}
	if isDigit(r) {
		if i := strings.LastIndex(buffer, "\\"); i >= 0 {
			tail := buffer[i+1:]
			if tail == "h" {
				return true
			}
			for _, t := range tail {
				if !isDigit(t) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// Punctuation gives the Arabic for a key that is not part of a word ("" = leave it).
func Punctuation(r rune, o Options) string {
	switch r {
	case ',':
		return "،"
	case ';':
		return "؛"
	case '?':
		return "؟"
	}
	if o.ArabicDigits && isDigit(r) {
		return arabicDigits(string(r))
	}
	return ""
}

// MARK: building letters

func buildUnits(toks []tok) []unit {
	chunks := [][]tok{{}}
	for _, t := range toks {
		if t.k == tJoiner {
			chunks = append(chunks, []tok{})
		} else {
			chunks[len(chunks)-1] = append(chunks[len(chunks)-1], t)
		}
	}

	units := []unit{}
	for ci := 0; ci < len(chunks); ci++ {
		ch := chunks[ci]
		isLast := ci == len(chunks)-1
		start := len(units)

		if !isLast {
			if len(ch) == 2 && ch[0].k == tVowel && ch[0].v == vA && ch[1].k == tLetter && ch[1].s == lam {
				units = addArticle(units, start)
				continue
			}
			if len(ch) == 2 && ch[0].k == tVowel && ch[0].v == vA && ch[1].k == tLetter && sunLetters[ch[1].s] &&
				len(chunks[ci+1]) > 0 && chunks[ci+1][0].k == tLetter && chunks[ci+1][0].s == ch[1].s {
				units = addArticle(units, start)
				continue
			}
			if len(ch) == 3 && ch[0].k == tLetter && set("و", "ف", "ب", "ك", "ل")[ch[0].s] &&
				ch[1].k == tVowel && ch[2].k == tLetter && ch[2].s == lam {
				p := unit{kind: kConsonant, base: ch[0].s, vowel: ch[1].v}
				units = add(units, p, start)
				if ch[0].s != lam {
					units = add(units, unit{kind: kWasl, base: alif}, start)
				}
				units = add(units, unit{kind: kArticleLam, base: lam}, start)
				continue
			}
		}

		rest := ch
		if len(ch) >= 3 && ch[0].k == tVowel && ch[0].v == vA && ch[1].k == tLetter && ch[1].s == lam && articleFollows(ch, 2) {
			units = addArticle(units, start)
			rest = ch[2:]
		}
		units = process(rest, units, start)
	}
	return units
}

func articleFollows(ch []tok, i int) bool {
	if i >= len(ch) {
		return false
	}
	switch ch[i].k {
	case tLetter, tHamzaKey, tHamzaSeat:
		return true
	case tAlifKey:
		return i+1 < len(ch) && ch[i+1].isVowel()
	}
	return false
}

func addArticle(units []unit, start int) []unit {
	units = add(units, unit{kind: kWasl, base: alif}, start)
	return add(units, unit{kind: kArticleLam, base: lam}, start)
}

func add(units []unit, u unit, chunkStart int) []unit {
	u.chunkStart = len(units) == chunkStart
	u.wordStart = len(units) == 0
	return append(units, u)
}

func process(toks []tok, units []unit, start int) []unit {
	lastWasSep := false
	for k := 0; k < len(toks); k++ {
		t := toks[k]
		var next *tok
		if k+1 < len(toks) {
			next = &toks[k+1]
		}
		inChunk := len(units) > start
		li := len(units) - 1

		switch t.k {
		case tLetter:
			if !lastWasSep && inChunk && units[li].kind == kConsonant && units[li].base == t.s && units[li].isBare() {
				units[li].shadda = true
			} else {
				units = add(units, unit{kind: kConsonant, base: t.s}, start)
			}
		case tVowel:
			if !inChunk {
				units = add(units, unit{kind: kWasl, base: alif, vowel: t.v}, start)
			} else if units[li].takesVowel() {
				units[li].vowel = t.v
			} else {
				kd := kConsonant
				if t.v == vA {
					kd = kAlif
				}
				units = add(units, unit{kind: kd, base: t.v.maddLetter()}, start)
			}
		case tTanween:
			if inChunk && units[li].takesVowel() {
				units[li].tanween = t.v
			}
		case tAlifKey:
			if next != nil && next.isVowel() {
				units = add(units, unit{kind: kHamza, base: hamza}, start)
			} else if next != nil && next.k == tAlifKey && !(k+2 < len(toks) && toks[k+2].isVowel()) {
				units = add(units, unit{kind: kMadda, base: alifMadda}, start)
				k++
			} else {
				autoFatha(units, start)
				units = add(units, unit{kind: kAlif, base: alif}, start)
			}
		case tHamzaKey:
			units = add(units, unit{kind: kHamza, base: hamza}, start)
		case tHamzaSeat:
			units = add(units, unit{kind: kHamza, base: hamza, forcedSeat: t.s}, start)
		case tTaMarbuta:
			units = add(units, unit{kind: kTaMarbuta, base: taMarbuta}, start)
		case tMaqsura:
			autoFatha(units, start)
			units = add(units, unit{kind: kMaqsura, base: alifMaqsura}, start)
		case tShadda:
			if li >= 0 {
				units[li].shadda = true
			}
		case tSukun:
			if li >= 0 {
				units[li].sukun = true
			}
		case tDagger:
			if li >= 0 {
				units[li].marks = append(units[li].marks, dagger)
			}
		case tMaddah:
			if li >= 0 {
				units[li].marks = append(units[li].marks, maddah)
			}
		case tMark:
			if li >= 0 {
				units[li].marks = append(units[li].marks, t.s)
			} else {
				units = add(units, unit{kind: kLiteral, base: t.s}, start)
			}
		case tTatweel:
			units = add(units, unit{kind: kTatweel, base: tatweel}, start)
		case tSep:
			lastWasSep = true
			continue
		case tLiteral:
			units = add(units, unit{kind: kLiteral, base: t.s}, start)
		case tRaw:
			units = add(units, unit{kind: kRaw, base: t.s}, start)
		}
		lastWasSep = false
	}
	return units
}

func autoFatha(units []unit, start int) {
	if len(units) <= start {
		return
	}
	li := len(units) - 1
	u := &units[li]
	if li > 0 && ((u.base == waw && units[li-1].vowel == vU) || (u.base == ya && units[li-1].vowel == vI)) {
		return
	}
	if (u.kind == kConsonant || u.kind == kHamza) && u.vowel == vNone && u.tanween == vNone && !u.sukun {
		u.vowel = vA
	}
}

// MARK: rules

func applyArticle(units []unit) {
	for i := range units {
		if units[i].kind == kArticleLam && i+1 < len(units) && units[i+1].kind == kConsonant && sunLetters[units[i+1].base] {
			units[i].assimilated = true
			units[i+1].shadda = true
		}
	}
}

func markMadd(units []unit) {
	for i := 1; i < len(units); i++ {
		u, p := &units[i], units[i-1]
		if u.kind != kConsonant || u.vowel != vNone || u.tanween != vNone || u.shadda || u.sukun {
			continue
		}
		if (u.base == waw && p.vowel == vU) || (u.base == ya && p.vowel == vI) {
			u.isMadd = true
		}
	}
}

type prevV int

const (
	pA prevV = iota
	pI
	pU
	pSukun
	pLongA
	pLongI
	pLongU
)

func prevVowel(p unit) prevV {
	if p.kind == kAlif || p.kind == kMadda {
		return pLongA
	}
	if p.isMadd {
		if p.base == waw {
			return pLongU
		}
		return pLongI
	}
	if p.hasDagger() {
		return pLongA
	}
	v := p.vowel
	if v == vNone {
		v = p.tanween
	}
	switch v {
	case vA:
		return pA
	case vI:
		return pI
	case vU:
		return pU
	}
	return pSukun
}

func seatHamzas(units []unit) []unit {
	for i := 0; i < len(units); i++ {
		if units[i].kind == kHamza {
			units = seat(i, units)
		}
	}
	return units
}

func seat(i int, units []unit) []unit {
	u := units[i]
	if u.forcedSeat != "" {
		units[i].base = u.forcedSeat
		return units
	}
	own := u.vowel
	if own == vNone {
		own = u.tanween
	}
	var prev *unit
	if i > 0 {
		prev = &units[i-1]
	}
	initial := u.chunkStart || prev == nil || prev.kind == kArticleLam || prev.kind == kWasl
	final := i == len(units)-1
	var s string

	if initial {
		if own == vI {
			s = hamzaUnderAlif
		} else {
			s = hamzaOnAlif
		}
	} else {
		pv := prevVowel(*prev)
		if final {
			switch pv {
			case pI:
				s = hamzaOnYa
			case pU:
				s = hamzaOnWaw
			case pA:
				s = hamzaOnAlif
			default:
				s = hamza
			}
			if s == hamza && u.tanween == vA && (pv == pSukun || pv == pLongI) && !nonJoining[prev.base] {
				s = hamzaOnYa
			}
		} else {
			switch {
			case own == vI || pv == pI || pv == pLongI || (prev.base == ya && pv == pSukun):
				s = hamzaOnYa
			case pv == pLongA && own == vA:
				s = hamza
			case pv == pLongU:
				s = hamza
			case own == vU || pv == pU:
				s = hamzaOnWaw
			default:
				s = hamzaOnAlif
			}
		}
	}
	units[i].base = s

	if s == hamzaOnAlif && u.vowel == vA && i+1 < len(units) && units[i+1].kind == kAlif {
		units[i].kind = kMadda
		units[i].base = alifMadda
		units[i].vowel = vNone
		units = append(units[:i+1], units[i+2:]...)
	}
	return units
}

func insertTanweenAlif(units []unit) []unit {
	for i := 0; i < len(units); i++ {
		u := units[i]
		if u.tanween != vA {
			continue
		}
		var next *unit
		if i+1 < len(units) {
			next = &units[i+1]
		}
		afterLongA := i > 0 && (units[i-1].kind == kAlif || units[i-1].kind == kMadda)
		skip := u.kind == kTaMarbuta ||
			(next != nil && (next.kind == kMaqsura || next.kind == kAlif)) ||
			(u.kind == kHamza && u.base == hamza && afterLongA)
		if !skip {
			units = append(units[:i+1], append([]unit{{kind: kAlif, base: alif}}, units[i+1:]...)...)
		}
	}
	return units
}

func autoSukun(units []unit) {
	for i := range units {
		u := &units[i]
		if u.kind != kConsonant && u.kind != kHamza && u.kind != kArticleLam {
			continue
		}
		if u.vowel != vNone || u.tanween != vNone || u.shadda || u.sukun || u.isMadd || u.assimilated || u.blocksSukun() {
			continue
		}
		u.sukun = true
	}
}

// MARK: output

func render(units []unit, o Options) string {
	var b strings.Builder
	for _, u := range units {
		if u.kind == kWasl {
			if o.Style == Quran {
				b.WriteString(alifWasla)
			} else {
				b.WriteString(alif)
				if u.wordStart && u.vowel != vNone {
					b.WriteString(u.vowel.mark())
				}
			}
			b.WriteString(strings.Join(u.marks, ""))
			continue
		}
		b.WriteString(u.base)
		if u.shadda {
			b.WriteString(shadda)
		}
		switch {
		case u.vowel != vNone:
			b.WriteString(u.vowel.mark())
		case u.tanween != vNone:
			b.WriteString(u.tanween.tanweenMark())
		case u.sukun:
			b.WriteString(sukun)
		}
		b.WriteString(strings.Join(u.marks, ""))
	}
	return b.String()
}

func finish(s string, o Options) string {
	if o.BlockAllahLigature {
		s = BlockAllahLigature(s)
	}
	if o.Harakat == NoHarakat {
		return stripHarakat(s)
	}
	if o.Style == Quran && o.QuranSmallSukun {
		return strings.ReplaceAll(s, sukun, quranSukun)
	}
	return s
}

// BlockAllahLigature inserts U+034F between the two lāms of ل‌ل‌ه (marks ignored).
func BlockAllahLigature(s string) string {
	rs := []rune(s)
	isMark := func(r rune) bool {
		return (r >= 0x064B && r <= 0x065F) || r == 0x0670 || (r >= 0x06D6 && r <= 0x06ED) || r == 0x034F
	}
	bases := []int{}
	for i, r := range rs {
		if !isMark(r) {
			bases = append(bases, i)
		}
	}
	insert := map[int]bool{}
	for k := 0; k+2 < len(bases); k++ {
		if rs[bases[k]] == 0x0644 && rs[bases[k+1]] == 0x0644 && rs[bases[k+2]] == 0x0647 {
			insert[bases[k+1]] = true
		}
	}
	if len(insert) == 0 {
		return s
	}
	out := make([]rune, 0, len(rs)+len(insert))
	for i, r := range rs {
		if insert[i] {
			out = append(out, 0x034F)
		}
		out = append(out, r)
	}
	return string(out)
}
