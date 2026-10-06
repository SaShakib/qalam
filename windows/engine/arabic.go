// Package engine is the Go port of the Swift engine in ../../engine.
// Both must pass ../../tests/cases.tsv.
package engine

const (
	fatha      = "َ"
	damma      = "ُ"
	kasra      = "ِ"
	fathatan   = "ً"
	dammatan   = "ٌ"
	kasratan   = "ٍ"
	shadda     = "ّ"
	sukun      = "ْ"
	quranSukun = "ۡ"
	dagger     = "ٰ"
	maddah     = "ٓ"

	alif           = "ا"
	alifWasla      = "ٱ"
	alifMadda      = "آ"
	hamza          = "ء"
	hamzaOnAlif    = "أ"
	hamzaUnderAlif = "إ"
	hamzaOnWaw     = "ؤ"
	hamzaOnYa      = "ئ"
	taMarbuta      = "ة"
	alifMaqsura    = "ى"
	tatweel        = "ـ"
	lam            = "ل"
	waw            = "و"
	ya             = "ي"

	silent    = "۟"
	smallWaw  = "ۥ"
	smallYa   = "ۦ"
	iqlabMeem = "ۢ"
	lowMeem   = "ۭ"
)

var sukunBlockers = set(dagger, maddah, silent, smallWaw, smallYa, iqlabMeem, lowMeem)
var sunLetters = set("ت", "ث", "د", "ذ", "ر", "ز", "س", "ش", "ص", "ض", "ط", "ظ", "ل", "ن")
var nonJoining = set("ا", "أ", "إ", "آ", "ٱ", "د", "ذ", "ر", "ز", "و", "ؤ", "ء", "ة", "ى")

func set(items ...string) map[string]bool {
	m := map[string]bool{}
	for _, s := range items {
		m[s] = true
	}
	return m
}

func stripHarakat(s string) string {
	out := []rune{}
	for _, r := range s {
		if (r >= 0x064B && r <= 0x0652) || r == 0x0670 || r == 0x0653 || r == 0x06E1 {
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

func arabicDigits(s string) string {
	out := []rune{}
	for _, r := range s {
		if r >= '0' && r <= '9' {
			r = 0x0660 + (r - '0')
		}
		out = append(out, r)
	}
	return string(out)
}

// V is a short vowel; vNone means none.
type V int

const (
	vNone V = iota
	vA
	vI
	vU
)

func (v V) mark() string {
	switch v {
	case vA:
		return fatha
	case vI:
		return kasra
	case vU:
		return damma
	}
	return ""
}

func (v V) tanweenMark() string {
	switch v {
	case vA:
		return fathatan
	case vI:
		return kasratan
	case vU:
		return dammatan
	}
	return ""
}

func (v V) maddLetter() string {
	switch v {
	case vA:
		return alif
	case vI:
		return ya
	}
	return waw
}
