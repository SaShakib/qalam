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

type Options struct {
	Style           Style   `json:"style"`
	Harakat         Harakat `json:"harakat"`
	QuranSmallSukun bool    `json:"quranSmallSukun"`
	SpellingWords   bool    `json:"spellingWords"`
	ArabicDigits    bool    `json:"arabicDigits"`
	// BlockAllahLigature inserts U+034F between the lāms of لله so fonts don't draw
	// their built-in Allah ligature over our harakat (Chrome-based apps).
	BlockAllahLigature bool `json:"blockAllahLigature"`
}

func DefaultOptions() Options { return Options{SpellingWords: true} }
