# Qalam (قَلَم): Phonetic Arabic Keyboard

Type Arabic the way you spell it: each letter, then its haraka. Sukūn and shadda appear by themselves. Hamza seats, the article, tanwīn and Qur'anic spelling are handled for you.

```
kitaAbuN → كِتَابٌ     saAala → سَأَلَ     eilmuN → عِلْمٌ     madrasat'uN → مَدْرَسَةٌ     alshamsu → الشَّمْسُ
```

Works in every app, on **macOS** (a native input method, like [Lekho](https://github.com/ARahim3/Lekho)) and **Windows** (a tray app). There is an **Everyday** style and a **Qur'an (Uthmani)** style with ٱ, dagger alif, maddah, waqf signs and āyah numbers.

The full key chart is in **[LAYOUT.md](LAYOUT.md)**.

## Install on macOS

macOS 13 (Ventura) or later.

**Homebrew**

```bash
brew tap sashakib/qalam https://github.com/SaShakib/qalam && brew install --cask qalam
```

**One command (no Homebrew)**

```bash
curl -fsSL https://raw.githubusercontent.com/SaShakib/qalam/main/install.sh | bash
```

**Installer package:** download `Qalam-1.0.0.pkg` from [Releases](https://github.com/SaShakib/qalam/releases/latest) and double-click it. The package isn't notarized yet. If macOS blocks it, go to System Settings → Privacy & Security and click **Open Anyway**.

The Qalam app then opens. It adds the keyboard to your input menu and walks you through a **20-word practice** that covers every special key. Switch keyboards with **🌐 Globe** or **Control-Space**.

## Install on Windows

Download **`Qalam-Windows-x64.exe`** (or `-arm64` for ARM laptops) from [Releases](https://github.com/SaShakib/qalam/releases/latest) and run it.

- It offers to install for your user: Start menu, start with Windows, and an uninstall entry in Settings → Apps. No admin rights are needed.
- **Ctrl + Alt + A** (or a click on the ق tray icon) switches Arabic on and off.
- Right-click the tray icon for style, harakat, Arabic digits, and the guide and practice page.
- The exe isn't code-signed yet. If SmartScreen warns, click **More info → Run anyway**.

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

The app includes four open-licence Arabic fonts: **Noto Naskh Arabic** (default), **Amiri Quran** (default for Qur'an text), **Scheherazade New** (SIL) and **Noto Sans Arabic**. If **KFGQPC Hafs** is installed (from [fonts.qurancomplex.gov.sa](https://fonts.qurancomplex.gov.sa/)), you can pick it in Settings. Font licences are in `fonts/`.

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

- `make test` runs 175 golden cases against both engines.
- `make install` installs on this Mac.
- `make release` builds the Mac zip and .pkg and the Windows exes into `dist/`.

| Path | What |
|---|---|
| `engine/` | Transliteration rules in Swift (keys in `Tokenizer.swift`, word lists in `Content.swift`) |
| `mac/QalamInput/` | macOS input method (InputMethodKit) |
| `mac/QalamApp/` | macOS guide and practice app (SwiftUI) |
| `windows/` | Windows tray app in Go, with `windows/engine/` (the same rules ported to Go) |
| `cli/` | `qalam` command: convert text, run tests, generate the Windows guide page |
| `tests/cases.tsv` | Input ⇥ expected Arabic ⇥ options; both engines must pass |
| `Casks/qalam.rb` | Homebrew cask |
