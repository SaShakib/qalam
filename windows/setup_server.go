package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"qalam/engine"
)

// The setup and settings page (assets/setup.html) is served by Qalam itself on 127.0.0.1 and opened
// in the browser: on first run it walks through the settings step by step, later it is the settings
// window ("Setup & settings…" in the tray menu). Every request must carry a random token, and the
// Host header must be the loopback address, so other web pages can't read or change the settings.

// setupState is what the page shows. Names match the page's form fields.
type setupState struct {
	Version         string            `json:"version"`
	Enabled         bool              `json:"enabled"`
	Style           string            `json:"style"`   // everyday, quran
	Harakat         string            `json:"harakat"` // full, astyped, none
	Sukun           string            `json:"sukun"`   // smart, full, off
	QuranSmallSukun bool              `json:"quranSmallSukun"`
	SpellingWords   bool              `json:"spellingWords"`
	ArabicDigits    bool              `json:"arabicDigits"`
	ShowOptions     bool              `json:"showOptions"`
	Startup         bool              `json:"startup"`
	AutoUpdate      bool              `json:"autoUpdate"`
	SetupDone       bool              `json:"setupDone"`
	Keys            map[string]string `json:"keys"`      // name → "Ctrl+Alt+A" or "none"
	Recording       string            `json:"recording"` // shortcut being recorded, or ""
	Fonts           []fontState       `json:"fonts"`     // Qalam's fonts, and whether each is installed for all apps
}

type fontState struct {
	Name      string `json:"name"`
	Quran     bool   `json:"quran"`
	Installed bool   `json:"installed"`
}

// setupPatch changes some settings; nil fields stay as they are.
type setupPatch struct {
	Enabled         *bool   `json:"enabled"`
	Style           *string `json:"style"`
	Harakat         *string `json:"harakat"`
	Sukun           *string `json:"sukun"`
	QuranSmallSukun *bool   `json:"quranSmallSukun"`
	SpellingWords   *bool   `json:"spellingWords"`
	ArabicDigits    *bool   `json:"arabicDigits"`
	ShowOptions     *bool   `json:"showOptions"`
	Startup         *bool   `json:"startup"`
	AutoUpdate      *bool   `json:"autoUpdate"`
	SetupDone       *bool   `json:"setupDone"`
}

// setupHost is the running app, as the page sees it.
type setupHost interface {
	State() setupState
	Apply(p setupPatch) setupState
	// Action runs a button: "record:<name>", "unset:<name>", "resetKeys", "cancelRecord", "guide", "installFonts".
	Action(name string) error
}

var styleNames = map[string]engine.Style{"everyday": engine.Everyday, "quran": engine.Quran}
var harakatNames = map[string]engine.Harakat{"full": engine.Full, "astyped": engine.AsTyped, "none": engine.NoHarakat}
var sukunNames = map[string]engine.SukunMode{"smart": engine.SukunSmart, "full": engine.SukunFull, "off": engine.SukunOff}

func nameOf[T comparable](m map[string]T, v T) string {
	for k, x := range m {
		if x == v {
			return k
		}
	}
	return ""
}

// stateFrom describes settings for the page (the app-only parts are filled in by the host).
func stateFrom(s Settings) setupState {
	st := setupState{
		Version:         version,
		Enabled:         s.Enabled,
		Style:           nameOf(styleNames, s.Style),
		Harakat:         nameOf(harakatNames, s.Harakat),
		Sukun:           nameOf(sukunNames, s.Sukun),
		QuranSmallSukun: s.QuranSmallSukun,
		SpellingWords:   s.SpellingWords,
		ArabicDigits:    s.ArabicDigits,
		ShowOptions:     !s.NoOptions,
		AutoUpdate:      !s.NoAutoUpdate,
		SetupDone:       s.SetupDone,
		Keys:            map[string]string{},
	}
	for _, n := range hotkeyNames {
		st.Keys[n.Name] = s.hotkeyText(n.Name)
	}
	return st
}

// applyPatch changes s; unknown names are ignored. Startup is the host's job (it is in the registry).
func applyPatch(s *Settings, p setupPatch) {
	if p.Enabled != nil {
		s.Enabled = *p.Enabled
	}
	if p.Style != nil {
		if v, ok := styleNames[*p.Style]; ok {
			s.Style = v
		}
	}
	if p.Harakat != nil {
		if v, ok := harakatNames[*p.Harakat]; ok {
			s.Harakat = v
		}
	}
	if p.Sukun != nil {
		if v, ok := sukunNames[*p.Sukun]; ok {
			s.Sukun = v
		}
	}
	if p.QuranSmallSukun != nil {
		s.QuranSmallSukun = *p.QuranSmallSukun
	}
	if p.SpellingWords != nil {
		s.SpellingWords = *p.SpellingWords
	}
	if p.ArabicDigits != nil {
		s.ArabicDigits = *p.ArabicDigits
	}
	if p.ShowOptions != nil {
		s.NoOptions = !*p.ShowOptions
	}
	if p.AutoUpdate != nil {
		s.NoAutoUpdate = !*p.AutoUpdate
	}
	if p.SetupDone != nil {
		s.SetupDone = *p.SetupDone
	}
}

// previewRequest asks how words come out with the current settings changed by Patch
// (so the page can show each choice before it is picked).
type previewRequest struct {
	Words []string   `json:"words"`
	Patch setupPatch `json:"patch"`
}

type setupServer struct {
	host  setupHost
	page  []byte
	font  []byte // the Arabic font the page shows its examples in
	token string
	addr  string // 127.0.0.1:port once listening
}

func newSetupServer(host setupHost, page, font []byte) *setupServer {
	b := make([]byte, 16)
	rand.Read(b)
	return &setupServer{host: host, page: page, font: font, token: hex.EncodeToString(b)}
}

// listen starts serving on a free loopback port.
func (s *setupServer) listen() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.addr = ln.Addr().String()
	go http.Serve(ln, s)
	return nil
}

// url opens the page at a step ("" = the start of the setup, "settings" = all settings).
func (s *setupServer) url(page string) string {
	u := "http://" + s.addr + "/?t=" + s.token
	if page != "" {
		u += "#" + page
	}
	return u
}

func (s *setupServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only this computer, by its loopback address (a DNS name pointing here is refused),
	// and only with the token Qalam put in the link it opened.
	if s.addr != "" && r.Host != s.addr {
		http.Error(w, "wrong host", http.StatusForbidden)
		return
	}
	token := r.Header.Get("X-Qalam-Token")
	if r.URL.Path == "/" || r.URL.Path == "/arabic.ttf" {
		token = r.URL.Query().Get("t") // links and the page's font can't send a header
	}
	if token != s.token {
		http.Error(w, "Open this page from the Qalam tray menu: Setup & settings…", http.StatusForbidden)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")

	switch {
	case r.URL.Path == "/" && r.Method == http.MethodGet:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(s.page)
	case r.URL.Path == "/arabic.ttf" && r.Method == http.MethodGet:
		w.Header().Set("Content-Type", "font/ttf")
		w.Write(s.font)
	case r.URL.Path == "/api/state" && r.Method == http.MethodGet:
		writeJSON(w, s.host.State())
	case r.URL.Path == "/api/settings" && r.Method == http.MethodPost:
		var p setupPatch
		if !readJSON(w, r, &p) {
			return
		}
		writeJSON(w, s.host.Apply(p))
	case r.URL.Path == "/api/preview" && r.Method == http.MethodPost:
		var p previewRequest
		if !readJSON(w, r, &p) {
			return
		}
		writeJSON(w, s.preview(p))
	case strings.HasPrefix(r.URL.Path, "/api/action/") && r.Method == http.MethodPost:
		if err := s.host.Action(strings.TrimPrefix(r.URL.Path, "/api/action/")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, s.host.State())
	default:
		http.NotFound(w, r)
	}
}

// preview writes the words with the settings the page shows, changed by the patch.
func (s *setupServer) preview(p previewRequest) []string {
	st := s.host.State()
	var set Settings
	set.Options = engine.DefaultOptions()
	applyPatch(&set, setupPatch{Style: &st.Style, Harakat: &st.Harakat, Sukun: &st.Sukun,
		QuranSmallSukun: &st.QuranSmallSukun, SpellingWords: &st.SpellingWords, ArabicDigits: &st.ArabicDigits})
	applyPatch(&set, p.Patch)
	out := make([]string, 0, len(p.Words))
	for i, word := range p.Words {
		if i >= 50 || len(word) > 200 {
			break
		}
		out = append(out, engine.Text(word, set.Options))
	}
	return out
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "send JSON", http.StatusUnsupportedMediaType)
		return false
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
