package engine

type Style int

const (
	Everyday Style = iota
	Quran
)

type Harakat int

const (
	Full Harakat = iota
	AsTyped
	NoHarakat
)

// SukunMode: Smart = only at a stop inside a word (`o` always adds one);
// Full = on every vowel-less letter; Off = none at all.
type SukunMode int

const (
	SukunSmart SukunMode = iota
	SukunFull
	SukunOff
)

type Options struct {
	Style           Style     `json:"style"`
	Harakat         Harakat   `json:"harakat"`
	Sukun           SukunMode `json:"sukun"`
	QuranSmallSukun bool      `json:"quranSmallSukun"`
	SpellingWords   bool      `json:"spellingWords"`
	ArabicDigits    bool      `json:"arabicDigits"`
	// BlockAllahLigature inserts U+200D (ZWJ) between the lāms of لله so fonts don't draw
	// their built-in Allah ligature over our harakat (Chrome-based apps).
	BlockAllahLigature bool `json:"blockAllahLigature"`
}

func DefaultOptions() Options { return Options{SpellingWords: true} }
