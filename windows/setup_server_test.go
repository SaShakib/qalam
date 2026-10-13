package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qalam/engine"
)

// memHost is the app as the settings page sees it, kept in memory.
type memHost struct {
	s       Settings
	actions []string
}

func (h *memHost) State() setupState { return stateFrom(h.s) }
func (h *memHost) Apply(p setupPatch) setupState {
	applyPatch(&h.s, p)
	return h.State()
}
func (h *memHost) Action(name string) error { h.actions = append(h.actions, name); return nil }

func newTestServer() (*setupServer, *memHost) {
	h := &memHost{s: Settings{Options: engine.DefaultOptions(), Enabled: true}}
	srv := newSetupServer(h, []byte("<title>Qalam Setup</title>"), []byte("font"))
	srv.addr = "127.0.0.1:5555"
	return srv, h
}

func call(srv *setupServer, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "http://127.0.0.1:5555"+path, strings.NewReader(body))
	req.Host = "127.0.0.1:5555"
	if token != "" {
		req.Header.Set("X-Qalam-Token", token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func TestSetupPageNeedsToken(t *testing.T) {
	srv, _ := newTestServer()
	if w := call(srv, "GET", "/?t="+srv.token, "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Qalam") {
		t.Fatalf("page with token: %d %q", w.Code, w.Body.String())
	}
	if w := call(srv, "GET", "/?t=nope", "", ""); w.Code != 403 {
		t.Errorf("page with a wrong token: %d", w.Code)
	}
	if w := call(srv, "GET", "/api/state", "", ""); w.Code != 403 {
		t.Errorf("API without token: %d", w.Code)
	}
	if w := call(srv, "GET", "/api/state?t="+srv.token, "", ""); w.Code != 403 {
		t.Errorf("API must take the token from the header only: %d", w.Code)
	}
	if w := call(srv, "GET", "/arabic.ttf?t="+srv.token, "", ""); w.Code != 200 || w.Body.String() != "font" {
		t.Errorf("font: %d", w.Code)
	}
}

func TestSetupRefusesOtherHosts(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest("GET", "http://evil.example/api/state", nil)
	req.Header.Set("X-Qalam-Token", srv.token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Errorf("a DNS-rebinding host got %d", w.Code)
	}
}

func TestSetupChangesSettings(t *testing.T) {
	srv, h := newTestServer()
	w := call(srv, "POST", "/api/settings", srv.token, `{"style":"quran","sukun":"off","showOptions":false,"autoUpdate":false,"setupDone":true}`)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if h.s.Style != engine.Quran || h.s.Sukun != engine.SukunOff || !h.s.NoOptions || !h.s.NoAutoUpdate || !h.s.SetupDone {
		t.Errorf("settings not applied: %+v", h.s)
	}
	var st setupState
	json.Unmarshal(w.Body.Bytes(), &st)
	if st.Style != "quran" || st.Sukun != "off" || st.ShowOptions || st.Keys["toggle"] != "Ctrl+Alt+A" {
		t.Errorf("state: %+v", st)
	}
	// Unknown values are ignored, not saved.
	call(srv, "POST", "/api/settings", srv.token, `{"style":"klingon"}`)
	if h.s.Style != engine.Quran {
		t.Errorf("unknown style changed the setting")
	}
	// A form post from another page (no JSON content type) is refused.
	req := httptest.NewRequest("POST", "http://127.0.0.1:5555/api/settings", strings.NewReader("style=everyday"))
	req.Host = "127.0.0.1:5555"
	req.Header.Set("X-Qalam-Token", srv.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("form post: %d", rec.Code)
	}
}

func TestSetupPreviewAndActions(t *testing.T) {
	srv, h := newTestServer()
	w := call(srv, "POST", "/api/preview", srv.token, `{"words":["maktab","kataba"],"patch":{"sukun":"full"}}`)
	var out []string
	json.Unmarshal(w.Body.Bytes(), &out)
	if len(out) != 2 || out[0] != "مَكْتَبْ" || out[1] != "كَتَبَ" {
		t.Errorf("preview = %q", out)
	}
	if h.s.Sukun != engine.SukunSmart {
		t.Errorf("preview changed the saved sukūn setting")
	}
	if w := call(srv, "POST", "/api/action/record:sukun", srv.token, `{}`); w.Code != 200 || h.actions[0] != "record:sukun" {
		t.Errorf("action: %d %v", w.Code, h.actions)
	}
}
