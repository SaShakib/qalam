# Plan: Avro-style options and a smarter sukūn

**Status:** plan for review (2026-10-08). Nothing is built yet.

## Why

Full harakat suits Qur'an and study. For day-to-day writing, hadith, and dictionary search, people often want:
- **no sukūn** (قرأ, not قَرَأْ)
- **only some harakat**, or **just the consonants** (كتب to search a dictionary)
- **a different hamza or alif** from the one the rules chose (قرأ with alif-hamza, not قَرَأَ with hamza-fatḥa)

One fixed output can't serve all of these. So Qalam should **offer choices while you type**, as Avro does, and **learn what you pick**.

---

## 1. Smarter sukūn (rules)

These rules apply in **Everyday** style. **Qur'an** style keeps full sukūn, because a mushaf needs it on every stop.

| Rule | Example | Now | New |
|---|---|---|---|
| **R1. Sukūn only at a real stop inside the word:** a consonant that has a vowel before it and another consonant right after it. | `maktab` | مَكْتَبْ | مَكْتَب |
| **R2. No sukūn on the last letter** (setting, see Q1) | `qul` | قُلْ | قُل |
| **R3. No sukūn on ي / و after fatḥa** (ay, aw): they need no mark | `bayt` | بَيْتْ | بَيت |
| **R4. No vowel typed in the whole word → bare consonants.** No sukūn and no marks at all. Long vowels (`aa`, `ii`, `uu`) don't count as typed vowels. | `ktb` | كْتْبْ | كتب |
| **R5. A single letter on its own gets no mark** | `b` | بْ | ب |

Sukūn you type yourself (`o`) is always kept.

## 2. The options panel (like Avro)

While you type a word, a small panel under it shows **3–4 versions**:

```
 qaraA
 ┌──────────────────────────────┐
 │ 1  قَرَا     as typed (alif)   │  ← highlighted
 │ 2  قَرَأ     alif with hamza   │
 │ 3  قَرَأَ     hamza + fatḥa     │
 │ 4  قرأ      no harakat        │
 ├──────────────────────────────┤
 │ [ Sukūn: Auto ▾ ]  [ Harakat ] │  ← buttons, mouse only (no shortcut)
 └──────────────────────────────┘
```

**Row 1 is always exactly what you typed.** The other rows come from these sources, best first, never duplicated, at most 4 rows:

| Source | When | Example options |
|---|---|---|
| **Hamza / alif seat** | the word has `A` or `'` | ا · أ · إ · ء · آ: `qaraA` → قَرَا / قَرَأ / قَرَأَ, `qr'` → قرء / قرأ / قرئ |
| **Harakat level** | always | full → only what you typed → none: `kitaAb` → كِتَاب / كتاب |
| **Long vowels** | `aa ii uu` (or `aA iy uw`) | with and without harakat: `fii` → فِي / في |
| **Special spellings** | Allah words, هَٰذَا, ذَٰلِكَ … | لِلَّهِ / لله / لِلّٰهِ, هَٰذَا / هذا |
| **What you picked before** | the same word was typed earlier | your earlier choice goes to the top |

**Keys:**
- **↑ / ↓** move the highlight.
- **Space, Enter or punctuation** puts in the highlighted version.
- **Esc** gives back the English letters.
- **Backspace** edits the word.
- **Clicking a row** picks it.
- There are no number shortcuts, so typing digits still works (āyah codes `\7`).

**Learning (stored only on your Mac, can be cleared):**
- **Per word:** if you pick version 3 for `qaraA`, version 3 is highlighted the next time you type `qaraA`.
- **Harakat habit:** the harakat level you last picked (full / some / none) becomes the default for the next words. If you pick a version with harakat, the next word defaults to harakat too.
- **"Forget what I picked"** in Settings clears everything.

**Panel buttons** (below the options, mouse only):
- **Sukūn: Auto / Off.** Auto uses the rules in §1; Off never adds sukūn.
- **Harakat: Full / Typed / None.** The same choice as in the menu, one click away.

**Turning the panel off:** if you don't want options, set Settings → **Show options while typing** to off. Qalam then behaves as it does today, with the new sukūn rules.

## 3. Windows

The preview box becomes the same panel (arrow keys, click, the same buttons and learning). This comes after the Mac version works.

## 4. Things that can go wrong, and how the plan handles them

| Risk | Handling |
|---|---|
| **The panel steals focus** or makes the app lose the cursor | Use a non-activating panel (it never becomes the active window); clicks on it still work. |
| **Some apps don't report where the cursor is** (some Electron apps, full-screen games) | Put the panel near the mouse, or at the bottom centre of the window. Never off-screen; multiple monitors handled. |
| **Arrow keys** currently finish the word | While the panel is open, ↑/↓ move the highlight; ←/→ still finish the word and move the cursor. |
| **Enter** in chat apps sends the message | Enter only chooses the word, as now; press it again to send. |
| **Learning sticks to a mistake** | Row 1 is always what you typed, so it's never more than one ↓ away; "Forget" in Settings clears it. |
| **Privacy** of learned words | Saved locally in one file (`~/Library/Application Support/Qalam/choices.json`), never uploaded. |
| **The new sukūn rules change existing behaviour** | Qur'an style keeps full sukūn, and a setting ("Sukūn: Full / Smart / Off") brings back the old behaviour. All golden tests get updated, with new ones for R1–R5. |
| **Hamza seat with no vowels** (`qr'`) has no rule to follow | Default to ء, and offer أ / ئ / ؤ as options. |
| **Too many options** become noise | At most 4 rows, with duplicates removed. A word with only one form shows one row, plus the buttons. |
| **The practice and word-list checks** expect full harakat | The practice accepts any of the options for a word (it already accepts several forms). |
| **Speed** | Options are a handful of transliterations per keystroke (microseconds); the panel redraw is cheap. |
| **Windows** behaves differently | The same engine code (Go port) produces the options; only the panel is new. |

## 5. Build order

1. **Engine:** the sukūn rules R1–R5 and a "Sukūn: Full / Smart / Off" option, plus an `options(for: word)` function that returns the 3–4 versions. Golden tests are updated and extended. (Swift and Go.)
2. **Mac keyboard:** the options panel, keys, buttons, and learning.
3. **Mac app:** Settings for options on/off, the sukūn mode and "Forget what I picked". The practice accepts the options.
4. **Windows:** the options in the preview box.
5. **Release 1.2.0:** it reaches everyone through the updater.

## 6. Decisions needed

- **Q1. Last letter:** sukūn or not? (`qul` → قُلْ or قُل). *Suggested: no sukūn on the last letter in Everyday style.*
- **Q2. Bare consonants (R4):** remove shadda too (`rbb` → ربب), or keep shadda? *Suggested: remove everything (pure skeleton, best for dictionary search).*
- **Q3. Panel default:** you said "always give options", so the panel shows on every word. A setting can limit it to words with a real choice (hamza/alif, long vowels, special words). *Suggested: always, as you asked.*
- **Q4. Learning:** both "per word" and "harakat habit"? *Suggested: yes, both.*
