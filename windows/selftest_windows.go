//go:build windows

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"qalam/engine"
	"qalam/quran"
)

// selfTest checks a built Qalam.exe without touching the keyboard or settings:
//
//	Qalam.exe --selftest result.txt
//
// It writes "ok" (or what failed) to the file and exits 0 on success. CI runs it on x64 and ARM64.
func selfTest(args []string) int {
	out := "qalam-selftest.txt"
	for i, a := range args {
		if a == "--selftest" && i+1 < len(args) {
			out = args[i+1]
		}
	}
	var fails []string
	check := func(ok bool, what string) {
		if !ok {
			fails = append(fails, what)
		}
	}

	o := engine.DefaultOptions()
	o.Sukun = engine.SukunFull
	check(engine.Text("kataba", o) == "كَتَبَ", "engine: kataba")
	check(engine.Text("kitaAbuN", o) == "كِتَابٌ", "engine: kitaAbuN")
	o.BlockAllahLigature = true
	check(strings.ContainsRune(engine.Text("lillaAhi", o), 0x200D), "engine: Allah-ligature guard in Chrome apps")

	check(len(guideHTML) > 1000 && len(setupHTML) > 1000, "embedded guide and setup pages")
	check(len(iconOn) > 0 && len(iconOff) > 0 && len(previewFont) > 100000, "embedded icons and font")
	check(typingFont(fonts) != nil && typingFont(fonts).Name == "Scheherazade New", "the typing box uses Scheherazade New (SIL)")
	if lib := quranLibrary(); lib == nil {
		check(false, "embedded Qur'an data")
	} else {
		p, ok := lib.Resolve("baqarah 255")
		check(ok && p.Reference() == "2:255" && strings.HasSuffix(lib.Text(p, quran.Arabic, ""), "(2:255)"), "Qur'an: baqarah 255")
		_, total := lib.SearchArabic("الرحمن", false, 0)
		check(total == 50 && len(lib.Available()) == 3, "Qur'an: search and the bundled translations")
	}
	check(!isChromiumExe(exePath()), "Qalam.exe itself is not a Chrome app")
	check(isChromiumExe(`C:\Program Files\WindowsApps\x\msedgewebview2.exe`), "WebView2 apps are Chrome apps")
	check(strings.HasPrefix(assetName(), "Qalam-Windows-"), "update file name")

	// The settings page answers only with its token.
	srv := newSetupServer(fakeHost{}, setupHTML, previewFont)
	if err := srv.listen(); err != nil {
		check(false, "settings page: "+err.Error())
	} else {
		resp, err := http.Get(srv.url(""))
		check(err == nil && resp.StatusCode == 200, "settings page opens with its token")
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			check(strings.Contains(string(body), "Qalam"), "settings page content")
		}
		resp, err = http.Get("http://" + srv.addr + "/?t=wrong")
		check(err == nil && resp.StatusCode == 403, "settings page refuses a wrong token")
		if err == nil {
			resp.Body.Close()
		}
	}

	result := "ok " + version + "\n"
	code := 0
	if len(fails) > 0 {
		result = "FAILED:\n" + strings.Join(fails, "\n") + "\n"
		code = 1
	}
	if err := os.WriteFile(out, []byte(result), 0o644); err != nil {
		fmt.Println(err)
		return 2
	}
	return code
}

type fakeHost struct{}

func (fakeHost) State() setupState             { return stateFrom(Settings{Options: engine.DefaultOptions()}) }
func (h fakeHost) Apply(setupPatch) setupState { return h.State() }
func (fakeHost) Action(string) error           { return nil }
