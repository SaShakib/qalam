package main

import (
	"embed"
	"io/fs"
)

// The Qur'an (Tanzil Uthmani) and the bundled translations for the Ctrl+Alt+Q box. assets/quran is
// made from the repository's quran/ folder by `make guide` (tools/packquran: the texts gzipped).
//
//go:embed assets/quran
var quranFiles embed.FS

func quranFS() fs.FS {
	sub, _ := fs.Sub(quranFiles, "assets/quran")
	return sub
}
