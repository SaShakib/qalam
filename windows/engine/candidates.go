package engine

import "strings"

// Candidate is one version of the word offered in the options panel (Go port of Candidates.swift).
type Candidate struct {
	Text  string
	Kind  string // typed, seat, plain, dagger, sukun
	Label string
}

func isMarkRune(r rune) bool {
	return (r >= 0x064B && r <= 0x065F) || r == 0x0670 || (r >= 0x06D6 && r <= 0x06ED) || r == 0x034F || r == 0x200D
}

// IsSkeleton: the word was typed with no vowels at all (it comes out as bare consonants).
func IsSkeleton(typed string, o Options) bool {
	if o.SpellingWords {
		if _, ok := lookupSpelling(typed, o); ok {
			return false
		}
	}
	return !hasTypedMarks(tokenize(typed, o))
}

// Candidates returns up to 4 versions of a word; the first is always what was typed.
func Candidates(typed string, o Options) []Candidate {
	main := Word(typed, o)
	out := []Candidate{{Text: main, Kind: "typed", Label: "as typed"}}
	add := func(text, kind, label string) {
		if text == "" {
			return
		}
		for _, c := range out {
			if c.Text == text {
				return
			}
		}
		out = append(out, Candidate{text, kind, label})
	}
	skeleton := IsSkeleton(typed, o)
	seats := seatVariants(main, typed, skeleton)
	if len(seats) > 0 {
		add(seats[0][0], "seat", seats[0][1])
	}
	if strings.HasSuffix(typed, "A") && !skeleton {
		add(Word(typed+"a", o), "seat", "أَ hamza + fatḥa")
	}
	for _, v := range seats[min(1, len(seats)):] {
		add(v[0], "seat", v[1])
	}
	if o.Harakat != NoHarakat {
		add(stripHarakat(main), "plain", "no harakat")
	}
	if d, ok := daggerVariant(main); ok {
		add(d, "dagger", "small alif ـٰ")
	}
	if o.Harakat == Full && !skeleton {
		alt := o
		if EffectiveSukun(o) == SukunFull {
			alt.Sukun = SukunSmart
		} else {
			alt.Sukun = SukunFull
		}
		if o.Style == Quran && alt.Sukun == SukunSmart {
			alt.Sukun = SukunOff
		}
		label := "less sukūn"
		if alt.Sukun == SukunFull {
			label = "with sukūn"
		}
		add(Word(typed, alt), "sukun", label)
	}
	if len(out) > 4 {
		out = out[:4]
	}
	return out
}

var seatNames = map[rune]string{
	0x0623: "أ alif + hamza", 0x0625: "إ hamza below", 0x0621: "ء hamza alone",
	0x0626: "ئ hamza on yāʾ", 0x0624: "ؤ hamza on wāw",
}

func seatVariants(s, typed string, skeleton bool) [][2]string {
	rs := []rune(s)
	bases := []int{}
	for i, r := range rs {
		if !isMarkRune(r) {
			bases = append(bases, i)
		}
	}
	if len(bases) == 0 {
		return nil
	}
	replacing := func(i int, r rune) string {
		c := append([]rune{}, rs...)
		c[i] = r
		return string(c)
	}
	out := [][2]string{}
	if last := bases[len(bases)-1]; rs[last] == 0x0627 && len(bases) > 1 {
		out = append(out, [2]string{replacing(last, 0x0623), seatNames[0x0623]})
	}
	article := strings.HasPrefix(typed, "al") || strings.HasPrefix(typed, "Al")
	if first := bases[0]; rs[first] == 0x0627 && !article {
		var mark rune
		if first+1 < len(rs) {
			mark = rs[first+1]
		}
		seats := []rune{0x0623, 0x0625}
		if mark == 0x0650 {
			seats = []rune{0x0625}
		} else if mark == 0x064E || mark == 0x064F {
			seats = []rune{0x0623}
		}
		for _, v := range seats {
			out = append(out, [2]string{replacing(first, v), seatNames[v]})
		}
	}
	if skeleton {
		hamzas := []rune{0x0623, 0x0621, 0x0626, 0x0624, 0x0625}
		for _, b := range bases {
			cur := rs[b]
			if strings.ContainsRune(string(hamzas), cur) {
				for _, v := range hamzas {
					if v != cur {
						out = append(out, [2]string{replacing(b, v), seatNames[v]})
					}
				}
				break
			}
		}
	}
	return out
}

// daggerVariant: لَّه (fatḥa + shadda on the second lām) → لّٰه.
func daggerVariant(s string) (string, bool) {
	rs := []rune(s)
	var b strings.Builder
	changed := false
	for i := 0; i < len(rs); i++ {
		b.WriteRune(rs[i])
		if rs[i] == 0x0644 {
			j := i + 1
			hasF, hasS := false, false
			for j < len(rs) && (rs[j] == 0x064E || rs[j] == 0x0651) {
				if rs[j] == 0x064E {
					hasF = true
				} else {
					hasS = true
				}
				j++
			}
			if hasF && hasS && j < len(rs) && rs[j] == 0x0647 {
				b.WriteRune(0x0651)
				b.WriteRune(0x0670)
				changed = true
				i = j - 1
			}
		}
	}
	return b.String(), changed
}
