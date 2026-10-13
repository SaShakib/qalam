//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"strings"
	"time"
)

//go:embed assets/setup.html
var setupHTML []byte

const wmRunOnUI = wmApp + 2

var (
	setupSrv *setupServer
	uiCalls  = make(chan func(), 16)
)

// runOnUI runs f on the window thread (where the keyboard hook reads the settings) and waits for it.
func runOnUI(f func()) {
	done := make(chan struct{})
	uiCalls <- func() { f(); close(done) }
	pPostMessageW.Call(trayHwnd, wmRunOnUI, 0, 0)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

// drainUICalls is called by the tray window for wmRunOnUI.
func drainUICalls() {
	for {
		select {
		case f := <-uiCalls:
			f()
		default:
			return
		}
	}
}

// openSetup opens the setup and settings page in the browser ("" = the steps, "settings" = all settings).
func openSetup(page string) {
	if setupSrv == nil {
		srv := newSetupServer(appHost{}, setupHTML, previewFont)
		if err := srv.listen(); err != nil {
			messageBox("Could not open the settings page: "+err.Error(), "Qalam", mbIconInfo)
			return
		}
		setupSrv = srv
	}
	shellOpen(setupSrv.url(page))
}

// appHost connects the page to the running app.
type appHost struct{}

func (appHost) State() setupState {
	var st setupState
	runOnUI(func() {
		st = stateFrom(settings)
		st.Recording = recording
	})
	st.Startup = startupEnabled()
	for _, f := range fonts {
		st.Fonts = append(st.Fonts, fontState{Name: f.Name, Quran: isQuranFont(f), Installed: fontInstalled(f)})
	}
	return st
}

func (h appHost) Apply(p setupPatch) setupState {
	runOnUI(func() {
		wasEnabled := settings.Enabled
		applyPatch(&settings, p)
		if settings.Enabled != wasEnabled {
			commitSelected()
			composer.Buffer = ""
			hidePreview()
		}
		saveSettings(settings)
		updateTray()
	})
	if p.Startup != nil {
		setStartup(*p.Startup)
	}
	return h.State()
}

func (appHost) Action(name string) error {
	var err error
	runOnUI(func() {
		switch {
		case strings.HasPrefix(name, "record:"):
			n := strings.TrimPrefix(name, "record:")
			for _, k := range hotkeyNames {
				if k.Name == n {
					startRecording(k.Name, k.Title)
					return
				}
			}
			err = fmt.Errorf("no shortcut called %q", n)
		case strings.HasPrefix(name, "unset:"):
			n := strings.TrimPrefix(name, "unset:")
			for _, k := range hotkeyNames {
				if k.Name == n {
					setHotkey(n, nil)
					updateTray()
					return
				}
			}
			err = fmt.Errorf("no shortcut called %q", n)
		case name == "cancelRecord":
			if recording != "" {
				stopRecording()
			}
		case name == "resetKeys":
			settings.Keys = nil
			saveSettings(settings)
			updateTray()
		case name == "guide":
			openGuide()
		case name == "installFonts":
			err = installFonts()
		default:
			err = fmt.Errorf("unknown action %q", name)
		}
	})
	return err
}
