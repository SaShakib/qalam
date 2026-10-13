# Building from source

## What you need

| To build | Needs |
|---|---|
| Mac apps | macOS 13+, Apple's Command Line Tools (`xcode-select --install`). Xcode isn't needed. |
| Windows exes | Go 1.22+ (cross-compiles from Mac or Linux) |
| Windows installer | Windows with [Inno Setup 6](https://jrsoftware.org/isinfo.php) (`iscc`), or let CI build it |
| Website demo | Go (compiles the engine to WebAssembly) |

## Commands

```bash
make test      # 175+ golden cases against both engines, plus the Windows app's tests
make install   # build and install on this Mac (universal: Apple Silicon and Intel)
make release   # Mac zip and .pkg, and the Windows exes, into dist/
make site      # the website with the live demo, into build/site
```

`make ARCHS=arm64 install` builds for Apple Silicon only, which is faster while developing.

Without a Mac you can still work on the engine, the Windows app and the website:

```bash
cd windows && go test ./...                         # engine golden cases + Windows app tests
GOOS=windows GOARCH=amd64 go build -o Qalam.exe .   # a Windows build
make site && python3 -m http.server -d build/site   # the website at http://localhost:8000
```

## Where things are

| Path | What |
|---|---|
| `engine/` | Transliteration rules in Swift (keys in `Tokenizer.swift`, word lists in `Content.swift`) |
| `mac/QalamInput/` | macOS input method (InputMethodKit) |
| `mac/QalamApp/` | macOS app: setup steps (`SetupView.swift`), letters, words, practice, settings, updater |
| `windows/` | Windows tray app in Go |
| `windows/engine/` | the same rules ported to Go (must pass the same tests) |
| `windows/assets/setup.html` | the Windows setup and settings page (served by the app on 127.0.0.1) |
| `windows/installer/qalam.iss` | the Windows installer (Inno Setup) |
| `windows/wasm/` | the engine for the website demo (WebAssembly) |
| `cli/` | `qalam` command: convert text, run tests, generate the Windows guide page |
| `tests/cases.tsv` | input ⇥ expected Arabic ⇥ options; both engines must pass |
| `website/` | the website (GitHub Pages) |
| `docs/` | these docs, shown on the website as the wiki |
| `Casks/qalam.rb` | Homebrew cask |
| `.github/workflows/` | CI, release builds and the website deploy |

## Changing the typing rules

1. Add the case to `tests/cases.tsv` first (input, expected Arabic, options such as `quran`, `smart`, `chromium`).
2. Change the Swift engine (`engine/`) **and** the Go engine (`windows/engine/`) the same way.
3. `make test` must pass for both.

The Go test `TestCandidatesMatchSwift` also checks that the options panel offers the same versions on both platforms.
