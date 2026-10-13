//go:build js && wasm

// The Qalam engine for the website's live demo (make wasm). It exposes window.qalam:
//
//	qalam.text(s, opts)            whole text → Arabic
//	qalam.candidates(word, opts)   [{text, kind, label}] (the options panel)
//	qalam.accepts(ch, buffer)      whether ch continues the word being typed
//	qalam.punctuation(ch, opts)    the Arabic punctuation for ch, or ""
//
// opts: {style: "everyday"|"quran", harakat: "full"|"astyped"|"none", sukun: "smart"|"full"|"off",
// spelling: bool, digits: bool}
package main

import (
	"syscall/js"

	"qalam/engine"
)

func options(v js.Value) engine.Options {
	o := engine.DefaultOptions()
	if v.Type() != js.TypeObject {
		return o
	}
	switch v.Get("style").String() {
	case "quran":
		o.Style = engine.Quran
	}
	switch v.Get("harakat").String() {
	case "astyped":
		o.Harakat = engine.AsTyped
	case "none":
		o.Harakat = engine.NoHarakat
	}
	switch v.Get("sukun").String() {
	case "full":
		o.Sukun = engine.SukunFull
	case "off":
		o.Sukun = engine.SukunOff
	}
	if s := v.Get("spelling"); s.Type() == js.TypeBoolean {
		o.SpellingWords = s.Bool()
	}
	if d := v.Get("digits"); d.Type() == js.TypeBoolean {
		o.ArabicDigits = d.Bool()
	}
	// Browsers are Chrome-based, or draw like them: keep the harakat on اللَّه.
	o.BlockAllahLigature = true
	return o
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

func main() {
	api := js.Global().Get("Object").New()
	api.Set("text", js.FuncOf(func(_ js.Value, a []js.Value) any {
		return engine.Text(a[0].String(), options(a[1]))
	}))
	api.Set("candidates", js.FuncOf(func(_ js.Value, a []js.Value) any {
		out := []any{}
		for _, c := range engine.Candidates(a[0].String(), options(a[1])) {
			out = append(out, map[string]any{"text": c.Text, "kind": c.Kind, "label": c.Label})
		}
		return out
	}))
	api.Set("accepts", js.FuncOf(func(_ js.Value, a []js.Value) any {
		return engine.Accepts(firstRune(a[0].String()), a[1].String())
	}))
	api.Set("punctuation", js.FuncOf(func(_ js.Value, a []js.Value) any {
		return engine.Punctuation(firstRune(a[0].String()), options(a[1]))
	}))
	js.Global().Set("qalam", api)
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("qalam-ready"))
	select {}
}
