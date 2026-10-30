package main

// While a word is being typed Qalam keeps it in the app itself, as the Mac's marked text does:
// every key replaces the end of what is showing. inlineEdit says how: delete this many characters
// (one Backspace each; a haraka counts as one), then type this text. Only the part that changed is
// touched, so typing a letter usually adds to the word and nothing is deleted.
func inlineEdit(old, new string) (int, string) {
	a, b := []rune(old), []rune(new)
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return len(a) - n, string(b[n:])
}
