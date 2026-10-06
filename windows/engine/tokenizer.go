package engine

type tokKind int

const (
	tLetter tokKind = iota
	tVowel
	tTanween
	tAlifKey
	tHamzaKey
	tHamzaSeat
	tTaMarbuta
	tMaqsura
	tShadda
	tSukun
	tDagger
	tMaddah
	tTatweel
	tSep
	tJoiner
	tMark
	tLiteral
	tRaw
)

type tok struct {
	k tokKind
	s string
	v V
}

func (t tok) isVowel() bool { return t.k == tVowel || t.k == tTanween }

type codeKind int

const (
	cMark codeKind = iota
	cStandalone
	cSeat
	cPhrase
)

type code struct {
	k        codeKind
	s, quran string
}

var codes = map[string]code{
	"0":   {k: cMark, s: silent},
	"w":   {k: cMark, s: smallWaw},
	"y":   {k: cMark, s: smallYa},
	"mi":  {k: cMark, s: iqlabMeem},
	"ml":  {k: cMark, s: lowMeem},
	"mm":  {k: cMark, s: "ۘ"},
	"la":  {k: cMark, s: "ۙ"},
	"j":   {k: cMark, s: "ۚ"},
	"sl":  {k: cMark, s: "ۖ"},
	"ql":  {k: cMark, s: "ۗ"},
	"mu":  {k: cMark, s: "ۛ"},
	"sk":  {k: cMark, s: "ۜ"},
	"sj":  {k: cStandalone, s: "۩"},
	"hz":  {k: cStandalone, s: "۞"},
	"saw": {k: cStandalone, s: "ﷺ"},
	"hA":  {k: cSeat, s: hamzaOnAlif},
	"hI":  {k: cSeat, s: hamzaUnderAlif},
	"hW":  {k: cSeat, s: hamzaOnWaw},
	"hY":  {k: cSeat, s: hamzaOnYa},
	"h0":  {k: cSeat, s: hamza},
	"bism": {k: cPhrase, s: "بِسْمِ اللَّهِ الرَّحْمَٰنِ الرَّحِيمِ",
		quran: "بِسْمِ ٱللَّهِ ٱلرَّحْمَٰنِ ٱلرَّحِيمِ"},
	"swt": {k: cPhrase, s: "سُبْحَانَهُ وَتَعَالَى", quran: "سُبْحَانَهُ وَتَعَالَىٰ"},
	"as":  {k: cPhrase, s: "عَلَيْهِ السَّلَامُ", quran: "عَلَيْهِ ٱلسَّلَامُ"},
	"ra":  {k: cPhrase, s: "رَضِيَ اللَّهُ عَنْهُ", quran: "رَضِيَ ٱللَّهُ عَنْهُ"},
}

const codeMaxLen = 4

var digraphs = map[string]string{
	"th": "ث", "dh": "ذ", "kh": "خ", "sh": "ش", "gh": "غ", "zh": "ز",
}

var singles = map[rune]string{
	'b': "ب", 't': "ت", 'j': "ج", 'H': "ح", 'd': "د", 'r': "ر", 'z': "ز", 's': "س",
	'S': "ص", 'D': "ض", 'T': "ط", 'Z': "ظ", 'e': "ع", 'E': "ع", 'g': "غ", 'f': "ف",
	'q': "ق", 'k': "ك", 'l': "ل", 'm': "م", 'n': "ن", 'N': "ن", 'h': "ه", 'w': "و", 'y': "ي",
	'B': "ب", 'J': "ج", 'R': "ر", 'G': "غ", 'F': "ف", 'Q': "ق", 'K': "ك", 'L': "ل", 'M': "م", 'W': "و",
}

func vowelOf(r rune) V {
	switch r {
	case 'a':
		return vA
	case 'i', 'I':
		return vI
	case 'u', 'U':
		return vU
	}
	return vNone
}

func isASCIILetter(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }
func isDigit(r rune) bool       { return r >= '0' && r <= '9' }

func tokenize(input string, o Options) []tok {
	c := []rune(input)
	out := []tok{}
	for i := 0; i < len(c); {
		ch := c[i]

		if ch == '\\' {
			j := i + 1
			for j < len(c) && (isASCIILetter(c[j]) || isDigit(c[j])) {
				j++
			}
			run := c[i+1 : j]
			if len(run) > 0 && isDigit(run[0]) && run[0] != '0' {
				k := 0
				for k < len(run) && isDigit(run[k]) {
					k++
				}
				out = append(out, tok{k: tLiteral, s: "﴿" + arabicDigits(string(run[:k])) + "﴾"})
				i += 1 + k
				continue
			}
			n := len(run)
			if n > codeMaxLen {
				n = codeMaxLen
			}
			found := false
			for ; n > 0; n-- {
				if cd, ok := codes[string(run[:n])]; ok {
					switch cd.k {
					case cMark:
						out = append(out, tok{k: tMark, s: cd.s})
					case cStandalone:
						out = append(out, tok{k: tLiteral, s: cd.s})
					case cSeat:
						out = append(out, tok{k: tHamzaSeat, s: cd.s})
					case cPhrase:
						if o.Style == Quran {
							out = append(out, tok{k: tLiteral, s: cd.quran})
						} else {
							out = append(out, tok{k: tLiteral, s: cd.s})
						}
					}
					i += 1 + n
					found = true
					break
				}
			}
			if !found {
				out = append(out, tok{k: tRaw, s: "\\"})
				i++
			}
			continue
		}

		if i+1 < len(c) {
			pair := string(c[i : i+2])
			if pair == "t'" {
				out = append(out, tok{k: tTaMarbuta})
				i += 2
				continue
			}
			if l, ok := digraphs[pair]; ok {
				out = append(out, tok{k: tLetter, s: l})
				i += 2
				continue
			}
		}

		if v := vowelOf(ch); v != vNone {
			if i+1 < len(c) && c[i+1] == 'N' {
				out = append(out, tok{k: tTanween, v: v})
				i += 2
			} else {
				out = append(out, tok{k: tVowel, v: v})
				i++
			}
			continue
		}

		switch ch {
		case 'A':
			out = append(out, tok{k: tAlifKey})
		case '\'':
			out = append(out, tok{k: tHamzaKey})
		case 'Y':
			out = append(out, tok{k: tMaqsura})
		case '~':
			out = append(out, tok{k: tShadda})
		case 'o', 'O':
			out = append(out, tok{k: tSukun})
		case '^':
			out = append(out, tok{k: tDagger})
		case '=':
			out = append(out, tok{k: tMaddah})
		case '_':
			out = append(out, tok{k: tTatweel})
		case '`':
			out = append(out, tok{k: tSep})
		case '-':
			out = append(out, tok{k: tJoiner})
		default:
			if l, ok := singles[ch]; ok {
				out = append(out, tok{k: tLetter, s: l})
			} else {
				out = append(out, tok{k: tRaw, s: string(ch)})
			}
		}
		i++
	}
	return out
}
