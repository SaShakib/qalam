package engine

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

// Runs the same golden cases as the Swift engine.
func TestGoldenCases(t *testing.T) {
	data, err := os.ReadFile("../../tests/cases.tsv")
	if err != nil {
		t.Fatal(err)
	}
	pass := 0
	for n, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "\t")
		o := DefaultOptions()
		o.Sukun = SukunFull // older cases were written for full sukūn
		if len(cols) >= 3 {
			for _, f := range strings.Split(cols[2], ",") {
				switch f {
				case "quran":
					o.Style = Quran
				case "astyped":
					o.Harakat = AsTyped
				case "none":
					o.Harakat = NoHarakat
				case "smallsukun":
					o.QuranSmallSukun = true
				case "nospelling":
					o.SpellingWords = false
				case "digits":
					o.ArabicDigits = true
				case "chromium":
					o.BlockAllahLigature = true
				case "smart":
					o.Sukun = SukunSmart
				case "sukunoff":
					o.Sukun = SukunOff
				case "fullsukun":
					o.Sukun = SukunFull
				}
			}
		}
		got := Text(cols[0], o)
		if canonical(got) != canonical(cols[1]) {
			t.Errorf("line %d: %q\n  expected %s  %s\n  got      %s  %s", n+1, cols[0], cols[1], hex(cols[1]), got, hex(got))
		} else {
			pass++
		}
	}
	t.Logf("%d passed", pass)
}

// The Go options must match the Swift ones (tests/candidates.tsv is written by `qalam --cands-tsv`).
func TestCandidatesMatchSwift(t *testing.T) {
	data, err := os.ReadFile("../../tests/candidates.tsv")
	if err != nil {
		t.Skip("no candidates.tsv")
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		cols := strings.SplitN(line, "\t", 2)
		if len(cols) != 2 {
			continue
		}
		got := []string{}
		for _, c := range Candidates(cols[0], DefaultOptions()) {
			got = append(got, canonical(c.Text))
		}
		want := []string{}
		for _, w := range strings.Split(cols[1], "|") {
			want = append(want, canonical(w))
		}
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("%s:\n  swift %s\n  go    %s", cols[0], strings.Join(want, " | "), strings.Join(got, " | "))
		}
	}
}

// canonical decomposes the precomposed hamza/madda letters and orders marks by
// combining class, so strings that look the same compare equal (as Swift does).
func canonical(s string) string {
	decomp := map[rune][]rune{
		0x0622: {0x0627, 0x0653}, 0x0623: {0x0627, 0x0654}, 0x0625: {0x0627, 0x0655},
		0x0624: {0x0648, 0x0654}, 0x0626: {0x064A, 0x0654},
	}
	ccc := func(r rune) int {
		switch {
		case r >= 0x064B && r <= 0x0652:
			return int(r-0x064B) + 27
		case r == 0x0670:
			return 35
		case r == 0x0655 || r == 0x06ED:
			return 220
		case r == 0x0653 || r == 0x0654 || (r >= 0x06D6 && r <= 0x06DC) || (r >= 0x06DF && r <= 0x06E4) || r == 0x06E7 || r == 0x06E8:
			return 230
		}
		return 0
	}
	rs := []rune{}
	for _, r := range s {
		if d, ok := decomp[r]; ok {
			rs = append(rs, d...)
		} else {
			rs = append(rs, r)
		}
	}
	for i := 0; i < len(rs); {
		if ccc(rs[i]) == 0 {
			i++
			continue
		}
		j := i
		for j < len(rs) && ccc(rs[j]) != 0 {
			j++
		}
		seg := rs[i:j]
		sort.SliceStable(seg, func(a, b int) bool { return ccc(seg[a]) < ccc(seg[b]) })
		i = j
	}
	return string(rs)
}

func hex(s string) string {
	parts := []string{}
	for _, r := range s {
		parts = append(parts, fmt.Sprintf("%04X", r))
	}
	return strings.Join(parts, " ")
}
