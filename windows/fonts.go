package main

import (
	"io/fs"
	"path"
	"sort"
	"strings"
)

// The Arabic fonts Qalam carries (assets/fonts): SIL Scheherazade New, and KFGQPC Hafs (the Madinah
// mushaf font) when it's in the repository's fonts/ folder at build time. Qalam's typing box uses
// them, and Qalam can install them for every app (Word, Notepad, browsers…).

type bundledFont struct {
	File string // e.g. ScheherazadeNew.ttf
	Name string // the family name, e.g. Scheherazade New
	Data []byte
}

// knownFontNames gives the family name of the fonts we expect; others use their file name.
var knownFontNames = map[string]string{
	"scheherazadenew": "Scheherazade New",
	"uthmanichafs":    "KFGQPC HAFS Uthmanic Script",
	"kfgqpchafs":      "KFGQPC HAFS Uthmanic Script",
}

func fontName(file string) string {
	base := strings.TrimSuffix(path.Base(file), path.Ext(file))
	key := strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(base))
	for prefix, name := range knownFontNames {
		if strings.HasPrefix(key, prefix) {
			return name
		}
	}
	return base
}

// isQuranFont: KFGQPC Hafs.
func isQuranFont(f bundledFont) bool {
	u := strings.ToUpper(f.File + f.Name)
	return strings.Contains(u, "KFGQPC") || strings.Contains(u, "HAFS")
}

// loadBundledFonts reads the .ttf/.otf files in dir of fsys, Scheherazade New first.
func loadBundledFonts(fsys fs.FS, dir string) []bundledFont {
	entries, _ := fs.ReadDir(fsys, dir)
	var out []bundledFont
	for _, e := range entries {
		ext := strings.ToLower(path.Ext(e.Name()))
		if e.IsDir() || (ext != ".ttf" && ext != ".otf") {
			continue
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, e.Name()))
		if err != nil || len(data) == 0 {
			continue
		}
		out = append(out, bundledFont{File: e.Name(), Name: fontName(e.Name()), Data: data})
	}
	sort.SliceStable(out, func(i, j int) bool { return !isQuranFont(out[i]) && isQuranFont(out[j]) })
	return out
}

// typingFont is the font for Qalam's typing box: Scheherazade New (SIL).
func typingFont(fonts []bundledFont) *bundledFont {
	for i := range fonts {
		if fonts[i].Name == "Scheherazade New" {
			return &fonts[i]
		}
	}
	if len(fonts) > 0 {
		return &fonts[0]
	}
	return nil
}
