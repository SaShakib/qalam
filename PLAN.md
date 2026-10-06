# Qalam (قَلَم): Phonetic Arabic Keyboard: Plan

**Status (2026-10-07):** Phases 1 and 2 are built and installed (engine and tests, the macOS keyboard, and the Qalam app). Phase 3 (Windows) is built too. Releases: https://github.com/SaShakib/qalam/releases. How to use it: [README.md](README.md).

---

## 1. What it is

A small input method, like Avro or Lekho for Bangla. You type Arabic sounds with an ordinary English keyboard and get correct Arabic script with full harakat:

```
you type:   kataba          →  كَتَبَ
you type:   alkitaabu       →  الْكِتَابُ
you type:   qur'aanun       →  قُرْآنٌ
```

- **Works in every app** (Notes, Pages, Word, Safari, Chrome, WhatsApp desktop, VS Code…), the same way Lekho does.
- **Installed in your system:** it shows up in the macOS input menu next to "ABC". Switch with **Globe** or **Ctrl + Space**.
- **Tiny:** no dictionary, no internet, no telemetry. It's a pure rule engine (about 1 MB).
- **Two writing styles:**
  - **Everyday (MSA):** standard modern spelling with full harakat (اللَّه, الْكِتَاب).
  - **Qur'an (Uthmani):** alif waṣla ٱ, dagger alif ٰ, maddah ٓ, silent-letter marks, waqf signs, āyah numbers (ٱللَّهُ، ٱلْعَٰلَمِينَ).

## 2. How typing feels

Like Lekho's "Phonetic-only" mode:

1. While you type a word, the Arabic shows **underlined** in place (the "preview"). Every keystroke re-reads the whole word, so later letters can fix earlier ones. For example, typing `'` then `i` turns أ into إ.
2. **Space, Enter, Tab or punctuation** commits the word.
3. **Backspace** removes the last *English* key you typed and the preview updates.
4. **Esc** cancels the Arabic and drops in the raw English letters, so you can type an English word without switching keyboards.

No suggestion popup. What you type is what you get, which suits learning Sarf, since you spell the vowels yourself.

## 3. Design principles for an English typist

**Spell it:** letter, then its haraka (layout v2).

1. **Every letter has its own key or two-letter combo.** `A` alif, `e` ʿayn, `'` hamza, `t'` ة, `Y` ى, `th` ث, `dh` ذ, `kh` خ, `sh` ش, `z`/`zh` ز, `Z` ظ, `g` غ.
2. **Harakat:** `a` fatḥa, `i` kasra, `u` ḍamma. Tanwīn is `aN` `iN` `uN`. Long vowels are spelled with their letter (`aA` `iy` `uw`) or by doubling the vowel (`aa` `ii` `uu`).
3. **Sukūn and shadda are automatic.** A letter with no haraka gets ْ, and a doubled letter gets ّ.
4. **The `A` rule:** `A` followed by a haraka is a hamza (an alif can't carry a haraka). Otherwise `A` is the alif letter. `'` gives a hamza with no haraka. The seat (أ إ ؤ ئ ء) is always chosen for you.
5. **Start of a word:** a bare vowel gives hamzat al-waṣl (`ismu`), and `A` + haraka gives hamzat al-qaṭʿ (`AaHmadu`).
6. **Capitals are the heavy letters:** `H` ح, `S` ص, `D` ض, `T` ط, `Z` ظ.
7. **The article is smart:** sun letters get shadda, moon letters get sukūn on the lām.

The full key chart is in **[LAYOUT.md](LAYOUT.md)**.

## 4. The engine (the part that does the Arabic)

Pure logic, no UI, so the same rules can be tested on their own and later reused for Windows.

For each word, on every keystroke:

| Step | What happens |
|---|---|
| 1. Tokenize | Read the English letters by longest match (`sh` before `s`, `aa` before `a`, `aN` before `a`). |
| 2. Build syllables | Attach to each consonant its vowel, shadda (from doubling or `~`), or tanwīn. |
| 3. Article and prefixes | `al…`, `wal-`, `fal-`, `bil-`, `kal-`, `lil-`: sun-letter / moon-letter handling. |
| 4. Hamza seat | Choose أ إ ؤ ئ ء آ from position (start / middle / end) and the vowels around it. Strength order: i > u > a > sukūn. |
| 5. Auto marks | Sukūn on vowel-less consonants (but not on ة, ى, the madd letters, or the lām of a sun-letter article). Tanwīn fatḥ gets its alif (ًا), except after ة and ـاء. |
| 6. Spelling words | A short built-in list of words whose spelling isn't phonetic: اللَّه، إِلَٰه، هَٰذَا، ذَٰلِكَ، لَٰكِنْ، هَٰؤُلَاءِ، أُولَٰئِكَ، الرَّحْمَٰن. (Can be switched off.) |
| 7. Style | In Qur'an style: initial alif → ٱ, sukūn glyph setting, dagger-alif forms. In Everyday style: standard forms. |
| 8. Output | Unicode text in a fixed mark order (letter, then shadda, then vowel), so fonts render it correctly. |

**Settings** (in the input menu, the "ق" icon in the menu bar):
- Style: **Everyday** / **Qur'an**
- Harakat: **Full** (auto sukūn) / **As typed** (only marks you typed) / **None** (plain unvowelled text, for casual writing)
- Digits: **123** / **١٢٣**
- Qur'an sukūn glyph: **ْ** (standard; works in every font) / **ۡ** (KFGQPC Madinah-Mushaf style)
- Open the key chart

## 5. Platforms

### macOS (main target: build first)
- Native **InputMethodKit** input method written in Swift, the same technology Lekho uses. Apple Silicon native.
- I checked: your Mac has Swift 6.3 and the InputMethodKit SDK, so **Xcode isn't needed**. It builds with `swiftc` plus a Makefile, as Lekho does.
- Install: `make install` builds `Qalam.app`, ad-hoc signs it for your own Mac, and copies it to `~/Library/Input Methods/`. Then **System Settings → Keyboard → Input Sources → Edit → + → Arabic → Qalam**. (The first time, macOS sometimes needs a log-out and log-in before it appears.)
- Uninstall: `make uninstall`.

### Windows (built)
- A Go tray app: one portable `.exe` per architecture (x64, ARM64), cross-compiled from macOS. It uses the same rules (ported to Go in `windows/engine/`) and passes the same `tests/cases.tsv`.
- A low-level keyboard hook turns typing into Arabic. A small preview box near the caret shows the word, which is sent on Space or punctuation. Ctrl + Left Alt + A switches Arabic on and off.
- On first run it offers a per-user install (Programs folder, Start menu, start with Windows, an uninstall entry in Apps). The tray menu opens an HTML guide with the required 20-word starter practice.

## 6. Folder layout (after the build)

```
Qalam - Arabic Phonetic Keyboard/
├── PLAN.md                ← this file
├── LAYOUT.md              ← key chart / cheat sheet
├── README.md              ← install + use (written at build time)
├── Makefile               ← build · test · install · uninstall
├── engine/                ← pure Swift transliteration engine (no UI)
│   ├── Rules.swift        ← the key table (one place to change a key)
│   ├── Transliterator.swift
│   ├── Hamza.swift
│   └── SpellingWords.swift
├── mac/                   ← InputMethodKit app
│   ├── Info.plist
│   ├── main.swift
│   ├── InputController.swift   ← keys, preview, commit
│   ├── Menu.swift              ← settings in the input menu
│   └── icon.tiff
├── cli/qalam.swift        ← `qalam kataba` → كَتَبَ (quick testing in Terminal)
├── tests/cases.tsv        ← ~150 golden cases: input ⇥ style ⇥ expected
└── windows/               ← phase 3
```

## 7. Build phases

| Phase | Deliverable | How it's checked |
|---|---|---|
| **1. Engine + tests** | `engine/`, `cli/`, `tests/cases.tsv` | Every golden case passes (`make test`). The cases include every example in LAYOUT.md plus hamza, article, tanwīn and Qur'an-style edge cases. |
| **2. macOS input method** | `Qalam.app`, `make install`, input-menu settings | Installed on your Mac and typed into Notes, TextEdit, Safari and VS Code. |
| **3. Windows** | `windows/Qalam.exe` (AutoHotkey v2) | Same `cases.tsv` passes. You test it on a Windows machine. |
| *(optional)* | A web "typing pad" page for phones / any computer, with copy-paste output | Same tests. |

Phases 1 and 2 are one sitting of work. Phase 3 waits for your go-ahead.

## 8. Decisions I need from you

My recommended default is listed first. Reply with just the numbers you want changed.

1. **Name:** *Qalam (قَلَم)*. Or suggest another.
2. ~~Key for ة~~ **Decided:** `t'`.
3. ~~ز and ظ~~ **Decided:** `z` or `zh` = ز, `Z` = ظ, `dh` = ذ.
4. **Sukūn at the end of a word** (`qul` → قُلْ): **on** in Full mode. Turn it off if you'd rather leave endings bare (قُل).
5. **Default style:** **Everyday**, switched to Qur'an from the menu. Or should Qur'an be the default?
6. **Windows route:** **AutoHotkey `.exe`** (no other install needed), or Keyman (more robust, but you install the Keyman app).
7. **Preview:** **underlined preview** (Lekho style). The alternative is letters appearing directly as you type (Avro-classic style, no underline).
