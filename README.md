# Qalam (قَلَم): Phonetic Arabic Keyboard

Type Arabic the way you spell it: each letter, then its haraka. Sukūn and shadda appear by themselves. Hamza seats, the article, tanwīn and Qur'anic spelling are handled for you.

```
kitaAbuN → كِتَابٌ     saAala → سَأَلَ     eilmuN → عِلْمٌ     madrasat'uN → مَدْرَسَةٌ     alshamsu → الشَّمْسُ
```

Works in every app, on **macOS** (a native input method, like [Lekho](https://github.com/ARahim3/Lekho)) and **Windows** (a tray app). There is an **Everyday** style and a **Qur'an (Uthmani)** style with ٱ, dagger alif, maddah, waqf signs and āyah numbers.

**Website, download and live demo: https://sashakib.github.io/qalam/** · **[Docs](docs/README.md)** · **[Questions](https://sashakib.github.io/qalam/questions.html)** · The full key chart is in **[LAYOUT.md](LAYOUT.md)**.

## Install on macOS

macOS 13 (Ventura) or later, Apple Silicon or Intel (universal from 1.3.0). Step-by-step guide: [docs/install-mac.md](docs/install-mac.md).

**Homebrew**

```bash
brew tap sashakib/qalam https://github.com/SaShakib/qalam
```

```bash
brew trust sashakib/qalam
```

```bash
brew install --cask qalam
```

Homebrew 7 asks you to trust a tap before installing from it (the second command). Update later with `brew upgrade --cask qalam`.

**One command (no Homebrew)**

```bash
curl -fsSL https://raw.githubusercontent.com/SaShakib/qalam/main/install.sh | bash
```

**Installer package:** download `Qalam-mac.pkg` from [Releases](https://github.com/SaShakib/qalam/releases/latest) and double-click it. The package isn't notarized yet. If macOS blocks it, go to System Settings → Privacy & Security and click **Open Anyway** (once; in-app updates don't ask again).

The Qalam app then opens a **short guided setup**: it adds the keyboard to your input menu, then lets you pick the style, harakat, sukūn, typing options, shortcuts, fonts and updates, and offers an optional **20-word practice**. Switch keyboards with **🌐 Globe** or **Control-Space**.

## Install on Windows

Download **`Qalam-Setup-x64.exe`** (or `-arm64` for ARM laptops) from [Releases](https://github.com/SaShakib/qalam/releases/latest) and run it. Step-by-step guide: [docs/install-windows.md](docs/install-windows.md).

- It installs for your user (no admin rights): Start menu, start with Windows (optional), and an uninstall entry in Settings → Apps. Silent install: `/VERYSILENT`.
- Qalam then opens a **setup page** in your browser (served by Qalam on your PC) with the same steps as on the Mac. Open it again from the tray: right-click **ق** → **Setup & settings…**
- **Ctrl + Alt + A** (or a click on the ق tray icon) switches Arabic on and off.
- The single-file `Qalam-Windows-x64.exe` / `-arm64.exe` still works without an installer.
- The downloads aren't code-signed yet. If SmartScreen warns, click **More info → Run anyway** (once; updates don't ask again).

## Updates

Qalam checks GitHub for a new version when it starts and once a day.

- **macOS:** Settings → Updates has **Check for updates now**, **Install update**, and switches for automatic checking and installing. How it installs depends on how you installed Qalam: an install.sh copy updates itself in place, Homebrew runs `brew upgrade --cask qalam`, and the .pkg downloads the new installer for you. If you had the keyboard turned on, it stays on after an update.
- **Windows:** the tray menu has **Check for updates…** and **Install updates automatically**. Qalam swaps in the new .exe and restarts by itself.

## Options and sukūn (1.2)

- **Options while you type, like Avro.** ↑/↓ choose between up to 4 versions: as typed, another alif/hamza (قَرَا / قَرَأ / قَرَأَ), no harakat, small alif (لِلّٰهِ). Qalam remembers what you pick.
- **Smart sukūn** (default): only at a stop inside a word (مَكْتَب، قُل، بَيت). `o` adds one by hand. **Off** (Control-Shift-O / Ctrl+Alt+O, or the panel button) means no sukūn at all.
- **Shortcuts can be changed** (in case they clash with yours): Mac: Qalam → Settings → Shortcuts. Windows: tray menu → Shortcuts. Harakat on/off can get a shortcut too. The panel buttons show the current ones.
- **No vowels typed → bare letters:** `ktb` → كتب.

## Qur'an (Mac)

- **⌃⇧Q** opens a box at the cursor: type `2:255`, `2:255-257`, `baqarah 255` or search (`الرحمن`, `rahmaan`, harakat optional), pick, Return. Tab: Arabic only / Arabic + translation / translation only.
- **Qalam → Qur'an** shows every sūrah with a translation, searches the whole Qur'an, copies āyāt (rich text in the Qur'an font) and downloads more translations.
- Bundled: Tanzil Uthmani text, Saheeh International, Muhiuddin Khan (Bangla), Zohurul Hoque (Bangla). Command line: `qalam ayah 2:255 --trans bn.bengali`, `qalam search …`. Details: [docs/quran.md](docs/quran.md).

## Typing in short

| Type | Gets | Type | Gets |
|---|---|---|---|
| `a i u` | ـَ ـِ ـُ | `aN iN uN` | ـً ـٍ ـٌ |
| `A` | ا (alif) | `A` + haraka | hamza (seat automatic) |
| `'` | hamza with no haraka | `e` | ع |
| `t'` | ة | `Y` | ى |
| `AA` | آ | doubled letter | shadda |
| `th dh kh sh` | ث ذ خ ش | `g` / `gh` | غ |
| `H S D T Z` | ح ص ض ط ظ | `z` / `zh` | ز |
| `aA` = `aa` | long ā | `iy` = `ii`, `uw` = `uu` | long ī, ū |

While you type, the word is shown underlined (macOS) or in a small preview box (Windows). **Space** or punctuation finishes it, **Backspace** removes the last English key, and **Esc** gives back the English letters.

## Fonts

Qalam includes **Scheherazade New** (SIL, the default Arabic font) and **KFGQPC HAFS Uthmanic Script** (the Madinah mushaf font from the [King Fahd Complex](https://fonts.qurancomplex.gov.sa/), the default for Qur'an text), and installs both for every app: automatically with the Mac `.pkg`, `install.sh` and the Windows installer, or with **Install fonts for all apps** in the setup. The Mac app also includes Noto Naskh Arabic, Amiri Quran and Noto Sans Arabic. Font licences are in `fonts/`. More: [docs/fonts.md](docs/fonts.md).

## Build from source

Needs Apple's Command Line Tools (Swift) and Go. Xcode isn't needed.

```bash
make test
```

```bash
make install
```

```bash
make release
```

- `make test` runs 175+ golden cases against both engines, plus the Windows app's tests.
- `make install` installs on this Mac.
- `make release` builds the Mac zip and .pkg (universal) and the Windows exes into `dist/`. Pushing a `v*` tag builds everything in CI, including the Windows installers, as a draft release ([docs/developers/releasing.md](docs/developers/releasing.md)).
- `make site` builds the website with the live demo. Signing and notarization: [docs/developers/signing.md](docs/developers/signing.md).

| Path | What |
|---|---|
| `engine/` | Transliteration rules in Swift (keys in `Tokenizer.swift`, word lists in `Content.swift`) |
| `mac/QalamInput/` | macOS input method (InputMethodKit) |
| `mac/QalamApp/` | macOS guide and practice app (SwiftUI) |
| `windows/` | Windows tray app in Go, with `windows/engine/` (the same rules ported to Go), `windows/installer/` (Inno Setup) and `windows/wasm/` (the engine for the website demo) |
| `cli/` | `qalam` command: convert text, run tests, generate the Windows guide page |
| `tests/cases.tsv` | Input ⇥ expected Arabic ⇥ options; both engines must pass |
| `Casks/qalam.rb` | Homebrew cask |
| `website/`, `docs/` | The website (GitHub Pages) and the docs it shows as the wiki |

## Licence and code signing

Qalam is free software under the **GNU General Public License v3.0** ([LICENSE](LICENSE)). The fonts in `fonts/` keep their own licences (SIL Open Font License, and the King Fahd Complex licence for KFGQPC Hafs).

Windows code signing: free code signing provided by [SignPath.io](https://about.signpath.io/), certificate by [SignPath Foundation](https://signpath.org/) (being set up; releases are signed once it's approved). See the [code signing policy](docs/code-signing-policy.md), which also covers privacy (Qalam collects no data).

