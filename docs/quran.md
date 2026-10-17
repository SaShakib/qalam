# Qur'an: insert, read and search

Qalam carries the whole Qur'an (Uthmani script, full harakat, from the [Tanzil Project](https://tanzil.net)) with
three translations: **English: Saheeh International**, **Bangla: Muhiuddin Khan** and **Bangla: Zohurul Hoque**.
On the Mac more can be downloaded in the app. Windows has the insert box with the three bundled translations.

## Insert āyāt while typing (Mac: Control-Shift-Q · Windows: Ctrl+Alt+Q)

1. **Mac:** with the Qalam keyboard on, press **⌃⇧Q** (Control-Shift-Q) in any app.
   **Windows:** with Qalam running (the ق in the tray), press **Ctrl+Alt+Q** in any app. It works even
   while Arabic typing is off. A box opens at the cursor.
2. Pick with the three dropdowns, by mouse or keyboard:

   **Sūrah ▾**  **From āyah ▾**  **To āyah ▾**

   The preview underneath shows the āyāt in the Qur'an font. Click **Insert** (or press Return).
   The box opens on your last choice next time.

   | Key | Does |
   |---|---|
   | ↑ / ↓ | change the highlighted dropdown |
   | ← / → or Tab | go to the next dropdown |
   | type `36` or `yasin` | jump to that sūrah (spelling is loose: `baqarah`, `cow`) |
   | type `2:255-257` | fill all three at once |
   | Return | insert |
   | `/` | open **Search** |
   | Esc | close (the shortcut again closes it too) |

3. **Search** (the tab at the top, or `/`): type a word and pick an āyah. Harakat are optional
   (`الرحمن` = `ٱلرَّحْمَٰنِ`), and Latin is read the Qalam way (`rahmaan`, with similar letters counting too).
   ↑ / ↓ choose, Return inserts, Tab goes back to **Choose**.
4. At the bottom choose what goes in: **Arabic only**, **Arabic + translation** or **Translation only**, and
   which translation. Qalam remembers both.

Each āyah ends with its reference, like `(2:255)`. With a translation, the Arabic line comes first and the
translation under it.

The shortcut can be changed in **Qalam → Settings → Shortcuts** on the Mac (the ق menu has
**Insert from the Qur'an…** too), and on Windows in the tray menu → **Shortcuts** → **Insert from the Qur'an**
(or **Setup & settings… → Shortcuts**).

**Windows notes**

- Keys are read on the US layout, like Qalam's typing. To search in Arabic letters, copy them and press
  **Ctrl+V** in the box; or type the Latin (`rahmaan`).
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
