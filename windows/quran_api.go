package main

import (
	"net/http"
	"strconv"
	"strings"

	"qalam/quran"
)

// The app window's Qur'an page reads the text through these (the page itself is assets/app.html).

type apiSurah struct {
	N       int    `json:"n"`
	Ayahs   int    `json:"ayahs"`
	Arabic  string `json:"arabic"`
	Name    string `json:"name"`
	English string `json:"english"`
	Type    string `json:"type"`
}

type apiTranslation struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type apiAyah struct {
	S  int    `json:"s"`
	A  int    `json:"a"`
	Ar string `json:"ar"`
	Tr string `json:"tr,omitempty"`
}

type apiRow struct {
	Kind string `json:"kind"` // surah, ayah, problem
	S    int    `json:"s,omitempty"`
	From int    `json:"from,omitempty"`
	To   int    `json:"to,omitempty"`
	Name string `json:"name,omitempty"`
	Ar   string `json:"ar,omitempty"`
	Tr   string `json:"tr,omitempty"`
	Text string `json:"text,omitempty"`
}

func (s *setupServer) quranAPI(w http.ResponseWriter, r *http.Request) {
	if s.quran == nil || s.quran() == nil {
		http.Error(w, "the Qur'an data did not load", http.StatusInternalServerError)
		return
	}
	lib := s.quran()
	q := r.URL.Query()
	trans := q.Get("t")
	if lib.Translation(trans) == nil {
		trans = "en.sahih"
	}
	switch strings.TrimPrefix(r.URL.Path, "/api/quran/") {
	case "index":
		out := struct {
			Surahs       []apiSurah       `json:"surahs"`
			Translations []apiTranslation `json:"translations"`
		}{}
		for _, x := range lib.Surahs {
			out.Surahs = append(out.Surahs, apiSurah{x.Number, x.Ayahs, x.Arabic, x.Name, x.English, x.Type})
		}
		for _, t := range lib.Available() {
			out.Translations = append(out.Translations, apiTranslation{t.ID, t.Title()})
		}
		writeJSON(w, out)
	case "surah":
		n, _ := strconv.Atoi(q.Get("n"))
		x := lib.Surah(n)
		if x == nil {
			http.NotFound(w, r)
			return
		}
		out := make([]apiAyah, 0, x.Ayahs)
		for a := 1; a <= x.Ayahs; a++ {
			out = append(out, apiAyah{n, a, quran.ForKFGQPC(lib.Arabic(n, a)), lib.TranslationText(trans, n, a)})
		}
		writeJSON(w, out)
	case "search":
		in := quran.InArabic
		if q.Get("in") == "translation" {
			in = quran.InTranslation
		}
		res := lib.Search(q.Get("q"), in, trans, 60)
		out := struct {
			Rows       []apiRow `json:"rows"`
			Total      int      `json:"total"`
			SearchedIn string   `json:"searchedIn"`
			Arabic     string   `json:"arabic,omitempty"`
		}{Rows: []apiRow{}, Total: res.Total, SearchedIn: res.SearchedIn, Arabic: res.Arabic}
		for _, row := range res.Rows {
			switch row.Kind {
			case quran.KindSurah:
				out.Rows = append(out.Rows, apiRow{Kind: "surah", S: row.Surah.Number, Name: row.Surah.Name + " · " + row.Surah.English, Ar: row.Surah.Arabic})
			case quran.KindPassage, quran.KindHit:
				p := row.Passage
				name := ""
				if x := lib.Surah(p.Surah); x != nil {
					name = x.Name
				}
				out.Rows = append(out.Rows, apiRow{Kind: "ayah", S: p.Surah, From: p.From, To: p.To, Name: name,
					Ar: quran.ForKFGQPC(lib.Arabic(p.Surah, p.From)), Tr: lib.TranslationText(trans, p.Surah, p.From)})
			default:
				out.Rows = append(out.Rows, apiRow{Kind: "problem", Text: row.Text})
			}
		}
		writeJSON(w, out)
	default:
		http.NotFound(w, r)
	}
}

// guidePage is the Letters / Word list / Practice guide inside the app window: the sidebar replaces
// its header, and the part after # picks the page.
func guidePage(html []byte) []byte {
	s := string(html)
	s = strings.Replace(s, "</head>", "<style>header{display:none}main{padding-top:8px}</style></head>", 1)
	s = strings.ReplaceAll(s, "(or click the ق tray icon; change it in the tray menu → Shortcuts)", "(change it in Settings → Shortcuts)")
	if i := strings.LastIndex(s, "render();"); i >= 0 {
		s = s[:i] + "starterDone=true;page=location.hash.slice(1)||'practice';addEventListener('hashchange',()=>{page=location.hash.slice(1)||'practice';render()});" + s[i:]
	}
	return []byte(s)
}
