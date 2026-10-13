package main

import (
	"testing"
	"testing/fstest"
)

func TestBundledFonts(t *testing.T) {
	fsys := fstest.MapFS{
		"f/UthmanicHafs1Ver18.ttf": {Data: []byte("hafs")},
		"f/ScheherazadeNew.ttf":    {Data: []byte("sil")},
		"f/OFL.txt":                {Data: []byte("licence")},
		"f/Empty.ttf":              {Data: nil},
	}
	fonts := loadBundledFonts(fsys, "f")
	if len(fonts) != 2 {
		t.Fatalf("fonts = %+v", fonts)
	}
	if fonts[0].Name != "Scheherazade New" || fonts[1].Name != "KFGQPC HAFS Uthmanic Script" || !isQuranFont(fonts[1]) {
		t.Errorf("names/order: %q, %q", fonts[0].Name, fonts[1].Name)
	}
	if f := typingFont(fonts); f == nil || f.Name != "Scheherazade New" {
		t.Errorf("typing font = %+v", f)
	}
	if fontName("KFGQPC-Hafs.ttf") != "KFGQPC HAFS Uthmanic Script" || fontName("Other Font.otf") != "Other Font" {
		t.Errorf("fontName")
	}
}
