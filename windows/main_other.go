//go:build !windows

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"qalam/engine"
	"qalam/quran"
)

// The keyboard itself only runs on Windows; on other systems this just explains that.
// `go run . --appdev` serves the app window's page here, for working on it.
func main() {
	if len(os.Args) == 2 && os.Args[1] == "--appdev" {
		lib, _ := quran.Load(quranFS())
		srv := newSetupServer(devHost{}, setupHTML, nil)
		srv.shell, srv.guide = appHTML, guidePage(guideHTML)
		srv.quran = func() *quran.Library { return lib }
		srv.quranFont, _ = os.ReadFile("assets/fonts/KFGQPC-HAFS-Uthmanic-Script.ttf")
		srv.font, _ = os.ReadFile("assets/fonts/ScheherazadeNew.ttf")
		ln, err := net.Listen("tcp", "127.0.0.1:8799")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		srv.addr = ln.Addr().String()
		fmt.Println(srv.url(""))
		http.Serve(ln, srv)
		return
	}
	fmt.Println("Qalam for Windows: build with GOOS=windows (see the Makefile's `windows` target).")
}

type devHost struct{}

var devSettings = Settings{Options: engine.DefaultOptions(), Enabled: true}

// Keep the non-Windows test/dev build aware of the same choices as the Windows executable.
// The Windows build replaces this with the embedded font files in fonts_windows.go.
var fonts = []bundledFont{
	{Name: "Scheherazade New"},
	{Name: "Amiri Quran"},
	{Name: "Noto Naskh Arabic"},
	{Name: "Noto Sans Arabic"},
	{Name: "KFGQPC HAFS Uthmanic Script"},
}

func (devHost) State() setupState {
	st := stateFrom(devSettings)
	st.Fonts = []fontState{{Name: "Scheherazade New", Installed: true}, {Name: "KFGQPC HAFS Uthmanic Script", Quran: true}}
	return st
}
func (h devHost) Apply(p setupPatch) setupState { applyPatch(&devSettings, p); return h.State() }
func (devHost) Action(string) error             { return nil }
