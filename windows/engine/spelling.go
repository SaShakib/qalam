package engine

import "strings"

type spellEntry struct {
	everyday, quran, plain string
	open                   bool
}

var spelling = func() map[string]spellEntry {
	t := map[string]spellEntry{}
	put := func(keys []string, e, q, p string, open bool) {
		for _, k := range keys {
			t[k] = spellEntry{e, q, p, open}
		}
	}
	put([]string{"allaah", "allah"}, "اللَّه", "ٱللَّه", "الله", true)
	put([]string{"wallaah", "wallah"}, "وَاللَّه", "وَٱللَّه", "والله", true)
	put([]string{"billaah", "billah"}, "بِاللَّه", "بِٱللَّه", "بالله", true)
	put([]string{"lillaah", "lillah"}, "لِلَّه", "لِلَّه", "لله", true)
	put([]string{"tallaah"}, "تَاللَّه", "تَٱللَّه", "تالله", true)
	put([]string{"allaahumma", "allahumma"}, "اللَّهُمَّ", "ٱللَّهُمَّ", "اللهم", false)
	put([]string{"alladhii"}, "الَّذِي", "ٱلَّذِي", "الذي", false)
	put([]string{"allatii"}, "الَّتِي", "ٱلَّتِي", "التي", false)
	put([]string{"alladhiina", "alladhiin"}, "الَّذِينَ", "ٱلَّذِينَ", "الذين", false)
	put([]string{"'ilaah"}, "إِلَٰه", "إِلَٰه", "إله", true)
	put([]string{"haadhaa"}, "هَٰذَا", "هَٰذَا", "هذا", false)
	put([]string{"haadhihi"}, "هَٰذِهِ", "هَٰذِهِ", "هذه", false)
	put([]string{"dhaalik"}, "ذَٰلِك", "ذَٰلِك", "ذلك", true)
	put([]string{"dhaalikum"}, "ذَٰلِكُمْ", "ذَٰلِكُمْ", "ذلكم", false)
	put([]string{"kadhaalik"}, "كَذَٰلِك", "كَذَٰلِك", "كذلك", true)
	put([]string{"haakadhaa"}, "هَٰكَذَا", "هَٰكَذَا", "هكذا", false)
	put([]string{"laakin"}, "لَٰكِن", "لَٰكِن", "لكن", true)
	put([]string{"laakinna"}, "لَٰكِنَّ", "لَٰكِنَّ", "لكن", false)
	put([]string{"haa'ulaa'"}, "هَٰؤُلَاء", "هَٰٓؤُلَآء", "هؤلاء", true)
	put([]string{"'ulaa'ik", "'uulaa'ik"}, "أُولَٰئِك", "أُو۟لَٰٓئِك", "أولئك", true)
	put([]string{"alraHmaan", "al-raHmaan", "ar-raHmaan"}, "الرَّحْمَٰن", "ٱلرَّحْمَٰن", "الرحمن", true)
	put([]string{"raHmaan"}, "رَحْمَٰن", "رَحْمَٰن", "رحمن", true)
	return t
}()

var suffixes = []struct {
	s       string
	v       V
	tanween bool
}{
	{"aN", vA, true}, {"iN", vI, true}, {"uN", vU, true},
	{"a", vA, false}, {"i", vI, false}, {"u", vU, false},
}

func lookupSpelling(input string, o Options) (string, bool) {
	n := normalize(input)
	if e, ok := spelling[n]; ok {
		return buildSpelling(e, vNone, false, o), true
	}
	for _, suf := range suffixes {
		if strings.HasSuffix(n, suf.s) {
			if e, ok := spelling[strings.TrimSuffix(n, suf.s)]; ok && e.open {
				return buildSpelling(e, suf.v, suf.tanween, o), true
			}
		}
	}
	return "", false
}

func buildSpelling(e spellEntry, v V, tanween bool, o Options) string {
	if o.Harakat == NoHarakat {
		if tanween && v == vA {
			return e.plain + alif
		}
		return e.plain
	}
	s := e.everyday
	if o.Style == Quran {
		s = e.quran
	}
	if v != vNone {
		if tanween {
			s += v.tanweenMark()
			if v == vA {
				s += alif
			}
		} else {
			s += v.mark()
		}
	} else if e.open && o.Harakat == Full && EffectiveSukun(o) == SukunFull {
		s += sukun
	}
	return s
}

func normalize(s string) string {
	c := []rune(s)
	var b strings.Builder
	in := func(r rune, set string) bool { return strings.ContainsRune(set, r) }
	for i, ch := range c {
		var prev, next rune
		if i > 0 {
			prev = c[i-1]
		}
		if i+1 < len(c) {
			next = c[i+1]
		}
		switch {
		case ch == 'A':
			if next != 0 && in(next, "aiu") {
				b.WriteRune('\'')
			} else if prev == 'a' {
				b.WriteRune('a')
			} else {
				b.WriteRune('A')
			}
		case ch == 'y' && prev == 'i' && !(next != 0 && in(next, "aiuy")):
			b.WriteRune('i')
		case ch == 'w' && prev == 'u' && !(next != 0 && in(next, "aiuw")):
			b.WriteRune('u')
		case ch == 'E':
			b.WriteRune('e')
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}
