# Qur'an: insert, read and search

Qalam carries the whole Qur'an (Uthmani script, full harakat, from the [Tanzil Project](https://tanzil.net)) with
translations: **English**: Saheeh International, Mufti Taqi Usmani; **Bangla**: Muhiuddin Khan, Bayaan Foundation,
Sheikh Mujibur Rahman, Taisirul Quran (Tawheed Publication), Dr. Abu Bakr Zakaria, Zohurul Hoque; **Urdu**: Mufti Taqi Usmani.
On the Mac more can be downloaded in the app. Windows has the insert box with the three bundled translations.

## Insert āyāt while typing (Mac: Control-Shift-Q · Windows: Ctrl+Alt+Q)

**Mac.** Press **⌃⇧Q** in any app, with any keyboard on (Qalam stays in the menu bar as ق and starts at
login for this; switch it off in **Settings → Qur'an box**). A box opens:

1. **Search** at the top. What you type is searched in the language you type it in:
   - English letters → the English translation (and sūrah names: `baqarah`, `cow`).
   - Switch the keyboard to **Qalam** (or click **Arabic** next to the field) and type part of an āyah the
     Qalam way, e.g. `qul huwa allaahu` → the Arabic text. The box shows the Arabic it searched for.
     Harakat are optional, and words the mushaf joins (يَٰٓأَيُّهَا) match when typed apart (يا أيها).
   - Bangla letters → the Bangla translation.
   - `2:255` or `baqarah 255-257` → that passage.
2. Choose a result with **↑ / ↓** or a **click** (double-click inserts). It fills **Sūrah**, **From āyah** and
   **To āyah**.
3. Set the range with the **From āyah** / **To āyah** dropdowns. The preview shows exactly what goes in.
4. Choose **Arabic only**, **Arabic + translation** or **Translation only**, and the translation.
5. **Insert** (Return). **Esc** clears the search, then closes the box.

The text goes back into the app you were in. With the Qalam keyboard on there, it is typed in directly.
With another keyboard, Qalam pastes it (⌘V) once you allow it in **System Settings → Privacy & Security →
Accessibility** (the box has an **Allow…** button); until then it is copied and you press ⌘V.
The setup's **Qur'an box** step (also in **Settings → Qur'an box**) shows both permissions with a ✓ when
they're on. After a Qalam update macOS forgets the Accessibility permission (it belongs to the exact build);
**Allow…** clears the old entry so you can switch it on again.

**Windows.** Press **Ctrl+Alt+Q** (Qalam running in the tray, Arabic on or off):

1. Pick with the three dropdowns, by mouse or keyboard: **Sūrah ▾**  **From āyah ▾**  **To āyah ▾**.
   The preview underneath shows the āyāt in the Qur'an font. Click **Insert** (or press Enter).

   | Key | Does |
   |---|---|
   | ↑ / ↓ | change the highlighted dropdown |
   | ← / → or Tab | go to the next dropdown |
   | type `36` or `yasin` | jump to that sūrah |
   | type `2:255-257` | fill all three at once |
   | `/` | open **Search** (a word, harakat optional) |
   | Esc | close |

2. At the bottom choose what goes in and which translation.

Each āyah ends with its reference, like `(2:255)`. With a translation, the Arabic line comes first and the
translation under it.

The shortcut can be changed in **Qalam → Settings → Shortcuts** on the Mac (the ق menu has
**Insert from the Qur'an…** too), and on Windows in the tray menu → **Shortcuts** → **Insert from the Qur'an**
(or **Setup & settings… → Shortcuts**).

**Windows notes**

- Keys are read on the US layout, like Qalam's typing. To search in Arabic letters, copy them and press
  **Ctrl+V** in the Search tab; or type the Latin (`rahmaan`).
- A click on a dropdown opens a Windows menu (long lists in columns). The menu has the focus while it is
  open; your app gets it back as soon as you pick.
- Lines are separated with **Shift+Enter**, so in chat apps (WhatsApp, Teams, Slack) the āyah and its
  translation stay in one message instead of sending it.
- Clicking another window or pressing another Ctrl/Alt shortcut closes the box.
- In Chrome-based apps the same invisible fix as for typing keeps اللَّه from turning into the font's ligature.
- Searching inside the translations is Mac only.

The box shows the Arabic in the Qur'an font (KFGQPC Hafs); what goes into your app takes that app's font.
For the Qur'an font in Pages, Word or Notes on the Mac, use **Copy** on the app's Qur'an page (it copies
rich text). On Windows, install the fonts (**Setup & settings… → Fonts**) and pick KFGQPC HAFS Uthmanic Script.

## The Qur'an page in the app (Mac)

**Qalam → Qur'an**:

- Every sūrah, with the translation under each āyah. Choose a range (Āyāt *from* … *to*) and **Copy** it.
- **Search the whole Qur'an** at the top: Arabic with or without harakat, or Latin letters. Type `2:255` and
  Return to jump there.
- **Translation**: pick any translation on this Mac. **More translations** downloads others from tanzil.net
  (Pickthall, Yusuf Ali, Hilali & Khan, Wahiduddin Khan, Urdu, Hindi, Indonesian, Turkish, French…).
- **⌃⇧Q inserts**: the same choice as Tab in the box.

## From the command line (Mac; for scripts and AI tools)

```
qalam ayah 2:255 "baqarah 1-5" --trans en.sahih,bn.bengali
qalam ayah 2:255-257 --trans bn.hoque --insert     # exactly what ⌃⇧Q would insert
qalam search "الرحمن" --max 20 --trans none
qalam surahs
qalam translations
```

`qalam` is built by `swift build -c release --product qalam`. It finds the data in `quran/` next to the
repo, in an installed Qalam.app, or wherever `QALAM_QURAN` points.

## Sources

The Qur'an text and translations are verbatim from tanzil.net; see [quran/README.md](https://github.com/SaShakib/qalam/blob/main/quran/README.md).
